package model

type ScheduleRequest struct {
	DayOfWeek int32  `json:"day_of_week"`
	StartTime string `json:"start_time"` // "HH:MM:SS" or "HH:MM"
	EndTime   string `json:"end_time"`
	Venue     string `json:"venue"`
}

type ScheduleResponse struct {
	ID        string `json:"id"`
	CourseID  string `json:"course_id"`
	DayOfWeek int32  `json:"day_of_week"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Venue     string `json:"venue"`
}
