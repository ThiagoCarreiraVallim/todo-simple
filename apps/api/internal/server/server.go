package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/thiago/todo-simple-api/internal/config"
	"github.com/thiago/todo-simple-api/internal/health"
	"github.com/thiago/todo-simple-api/internal/lists"
)

// New builds the HTTP server: middleware, routes, and dependency wiring.
// Business logic lives in each domain's service — handlers only translate
// HTTP <-> Go types.
func New(cfg config.Config, pool *pgxpool.Pool) *http.Server {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.WebOrigin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: false,
		MaxAge:           600,
	}))

	r.Get("/health", health.Handler(pool))

	listsHandler := lists.NewHandler(lists.NewService(lists.NewRepository(pool)))
	r.Route("/api/lists", listsHandler.Routes)

	return &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}
}
