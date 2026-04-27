package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/niyiayooluwa/geotas/internal/repository"
)

// QRRotationManager manages all active QR rotation goroutines
// one goroutine per active session
type QRRotationManager struct {
	qrRepo   *repository.QRTokenRepository
	stopChans map[string]chan struct{}  // sessionID → stop signal channel
	mu        sync.Mutex               // protects the map from concurrent access
}

func NewQRRotationManager(qrRepo *repository.QRTokenRepository) *QRRotationManager {
	return &QRRotationManager{
		qrRepo:    qrRepo,
		stopChans: make(map[string]chan struct{}),
	}
}

// generates a signed HMAC token for a session at a given timestamp
func generateQRToken(sessionID string, timestamp time.Time) string {
	var secret string = os.Getenv("JWT_SECRET")
	var mac = hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("%s:%d", sessionID, timestamp.Unix())))
	return hex.EncodeToString(mac.Sum(nil))
}

// StartRotation launches a goroutine that rotates QR tokens for a session
func (m *QRRotationManager) StartRotation(sessionID string, rotationSecs int32) {
	// create a stop channel for this session
	var stopChan chan struct{} = make(chan struct{})

	// store it so we can stop it later
	m.mu.Lock()
	m.stopChans[sessionID] = stopChan
	m.mu.Unlock()

	// parse sessionID into pgtype.UUID for DB calls
	var sessionUUID pgtype.UUID
	sessionUUID.Scan(sessionID)

	// launch the goroutine
	go func() {
		// generate first token immediately
		m.rotateToken(sessionUUID, sessionID)

		// set up ticker for subsequent rotations
		var ticker *time.Ticker = time.NewTicker(time.Duration(rotationSecs) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// ticker fired — rotate the token
				m.rotateToken(sessionUUID, sessionID)

			case <-stopChan:
				// stop signal received — session closed, exit goroutine
				return
			}
		}
	}()
}

// StopRotation sends a stop signal to the goroutine for a session
func (m *QRRotationManager) StopRotation(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if stopChan, exists := m.stopChans[sessionID]; exists {
		close(stopChan)
		delete(m.stopChans, sessionID)
	}
}

// rotateToken invalidates old tokens and generates a new one
func (m *QRRotationManager) rotateToken(sessionUUID pgtype.UUID, sessionID string) {
	var ctx context.Context = context.Background()

	// invalidate all previous tokens for this session
	m.qrRepo.InvalidatePreviousTokens(ctx, sessionUUID)

	// generate new signed token
	var now time.Time = time.Now()
	var token string = generateQRToken(sessionID, now)

	// store in DB with expiry = now + rotation window + small buffer
	m.qrRepo.CreateQRToken(ctx, sessionUUID, token, now.Add(40*time.Second))
}

// GetCurrentToken returns the latest valid token for a session
func (m *QRRotationManager) GetCurrentToken(sessionID string) (string, error) {
	var sessionUUID pgtype.UUID
	if err := sessionUUID.Scan(sessionID); err != nil {
		return "", fmt.Errorf("invalid session id")
	}

	// we generate a fresh token based on current time
	// this is what the lecturer's dashboard displays
	var token string = generateQRToken(sessionID, time.Now().Truncate(40*time.Second))
	return token, nil
}