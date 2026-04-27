package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/niyiayooluwa/geotas/internal/db"
)

type CourseRepository struct {
	queries *db.Queries
}

func NewCourseRepository(queries *db.Queries) *CourseRepository {
	return &CourseRepository{queries: queries}
}

func (r *CourseRepository) CreateCourse(ctx context.Context, params db.CreateCourseParams) (db.Course, error) {
	return r.queries.CreateCourse(ctx, params)
}

func (r *CourseRepository) GetCourseByID(ctx context.Context, id pgtype.UUID) (db.Course, error) {
	return r.queries.GetCourseByID(ctx, id)
}

func (r *CourseRepository) GetCoursesByOwner(ctx context.Context, ownerID pgtype.UUID) ([]db.Course, error) {
	return r.queries.GetCoursesByOwner(ctx, ownerID)
}

func (r *CourseRepository) GetCourseByCode(ctx context.Context, code string) (db.Course, error) {
	return r.queries.GetCourseByCode(ctx, code)
}

func (r *CourseRepository) AddCourseMember(ctx context.Context, params db.AddCourseMemberParams) (db.CourseMember, error) {
	return r.queries.AddCourseMember(ctx, params)
}

func (r *CourseRepository) GetCourseMembersByCourse(ctx context.Context, courseID pgtype.UUID) ([]db.CourseMember, error) {
	return r.queries.GetCourseMembersByCourse(ctx, courseID)
}

func (r *CourseRepository) GetCoursesByMember(ctx context.Context, userID pgtype.UUID) ([]db.CourseMember, error) {
	return r.queries.GetCoursesByMember(ctx, userID)
}

func (r *CourseRepository) GetCourseMember(ctx context.Context, params db.GetCourseMemberParams) (db.CourseMember, error) {
	return r.queries.GetCourseMember(ctx, params)
}

func (r *CourseRepository) GetCourseByInviteCode(ctx context.Context, inviteCode string) (db.Course, error) {
	return r.queries.GetCourseByInviteCode(ctx, inviteCode)
}

func (r *CourseRepository) DeleteCourse(ctx context.Context, id pgtype.UUID) error {
	return r.queries.DeleteCourse(ctx, id)
}
