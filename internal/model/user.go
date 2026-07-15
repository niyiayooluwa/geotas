package model

type GoogleLoginRequest struct {
	IDToken string `json:"id_token"`
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
    ID        string  `json:"id"`
    FirstName string  `json:"first_name"`
    LastName  string  `json:"last_name"`
    Email     string  `json:"email"`
    AvatarURL *string `json:"avatar_url,omitempty"`
    CreatedAt string  `json:"created_at"`
    Role      string  `json:"role"`
}

type LecturerRegisterRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type LecturerLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}