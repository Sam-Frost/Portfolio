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
	"github.com/Sam-Frost/portfolio/internal/blobstore"
	"github.com/Sam-Frost/portfolio/internal/cms"
	"github.com/Sam-Frost/portfolio/internal/db"
	"github.com/Sam-Frost/portfolio/internal/diary"
	"github.com/Sam-Frost/portfolio/internal/document"
	"github.com/Sam-Frost/portfolio/internal/documentlabel"
	"github.com/Sam-Frost/portfolio/internal/drawingboard"
	"github.com/Sam-Frost/portfolio/internal/fitness"
	"github.com/Sam-Frost/portfolio/internal/label"
	"github.com/Sam-Frost/portfolio/internal/mailer"
	"github.com/Sam-Frost/portfolio/internal/notepad"
	"github.com/Sam-Frost/portfolio/internal/notepadlabel"
	"github.com/Sam-Frost/portfolio/internal/notification"
	"github.com/Sam-Frost/portfolio/internal/reminder"
	"github.com/Sam-Frost/portfolio/internal/reqlog"
	"github.com/Sam-Frost/portfolio/internal/scheduler"
	"github.com/Sam-Frost/portfolio/internal/settings"
	"github.com/Sam-Frost/portfolio/internal/spotify"
	"github.com/Sam-Frost/portfolio/internal/todo"
	"github.com/Sam-Frost/portfolio/internal/upskill"
	"github.com/Sam-Frost/portfolio/internal/workprofile"
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
		// ETag is exposed so the browser can read each multipart-upload
		// part's ETag off the PUT response (diary video log, local-disk
		// blob store). The S3 store's PUTs go straight to the bucket, whose
		// own CORS config must expose ETag too.
		w.Header().Set("Access-Control-Expose-Headers", "ETag")
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
	// The service worker's pushsubscriptionchange handler re-registers a
	// rotated Web Push subscription with no bearer token available — see
	// notification.Handler.resync.
	"/api/notifications/subscriptions/sync": true,
}

// withAuth gates every route except publicPaths behind auth.RequireAuth.
func withAuth(authService *auth.Service, next http.Handler) http.Handler {
	protected := auth.RequireAuth(authService)(next)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The local document blob store's upload/download URLs carry their
		// own HMAC signature (verified in document.ServeBlob), so they're
		// reached by a browser with no bearer token, like the OAuth callback.
		if publicPaths[r.URL.Path] ||
			strings.HasPrefix(r.URL.Path, "/api/document-blob/") ||
			strings.HasPrefix(r.URL.Path, "/api/blob/") {
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
	notepadLabelRepo notepadlabel.Repository,
	drawingBoardRepo drawingboard.Repository,
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
	diaryVideoRepo diary.VideoRepository,
	diaryVideoBlob blobstore.Store,
	notificationRepo notification.Repository,
	reminderRepo reminder.Repository,
	workProfileRepo workprofile.Repository,
	mail mailer.Mailer,
	vapid notification.VAPIDConfig,
	spotifyClientID, spotifyClientSecret, spotifyRedirectURI, spotifyFrontendURL, jwtSecret string,
) (*http.ServeMux, *scheduler.Scheduler) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthCheck)

	auth.NewHandler(authService).Register(mux)

	todoService := todo.NewService(todoRepo)
	todo.NewHandler(todoService).Register(mux)

	labelService := label.NewService(labelRepo)
	label.NewHandler(labelService).Register(mux)

	settingsService := settings.NewService(settingsRepo)
	settings.NewHandler(settingsService).Register(mux)

	var pushSender notification.PushSender
	if vapid.PublicKey != "" {
		pushSender = notification.NewWebPushSender(vapid)
	} else {
		pushSender = notification.NewNoopPushSender()
	}
	notificationService := notification.NewService(notificationRepo, pushSender, mail, settingsService)
	notification.NewHandler(notificationService).Register(mux)

	reminderService := reminder.NewService(reminderRepo)
	reminder.NewHandler(reminderService).Register(mux)

	workProfileService := workprofile.NewService(workProfileRepo)
	workprofile.NewHandler(workProfileService).Register(mux)

	sched := scheduler.New(scheduler.Deps{
		Notifications: notificationService,
		Todos:         todoService,
		Settings:      settingsService,
		Reminders:     reminderService,
	})

	notepadService := notepad.NewService(notepadRepo)
	notepad.NewHandler(notepadService).Register(mux)

	notepadLabelService := notepadlabel.NewService(notepadLabelRepo)
	notepadlabel.NewHandler(notepadLabelService).Register(mux)

	drawingBoardService := drawingboard.NewService(drawingBoardRepo)
	drawingboard.NewHandler(drawingBoardService).Register(mux)

	upskillService := upskill.NewService(upskillRepo)
	upskill.NewHandler(upskillService).Register(mux)

	diaryService := diary.NewService(diaryRepo)
	diary.NewHandler(diaryService).Register(mux)

	diaryVideoService := diary.NewVideoService(diaryVideoRepo, diaryVideoBlob)
	diary.NewVideoHandler(diaryVideoService).Register(mux)
	if local, ok := diaryVideoBlob.(blobstore.LocalServer); ok {
		mux.HandleFunc(local.RoutePrefix(), local.ServeBlob)
	}

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

	return mux, sched
}

