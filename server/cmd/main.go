package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	_ "time/tzdata" // embeds the IANA zoneinfo DB so time.LoadLocation("Asia/Kolkata") (internal/diary) works even on a host/image without a system tzdata package (e.g. the alpine runtime image)

	"github.com/Sam-Frost/portfolio/internal/auth"
	"github.com/Sam-Frost/portfolio/internal/cms"
	"github.com/Sam-Frost/portfolio/internal/db"
	"github.com/Sam-Frost/portfolio/internal/diary"
	"github.com/Sam-Frost/portfolio/internal/document"
	"github.com/Sam-Frost/portfolio/internal/documentlabel"
	"github.com/Sam-Frost/portfolio/internal/fitness"
	"github.com/Sam-Frost/portfolio/internal/label"
	"github.com/Sam-Frost/portfolio/internal/notepad"
	"github.com/Sam-Frost/portfolio/internal/reqlog"
	"github.com/Sam-Frost/portfolio/internal/settings"
	"github.com/Sam-Frost/portfolio/internal/spotify"
	"github.com/Sam-Frost/portfolio/internal/todo"
	"github.com/Sam-Frost/portfolio/internal/upskill"
	"github.com/Sam-Frost/portfolio/internal/worksession"
)

func healthCheck(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Server is healthy, boi!!!")
}

// withCORS is env-configurable ("*" by default). Authorization is allowed
// alongside Content-Type now that /api/* routes are gated by a bearer JWT.
func withCORS(allowedOrigin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		// PUT is here for the local-disk document blob store, whose upload
		// URLs point back at this server; the S3 store's PUT goes straight
		// to the bucket and is governed by the bucket's own CORS config.
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// publicPaths bypass auth: the login endpoint (issues the token in the
// first place), the health check (used by container orchestration), and
// the Spotify OAuth callback (hit by a browser redirect from Spotify,
// which carries no bearer token — see internal/spotify's handler for how
// that endpoint protects itself instead).
var publicPaths = map[string]bool{
	"/health":               true,
	"/api/auth/login":       true,
	"/api/spotify/callback": true,
}

// withAuth gates every route except publicPaths behind auth.RequireAuth.
func withAuth(authService *auth.Service, next http.Handler) http.Handler {
	protected := auth.RequireAuth(authService)(next)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The local document blob store's upload/download URLs carry their
		// own HMAC signature (verified in document.ServeBlob), so they're
		// reached by a browser with no bearer token, like the OAuth callback.
		if publicPaths[r.URL.Path] || strings.HasPrefix(r.URL.Path, "/api/document-blob/") {
			next.ServeHTTP(w, r)
			return
		}
		protected.ServeHTTP(w, r)
	})
}

// newRouter wires every feature's repository -> service -> handler and
// registers its routes. Adding a feature means one more block here.
func newRouter(
	authService *auth.Service,
	todoRepo todo.Repository,
	labelRepo label.Repository,
	settingsRepo settings.Repository,
	notepadRepo notepad.Repository,
	upskillRepo upskill.Repository,
	diaryRepo diary.Repository,
	fitnessRepo fitness.Repository,
	spotifyRepo spotify.Repository,
	workSessionRepo worksession.Repository,
	cmsRepo cms.Repository,
	cmsPublisher cms.Publisher,
	documentRepo document.Repository,
	documentLabelRepo documentlabel.Repository,
	documentBlob document.BlobStore,
	spotifyClientID, spotifyClientSecret, spotifyRedirectURI, spotifyFrontendURL, jwtSecret string,
) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthCheck)

	auth.NewHandler(authService).Register(mux)

	todoService := todo.NewService(todoRepo)
	todo.NewHandler(todoService).Register(mux)

	labelService := label.NewService(labelRepo)
	label.NewHandler(labelService).Register(mux)

	settingsService := settings.NewService(settingsRepo)
	settings.NewHandler(settingsService).Register(mux)

	notepadService := notepad.NewService(notepadRepo)
	notepad.NewHandler(notepadService).Register(mux)

	upskillService := upskill.NewService(upskillRepo)
	upskill.NewHandler(upskillService).Register(mux)

	diaryService := diary.NewService(diaryRepo)
	diary.NewHandler(diaryService).Register(mux)

	fitnessService := fitness.NewService(fitnessRepo)
	fitness.NewHandler(fitnessService).Register(mux)

	spotifyClient := spotify.NewAPIClient(spotifyClientID, spotifyClientSecret, spotifyRedirectURI)
	spotifyService := spotify.NewService(spotifyRepo, spotifyClient, []byte(jwtSecret))
	spotify.NewHandler(spotifyService, spotifyFrontendURL).Register(mux)

	workSessionService := worksession.NewService(workSessionRepo)
	worksession.NewHandler(workSessionService).Register(mux)

	cmsService := cms.NewService(cmsRepo, cmsPublisher)
	cms.NewHandler(cmsService).Register(mux)

	documentLabelService := documentlabel.NewService(documentLabelRepo)
	documentlabel.NewHandler(documentLabelService).Register(mux)

	documentService := document.NewService(documentRepo, documentBlob)
	document.NewHandler(documentService).Register(mux)

	return mux
}

