package model

type NotificationResponse struct {
	ID        string `json:"id"`
	CourseID  string `json:"course_id"`
	Type      string `json:"type"`
	Payload   any    `json:"payload"`
	SeenAt    string `json:"seen_at,omitempty"`
	CreatedAt string `json:"created_at"`
}
