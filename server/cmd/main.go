package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sam-Frost/portfolio/internal/auth"
	"github.com/Sam-Frost/portfolio/internal/db"
	"github.com/Sam-Frost/portfolio/internal/label"
	"github.com/Sam-Frost/portfolio/internal/settings"
	"github.com/Sam-Frost/portfolio/internal/todo"
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
// first place) and the health check (used by container orchestration).
var publicPaths = map[string]bool{
	"/health":         true,
	"/api/auth/login": true,
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
func newRouter(authService *auth.Service, todoRepo todo.Repository, labelRepo label.Repository, settingsRepo settings.Repository) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthCheck)

	auth.NewHandler(authService).Register(mux)

	todoService := todo.NewService(todoRepo)
	todo.NewHandler(todoService).Register(mux)

	labelService := label.NewService(labelRepo)
	label.NewHandler(labelService).Register(mux)

	settingsService := settings.NewService(settingsRepo)
	settings.NewHandler(settingsService).Register(mux)

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

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           withCORS(allowedOrigin, withAuth(authService, newRouter(authService, todoRepo, labelRepo, settingsRepo))),
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
