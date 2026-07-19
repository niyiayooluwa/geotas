package service

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/niyiayooluwa/geotas/internal/model"
	"github.com/niyiayooluwa/geotas/internal/repository"
)

type OTPRotationManager struct {
	otpRepo   *repository.OTPRepository
	stopChans map[string]chan struct{}
	mu        sync.Mutex
}

func NewOTPRotationManager(otpRepo *repository.OTPRepository) *OTPRotationManager {
	return &OTPRotationManager{
		otpRepo:   otpRepo,
		stopChans: make(map[string]chan struct{}),
	}
}

func generateOTP() string {
	var b [4]byte
	rand.Read(b[:])
	code := binary.BigEndian.Uint32(b[:]) % 1000000
	return fmt.Sprintf("%06d", code)
}

func (m *OTPRotationManager) StartRotation(sessionID string, rotationSecs int32) {
	stopChan := make(chan struct{})

	m.mu.Lock()
	m.stopChans[sessionID] = stopChan
	m.mu.Unlock()

	var sessionUUID pgtype.UUID
	sessionUUID.Scan(sessionID)

	go func() {
		m.rotateToken(sessionUUID, rotationSecs)

		ticker := time.NewTicker(time.Duration(rotationSecs) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.rotateToken(sessionUUID, rotationSecs)
			case <-stopChan:
				return
			}
		}
	}()
}

func (m *OTPRotationManager) StopRotation(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if stopChan, exists := m.stopChans[sessionID]; exists {
		close(stopChan)
		delete(m.stopChans, sessionID)
	}
}

func (m *OTPRotationManager) rotateToken(sessionUUID pgtype.UUID, rotationSecs int32) {
	ctx := context.Background()

	// Invalidate previous OTPs for this session
	m.otpRepo.InvalidatePreviousOTPs(ctx, sessionUUID)

	// Generate and save new OTP with some grace period buffer (15s)
	now := time.Now()
	code := generateOTP()
	expiry := now.Add(time.Duration(rotationSecs)*time.Second + 15*time.Second)
	m.otpRepo.CreateOTP(ctx, sessionUUID, code, expiry)
}

func (m *OTPRotationManager) GetCurrentToken(sessionID string) (model.OTPResponse, error) {
	var sessionUUID pgtype.UUID
	if err := sessionUUID.Scan(sessionID); err != nil {
		return model.OTPResponse{}, fmt.Errorf("Invalid Session ID")
	}

	token, err := m.otpRepo.GetLatestOTPBySession(context.Background(), sessionUUID)
	if err != nil {
		return model.OTPResponse{}, fmt.Errorf("No active OTP found for session")
	}

	return model.OTPResponse{
		OTPCode:   token.Code,
		ExpiresAt: token.ExpiresAt.Time.Format(time.RFC3339),
	}, nil
}
