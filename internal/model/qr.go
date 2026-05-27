package model

type QRTokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"` // RFC3339, Flutter parses this
}
