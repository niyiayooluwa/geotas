package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/niyiayooluwa/geotas/internal/db"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	FirstName           string `json:"first_name"`
	LastName            string `json:"last_name"`
	Email               string `json:"email"`
	Password            string `json:"password"`
	MatriculationNumber string `json:"matriculation_number"`
	Department          string `json:"department"`
}

// add this struct — controls what we send back to the client
type RegisterResponse struct {
	ID                  string `json:"id"`
	FirstName           string `json:"first_name"`
	LastName            string `json:"last_name"`
	Email               string `json:"email"`
	MatriculationNumber string `json:"matriculation_number"`
	Department          string `json:"department"`
	CreatedAt           string `json:"created_at"`
}

func RegisterHandler(queries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//decode the JSON body into a RegisterRequest struct
		var req RegisterRequest
		
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// hash the password before storing it in the database
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// insert user into db
		user, err := queries.CreateUser(r.Context(), db.CreateUserParams{
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
			// check for unique constraint violation
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				http.Error(w, "email or matric number already exists", http.StatusConflict)
				return
			}
			http.Error(w, "could not create user", http.StatusInternalServerError)
			return
		}

		var response RegisterResponse = RegisterResponse{
			ID:                  user.ID.String(),
			FirstName:           user.FirstName,
			LastName:            user.LastName,
			Email:               user.Email,
			MatriculationNumber: user.MatriculationNumber,
			Department:          user.Department.String,
			CreatedAt:           user.CreatedAt.Time.Format("2006-01-02 15:04:05"),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}
