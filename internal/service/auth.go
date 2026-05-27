package service

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/niyiayooluwa/geotas/internal/db"
	"github.com/niyiayooluwa/geotas/internal/model"
	"github.com/niyiayooluwa/geotas/internal/repository"
	"google.golang.org/api/idtoken"
)

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) GoogleLogin(ctx context.Context, req model.GoogleLoginRequest) (model.LoginResponse, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	log.Printf("Validating Google ID token for client ID: %s", clientID)

	// 1. Verify token with Google
	payload, err := idtoken.Validate(ctx, req.IDToken, clientID)
	if err != nil {
		return model.LoginResponse{}, errors.New("invalid Google token")
	}

	// 2. Safely extract claims
	googleID := payload.Subject
	email, _ := payload.Claims["email"].(string)
	avatarURL, _ := payload.Claims["picture"].(string)

	// Google's payload structure can vary slightly depending on the user's profile
	firstName, _ := payload.Claims["given_name"].(string)
	lastName, _ := payload.Claims["family_name"].(string)
	
	if firstName == "" {
		name, _ := payload.Claims["name"].(string)
		parts := strings.SplitN(name, " ", 2)
		firstName = parts[0]
		if len(parts) > 1 {
			lastName = parts[1]
		}
	}

	// 3. Upsert user
	user, err := s.userRepo.UpsertGoogleUser(ctx, db.UpsertGoogleUserParams{
		Email:     email,
		GoogleID:  pgtype.Text{String: googleID, Valid: true},
		FirstName: firstName,
		LastName:  lastName,
		AvatarUrl: pgtype.Text{String: avatarURL, Valid: avatarURL != ""},
	})
	if err != nil {
		return model.LoginResponse{}, errors.New("could not save user data")
	}

	// 4. Generate GEOTAS JWT
	claims := model.Claims{
		UserID: user.ID.String(),
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return model.LoginResponse{}, errors.New("could not generate token")
	}

	return model.LoginResponse{
		Token:     signedToken,
		ID:        user.ID.String(),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		AvatarURL: user.AvatarUrl.String,
	}, nil
}

// Keep GetUserByID and GetUserProfile exactly as they were
func (s *AuthService) GetUserByID(ctx context.Context, userID string) (db.User, error) {
	var uuid pgtype.UUID
	if err := uuid.Scan(userID); err != nil {
		return db.User{}, errors.New("invalid user id")
	}
	return s.userRepo.GetUserByID(ctx, uuid)
}

func (s *AuthService) GetUserProfile(ctx context.Context, userID string) (model.UserResponse, error) {
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return model.UserResponse{}, err
	}

	return model.UserResponse{
		ID:         user.ID.String(),
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Email:      user.Email,
		Department: user.Department.String,
		CreatedAt:  user.CreatedAt.Time.Format(time.RFC3339),
	}, nil
}