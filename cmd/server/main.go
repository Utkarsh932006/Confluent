package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"confluent/internal/auth"
	"confluent/internal/config"
	"confluent/internal/database"
	"confluent/internal/handlers"
	"confluent/internal/models"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg := config.Load()

	db, err := database.Open(cfg.DBPath, cfg.AdminPass)
	if err != nil {
		slog.Error("Failed to initialize database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	// Instantiate Handlers
	authH := handlers.NewAuthHandler(db, cfg.JWTSecret)
	spacesH := handlers.NewSpacesHandler(db)
	pagesH := handlers.NewPagesHandler(db)
	commentsH := handlers.NewCommentsHandler(db)
	searchH := handlers.NewSearchHandler(db)
	adminH := handlers.NewAdminHandler(db)

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		handlers.JSON(w, http.StatusOK, map[string]string{"status": "ok", "app": "confluent"})
	})

	// Auth Endpoints
	mux.HandleFunc("POST /api/v1/auth/register", authH.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authH.Login)
	mux.HandleFunc("GET /api/v1/auth/me", auth.RequireAuth(authH.Me))

	// Spaces Endpoints
	mux.HandleFunc("GET /api/v1/spaces", auth.RequireAuth(spacesH.List))
	mux.HandleFunc("GET /api/v1/spaces/{slug}", auth.RequireAuth(spacesH.Get))
	mux.HandleFunc("POST /api/v1/spaces", auth.RequireRole(models.RoleAdmin)(spacesH.Create))
	mux.HandleFunc("DELETE /api/v1/spaces/{slug}", auth.RequireRole(models.RoleAdmin)(spacesH.Delete))

	// Pages Endpoints
	mux.HandleFunc("GET /api/v1/spaces/{slug}/pages", auth.RequireAuth(pagesH.Tree))
	mux.HandleFunc("POST /api/v1/spaces/{slug}/pages", auth.RequireRole(models.RoleAdmin, models.RoleEditor)(pagesH.Create))
	mux.HandleFunc("GET /api/v1/pages/{id}", auth.RequireAuth(pagesH.Get))
	mux.HandleFunc("PUT /api/v1/pages/{id}", auth.RequireRole(models.RoleAdmin, models.RoleEditor)(pagesH.Update))
	mux.HandleFunc("DELETE /api/v1/pages/{id}", auth.RequireRole(models.RoleAdmin, models.RoleEditor)(pagesH.Delete))
	mux.HandleFunc("GET /api/v1/pages/{id}/versions", auth.RequireAuth(pagesH.Versions))
	mux.HandleFunc("GET /api/v1/pages/{id}/versions/{num}", auth.RequireAuth(pagesH.GetVersion))

	// Comments Endpoints
	mux.HandleFunc("GET /api/v1/pages/{id}/comments", auth.RequireAuth(commentsH.List))
	mux.HandleFunc("POST /api/v1/pages/{id}/comments", auth.RequireAuth(commentsH.Create))
	mux.HandleFunc("DELETE /api/v1/comments/{id}", auth.RequireAuth(commentsH.Delete))

	// Search Endpoint
	mux.HandleFunc("GET /api/v1/search", auth.RequireAuth(searchH.Search))

	// Admin Endpoints
	mux.HandleFunc("GET /api/v1/admin/users", auth.RequireRole(models.RoleAdmin)(adminH.ListUsers))
	mux.HandleFunc("PUT /api/v1/admin/users/{id}/role", auth.RequireRole(models.RoleAdmin)(adminH.UpdateRole))
	mux.HandleFunc("GET /api/v1/admin/stats", auth.RequireRole(models.RoleAdmin)(adminH.Stats))

	// Wrap with Global Middlewares: CORS + Auth Token Extraction
	handler := corsMiddleware(auth.Middleware(cfg.JWTSecret)(mux))

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown channel
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("Confluent server listening", "port", cfg.Port, "addr", "http://localhost:"+cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server failed", "err", err)
			os.Exit(1)
		}
	}()

	<-stop
	slog.Info("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server shutdown failed", "err", err)
	}
	slog.Info("Server stopped cleanly")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
