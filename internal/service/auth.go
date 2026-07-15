package service

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"
	"unicode"

	firebaseAuth "firebase.google.com/go/v4/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/niyiayooluwa/geotas/internal/db"
	"github.com/niyiayooluwa/geotas/internal/model"
	"github.com/niyiayooluwa/geotas/internal/repository"
	"golang.org/x/crypto/bcrypt"
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

	if user.Role != db.UserRoleStudent {
		return model.LoginResponse{}, errors.New("student app login only")
	}

	// 4. Generate GEOTAS JWT
	claims := model.Claims{
		UserID: user.ID.String(),
		Email:  user.Email,
		Role:   string(user.Role),
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
		Role:      string(user.Role),
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
		Role:      string(user.Role),
	}, nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	var hasUpper, hasNumber, hasSpecial bool
	for _, char := range password {
		if unicode.IsUpper(char) {
			hasUpper = true
		} else if unicode.IsNumber(char) {
			hasNumber = true
		} else if unicode.IsPunct(char) || unicode.IsSymbol(char) {
			hasSpecial = true
		}
	}
	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasNumber {
		return errors.New("password must contain at least one number")
	}
	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}
	return nil
}

func (s *AuthService) RegisterLecturer(ctx context.Context, req model.LecturerRegisterRequest) (model.UserResponse, error) {
	if !strings.HasSuffix(req.Email, ".edu.ng") {
		return model.UserResponse{}, errors.New("Email must be a valid .edu.ng address")
	}

	if err := validatePassword(req.Password); err != nil {
		return model.UserResponse{}, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return model.UserResponse{}, errors.New("failed to hash password")
	}

	user, err := s.userRepo.CreateUser(ctx, db.CreateUserParams{
		Email:        req.Email,
		PasswordHash: pgtype.Text{String: string(hashedPassword), Valid: true},
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         db.UserRoleLecturer,
	})
	if err != nil {
		return model.UserResponse{}, errors.New("could not create user")
	}

	return model.UserResponse{
		ID:        user.ID.String(),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Time.Format(time.RFC3339),
		Role:      string(user.Role),
	}, nil
}

func (s *AuthService) LoginLecturer(ctx context.Context, req model.LecturerLoginRequest) (model.LoginResponse, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return model.LoginResponse{}, errors.New("invalid email or password")
	}

	if user.Role != db.UserRoleLecturer {
		return model.LoginResponse{}, errors.New("lecturer web dashboard only")
	}

	if !user.PasswordHash.Valid {
		return model.LoginResponse{}, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(req.Password)); err != nil {
		return model.LoginResponse{}, errors.New("invalid email or password")
	}

	claims := model.Claims{
		UserID: user.ID.String(),
		Email:  user.Email,
		Role:   string(user.Role),
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
		Role:      string(user.Role),
	}, nil
}