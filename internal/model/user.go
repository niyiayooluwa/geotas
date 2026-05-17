package model

// incoming request to register a new user
type RegisterRequest struct {
	FirstName           string `json:"first_name"`
	LastName            string `json:"last_name"`
	Email               string `json:"email"`
	Password            string `json:"password"`
	Department          string `json:"department"`
}

// what we send back after registration
type RegisterResponse struct {
	ID                  string `json:"id"`
	FirstName           string `json:"first_name"`
	LastName            string `json:"last_name"`
	Email               string `json:"email"`
	Department          string `json:"department"`
	CreatedAt           string `json:"created_at"`
}

// what we send back for profile info
type UserResponse struct {
	ID         string `json:"id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Email      string `json:"email"`
	Department string `json:"department"`
	CreatedAt  string `json:"created_at"`
}

// incoming request to login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// what we send back after login
type LoginResponse struct {
	Token     string `json:"token"`
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}