// newMailer builds the outbound-email sender from SMTP_* env. Unset
// SMTP_USERNAME ⇒ a no-op mailer (local dev / no-credentials deploy), the
// same "unconfigured is fine" pattern as newCMSPublisher.
//
//	SMTP_USERNAME / SMTP_PASSWORD  Gmail address + app password (required to enable)
//	SMTP_FROM                      From header; defaults to SMTP_USERNAME
//	SMTP_HOST / SMTP_PORT          default smtp.gmail.com / 587
func newMailer() mailer.Mailer {
	username := os.Getenv("SMTP_USERNAME")
	if username == "" {
		log.Println("SMTP_USERNAME not set — outbound email is disabled")
		return mailer.NewNoopMailer()
	}
	log.Printf("outbound email enabled → %s via %s", username, envOr("SMTP_HOST", "smtp.gmail.com"))
	return mailer.NewSMTPMailer(mailer.SMTPConfig{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     os.Getenv("SMTP_PORT"),
		Username: username,
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     os.Getenv("SMTP_FROM"),
	})
}

// newVAPIDConfig reads the Web Push signing keys from env. Unset
// VAPID_PUBLIC_KEY ⇒ a zero config, which makes newRouter pick the no-op
// push sender and the client treat push as unavailable. Generate a keypair
// with `go run ./scripts/genvapid/main.go`.
//
// VAPID_SUBJECT must be an https: URL: Apple's Web Push service rejects a
// mailto: subject with 403 BadJwtToken (the spec allows both, Apple does
// not), so the default and any mailto: override are coerced to https.
func newVAPIDConfig() notification.VAPIDConfig {
	pub := os.Getenv("VAPID_PUBLIC_KEY")
	if pub == "" {
		log.Println("VAPID_PUBLIC_KEY not set — Web Push is disabled")
		return notification.VAPIDConfig{}
	}

	subject := envOr("VAPID_SUBJECT", "https://sat0ru.dev")
	if strings.HasPrefix(subject, "mailto:") {
		log.Printf("VAPID_SUBJECT %q is a mailto: URI — Apple Web Push rejects those; using https://sat0ru.dev instead", subject)
		subject = "https://sat0ru.dev"
	}

	log.Printf("Web Push enabled (subject %s)", subject)
	return notification.VAPIDConfig{
		PublicKey:  pub,
		PrivateKey: os.Getenv("VAPID_PRIVATE_KEY"),
		Subject:    subject,
	}
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

// newDiaryVideoBlobStore builds the store the diary video log keeps clip
// bytes in. It reuses the Document Storage S3 bucket (same AWS creds / IAM)
// under a separate key prefix; without DOCUMENTS_S3_BUCKET it's the same
// local-disk fallback as documents, so dev and the no-AWS deployment work
// unchanged. Clips upload via a multipart upload (see internal/blobstore).
//
//	DOCUMENTS_S3_BUCKET       bucket for clip objects (enables S3 mode; shared with documents)
//	DIARY_VIDEO_S3_PREFIX     key prefix in that bucket (default "diary-videos")
//	DIARY_VIDEO_LOCAL_DIR     on-disk root when S3 is not configured (default "./.data/diary-videos")
//	PUBLIC_API_URL            this server's external base URL, for local-store signed URLs
func newDiaryVideoBlobStore(ctx context.Context, jwtSecret string) blobstore.Store {
	bucket := os.Getenv("DOCUMENTS_S3_BUCKET")
	if bucket == "" {
		dir := envOr("DIARY_VIDEO_LOCAL_DIR", "./.data/diary-videos")
		publicURL := envOr("PUBLIC_API_URL", "http://localhost:8080")
		store, err := blobstore.NewLocal(dir, []byte(jwtSecret), publicURL)
		if err != nil {
			log.Fatalf("diary video local blob store init failed: %v", err)
		}
		log.Printf("DOCUMENTS_S3_BUCKET not set — diary videos using local disk blob store at %s", dir)
		return store
	}

	prefix := envOr("DIARY_VIDEO_S3_PREFIX", "diary-videos")
	store, err := blobstore.NewS3(ctx, blobstore.S3Config{
		Bucket: bucket,
		Prefix: prefix,
		Region: os.Getenv("AWS_REGION"),
	})
	if err != nil {
		log.Fatalf("diary video S3 blob store init failed: %v", err)
	}
	log.Printf("diary video storage enabled → s3://%s/%s", bucket, prefix)
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
	notepadLabelRepo := notepadlabel.NewPostgresRepository(sqlDB)
	drawingBoardRepo := drawingboard.NewPostgresRepository(sqlDB)
	upskillRepo := upskill.NewPostgresRepository(sqlDB)
	diaryRepo := diary.NewPostgresRepository(sqlDB)
	diaryVideoRepo := diary.NewPostgresVideoRepository(sqlDB)
	fitnessRepo := fitness.NewPostgresRepository(sqlDB)
	spotifyRepo := spotify.NewPostgresRepository(sqlDB, spotifyCipher)
	workSessionRepo := worksession.NewPostgresRepository(sqlDB)
	cmsRepo := cms.NewPostgresRepository(sqlDB)
	cmsPublisher := newCMSPublisher(context.Background())
	documentRepo := document.NewPostgresRepository(sqlDB)
	documentLabelRepo := documentlabel.NewPostgresRepository(sqlDB)
	documentBlob := newDocumentBlobStore(context.Background(), jwtSecret)
	diaryVideoBlob := newDiaryVideoBlobStore(context.Background(), jwtSecret)
	notificationRepo := notification.NewPostgresRepository(sqlDB)
	reminderRepo := reminder.NewPostgresRepository(sqlDB)
	workProfileRepo := workprofile.NewPostgresRepository(sqlDB)
	mail := newMailer()
	vapid := newVAPIDConfig()

	router, sched := newRouter(
		authService, todoRepo, labelRepo, settingsRepo, notepadRepo, notepadLabelRepo, drawingBoardRepo, upskillRepo, diaryRepo,
		fitnessRepo, spotifyRepo, workSessionRepo, cmsRepo, cmsPublisher,
		documentRepo, documentLabelRepo, documentBlob,
		diaryVideoRepo, diaryVideoBlob,
		notificationRepo, reminderRepo, workProfileRepo, mail, vapid,
		spotifyClientID, spotifyClientSecret, spotifyRedirectURI, spotifyFrontendURL, jwtSecret,
	)

	schedCtx, schedCancel := context.WithCancel(context.Background())
	go sched.Run(schedCtx)

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
	schedCancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
}
