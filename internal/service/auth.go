package service

import (
	"context"
	"errors"
	"strings"
	"time"

	firebaseAuth "firebase.google.com/go/v4/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/niyiayooluwa/geotas/internal/db"
	"github.com/niyiayooluwa/geotas/internal/model"
	"github.com/niyiayooluwa/geotas/internal/repository"
	"os"
)

type AuthService struct {
	userRepo     *repository.UserRepository
	firebaseAuth *firebaseAuth.Client
}

func NewAuthService(userRepo *repository.UserRepository, firebaseAuth *firebaseAuth.Client) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		firebaseAuth: firebaseAuth,
	}
}

func (s *AuthService) GoogleLogin(ctx context.Context, req model.GoogleLoginRequest) (model.LoginResponse, error) {
	// 1. Verify Firebase ID token
	token, err := s.firebaseAuth.VerifyIDToken(ctx, req.IDToken)
	if err != nil {
		return model.LoginResponse{}, errors.New("invalid Google token")
	}

	// 2. Extract claims
	email, _     := token.Claims["email"].(string)
	avatarURL, _ := token.Claims["picture"].(string)
	firstName, _ := token.Claims["given_name"].(string)
	lastName, _  := token.Claims["family_name"].(string)
	googleID     := token.UID

	if firstName == "" {
		name, _ := token.Claims["name"].(string)
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

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := jwtToken.SignedString([]byte(os.Getenv("JWT_SECRET")))
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

	var avatarURL *string
	if user.AvatarUrl.Valid {
		avatarURL = &user.AvatarUrl.String
	}

	return model.UserResponse{
		ID:        user.ID.String(),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		AvatarURL: avatarURL,
		CreatedAt: user.CreatedAt.Time.Format(time.RFC3339),
	}, nil
}