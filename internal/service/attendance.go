package service

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/niyiayooluwa/geotas/internal/db"
	"github.com/niyiayooluwa/geotas/internal/model"
	"github.com/niyiayooluwa/geotas/internal/repository"
)

type AttendanceService struct {
	attendanceRepo *repository.AttendanceRepository
	sessionRepo    *repository.SessionRepository
	courseRepo     *repository.CourseRepository
	qrRepo         *repository.QRTokenRepository
	otpRepo        *repository.OTPRepository
}

func NewAttendanceService(
	attendanceRepo *repository.AttendanceRepository,
	sessionRepo *repository.SessionRepository,
	courseRepo *repository.CourseRepository,
	qrRepo *repository.QRTokenRepository,
	otpRepo *repository.OTPRepository,
) *AttendanceService {
	return &AttendanceService{
		attendanceRepo: attendanceRepo,
		sessionRepo:    sessionRepo,
		courseRepo:     courseRepo,
		qrRepo:         qrRepo,
		otpRepo:        otpRepo,
	}
}

func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000 // meters
	dLat := (lat2 - lat1) * (math.Pi / 180)
	dLon := (lon2 - lon1) * (math.Pi / 180)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*(math.Pi/180))*math.Cos(lat2*(math.Pi/180))*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

func computeConfidenceScore(
	dist, radius float64,
	isQR bool,
	mockDetected bool,
	duplicateDevice bool,
	deviceSwitched bool,
	sessionStartedAt time.Time,
) float64 {
	score := 1.0

	if dist > radius {
		score -= 0.50
	} else {
		score -= (dist / radius) * 0.15
	}

	if !isQR {
		score -= 0.10
	}

	elapsed := time.Since(sessionStartedAt).Minutes()
	if elapsed >= 60 {
		score -= 0.15
	} else {
		score -= (elapsed / 60.0) * 0.10
	}

	if mockDetected {
		score -= 0.40
	}

	if duplicateDevice {
		score -= 0.30
	}

	if deviceSwitched {
		score -= 0.20
	}

	return math.Max(0, math.Min(1.0, score))
}

func (s *AttendanceService) MarkAttendanceQR(ctx context.Context, userID string, req model.MarkAttendanceQRRequest) (model.AttendanceResponse, error) {
	token, err := s.qrRepo.GetValidQRToken(ctx, req.QRToken)
	if err != nil {
		return model.AttendanceResponse{}, errors.New("Invalid or expired QR token")
	}

	sessionUUID, err := parseUUID(req.SessionID)
	if err != nil {
		return model.AttendanceResponse{}, errors.New("Invalid session ID")
	}

	if token.SessionID != sessionUUID {
		return model.AttendanceResponse{}, errors.New("Token does not belong to this session")
	}

	// invalidate token
	s.qrRepo.MarkQRTokenUsed(ctx, token.ID)

	return s.mark(ctx, userID, "qr", sessionUUID, req.Latitude, req.Longitude, req.DeviceID, req.DeviceModel, req.OsVersion, req.MockLocationDetected)
}

func (s *AttendanceService) MarkAttendanceOTP(ctx context.Context, userID string, req model.MarkAttendanceOTPRequest) (model.AttendanceResponse, error) {
	userUUID, err := parseUUID(userID)
	if err != nil {
		return model.AttendanceResponse{}, err
	}

	sessionUUID, err := parseUUID(req.SessionID)
	if err != nil {
		return model.AttendanceResponse{}, errors.New("Invalid session ID")
	}

	otp, err := s.otpRepo.GetValidOTP(ctx, db.GetValidOTPParams{
		UserID:    userUUID,
		SessionID: sessionUUID,
		Code:      req.OTPCode,
	})
	if err != nil {
		return model.AttendanceResponse{}, errors.New("Invalid or expired OTP")
	}

	s.otpRepo.MarkOTPUsed(ctx, otp.ID)

	return s.mark(ctx, userID, "otp", sessionUUID, req.Latitude, req.Longitude, req.DeviceID, req.DeviceModel, req.OsVersion, req.MockLocationDetected)
}

