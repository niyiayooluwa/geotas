package model

type GoogleLoginRequest struct {
	IDToken      string  `json:"id_token"`
	Department   *string `json:"department,omitempty"`
	MatricNumber *string `json:"matric_number,omitempty"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Role      string `json:"role"`
}

type UserResponse struct {
    ID           string  `json:"id"`
    FirstName    string  `json:"first_name"`
    LastName     string  `json:"last_name"`
    Email        string  `json:"email"`
    AvatarURL    *string `json:"avatar_url,omitempty"`
    Department   string  `json:"department"`
    MatricNumber *string `json:"matric_number,omitempty"`
    CreatedAt    string  `json:"created_at"`
    Role         string  `json:"role"`
}

type LecturerRegisterRequest struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	Department string `json:"department"`
}

type LecturerLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}