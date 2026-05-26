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
	router.Use(chiMiddleware.Compress(5)) // add this

	// wire up repositories
	var userRepo *repository.UserRepository = repository.NewUserRepository(queries)
	var courseRepo *repository.CourseRepository = repository.NewCourseRepository(queries)
	var qrRepo *repository.QRTokenRepository = repository.NewQRTokenRepository(queries)
	var sessionRepo *repository.SessionRepository = repository.NewSessionRepository(queries)
	var attendanceRepo *repository.AttendanceRepository = repository.NewAttendanceRepository(queries)
	var otpRepo *repository.OTPRepository = repository.NewOTPRepository(queries)

	// wire up services
	var authService *service.AuthService = service.NewAuthService(userRepo)
	var courseService *service.CourseService = service.NewCourseService(courseRepo)
	var qrManager *service.QRRotationManager = service.NewQRRotationManager(qrRepo)
	var sessionService *service.SessionService = service.NewSessionService(sessionRepo, courseRepo, qrManager)
	var otpService *service.OTPService = service.NewOTPService(otpRepo, sessionRepo, courseRepo)
	var attendanceService *service.AttendanceService = service.NewAttendanceService(attendanceRepo, sessionRepo, courseRepo, qrRepo, otpRepo)

	// wire up handlers
	var authHandler *AuthHandler = NewAuthHandler(authService)
	var userHandler *UserHandler = NewUserHandler(authService)
	var courseHandler *CourseHandler = NewCourseHandler(courseService)
	var sessionHandler *SessionHandler = NewSessionHandler(sessionService)
	var attendanceHandler *AttendanceHandler = NewAttendanceHandler(attendanceService, otpService)

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("GEOTAS is alive"))
	})

	// public routes
	router.Post("/auth/google", authHandler.GoogleLogin)

	// protected routes
	router.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleWare)
		r.Get("/me", userHandler.Me)

		// course routes
		r.Post("/courses", courseHandler.CreateCourse)
		r.Post("/courses/join", courseHandler.JoinCourse)
		r.Get("/courses", courseHandler.GetMyCourses)
		r.Delete("/courses/{id}", courseHandler.DeleteCourse)
		r.Get("/courses/enrolled", courseHandler.GetEnrolledCourses)

		// session routes
		r.Post("/sessions", sessionHandler.CreateSession)
		r.Get("/courses/{courseId}/sessions", sessionHandler.GetSessionsByCourse)
		r.Patch("/sessions/{id}/close", sessionHandler.CloseSession)
		r.Delete("/sessions/{id}", sessionHandler.DeleteSession)
		r.Get("/sessions/{id}/qr-token", sessionHandler.GetLiveQRToken)
		r.Get("/sessions/{id}/attendance", attendanceHandler.GetAttendanceBySession)

		// attendance routes
		r.Post("/attendance/qr", attendanceHandler.MarkAttendanceQR)
		r.Post("/attendance/otp/request", attendanceHandler.RequestOTP)
		r.Post("/attendance/otp/verify", attendanceHandler.MarkAttendanceOTP)

	})

	return router
}
