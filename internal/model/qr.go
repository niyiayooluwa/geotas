package model

// QRPayload is what gets JSON-encoded into the QR code itself
type QRPayload struct {
	Token     string `json:"token"`
	SessionID string `json:"session_id"`
	CourseID  string `json:"course_id"`
}

type QRTokenResponse struct {
	QRContent string `json:"qr_content"` // JSON string — encode this into the QR
	ExpiresAt string `json:"expires_at"` // RFC3339
}