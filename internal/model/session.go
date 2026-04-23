package model

// incoming request to create a session
type CreateSessionRequest struct {
	CourseID        string  `json:"course_id"`
	Title           string  `json:"title"`
	WeekNumber      int32   `json:"week_number"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	RadiusMeters    float64 `json:"radius_meters"`
	QrRotationSecs  int32   `json:"qr_rotation_secs"`
}

// what we send back after creating a session
type SessionResponse struct {
	ID             string  `json:"id"`
	CourseID       string  `json:"course_id"`
	CreatedBy      string  `json:"created_by"`
	Title          string  `json:"title"`
	WeekNumber     int32   `json:"week_number"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	RadiusMeters   float64 `json:"radius_meters"`
	QrRotationSecs int32   `json:"qr_rotation_secs"`
	Status         string  `json:"status"`
	StartedAt      string  `json:"started_at"`
	ClosedAt       string  `json:"closed_at"`
}