package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/niyiayooluwa/geotas/internal/db"
	"github.com/niyiayooluwa/geotas/internal/middleware"
	"github.com/niyiayooluwa/geotas/internal/repository"
	"github.com/niyiayooluwa/geotas/internal/service"
)

func NewRouter(queries *db.Queries, authService *service.AuthService) *chi.Mux {
	var router *chi.Mux = chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://geotas.vercel.app", "http://localhost:*", "http://127.0.0.1:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	router.Use(chiMiddleware.Logger)
	router.Use(chiMiddleware.Recoverer)
	router.Use(chiMiddleware.Compress(5))

	// wire up repositories
	var courseRepo *repository.CourseRepository = repository.NewCourseRepository(queries)
	var qrRepo *repository.QRTokenRepository = repository.NewQRTokenRepository(queries)
	var sessionRepo *repository.SessionRepository = repository.NewSessionRepository(queries)
	var attendanceRepo *repository.AttendanceRepository = repository.NewAttendanceRepository(queries)
	var otpRepo *repository.OTPRepository = repository.NewOTPRepository(queries)
	var notifRepo *repository.NotificationRepository = repository.NewNotificationRepository(queries)
	var scheduleRepo *repository.ScheduleRepository = repository.NewScheduleRepository(queries)

	// wire up services
	var notifService *service.NotificationService = service.NewNotificationService(notifRepo)
	var scheduleService *service.ScheduleService = service.NewScheduleService(scheduleRepo, courseRepo, notifService)
	var courseService *service.CourseService = service.NewCourseService(courseRepo, sessionRepo, attendanceRepo)
	var qrManager *service.QRRotationManager = service.NewQRRotationManager(qrRepo)
	var sessionService *service.SessionService = service.NewSessionService(sessionRepo, courseRepo, qrManager, notifService)
	var otpService *service.OTPService = service.NewOTPService(otpRepo, sessionRepo, courseRepo)
	var attendanceService *service.AttendanceService = service.NewAttendanceService(attendanceRepo, sessionRepo, courseRepo, qrRepo, otpRepo)

	// wire up handlers
	var authHandler *AuthHandler = NewAuthHandler(authService)
	var userHandler *UserHandler = NewUserHandler(authService)
	var courseHandler *CourseHandler = NewCourseHandler(courseService)
	var sessionHandler *SessionHandler = NewSessionHandler(sessionService)
	var attendanceHandler *AttendanceHandler = NewAttendanceHandler(attendanceService, otpService)
	var notifHandler *NotificationHandler = NewNotificationHandler(notifService)
	var scheduleHandler *ScheduleHandler = NewScheduleHandler(scheduleService)

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("GEOTAS is alive"))
	})
	router.Post("/auth/google", authHandler.GoogleLogin)
	router.Post("/auth/register", authHandler.RegisterLecturer)
	router.Post("/auth/login", authHandler.LoginLecturer)
	router.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleWare)
		r.Get("/me", userHandler.Me)

		// static course routes must come before wildcard {id} routes
		r.With(middleware.RequireRole("lecturer")).Post("/courses", courseHandler.CreateCourse)
		r.Get("/courses", courseHandler.GetMyCourses)
		r.Post("/courses/join", courseHandler.JoinCourse)
		r.Get("/courses/enrolled", courseHandler.GetEnrolledCourses) // must be above /{id}

		// wildcard course routes
		r.Delete("/courses/{id}", courseHandler.DeleteCourse)
		r.Delete("/courses/{id}/leave", courseHandler.LeaveCourse)
		r.Get("/courses/{id}/members", courseHandler.GetCourseMembers)
		r.Get("/courses/{id}/attendance", courseHandler.GetCourseAttendance)
		r.Delete("/courses/{id}/members/{targetId}", courseHandler.RemoveStudent)
		r.Post("/courses/{id}/invite-code/rotate", courseHandler.RotateInviteCode)
		r.With(middleware.RequireRole("lecturer")).Patch("/courses/{id}/settings", courseHandler.UpdateCourseSettings)

		r.Post("/sessions", sessionHandler.CreateSession)
		r.Get("/courses/{courseId}/sessions", sessionHandler.GetSessionsByCourse)
		r.Patch("/sessions/{id}/close", sessionHandler.CloseSession)
		r.Delete("/sessions/{id}", sessionHandler.DeleteSession)
		r.Get("/sessions/{id}/qr-token", sessionHandler.GetLiveQRToken)
		r.Get("/sessions/{id}/attendance", attendanceHandler.GetAttendanceBySession)

		r.Post("/attendance/qr", attendanceHandler.MarkAttendanceQR)
		r.Post("/attendance/otp/request", attendanceHandler.RequestOTP)
		r.Post("/attendance/otp/verify", attendanceHandler.MarkAttendanceOTP)

		r.With(middleware.RequireRole("lecturer")).Post("/courses/{id}/schedules", scheduleHandler.CreateSchedule)
		r.Get("/courses/{id}/schedules", scheduleHandler.GetSchedules)
		r.With(middleware.RequireRole("lecturer")).Patch("/schedules/{id}", scheduleHandler.UpdateSchedule)
		r.With(middleware.RequireRole("lecturer")).Delete("/schedules/{id}", scheduleHandler.DeleteSchedule)

		r.Get("/notifications", notifHandler.GetUnseenNotifications)
		r.Post("/notifications/{id}/seen", notifHandler.MarkSeen)
	})

	return router
}