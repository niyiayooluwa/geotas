package model

// MarkAttendanceQRRequest is what a student sends when scanning a QR code
type MarkAttendanceQRRequest struct {
	SessionID            string  `json:"session_id"`
	QRToken              string  `json:"qr_token"`
	Latitude             float64 `json:"latitude"`
	Longitude            float64 `json:"longitude"`
	DeviceID             string  `json:"device_id"`
	DeviceModel          string  `json:"device_model"`
	OsVersion            string  `json:"os_version"`
	MockLocationDetected bool    `json:"mock_location_detected"`
}

// MarkAttendanceOTPRequest is what a student sends when using an OTP fallback
type MarkAttendanceOTPRequest struct {
	SessionID            string  `json:"session_id"`
	OTPCode              string  `json:"otp_code"`
	Latitude             float64 `json:"latitude"`
	Longitude            float64 `json:"longitude"`
	DeviceID             string  `json:"device_id"`
	DeviceModel          string  `json:"device_model"`
	OsVersion            string  `json:"os_version"`
	MockLocationDetected bool    `json:"mock_location_detected"`
}

// RequestOTPRequest is what a student sends to get a new OTP
type RequestOTPRequest struct {
	SessionID string `json:"session_id"`
}

// OTPResponse is sent back when an OTP is requested (MVP)
type OTPResponse struct {
	OTPCode   string `json:"otp_code"`
	ExpiresAt string `json:"expires_at"`
}

// AttendanceResponse is the simplified "receipt" sent back to the student
type AttendanceResponse struct {
	ID         string `json:"id"`
	SessionID  string `json:"session_id"`
	UserID     string `json:"user_id"`
	MarkedAt   string `json:"marked_at"`
	Method     string `json:"method"`
	WeekNumber int32  `json:"week_number"`
}
