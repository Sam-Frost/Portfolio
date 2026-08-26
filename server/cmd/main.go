package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata" // embeds the IANA zoneinfo DB so time.LoadLocation("Asia/Kolkata") (internal/diary) works even on a host/image without a system tzdata package (e.g. the alpine runtime image)

	"github.com/Sam-Frost/portfolio/internal/auth"
	"github.com/Sam-Frost/portfolio/internal/db"
	"github.com/Sam-Frost/portfolio/internal/diary"
	"github.com/Sam-Frost/portfolio/internal/label"
	"github.com/Sam-Frost/portfolio/internal/notepad"
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
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
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
		if publicPaths[r.URL.Path] {
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
	spotifyRepo spotify.Repository,
	workSessionRepo worksession.Repository,
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

	spotifyClient := spotify.NewAPIClient(spotifyClientID, spotifyClientSecret, spotifyRedirectURI)
	spotifyService := spotify.NewService(spotifyRepo, spotifyClient, []byte(jwtSecret))
	spotify.NewHandler(spotifyService, spotifyFrontendURL).Register(mux)

	workSessionService := worksession.NewService(workSessionRepo)
	worksession.NewHandler(workSessionService).Register(mux)

	return mux
}

func requireEnv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		log.Fatalf("%s is required", name)
	}
	return v
}

func main() {
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
	spotifyRepo := spotify.NewPostgresRepository(sqlDB, spotifyCipher)
	workSessionRepo := worksession.NewPostgresRepository(sqlDB)

	router := newRouter(
		authService, todoRepo, labelRepo, settingsRepo, notepadRepo, upskillRepo, diaryRepo,
		spotifyRepo, workSessionRepo, spotifyClientID, spotifyClientSecret, spotifyRedirectURI, spotifyFrontendURL, jwtSecret,
	)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           withCORS(allowedOrigin, withAuth(authService, router)),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		fmt.Println("Server is starting to listen on port 8080...")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	fmt.Println("Shutting down...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
}