// newCMSPublisher builds the publisher the CMS Publish button uses. With no
// CMS_S3_BUCKET set (local dev), it's a no-op and POST /api/cms/publish
// returns a clear "not configured" 409 rather than pretending to publish.
//
//	CMS_S3_BUCKET                   S3 bucket serving the public site
//	CMS_S3_PREFIX                   key prefix in that bucket (default "sat0ru")
//	CMS_CONTENT_KEY                 live document basename (default "content.json")
//	CMS_CLOUDFRONT_DISTRIBUTION_ID  distribution to invalidate (optional)
//	AWS_REGION / AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY  standard AWS creds
func newCMSPublisher(ctx context.Context) cms.Publisher {
	bucket := os.Getenv("CMS_S3_BUCKET")
	if bucket == "" {
		log.Println("CMS_S3_BUCKET not set — CMS publishing is disabled")
		return cms.NewNoopPublisher()
	}

	prefix := os.Getenv("CMS_S3_PREFIX")
	if prefix == "" {
		prefix = "sat0ru"
	}

	publisher, err := cms.NewS3Publisher(ctx, cms.S3Config{
		Bucket:         bucket,
		Prefix:         prefix,
		ContentKey:     os.Getenv("CMS_CONTENT_KEY"),
		DistributionID: os.Getenv("CMS_CLOUDFRONT_DISTRIBUTION_ID"),
		Region:         os.Getenv("AWS_REGION"),
	})
	if err != nil {
		log.Fatalf("CMS publisher init failed: %v", err)
	}
	log.Printf("CMS publishing enabled → s3://%s/%s", bucket, prefix)
	return publisher
}

func requireEnv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		log.Fatalf("%s is required", name)
	}
	return v
}

func envOr(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}

// newDocumentBlobStore builds the store the Document Storage feature keeps
// file bytes in. With DOCUMENTS_S3_BUCKET set it's S3 (browser uploads
// straight to the bucket via presigned URLs); without it, a local-disk
// store whose signed URLs point back at this server — so local dev and the
// no-AWS deployment work unchanged.
//
//	DOCUMENTS_S3_BUCKET   bucket for document objects (enables S3 mode)
//	DOCUMENTS_S3_PREFIX   key prefix in that bucket (default "documents")
//	AWS_REGION / AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY  standard AWS creds
//	DOCUMENTS_LOCAL_DIR   on-disk root when S3 is not configured (default "./.data/documents")
//	PUBLIC_API_URL        this server's external base URL, for local-store signed URLs
func newDocumentBlobStore(ctx context.Context, jwtSecret string) document.BlobStore {
	bucket := os.Getenv("DOCUMENTS_S3_BUCKET")
	if bucket == "" {
		dir := envOr("DOCUMENTS_LOCAL_DIR", "./.data/documents")
		publicURL := envOr("PUBLIC_API_URL", "http://localhost:8080")
		store, err := document.NewLocalBlobStore(dir, []byte(jwtSecret), publicURL)
		if err != nil {
			log.Fatalf("document local blob store init failed: %v", err)
		}
		log.Printf("DOCUMENTS_S3_BUCKET not set — using local disk blob store at %s", dir)
		return store
	}

	store, err := document.NewS3BlobStore(ctx, document.S3Config{
		Bucket: bucket,
		Prefix: envOr("DOCUMENTS_S3_PREFIX", "documents"),
		Region: os.Getenv("AWS_REGION"),
	})
	if err != nil {
		log.Fatalf("document S3 blob store init failed: %v", err)
	}
	log.Printf("document storage enabled → s3://%s/%s", bucket, envOr("DOCUMENTS_S3_PREFIX", "documents"))
	return store
}

