package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/niyiayooluwa/geotas/internal/db"
)

type UserRepository struct {
	queries *db.Queries
}

func NewUserRepository(queries *db.Queries) *UserRepository {
	return &UserRepository{queries: queries}
}

func (r *UserRepository) UpsertGoogleUser(ctx context.Context, params db.UpsertGoogleUserParams) (db.User, error) {
	return r.queries.UpsertGoogleUser(ctx, params)
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	return r.queries.GetUserByEmail(ctx, email)
}

func (r *UserRepository) GetUserByID(ctx context.Context, id pgtype.UUID) (db.User, error) {
	return r.queries.GetUserByID(ctx, id)
}