package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/niyiayooluwa/geotas/internal/db"
	"github.com/niyiayooluwa/geotas/internal/middleware"

)

// NewRouter accepts queries so handlers can talk to the DB
func NewRouter(queries *db.Queries) *chi.Mux {
	var router *chi.Mux = chi.NewRouter()

	router.Use(chiMiddleware.Logger)
	router.Use(chiMiddleware.Recoverer)

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("GEOTAS is alive"))
	})

	router.Post("/auth/register", RegisterHandler(queries))

	router.Post("/auth/login", LoginHandler(queries))

	router.Group(func (r chi.Router)  {
		r.Use(middleware.AuthMiddleWare)
		// protected endpoints go here
	r.Get("/me", MeHandler(queries))
	})

	return router
}