func (s *AttendanceService) GetAttendanceBySession(ctx context.Context, userID, sessionID string) ([]model.DetailedAttendanceResponse, error) {
	sessionUUID, err := parseUUID(sessionID)
	if err != nil {
		return nil, errors.New("Invalid session ID")
	}

	userUUID, err := parseUUID(userID)
	if err != nil {
		return nil, err
	}

	// fetch session to confirm course ownership
	session, err := s.sessionRepo.GetSessionByID(ctx, sessionUUID)
	if err != nil {
		return nil, errors.New("Session not found")
	}

	course, err := s.courseRepo.GetCourseByID(ctx, session.CourseID)
	if err != nil {
		return nil, errors.New("Course not found")
	}

	if course.OwnerID != userUUID {
		return nil, errors.New("You do not own this course")
	}

	records, err := s.attendanceRepo.GetAttendanceBySession(ctx, sessionUUID)
	if err != nil {
		return nil, errors.New("Could not fetch attendance records")
	}

	var response []model.DetailedAttendanceResponse
	for _, r := range records {
		response = append(response, model.DetailedAttendanceResponse{
			ID:                   r.ID.String(),
			SessionID:            r.SessionID.String(),
			UserID:               r.UserID.String(),
			MarkedAt:             r.MarkedAt.Time.Format("2006-01-02 15:04:05"),
			Method:               r.Method,
			Latitude:             r.Latitude,
			Longitude:            r.Longitude,
			DistanceFromCenter:   r.DistanceFromCenter,
			MockLocationDetected: r.MockLocationDetected,
			ConfidenceScore:      r.ConfidenceScore,
			WeekNumber:           r.WeekNumber,
			DeviceID:             r.DeviceID.String,
			DeviceModel:          r.DeviceModel.String,
			OsVersion:            r.OsVersion.String,
			FirstName:            r.FirstName,
			LastName:             r.LastName,
			MatriculationNumber:  r.MatriculationNumber.String,
		})
	}

	return response, nil
}

func (s *AttendanceService) mark(
	ctx context.Context,
	userID string,
	method string,
	sessionUUID pgtype.UUID,
	lat, lon float64,
	deviceID, deviceModel, osVersion string,
	mockDetected bool,
) (model.AttendanceResponse, error) {
	userUUID, err := parseUUID(userID)
	if err != nil {
		return model.AttendanceResponse{}, err
	}

	session, err := s.sessionRepo.GetSessionByID(ctx, sessionUUID)
	if err != nil {
		return model.AttendanceResponse{}, errors.New("Session not found")
	}

	if session.Status != "active" {
		return model.AttendanceResponse{}, errors.New("Session is not active")
	}

	// hard block — mock location has no legitimate excuse
	if mockDetected {
		return model.AttendanceResponse{}, errors.New("Mock location detected — attendance cannot be marked")
	}

	// enrollment check
	_, err = s.courseRepo.GetCourseMember(ctx, db.GetCourseMemberParams{
		CourseID: session.CourseID,
		UserID:   userUUID,
	})
	if err != nil {
		return model.AttendanceResponse{}, errors.New("Not enrolled in course")
	}

	// duplicate attendance check
	_, err = s.attendanceRepo.GetAttendanceByUserAndSession(ctx, userUUID, sessionUUID)
	if err == nil {
		return model.AttendanceResponse{}, errors.New("Attendance already marked")
	}

	// duplicate device in this session — flag, do not block
	duplicateDevice := false
	_, err = s.attendanceRepo.CheckDuplicateDevice(ctx, sessionUUID, deviceID)
	if err == nil {
		duplicateDevice = true
	}

	// device switching — compare against prior closed sessions in this course
	deviceSwitched := false
	priorDeviceID, err := s.attendanceRepo.GetPrimaryDeviceForUser(ctx, userUUID, session.CourseID)
	if err == nil && priorDeviceID != deviceID {
		deviceSwitched = true
	}

	dist := calculateDistance(lat, lon, session.Latitude, session.Longitude)
	score := computeConfidenceScore(
		dist,
		session.RadiusMeters,
		method == "qr",
		mockDetected,
		duplicateDevice,
		deviceSwitched,
		session.StartedAt.Time,
	)

	record, err := s.attendanceRepo.CreateAttendanceRecord(ctx, db.CreateAttendanceRecordParams{
		SessionID:            sessionUUID,
		UserID:               userUUID,
		Method:               method,
		Latitude:             lat,
		Longitude:            lon,
		DistanceFromCenter:   dist,
		MockLocationDetected: mockDetected,
		ConfidenceScore:      score,
		WeekNumber:           session.WeekNumber,
		DeviceID:             pgtype.Text{String: deviceID, Valid: deviceID != ""},
		DeviceModel:          pgtype.Text{String: deviceModel, Valid: deviceModel != ""},
		OsVersion:            pgtype.Text{String: osVersion, Valid: osVersion != ""},
	})
	if err != nil {
		return model.AttendanceResponse{}, errors.New("Failed to record attendance")
	}

	return model.AttendanceResponse{
		ID:         record.ID.String(),
		SessionID:  record.SessionID.String(),
		UserID:     record.UserID.String(),
		MarkedAt:   record.MarkedAt.Time.Format("2006-01-02 15:04:05"),
		Method:     record.Method,
		WeekNumber: record.WeekNumber,
	}, nil
}

