package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/niyiayooluwa/geotas/internal/model"
	"github.com/niyiayooluwa/geotas/internal/repository"
)

// QRRotationManager generates tokens on demand (stateless)
type QRRotationManager struct {
	qrRepo *repository.QRTokenRepository
}

func NewQRRotationManager(qrRepo *repository.QRTokenRepository) *QRRotationManager {
	return &QRRotationManager{
		qrRepo: qrRepo,
	}
}

// generates a signed HMAC token for a session at a given timestamp
func generateQRToken(sessionID string, timestamp time.Time) string {
	var secret string = os.Getenv("JWT_SECRET")
	var mac = hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("%s:%d", sessionID, timestamp.Unix())))
	return hex.EncodeToString(mac.Sum(nil))
}


// GetCurrentToken lazily generates or returns the active token for a session
func (m *QRRotationManager) GetCurrentToken(sessionID string, courseID string, rotationSecs int32) (model.QRTokenResponse, error) {
    var sessionUUID pgtype.UUID
    if err := sessionUUID.Scan(sessionID); err != nil {
        return model.QRTokenResponse{}, fmt.Errorf("Invalid Session ID")
    }

    var ctx context.Context = context.Background()
    token, err := m.qrRepo.GetLatestQRTokenBySession(ctx, sessionUUID)
    
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
        newTokenStr := generateQRToken(sessionID, now)
        expiry := now.Add(time.Duration(rotationSecs)*time.Second + 10*time.Second)
        token, err = m.qrRepo.CreateQRToken(ctx, sessionUUID, newTokenStr, expiry)
        if err != nil {
            return model.QRTokenResponse{}, fmt.Errorf("Failed to generate QR token")
        }
    }

	payload:= model.QRPayload{
		Token: token.Token,
		SessionID: sessionID,
		CourseID: courseID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return model.QRTokenResponse{}, fmt.Errorf("Failed to encode QR payload")
	}

	// Hide the 10s grace period from the frontend so their UI timer is perfectly accurate
    return model.QRTokenResponse{
		QRContent: string(payloadBytes),
		ExpiresAt: token.ExpiresAt.Time.Add(-10 * time.Second).Format(time.RFC3339),
	}, nil
}