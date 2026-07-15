package repository

import (
	"context"

	"github.com/niyiayooluwa/geotas/internal/db"
)

type NotificationRepository struct {
	queries *db.Queries
}

func NewNotificationRepository(queries *db.Queries) *NotificationRepository {
	return &NotificationRepository{queries: queries}
}

func (r *NotificationRepository) CreateNotification(ctx context.Context, params db.CreateNotificationParams) (db.Notification, error) {
	return r.queries.CreateNotification(ctx, params)
}

func (r *NotificationRepository) CreateNotificationsForCourseMembers(ctx context.Context, params db.CreateNotificationsForCourseMembersParams) error {
	return r.queries.CreateNotificationsForCourseMembers(ctx, params)
}

func (r *NotificationRepository) GetUnseenNotificationsByUser(ctx context.Context, params db.GetUnseenNotificationsByUserParams) ([]db.Notification, error) {
	return r.queries.GetUnseenNotificationsByUser(ctx, params)
}

func (r *NotificationRepository) MarkNotificationAsSeen(ctx context.Context, params db.MarkNotificationAsSeenParams) error {
	return r.queries.MarkNotificationAsSeen(ctx, params)
}
