package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/niyiayooluwa/geotas/internal/db"
)

type SessionRepository struct {
	queries *db.Queries
}

func NewSessionRepository(queries *db.Queries) *SessionRepository {
	return &SessionRepository{queries: queries}
}

func (r *SessionRepository) CreateSession(ctx context.Context, params db.CreateSessionParams) (db.Session, error) {
	return r.queries.CreateSession(ctx, params)
}

func (r *SessionRepository) GetSessionByID(ctx context.Context, id pgtype.UUID) (db.Session, error) {
	return r.queries.GetSessionByID(ctx, id)
}

func (r *SessionRepository) GetSessionsByCourse(ctx context.Context, courseID pgtype.UUID) ([]db.Session, error) {
	return r.queries.GetSessionsByCourse(ctx, courseID)
}

func (r *SessionRepository) GetActiveSessionByCourse(ctx context.Context, courseID pgtype.UUID) (db.Session, error) {
	return r.queries.GetActiveSessionByCourse(ctx, courseID)
}

func (r *SessionRepository) CloseSession(ctx context.Context, id pgtype.UUID) (db.Session, error) {
	return r.queries.CloseSession(ctx, id)
}