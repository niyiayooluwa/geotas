package model

// incoming request to create a course
type CreateCourseRequest struct {
	Title      string `json:"title"`
	Code       string `json:"code"`
	Department string `json:"department"`
}

// what we send back after creating a course
type CourseResponse struct {
	ID         string `json:"id"`
	OwnerID    string `json:"owner_id"`
	Title      string `json:"title"`
	Code       string `json:"code"`
	InviteCode string `json:"invite_code"`
	Department string `json:"department"`
	CreatedAt  string `json:"created_at"`
}

// incoming request to join a course
type JoinCourseRequest struct {
	InviteCode string `json:"invite_code"`
}

// what we send back after joining a course
type CourseMemberResponse struct {
	ID       string `json:"id"`
	CourseID string `json:"course_id"`
	UserID   string `json:"user_id"`
	Role     string `json:"role"`
	JoinedAt string `json:"joined_at"`
}