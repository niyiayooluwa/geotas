package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/niyiayooluwa/geotas/internal/db"
	"github.com/niyiayooluwa/geotas/internal/model"
	"github.com/niyiayooluwa/geotas/internal/repository"
)

type ScheduleService struct {
	scheduleRepo *repository.ScheduleRepository
	courseRepo   *repository.CourseRepository
	notifService *NotificationService
}

func NewScheduleService(scheduleRepo *repository.ScheduleRepository, courseRepo *repository.CourseRepository, notifService *NotificationService) *ScheduleService {
	return &ScheduleService{
		scheduleRepo: scheduleRepo,
		courseRepo:   courseRepo,
		notifService: notifService,
	}
}

func parseTimeString(s string) (pgtype.Time, error) {
	parts := strings.Split(s, ":")
	if len(parts) < 2 {
		return pgtype.Time{}, errors.New("invalid time format")
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil {
		return pgtype.Time{}, err
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return pgtype.Time{}, err
	}
	s_val := 0
	if len(parts) == 3 {
		s_val, _ = strconv.Atoi(parts[2])
	}
	microseconds := int64(h*3600+m*60+s_val) * 1_000_000
	return pgtype.Time{Microseconds: microseconds, Valid: true}, nil
}

func formatTimeFromPg(t pgtype.Time) string {
	seconds := t.Microseconds / 1_000_000
	h := seconds / 3600
	m := (seconds % 3600) / 60
	return fmt.Sprintf("%02d:%02d", h, m)
}

func (s *ScheduleService) CreateSchedule(ctx context.Context, userID string, courseID string, req model.ScheduleRequest) (model.ScheduleResponse, error) {
	parsedUserID, err := parseUUID(userID)
	if err != nil {
		return model.ScheduleResponse{}, err
	}
	parsedCourseID, err := parseUUID(courseID)
	if err != nil {
		return model.ScheduleResponse{}, errors.New("invalid course id")
	}

	course, err := s.courseRepo.GetCourseByID(ctx, parsedCourseID)
	if err != nil {
		return model.ScheduleResponse{}, errors.New("course not found")
	}
	if course.OwnerID != parsedUserID {
		return model.ScheduleResponse{}, errors.New("only the course owner can create schedules")
	}

	startTime, err := parseTimeString(req.StartTime)
	if err != nil {
		return model.ScheduleResponse{}, errors.New("invalid start time")
	}
	endTime, err := parseTimeString(req.EndTime)
	if err != nil {
		return model.ScheduleResponse{}, errors.New("invalid end time")
	}

	schedule, err := s.scheduleRepo.CreateSchedule(ctx, db.CreateScheduleParams{
		CourseID:  parsedCourseID,
		DayOfWeek: req.DayOfWeek,
		StartTime: startTime,
		EndTime:   endTime,
		Venue:     req.Venue,
	})
	if err != nil {
		return model.ScheduleResponse{}, errors.New("could not create schedule")
	}

	// Trigger Notification
	payload, _ := json.Marshal(map[string]string{
		"schedule_id": schedule.ID.String(),
		"venue":       schedule.Venue,
	})
	_ = s.notifService.CreateNotificationsForCourseMembers(ctx, courseID, "schedule_created", payload)

	return model.ScheduleResponse{
		ID:        schedule.ID.String(),
		CourseID:  schedule.CourseID.String(),
		DayOfWeek: schedule.DayOfWeek,
		StartTime: formatTimeFromPg(schedule.StartTime),
		EndTime:   formatTimeFromPg(schedule.EndTime),
		Venue:     schedule.Venue,
	}, nil
}

func (s *ScheduleService) UpdateSchedule(ctx context.Context, userID string, scheduleID string, req model.ScheduleRequest) (model.ScheduleResponse, error) {
	parsedUserID, err := parseUUID(userID)
	if err != nil {
		return model.ScheduleResponse{}, err
	}
	parsedScheduleID, err := parseUUID(scheduleID)
	if err != nil {
		return model.ScheduleResponse{}, errors.New("invalid schedule id")
	}

	schedule, err := s.scheduleRepo.GetScheduleByID(ctx, parsedScheduleID)
	if err != nil {
		return model.ScheduleResponse{}, errors.New("schedule not found")
	}

	course, err := s.courseRepo.GetCourseByID(ctx, schedule.CourseID)
	if err != nil {
		return model.ScheduleResponse{}, errors.New("course not found")
	}
	if course.OwnerID != parsedUserID {
		return model.ScheduleResponse{}, errors.New("only the course owner can update schedules")
	}

	startTime, _ := parseTimeString(req.StartTime)
	endTime, _ := parseTimeString(req.EndTime)

	updated, err := s.scheduleRepo.UpdateSchedule(ctx, db.UpdateScheduleParams{
		ID:        parsedScheduleID,
		DayOfWeek: req.DayOfWeek,
		StartTime: startTime,
		EndTime:   endTime,
		Venue:     req.Venue,
	})
	if err != nil {
		return model.ScheduleResponse{}, errors.New("could not update schedule")
	}

	// Trigger Notification
	payload, _ := json.Marshal(map[string]string{
		"schedule_id": updated.ID.String(),
		"venue":       updated.Venue,
	})
	_ = s.notifService.CreateNotificationsForCourseMembers(ctx, updated.CourseID.String(), "schedule_updated", payload)

	return model.ScheduleResponse{
		ID:        updated.ID.String(),
		CourseID:  updated.CourseID.String(),
		DayOfWeek: updated.DayOfWeek,
		StartTime: formatTimeFromPg(updated.StartTime),
		EndTime:   formatTimeFromPg(updated.EndTime),
		Venue:     updated.Venue,
	}, nil
}

func (s *ScheduleService) GetSchedulesByCourse(ctx context.Context, userID string, courseID string) ([]model.ScheduleResponse, error) {
	parsedUserID, err := parseUUID(userID)
	if err != nil {
		return nil, err
	}
	parsedCourseID, err := parseUUID(courseID)
	if err != nil {
		return nil, errors.New("invalid course id")
	}

	_, err = s.courseRepo.GetCourseMember(ctx, db.GetCourseMemberParams{
		CourseID: parsedCourseID,
		UserID:   parsedUserID,
	})
	if err != nil {
		return nil, errors.New("you do not have permission to view schedules for this course")
	}

	schedules, err := s.scheduleRepo.GetSchedulesByCourse(ctx, parsedCourseID)
	if err != nil {
		return nil, errors.New("could not fetch schedules")
	}

	var response []model.ScheduleResponse
	for _, sched := range schedules {
		response = append(response, model.ScheduleResponse{
			ID:        sched.ID.String(),
			CourseID:  sched.CourseID.String(),
			DayOfWeek: sched.DayOfWeek,
			StartTime: formatTimeFromPg(sched.StartTime),
			EndTime:   formatTimeFromPg(sched.EndTime),
			Venue:     sched.Venue,
		})
	}
	return response, nil
}

func (s *ScheduleService) DeleteSchedule(ctx context.Context, userID string, scheduleID string) error {
	parsedUserID, err := parseUUID(userID)
	if err != nil {
		return err
	}
	parsedScheduleID, err := parseUUID(scheduleID)
	if err != nil {
		return errors.New("invalid schedule id")
	}

	schedule, err := s.scheduleRepo.GetScheduleByID(ctx, parsedScheduleID)
	if err != nil {
		return errors.New("schedule not found")
	}

	course, err := s.courseRepo.GetCourseByID(ctx, schedule.CourseID)
	if err != nil {
		return errors.New("course not found")
	}
	if course.OwnerID != parsedUserID {
		return errors.New("only the course owner can delete schedules")
	}

	return s.scheduleRepo.DeleteSchedule(ctx, parsedScheduleID)
}
