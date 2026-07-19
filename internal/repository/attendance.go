package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/niyiayooluwa/geotas/internal/db"
)

var ErrDuplicateAttendance = errors.New("Attendance already marked")

type AttendanceRepository struct {
	queries *db.Queries
}

func NewAttendanceRepository(queries *db.Queries) *AttendanceRepository {
	return &AttendanceRepository{queries: queries}
}

func (r *AttendanceRepository) CreateAttendanceRecord(ctx context.Context, params db.CreateAttendanceRecordParams) (db.AttendanceRecord, error) {
	record, err := r.queries.CreateAttendanceRecord(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return db.AttendanceRecord{}, ErrDuplicateAttendance
		}
		return db.AttendanceRecord{}, err
	}
	return record, nil
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

func (r *AttendanceRepository) GetPrimaryDeviceForUser(ctx context.Context, userID pgtype.UUID, courseID pgtype.UUID) (string, error) {
	deviceID, err := r.queries.GetPrimaryDeviceForUser(ctx, db.GetPrimaryDeviceForUserParams{
		UserID:   userID,
		CourseID: courseID,
	})
	if err != nil {
		return "", err
	}
	return deviceID.String, nil
}

func (r *AttendanceRepository) DeleteAttendanceRecordsByUserAndCourse(ctx context.Context, userID pgtype.UUID, courseID pgtype.UUID) error {
	return r.queries.DeleteAttendanceRecordsByUserAndCourse(ctx, db.DeleteAttendanceRecordsByUserAndCourseParams{
		UserID:   userID,
		CourseID: courseID,
	})
}