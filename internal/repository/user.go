package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/niyiayooluwa/geotas/internal/db"
)

type UserRepository struct {
	queries *db.Queries
}

// constructor — creates a new UserRepository
func NewUserRepository(queries *db.Queries) *UserRepository {
	return &UserRepository{queries: queries}
}

// inserts a new user into the DB
func (r *UserRepository) CreateUser(ctx context.Context, params db.CreateUserParams) (db.User, error) {
	return r.queries.CreateUser(ctx, params)
}

// fetches a user by email
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	return r.queries.GetUserByEmail(ctx, email)
}

// fetches a user by ID
func (r *UserRepository) GetUserByID(ctx context.Context, id pgtype.UUID) (db.User, error) {
	return r.queries.GetUserByID(ctx, id)
}
