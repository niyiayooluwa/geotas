package model

import "github.com/golang-jwt/jwt/v5"

// Claims defines what lives inside the JWT
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