func main() {
	reqlog.Init()

	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "*"
	}

	domainPassword := requireEnv("DOMAIN_PASSWORD")
	jwtSecret := requireEnv("JWT_SECRET")
	databaseURL := requireEnv("DATABASE_URL")

	spotifyClientID := requireEnv("SPOTIFY_CLIENT_ID")
	spotifyClientSecret := requireEnv("SPOTIFY_CLIENT_SECRET")
	spotifyRedirectURI := requireEnv("SPOTIFY_REDIRECT_URI")
	spotifyFrontendURL := requireEnv("SPOTIFY_FRONTEND_REDIRECT_URL")
	spotifyTokenKey, err := base64.StdEncoding.DecodeString(requireEnv("SPOTIFY_TOKEN_KEY"))
	if err != nil || len(spotifyTokenKey) != 32 {
		log.Fatalf("SPOTIFY_TOKEN_KEY must be a base64-encoded 32-byte key (openssl rand -base64 32)")
	}
	spotifyCipher, err := spotify.NewCipher(spotifyTokenKey)
	if err != nil {
		log.Fatalf("spotify cipher: %v", err)
	}

	connectCtx, connectCancel := context.WithTimeout(context.Background(), 10*time.Second)
	sqlDB, err := db.Connect(connectCtx, databaseURL)
	connectCancel()
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer sqlDB.Close()

	migrateCtx, migrateCancel := context.WithTimeout(context.Background(), 30*time.Second)
	err = db.Migrate(migrateCtx, sqlDB)
	migrateCancel()
	if err != nil {
		log.Fatalf("database migration failed: %v", err)
	}

	authService := auth.NewService(domainPassword, jwtSecret)
	todoRepo := todo.NewPostgresRepository(sqlDB)
	labelRepo := label.NewPostgresRepository(sqlDB)
	settingsRepo := settings.NewPostgresRepository(sqlDB)
	notepadRepo := notepad.NewPostgresRepository(sqlDB)
	upskillRepo := upskill.NewPostgresRepository(sqlDB)
	diaryRepo := diary.NewPostgresRepository(sqlDB)
	fitnessRepo := fitness.NewPostgresRepository(sqlDB)
	spotifyRepo := spotify.NewPostgresRepository(sqlDB, spotifyCipher)
	workSessionRepo := worksession.NewPostgresRepository(sqlDB)
	cmsRepo := cms.NewPostgresRepository(sqlDB)
	cmsPublisher := newCMSPublisher(context.Background())
	documentRepo := document.NewPostgresRepository(sqlDB)
	documentLabelRepo := documentlabel.NewPostgresRepository(sqlDB)
	documentBlob := newDocumentBlobStore(context.Background(), jwtSecret)

	router := newRouter(
		authService, todoRepo, labelRepo, settingsRepo, notepadRepo, upskillRepo, diaryRepo,
		fitnessRepo, spotifyRepo, workSessionRepo, cmsRepo, cmsPublisher,
		documentRepo, documentLabelRepo, documentBlob,
		spotifyClientID, spotifyClientSecret, spotifyRedirectURI, spotifyFrontendURL, jwtSecret,
	)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: reqlog.Middleware(withCORS(allowedOrigin, withAuth(authService, router))),
		// ReadHeaderTimeout still guards against slow-header (slowloris)
		// clients, but ReadTimeout/WriteTimeout are left unset: the
		// local-disk document blob store streams whole files through this
		// server, and a fixed 10s cap would sever large transfers.
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		slog.Info("server listening", "port", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("shutting down")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
}
