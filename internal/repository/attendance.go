package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/niyiayooluwa/geotas/internal/db"
)

type AttendanceRepository struct {
	queries *db.Queries
}

func NewAttendanceRepository(queries *db.Queries) *AttendanceRepository {
	return &AttendanceRepository{queries: queries}
}

func (r *AttendanceRepository) CreateAttendanceRecord(ctx context.Context, params db.CreateAttendanceRecordParams) (db.AttendanceRecord, error) {
	return r.queries.CreateAttendanceRecord(ctx, params)
}

func (r *AttendanceRepository) GetAttendanceBySession(ctx context.Context, sessionID pgtype.UUID) ([]db.GetAttendanceBySessionRow, error) {
	return r.queries.GetAttendanceBySession(ctx, sessionID)
}

func (r *AttendanceRepository) GetAttendanceByUserAndSession(ctx context.Context, userID pgtype.UUID, sessionID pgtype.UUID) (db.AttendanceRecord, error) {
	return r.queries.GetAttendanceByUserAndSession(ctx, db.GetAttendanceByUserAndSessionParams{
		UserID:    userID,
		SessionID: sessionID,
	})
}

func (r *AttendanceRepository) CheckDuplicateDevice(ctx context.Context, sessionID pgtype.UUID, deviceID string) (db.AttendanceRecord, error) {
	return r.queries.CheckDuplicateDevice(ctx, db.CheckDuplicateDeviceParams{
		SessionID: sessionID,
		DeviceID:  pgtype.Text{String: deviceID, Valid: deviceID != ""},
	})
}

func (r *AttendanceRepository) GetAttendanceByCourse(ctx context.Context, courseID pgtype.UUID) ([]db.GetAttendanceByCourseRow, error) {
	return r.queries.GetAttendanceByCourse(ctx, courseID)
}
