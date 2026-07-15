package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/niyiayooluwa/geotas/internal/db"
)

type ScheduleRepository struct {
	queries *db.Queries
}

func NewScheduleRepository(queries *db.Queries) *ScheduleRepository {
	return &ScheduleRepository{queries: queries}
}

func (r *ScheduleRepository) CreateSchedule(ctx context.Context, params db.CreateScheduleParams) (db.Schedule, error) {
	return r.queries.CreateSchedule(ctx, params)
}

func (r *ScheduleRepository) UpdateSchedule(ctx context.Context, params db.UpdateScheduleParams) (db.Schedule, error) {
	return r.queries.UpdateSchedule(ctx, params)
}

func (r *ScheduleRepository) GetSchedulesByCourse(ctx context.Context, courseID pgtype.UUID) ([]db.Schedule, error) {
	return r.queries.GetSchedulesByCourse(ctx, courseID)
}

func (r *ScheduleRepository) GetScheduleByID(ctx context.Context, id pgtype.UUID) (db.Schedule, error) {
	return r.queries.GetScheduleByID(ctx, id)
}

func (r *ScheduleRepository) DeleteSchedule(ctx context.Context, id pgtype.UUID) error {
	return r.queries.DeleteSchedule(ctx, id)
}
