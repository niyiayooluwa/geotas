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
}

// Keep UserResponse as is
type UserResponse struct {
	ID         string `json:"id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Email      string `json:"email"`
	Department string `json:"department"`
	CreatedAt  string `json:"created_at"`
}