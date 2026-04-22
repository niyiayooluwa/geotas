package service

import (
	"context"
	"errors"
	"os"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/niyiayooluwa/geotas/internal/db"
	"github.com/niyiayooluwa/geotas/internal/model"
	"github.com/niyiayooluwa/geotas/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo *repository.UserRepository
}

// constructor
func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

// validates the register request fields
func (s *AuthService) ValidateRegisterRequest(req model.RegisterRequest) error {
	if req.FirstName == "" || req.LastName == "" {
		return errors.New("first name and last name are required")
	}

	var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		return errors.New("invalid email address")
	}

	if len(req.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	if !regexp.MustCompile(`[A-Z]`).MatchString(req.Password) {
		return errors.New("password must contain at least one uppercase letter")
	}

	if !regexp.MustCompile(`[a-z]`).MatchString(req.Password) {
		return errors.New("password must contain at least one lowercase letter")
	}

	if !regexp.MustCompile(`[0-9]`).MatchString(req.Password) {
		return errors.New("password must contain at least one number")
	}

	if !regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(req.Password) {
		return errors.New("password must contain at least one special character")
	}

	return nil
}

// handles full registration logic
func (s *AuthService) Register(ctx context.Context, req model.RegisterRequest) (model.RegisterResponse, error) {
	// validate input
	if err := s.ValidateRegisterRequest(req); err != nil {
		return model.RegisterResponse{}, err
	}

	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return model.RegisterResponse{}, errors.New("could not hash password")
	}

	// insert into DB
	user, err := s.userRepo.CreateUser(ctx, db.CreateUserParams{
		FirstName:           req.FirstName,
		LastName:            req.LastName,
		Email:               req.Email,
		PasswordHash:        string(hashedPassword),
		MatriculationNumber: req.MatriculationNumber,
		Department: pgtype.Text{
			String: req.Department,
			Valid:  req.Department != "",
		},
	})
	if err != nil {
		// detect duplicate email or matric number
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.RegisterResponse{}, errors.New("email or matric number already exists")
		}
		return model.RegisterResponse{}, errors.New("could not create user")
	}

	// build and return response
	return model.RegisterResponse{
		ID:                  user.ID.String(),
		FirstName:           user.FirstName,
		LastName:            user.LastName,
		Email:               user.Email,
		MatriculationNumber: user.MatriculationNumber,
		Department:          user.Department.String,
		CreatedAt:           user.CreatedAt.Time.Format("2006-01-02 15:04:05"),
	}, nil
}

// handles full login logic
func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (model.LoginResponse, error) {
	// fetch user by email
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return model.LoginResponse{}, errors.New("invalid credentials")
	}

	// compare password with stored hash
	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(req.Password),
	); err != nil {
		return model.LoginResponse{}, errors.New("invalid credentials")
	}

	// build JWT claims
	var claims model.Claims = model.Claims{
		UserID: user.ID.String(),
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// sign the token
	var token *jwt.Token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	var secret string = os.Getenv("JWT_SECRET")
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return model.LoginResponse{}, errors.New("could not generate token")
	}

	return model.LoginResponse{
		Token:     signedToken,
		ID:        user.ID.String(),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}, nil
}

// GetUserByID fetches a user by their UUID string
func (s *AuthService) GetUserByID(ctx context.Context, userID string) (db.User, error) {
	var uuid pgtype.UUID
	if err := uuid.Scan(userID); err != nil {
		return db.User{}, errors.New("invalid user id")
	}
	return s.userRepo.GetUserByID(ctx, uuid)
}