package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/niyiayooluwa/geotas/internal/db"
)

type OTPRepository struct {
	queries *db.Queries
}

func NewOTPRepository(queries *db.Queries) *OTPRepository {
	return &OTPRepository{queries: queries}
}

func (r *OTPRepository) CreateOTP(ctx context.Context, params db.CreateOTPParams) (db.OtpCode, error) {
	return r.queries.CreateOTP(ctx, params)
}

func (r *OTPRepository) GetValidOTP(ctx context.Context, params db.GetValidOTPParams) (db.OtpCode, error) {
	return r.queries.GetValidOTP(ctx, params)
}

func (r *OTPRepository) MarkOTPUsed(ctx context.Context, id pgtype.UUID) error {
	return r.queries.MarkOTPUsed(ctx, id)
}
