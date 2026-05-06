package service

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/niyiayooluwa/geotas/internal/db"
	"github.com/niyiayooluwa/geotas/internal/model"
	"github.com/niyiayooluwa/geotas/internal/repository"
)

type OTPService struct {
	otpRepo     *repository.OTPRepository
	sessionRepo *repository.SessionRepository
	courseRepo  *repository.CourseRepository
}

func NewOTPService(otpRepo *repository.OTPRepository, sessionRepo *repository.SessionRepository, courseRepo *repository.CourseRepository) *OTPService {
	return &OTPService{
		otpRepo:     otpRepo,
		sessionRepo: sessionRepo,
		courseRepo:  courseRepo,
	}
}

func generateOTP() (string, error) {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	code := binary.BigEndian.Uint32(b[:]) % 1000000
	return fmt.Sprintf("%06d", code), nil
}

func (s *OTPService) RequestOTP(ctx context.Context, userID string, sessionID string) (model.OTPResponse, error) {
	userUUID, err := parseUUID(userID)
	if err != nil {
		return model.OTPResponse{}, err
	}

	sessionUUID, err := parseUUID(sessionID)
	if err != nil {
		return model.OTPResponse{}, errors.New("Invalid session ID")
	}

	// check session status
	session, err := s.sessionRepo.GetSessionByID(ctx, sessionUUID)
	if err != nil {
		return model.OTPResponse{}, errors.New("Session not found")
	}

	if session.Status != "active" {
		return model.OTPResponse{}, errors.New("Session is not active")
	}

	// check enrollment
	_, err = s.courseRepo.GetCourseMember(ctx, db.GetCourseMemberParams{
		CourseID: session.CourseID,
		UserID:   userUUID,
	})
	if err != nil {
		return model.OTPResponse{}, errors.New("You are not enrolled in this course")
	}

	// generate code
	code, err := generateOTP()
	if err != nil {
		return model.OTPResponse{}, err
	}

	// expiry
	expiry := time.Now().Add(5 * time.Minute)
	var pgExpiry pgtype.Timestamptz
	pgExpiry.Scan(expiry)

	_, err = s.otpRepo.CreateOTP(ctx, db.CreateOTPParams{
		SessionID: sessionUUID,
		UserID:    userUUID,
		Code:      code,
		ExpiresAt: pgExpiry,
	})
	if err != nil {
		return model.OTPResponse{}, errors.New("Failed to request OTP")
	}

	return model.OTPResponse{
		OTPCode:   code,
		ExpiresAt: expiry.Format("2006-01-02 15:04:05"),
	}, nil
}
