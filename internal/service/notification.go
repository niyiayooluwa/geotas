package service

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/niyiayooluwa/geotas/internal/db"
	"github.com/niyiayooluwa/geotas/internal/model"
	"github.com/niyiayooluwa/geotas/internal/repository"
)

type NotificationService struct {
	notifRepo *repository.NotificationRepository
}

func NewNotificationService(notifRepo *repository.NotificationRepository) *NotificationService {
	return &NotificationService{notifRepo: notifRepo}
}

func (s *NotificationService) CreateNotificationsForCourseMembers(ctx context.Context, courseID string, notifType string, payload []byte) error {
	parsedCourseID, err := parseUUID(courseID)
	if err != nil {
		return err
	}

	return s.notifRepo.CreateNotificationsForCourseMembers(ctx, db.CreateNotificationsForCourseMembersParams{
		CourseID: parsedCourseID,
		Type:     notifType,
		Payload:  payload,
	})
}

func (s *NotificationService) GetUnseenNotificationsByUser(ctx context.Context, userID string, since string) ([]model.NotificationResponse, error) {
	parsedUserID, err := parseUUID(userID)
	if err != nil {
		return nil, err
	}

	sinceTime := time.Time{} // Defaults to 0001-01-01
	if since != "" {
		t, err := time.Parse(time.RFC3339, since)
		if err == nil {
			sinceTime = t
		}
	}

	pgSince := pgtype.Timestamptz{Time: sinceTime, Valid: true}

	notifs, err := s.notifRepo.GetUnseenNotificationsByUser(ctx, db.GetUnseenNotificationsByUserParams{
		UserID:    parsedUserID,
		CreatedAt: pgSince,
	})
	if err != nil {
		return nil, errors.New("could not fetch notifications")
	}

	var response []model.NotificationResponse
	for _, n := range notifs {
		var payload any = n.Payload
		
		response = append(response, model.NotificationResponse{
			ID:        n.ID.String(),
			CourseID:  n.CourseID.String(),
			Type:      n.Type,
			Payload:   payload,
			CreatedAt: n.CreatedAt.Time.Format(time.RFC3339),
		})
	}
	return response, nil
}

func (s *NotificationService) MarkNotificationAsSeen(ctx context.Context, userID string, notificationID string) error {
	parsedUserID, err := parseUUID(userID)
	if err != nil {
		return err
	}
	parsedNotifID, err := parseUUID(notificationID)
	if err != nil {
		return err
	}

	return s.notifRepo.MarkNotificationAsSeen(ctx, db.MarkNotificationAsSeenParams{
		ID:     parsedNotifID,
		UserID: parsedUserID,
	})
}
