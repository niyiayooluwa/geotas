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

// calculateDistance computes the distance between two points using the Haversine formula (in meters)
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

// computeConfidenceScore calculates a score between 0.0 and 1.0 based on multiple trust factors
func computeConfidenceScore(
	dist, radius float64,
	isQR bool,
	mockDetected bool,
	duplicateDevice bool,
	sessionStartedAt time.Time,
) float64 {
	var score float64 = 0.0

	// 1. Geofence pass (+0.35)
	if dist <= radius {
		score += 0.35
	}

	// 2. Distance decay (+0.20)
	// Closer to center = higher score, linear within radius
	if dist <= radius {
		score += (1.0 - (dist / radius)) * 0.20
	}

	// 3. Mark method (+0.15 for QR, +0.05 for OTP)
	if isQR {
		score += 0.15
	} else {
		score += 0.05
	}

	// 4. Scan time (+0.15)
	// Early in session = higher, decays over 1 hour
	elapsed := time.Since(sessionStartedAt).Minutes()
	if elapsed < 60 {
		score += (1.0 - (elapsed / 60.0)) * 0.15
	}

	// 5. Mock location not detected (+0.10)
	if !mockDetected {
		score += 0.10
	}

	// 6. No duplicate device in session (+0.05)
	if !duplicateDevice {
		score += 0.05
	}

	return math.Max(0, math.Min(1.0, score))
}

func (s *AttendanceService) MarkAttendanceQR(ctx context.Context, userID string, req model.MarkAttendanceQRRequest) (model.AttendanceResponse, error) {
	// validate token
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

	// invalidate token after use (single-use)
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

	// fetch valid OTP
	otp, err := s.otpRepo.GetValidOTP(ctx, db.GetValidOTPParams{
		UserID:    userUUID,
		SessionID: sessionUUID,
		Code:      req.OTPCode,
	})
	if err != nil {
		return model.AttendanceResponse{}, errors.New("Invalid or expired OTP code")
	}

	// mark OTP used
	s.otpRepo.MarkOTPUsed(ctx, otp.ID)

	return s.mark(ctx, userID, "otp", sessionUUID, req.Latitude, req.Longitude, req.DeviceID, req.DeviceModel, req.OsVersion, req.MockLocationDetected)
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

	// fetch session
	session, err := s.sessionRepo.GetSessionByID(ctx, sessionUUID)
	if err != nil {
		return model.AttendanceResponse{}, errors.New("Session not found")
	}

	if session.Status != "active" {
		return model.AttendanceResponse{}, errors.New("Session is no longer active")
	}

	// check enrollment
	_, err = s.courseRepo.GetCourseMember(ctx, db.GetCourseMemberParams{
		CourseID: session.CourseID,
		UserID:   userUUID,
	})
	if err != nil {
		return model.AttendanceResponse{}, errors.New("You are not enrolled in this course")
	}

	// check duplicate attendance
	_, err = s.attendanceRepo.GetAttendanceByUserAndSession(ctx, userUUID, sessionUUID)
	if err == nil {
		return model.AttendanceResponse{}, errors.New("You have already marked attendance for this session")
	}

	// check duplicate device
	duplicateDevice := false
	_, err = s.attendanceRepo.CheckDuplicateDevice(ctx, sessionUUID, deviceID)
	if err == nil {
		duplicateDevice = true
	}

	// calculate geofence
	dist := calculateDistance(lat, lon, session.Latitude, session.Longitude)

	// compute confidence score
	score := computeConfidenceScore(dist, session.RadiusMeters, method == "qr", mockDetected, duplicateDevice, session.StartedAt.Time)

	// save record
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
