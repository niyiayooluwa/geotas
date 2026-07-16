package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/niyiayooluwa/geotas/internal/db"
)

type OTPRepository struct {
	queries *db.Queries
}

func NewOTPRepository(queries *db.Queries) *OTPRepository {
	return &OTPRepository{queries: queries}
}

func (r *OTPRepository) CreateOTP(ctx context.Context, sessionID pgtype.UUID, code string, expiresAt time.Time) (db.OtpCode, error) {
	var expires pgtype.Timestamptz
	expires.Scan(expiresAt)

	return r.queries.CreateOTP(ctx, db.CreateOTPParams{
		SessionID: sessionID,
		Code:      code,
		ExpiresAt: expires,
	})
}

func (r *OTPRepository) GetValidOTP(ctx context.Context, params db.GetValidOTPParams) (db.OtpCode, error) {
	return r.queries.GetValidOTP(ctx, params)
}

func (r *OTPRepository) GetLatestOTPBySession(ctx context.Context, sessionID pgtype.UUID) (db.OtpCode, error) {
	return r.queries.GetLatestOTPBySession(ctx, sessionID)
}

func (r *OTPRepository) InvalidatePreviousOTPs(ctx context.Context, sessionID pgtype.UUID) error {
	return r.queries.InvalidatePreviousOTPs(ctx, sessionID)
}

func (r *OTPRepository) MarkOTPUsed(ctx context.Context, id pgtype.UUID) error {
	return r.queries.MarkOTPUsed(ctx, id)
}
