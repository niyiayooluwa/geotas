package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/niyiayooluwa/geotas/internal/db"
)

type QRTokenRepository struct {
	queries *db.Queries
}

func NewQRTokenRepository(queries *db.Queries) *QRTokenRepository {
	return &QRTokenRepository{queries: queries}
}

func (r *QRTokenRepository) CreateQRToken(ctx context.Context, sessionID pgtype.UUID, token string, expiresAt time.Time) (db.QrToken, error) {
	var expires pgtype.Timestamptz
	expires.Scan(expiresAt)

	return r.queries.CreateQRToken(ctx, db.CreateQRTokenParams{
		SessionID: sessionID,
		Token:     token,
		ExpiresAt: expires,
	})
}

func (r *QRTokenRepository) GetValidQRToken(ctx context.Context, token string) (db.QrToken, error) {
	return r.queries.GetValidQRToken(ctx, token)
}

func (r *QRTokenRepository) MarkQRTokenUsed(ctx context.Context, id pgtype.UUID) error {
	return r.queries.MarkQRTokenUsed(ctx, id)
}

func (r *QRTokenRepository) InvalidatePreviousTokens(ctx context.Context, sessionID pgtype.UUID) error {
	return r.queries.InvalidatePreviousTokens(ctx, sessionID)
}