package service

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/niyiayooluwa/geotas/internal/db"
	"github.com/niyiayooluwa/geotas/internal/model"
	"github.com/niyiayooluwa/geotas/internal/repository"
)

type CourseService struct {
	courseRepo *repository.CourseRepository
}

func NewCourseService(courseRepo *repository.CourseRepository) *CourseService {
	return &CourseService{
		courseRepo: courseRepo,
	}
}

func generateInvitationCode() (string, error) {
	var bytes = make([]byte, 3)
	rand.Read(bytes)
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes)[:5], nil
}

func parseUUID(id string) (pgtype.UUID, error) {
	var uuid pgtype.UUID
	if err := uuid.Scan(id); err != nil {
		return pgtype.UUID{}, errors.New("Invalid ID")
	}
	return uuid, nil
}

func (s *CourseService) CreateCourse(ctx context.Context, userID string, req model.CreateCourseRequest) (model.CourseResponse, error) {
	// validate input
	if req.Title == "" || req.Code == "" {
		return model.CourseResponse{}, errors.New("Title and Code are required")
	}

	// parse userID to UUID
	ownerID, err := parseUUID(userID)
	if err != nil {
		return model.CourseResponse{}, err
	}

	// generate invitation code
	inviteCode, err := generateInvitationCode()
	if err != nil {
		return model.CourseResponse{}, err
	}

	// insert course into DB
	course, err := s.courseRepo.CreateCourse(ctx, db.CreateCourseParams{
		OwnerID: ownerID,
		Title:   req.Title,
		Code:    req.Code,
		Department: pgtype.Text{
			String: req.Department,
			Valid:  req.Department != "",
		},
		InviteCode: inviteCode,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.CourseResponse{}, errors.New("Course code already exists")
		}
		return model.CourseResponse{}, errors.New("Failed to create course")
	}

	// add lecturer as course owner automatially
	_, err = s.courseRepo.AddCourseMember(ctx, db.AddCourseMemberParams{
		CourseID: course.ID,
		UserID:   ownerID,
		Role:     "lecturer",
	})
	if err != nil {
		return model.CourseResponse{}, errors.New("Failed to add course member")
	}

	return model.CourseResponse{
		ID:         course.ID.String(),
		OwnerID:    course.OwnerID.String(),
		Title:      course.Title,
		Code:       course.Code,
		Department: course.Department.String,
		InviteCode: course.InviteCode,
		CreatedAt:  course.CreatedAt.Time.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *CourseService) JoinCourse(ctx context.Context, userID string, req model.JoinCourseRequest) (model.CourseMemberResponse, error) {
	if req.InviteCode == "" {
		return model.CourseMemberResponse{}, errors.New("invite code is required")
	}

	// look up course by invite code
	course, err := s.courseRepo.GetCourseByInviteCode(ctx, req.InviteCode)
	if err != nil {
		return model.CourseMemberResponse{}, errors.New("invalid invite code")
	}

	// parse student uuid
	studentID, err := parseUUID(userID)
	if err != nil {
		return model.CourseMemberResponse{}, err
	}

	// check if they own the course
	if course.OwnerID == studentID {
		return model.CourseMemberResponse{}, errors.New("you cannot join your own course")
	}

	// check if already a member
	_, err = s.courseRepo.GetCourseMember(ctx, db.GetCourseMemberParams{
		CourseID: course.ID,
		UserID:   studentID,
	})
	if err == nil {
		return model.CourseMemberResponse{}, errors.New("you are already a member of this course")
	}

	// add student as course member
	member, err := s.courseRepo.AddCourseMember(ctx, db.AddCourseMemberParams{
		CourseID: course.ID,
		UserID:   studentID,
		Role:     "student",
	})
	if err != nil {
		return model.CourseMemberResponse{}, errors.New("could not join course")
	}

	return model.CourseMemberResponse{
		ID:       member.ID.String(),
		CourseID: member.CourseID.String(),
		UserID:   member.UserID.String(),
		Role:     member.Role,
		JoinedAt: member.JoinedAt.Time.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *CourseService) GetCoursesByOwner(ctx context.Context, userID string) ([]model.CourseResponse, error) {
	ownerID, err := parseUUID(userID)
	if err != nil {
		return nil, err
	}

	courses, err := s.courseRepo.GetCoursesByOwner(ctx, ownerID)
	if err != nil {
		return nil, errors.New("could not fetch courses")
	}

	var response []model.CourseResponse
	for _, course := range courses {
		response = append(response, model.CourseResponse{
			ID:         course.ID.String(),
			OwnerID:    course.OwnerID.String(),
			Title:      course.Title,
			Code:       course.Code,
			InviteCode: course.InviteCode,
			Department: course.Department.String,
			CreatedAt:  course.CreatedAt.Time.Format("2006-01-02 15:04:05"),
		})
	}

	return response, nil
}
