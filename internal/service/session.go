package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/niyiayooluwa/geotas/internal/db"
	"github.com/niyiayooluwa/geotas/internal/model"
	"github.com/niyiayooluwa/geotas/internal/repository"
)

type SessionService struct {
	sessionRepo *repository.SessionRepository
	courseRepo  *repository.CourseRepository
}

func NewSessionService(sessionRepo *repository.SessionRepository, courseRepo *repository.CourseRepository) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
		courseRepo:  courseRepo,
	}
}

func (s *SessionService) CreateSession(ctx context.Context, userID string, req model.CreateSessionRequest) (model.SessionResponse, error) {
	// validate input
	if req.CourseID == "" {
		return model.SessionResponse{}, errors.New("course_id is required")
	}
	if req.WeekNumber <= 0 {
		return model.SessionResponse{}, errors.New("week_number must be greater than zero")
	}
	if req.RadiusMeters <= 0 {
		return model.SessionResponse{}, errors.New("radius_meters must be greater than zero")
	}
	if req.Latitude == 0 || req.Longitude == 0 {
		return model.SessionResponse{}, errors.New("valid coordinates are required")
	}

	// default QR rotation to 30 seconds if not provided
	if req.QrRotationSecs <= 0 {
		req.QrRotationSecs = 30
	}

	// parse UUIDs
	createdBy, err := parseUUID(userID)
	if err != nil {
		return model.SessionResponse{}, err
	}

	courseID, err := parseUUID(req.CourseID)
	if err != nil {
		return model.SessionResponse{}, errors.New("invalid course_id")
	}

	// confirm the course exists and belongs to this lecturer
	course, err := s.courseRepo.GetCourseByID(ctx, courseID)
	if err != nil {
		return model.SessionResponse{}, errors.New("course not found")
	}

	if course.OwnerID != createdBy {
		return model.SessionResponse{}, errors.New("you do not own this course")
	}

	// check no active session already exists for this course
	_, err = s.sessionRepo.GetActiveSessionByCourse(ctx, courseID)
	if err == nil {
		return model.SessionResponse{}, errors.New("an active session already exists for this course")
	}

	// create the session
	session, err := s.sessionRepo.CreateSession(ctx, db.CreateSessionParams{
		CourseID:  courseID,
		CreatedBy: createdBy,
		Title: pgtype.Text{
			String: req.Title,
			Valid:  req.Title != "",
		},
		WeekNumber:     req.WeekNumber,
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
		RadiusMeters:   req.RadiusMeters,
		QrRotationSecs: req.QrRotationSecs,
	})
	if err != nil {
		return model.SessionResponse{}, errors.New("could not create session")
	}

	return buildSessionResponse(session), nil
}

func (s *SessionService) CloseSession(ctx context.Context, userID string, sessionID string) (model.SessionResponse, error) {
	// parse UUIDs
	parsedSessionID, err := parseUUID(sessionID)
	if err != nil {
		return model.SessionResponse{}, errors.New("invalid session_id")
	}

	parsedUserID, err := parseUUID(userID)
	if err != nil {
		return model.SessionResponse{}, err
	}

	// fetch session
	session, err := s.sessionRepo.GetSessionByID(ctx, parsedSessionID)
	if err != nil {
		return model.SessionResponse{}, errors.New("session not found")
	}

	// confirm ownership
	if session.CreatedBy != parsedUserID {
		return model.SessionResponse{}, errors.New("you do not own this session")
	}

	// confirm it's still active
	if session.Status == "closed" {
		return model.SessionResponse{}, errors.New("session is already closed")
	}

	// close it
	closed, err := s.sessionRepo.CloseSession(ctx, parsedSessionID)
	if err != nil {
		return model.SessionResponse{}, errors.New("could not close session")
	}

	return buildSessionResponse(closed), nil
}

func (s *SessionService) GetSessionsByCourse(ctx context.Context, userID string, courseID string) ([]model.SessionResponse, error) {
	parsedCourseID, err := parseUUID(courseID)
	if err != nil {
		return nil, errors.New("invalid course_id")
	}

	sessions, err := s.sessionRepo.GetSessionsByCourse(ctx, parsedCourseID)
	if err != nil {
		return nil, errors.New("could not fetch sessions")
	}

	var response []model.SessionResponse
	for _, session := range sessions {
		response = append(response, buildSessionResponse(session))
	}

	return response, nil
}

// builds a SessionResponse from a db.Session
func buildSessionResponse(session db.Session) model.SessionResponse {
	var closedAt string
	if session.ClosedAt.Valid {
		closedAt = session.ClosedAt.Time.Format("2006-01-02 15:04:05")
	}

	return model.SessionResponse{
		ID:             session.ID.String(),
		CourseID:       session.CourseID.String(),
		CreatedBy:      session.CreatedBy.String(),
		Title:          session.Title.String,
		WeekNumber:     session.WeekNumber,
		Latitude:       session.Latitude,
		Longitude:      session.Longitude,
		RadiusMeters:   session.RadiusMeters,
		QrRotationSecs: session.QrRotationSecs,
		Status:         session.Status,
		StartedAt:      session.StartedAt.Time.Format("2006-01-02 15:04:05"),
		ClosedAt:       closedAt,
	}
}