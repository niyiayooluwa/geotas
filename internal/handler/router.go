package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/niyiayooluwa/geotas/internal/db"
	"github.com/niyiayooluwa/geotas/internal/middleware"
	"github.com/niyiayooluwa/geotas/internal/repository"
	"github.com/niyiayooluwa/geotas/internal/service"
)

func NewRouter(queries *db.Queries) *chi.Mux {
	var router *chi.Mux = chi.NewRouter()

	router.Use(chiMiddleware.Logger)
	router.Use(chiMiddleware.Recoverer)

	// wire up dependencies
	var userRepo *repository.UserRepository = repository.NewUserRepository(queries)
	var authService *service.AuthService = service.NewAuthService(userRepo)
	var authHandler *AuthHandler = NewAuthHandler(authService)
	var userHandler *UserHandler = NewUserHandler(authService)

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("GEOTAS is alive"))
	})

	// public routes
	router.Post("/auth/register", authHandler.Register)
	router.Post("/auth/login", authHandler.Login)

	// protected routes
	router.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleWare)
		r.Get("/me", userHandler.Me)
	})

	return router
}