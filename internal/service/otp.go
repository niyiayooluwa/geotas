package service

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/niyiayooluwa/geotas/internal/model"
	"github.com/niyiayooluwa/geotas/internal/repository"
)

// OTPRotationManager generates tokens on demand (stateless)
type OTPRotationManager struct {
	otpRepo *repository.OTPRepository
}

func NewOTPRotationManager(otpRepo *repository.OTPRepository) *OTPRotationManager {
	return &OTPRotationManager{
		otpRepo: otpRepo,
	}
}

func generateOTP() string {
	var b [4]byte
	rand.Read(b[:])
	code := binary.BigEndian.Uint32(b[:]) % 1000000
	return fmt.Sprintf("%06d", code)
}


func (m *OTPRotationManager) GetCurrentToken(sessionID string, rotationSecs int32) (model.OTPResponse, error) {
	var sessionUUID pgtype.UUID
	if err := sessionUUID.Scan(sessionID); err != nil {
		return model.OTPResponse{}, fmt.Errorf("Invalid Session ID")
	}

	ctx := context.Background()
	token, err := m.otpRepo.GetLatestOTPBySession(ctx, sessionUUID)
	
	var needsNewToken bool = false
	if err != nil {
	    needsNewToken = true
	} else {
	    strictExpiry := token.ExpiresAt.Time.Add(-10 * time.Second)
	    if time.Now().After(strictExpiry) {
	        needsNewToken = true
	    }
	}

	if needsNewToken {
	    now := time.Now()
	    code := generateOTP()
	    expiry := now.Add(time.Duration(rotationSecs)*time.Second + 10*time.Second)
	    token, err = m.otpRepo.CreateOTP(ctx, sessionUUID, code, expiry)
	    if err != nil {
	        return model.OTPResponse{}, fmt.Errorf("Failed to generate OTP")
	    }
	}

	// Hide the 10s grace period from the frontend so their UI timer is perfectly accurate
	return model.OTPResponse{
		OTPCode:   token.Code,
		ExpiresAt: token.ExpiresAt.Time.Add(-10 * time.Second).Format(time.RFC3339),
	}, nil
}
