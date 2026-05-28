package model

// incoming request to create a course
type CreateCourseRequest struct {
	Title      string `json:"title"`
	Code       string `json:"code"`
	Department string `json:"department"`
}

// what we send back after creating a course
type CourseResponse struct {
    ID           string `json:"id"`
    OwnerID      string `json:"owner_id"`
    Title        string `json:"title"`
    Code         string `json:"code"`
    InviteCode   string `json:"invite_code"`
    Department   string `json:"department"`
    StudentCount int64  `json:"student_count"`
    CreatedAt    string `json:"created_at"`
}

// incoming request to join a course
type JoinCourseRequest struct {
	InviteCode          string `json:"invite_code"`
	MatriculationNumber string `json:"matriculation_number"`
}

// what we send back after joining a course
type CourseMemberResponse struct {
	ID                  string `json:"id"`
	CourseID            string `json:"course_id"`
	UserID              string `json:"user_id"`
	Role                string `json:"role"`
	MatriculationNumber string `json:"matriculation_number"`
	JoinedAt            string `json:"joined_at"`
}

type MemberCourseResponse struct {
	ID                  string `json:"id"`
	OwnerID             string `json:"owner_id"`
	Title               string `json:"title"`
	Code                string `json:"code"`
	Department          string `json:"department"`
	InviteCode          string `json:"invite_code"`
	CreatedAt           string `json:"created_at"`
	Role                string `json:"role"`
	MatriculationNumber string `json:"matriculation_number"`
	StudentCount        int64  `json:"student_count"` // Add this line
}