package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/niyiayooluwa/geotas/internal/db"
)

// NewRouter accepts queries so handlers can talk to the DB
func NewRouter(queries *db.Queries) *chi.Mux {
	var router *chi.Mux = chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("GEOTAS is alive"))
	})

	router.Post("/auth/register", RegisterHandler(queries))

	router.Post("/auth/login", LoginHandler(queries))

	return router
}
