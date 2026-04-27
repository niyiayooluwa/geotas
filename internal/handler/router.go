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

	// wire up course dependencies
	var courseRepo *repository.CourseRepository = repository.NewCourseRepository(queries)
	var courseService *service.CourseService = service.NewCourseService(courseRepo)
	var courseHandler *CourseHandler = NewCourseHandler(courseService)

	// wire up session dependencies
	var sessionRepo *repository.SessionRepository = repository.NewSessionRepository(queries)
	var sessionService *service.SessionService = service.NewSessionService(sessionRepo, courseRepo)
	var sessionHandler *SessionHandler = NewSessionHandler(sessionService)

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

		// course routes
		r.Post("/courses", courseHandler.CreateCourse)
		r.Post("/courses/join", courseHandler.JoinCourse)
		r.Get("/courses", courseHandler.GetMyCourses)
		r.Delete("/courses/{id}", courseHandler.DeleteCourse)

		// session routes
		r.Post("/sessions", sessionHandler.CreateSession)
		r.Get("/courses/{courseId}/sessions", sessionHandler.GetSessionsByCourse)
		r.Patch("/sessions/{id}/close", sessionHandler.CloseSession)
		r.Delete("/sessions/{id}", sessionHandler.DeleteSession)

	})

	return router
}
