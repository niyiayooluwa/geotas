package service

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/niyiayooluwa/geotas/internal/db"
	"github.com/niyiayooluwa/geotas/internal/model"
	"github.com/niyiayooluwa/geotas/internal/repository"
)

type CourseService struct {
	courseRepo     *repository.CourseRepository
	sessionRepo    *repository.SessionRepository
	attendanceRepo *repository.AttendanceRepository
}

func NewCourseService(
	courseRepo *repository.CourseRepository,
	sessionRepo *repository.SessionRepository,
	attendanceRepo *repository.AttendanceRepository,
) *CourseService {
	return &CourseService{
		courseRepo:     courseRepo,
		sessionRepo:    sessionRepo,
		attendanceRepo: attendanceRepo,
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
	if req.Title == "" || req.Code == "" {
		return model.CourseResponse{}, errors.New("Title and Code are required")
	}

	ownerID, err := parseUUID(userID)
	if err != nil {
		return model.CourseResponse{}, err
	}

	inviteCode, err := generateInvitationCode()
	if err != nil {
		return model.CourseResponse{}, err
	}

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

	_, err = s.courseRepo.AddCourseMember(ctx, db.AddCourseMemberParams{
		CourseID: course.ID,
		UserID:   ownerID,
		Role:     "lecturer",
		MatriculationNumber: pgtype.Text{
			Valid: false,
		},
		CoLecturer: false,
	})
	if err != nil {
		return model.CourseResponse{}, errors.New("Failed to add course member")
	}

	return model.CourseResponse{
		ID:         course.ID.String(),
		OwnerID:    course.OwnerID.String(),
		Title:      course.Title,
		Code:                course.Code,
		Department:          course.Department.String,
		InviteCode:          course.InviteCode,
		ConfidenceThreshold: 0.75, // Default from DB
		CreatedAt:           course.CreatedAt.Time.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *CourseService) JoinCourse(ctx context.Context, userID string, userRole string, req model.JoinCourseRequest) (model.CourseMemberResponse, error) {
	if req.InviteCode == "" {
		return model.CourseMemberResponse{}, errors.New("invite code is required")
	}

	if userRole != "lecturer" && req.MatriculationNumber == "" {
		return model.CourseMemberResponse{}, errors.New("matriculation number is required to join a course")
	}

	course, err := s.courseRepo.GetCourseByInviteCode(ctx, req.InviteCode)
	if err != nil {
		return model.CourseMemberResponse{}, errors.New("invalid invite code")
	}

	studentID, err := parseUUID(userID)
	if err != nil {
		return model.CourseMemberResponse{}, err
	}

	if course.OwnerID == studentID {
		return model.CourseMemberResponse{}, errors.New("you cannot join your own course")
	}

	_, err = s.courseRepo.GetCourseMember(ctx, db.GetCourseMemberParams{
		CourseID: course.ID,
		UserID:   studentID,
	})
	if err == nil {
		return model.CourseMemberResponse{}, errors.New("you are already a member of this course")
	}

	roleToAssign := "student"
	coLecturer := false
	matNo := pgtype.Text{Valid: false}

	if userRole == "lecturer" {
		roleToAssign = "lecturer"
		coLecturer = true
	} else {
		matNo = pgtype.Text{String: req.MatriculationNumber, Valid: true}
	}

	member, err := s.courseRepo.AddCourseMember(ctx, db.AddCourseMemberParams{
		CourseID:            course.ID,
		UserID:              studentID,
		Role:                roleToAssign,
		MatriculationNumber: matNo,
		CoLecturer:          coLecturer,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.CourseMemberResponse{}, errors.New("this matriculation number is already registered for this course")
		}
		return model.CourseMemberResponse{}, errors.New("could not join course")
	}

	return model.CourseMemberResponse{
		ID:                  member.ID.String(),
		CourseID:            member.CourseID.String(),
		UserID:              member.UserID.String(),
		Role:                member.Role,
		MatriculationNumber: member.MatriculationNumber.String,
		JoinedAt:            member.JoinedAt.Time.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *CourseService) GetCoursesByOwner(ctx context.Context, userID string) ([]model.CourseResponse, error) {
	ownerID, err := parseUUID(userID)
	if err != nil {
		return nil, err
	}

	courses, err := s.courseRepo.GetCoursesWithStudentCountByOwner(ctx, ownerID)
	if err != nil {
		return nil, errors.New("could not fetch courses")
	}

	var response []model.CourseResponse
	for _, course := range courses {
		thresholdVal, _ := course.ConfidenceThreshold.Float64Value()

		response = append(response, model.CourseResponse{
			ID:                  course.ID.String(),
			OwnerID:             course.OwnerID.String(),
			Title:               course.Title,
			Code:                course.Code,
			InviteCode:          course.InviteCode,
			Department:          course.Department.String,
			StudentCount:        course.StudentCount,
			ConfidenceThreshold: thresholdVal.Float64,
			CreatedAt:           course.CreatedAt.Time.Format("2006-01-02 15:04:05"),
		})
	}

	return response, nil
}

// DeleteCourse deletes a course owned by userID.
// Blocked if an active session exists. DB cascades handle all child records.
func (s *CourseService) DeleteCourse(ctx context.Context, userID string, courseID string) error {
	parsedCourseID, err := parseUUID(courseID)
	if err != nil {
		return errors.New("invalid course_id")
	}

	parsedUserID, err := parseUUID(userID)
	if err != nil {
		return err
	}

	course, err := s.courseRepo.GetCourseByID(ctx, parsedCourseID)
	if err != nil {
		return errors.New("course not found")
	}

	if course.OwnerID != parsedUserID {
		return errors.New("you do not own this course")
	}

	// block deletion if an active session is running
	_, err = s.sessionRepo.GetActiveSessionByCourse(ctx, parsedCourseID)
	if err == nil {
		return errors.New("cannot delete a course with an active session — close the session first")
	}

	return s.courseRepo.DeleteCourse(ctx, parsedCourseID)
}

// LeaveCourse removes a student from a course and cascades their attendance records.
// Lecturers (course owners) are blocked — deleting the course is the correct action.
func (s *CourseService) LeaveCourse(ctx context.Context, userID string, courseID string) error {
	parsedCourseID, err := parseUUID(courseID)
	if err != nil {
		return errors.New("invalid course_id")
	}

	parsedUserID, err := parseUUID(userID)
	if err != nil {
		return err
	}

	// confirm the user is actually a member
	member, err := s.courseRepo.GetCourseMember(ctx, db.GetCourseMemberParams{
		CourseID: parsedCourseID,
		UserID:   parsedUserID,
	})
	if err != nil {
		return errors.New("you are not a member of this course")
	}

	// lecturers cannot leave their own course
	if member.Role == "lecturer" {
		return errors.New("lecturers cannot leave a course — delete the course instead")
	}

	// cascade-delete all attendance records for this student in this course
	if err := s.attendanceRepo.DeleteAttendanceRecordsByUserAndCourse(ctx, parsedUserID, parsedCourseID); err != nil {
		return errors.New("could not remove attendance records")
	}

	// remove from course_members
	if err := s.courseRepo.RemoveCourseMember(ctx, parsedCourseID, parsedUserID); err != nil {
		return errors.New("could not leave course")
	}

	return nil
}

func (s *CourseService) GetCoursesByMember(ctx context.Context, userID string) ([]model.MemberCourseResponse, error) {
	parsedID, err := parseUUID(userID)
	if err != nil {
		return nil, err
	}

	courses, err := s.courseRepo.GetCoursesByMember(ctx, parsedID)
	if err != nil {
		return nil, errors.New("could not fetch courses")
	}

	var response []model.MemberCourseResponse
	for _, c := range courses {
		thresholdVal, _ := c.ConfidenceThreshold.Float64Value()

		response = append(response, model.MemberCourseResponse{
			ID:                  c.ID.String(),
			OwnerID:             c.OwnerID.String(),
			Title:               c.Title,
			Code:                c.Code,
			Department:          c.Department.String,
			InviteCode:          c.InviteCode,
			CreatedAt:           c.CreatedAt.Time.Format("2006-01-02 15:04:05"),
			Role:                c.Role,
			MatriculationNumber: c.MatriculationNumber.String,
			StudentCount:        c.StudentCount,
			ConfidenceThreshold: thresholdVal.Float64,
		})
	}

	return response, nil
}

// GetCourseMembers returns the full roster of a course
func (s *CourseService) GetCourseMembers(ctx context.Context, userID string, courseID string) ([]model.CourseMemberDetailsResponse, error) {
	parsedUserID, err := parseUUID(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	parsedCourseID, err := parseUUID(courseID)
	if err != nil {
		return nil, errors.New("invalid course ID")
	}

	// Verify the caller is a member (either student or lecturer)
	_, err = s.courseRepo.GetCourseMember(ctx, db.GetCourseMemberParams{
		CourseID: parsedCourseID,
		UserID:   parsedUserID,
	})
	if err != nil {
		return nil, errors.New("you are not a member of this course")
	}

	members, err := s.courseRepo.GetCourseMembersByCourse(ctx, parsedCourseID)
	if err != nil {
		return nil, errors.New("could not fetch course members")
	}

	var response []model.CourseMemberDetailsResponse
	for _, m := range members {
		var avatarURL *string
		if m.AvatarUrl.Valid {
			avatarURL = &m.AvatarUrl.String
		}
		response = append(response, model.CourseMemberDetailsResponse{
			UserID:              m.UserID.String(),
			FirstName:           m.FirstName,
			LastName:            m.LastName,
			Email:               m.Email,
			AvatarURL:           avatarURL,
			Role:                string(m.Role),
			MatriculationNumber: m.MatriculationNumber.String,
			CoLecturer:          m.CoLecturer,
			JoinedAt:            m.JoinedAt.Time.Format("2006-01-02 15:04:05"),
		})
	}
	if response == nil {
		response = make([]model.CourseMemberDetailsResponse, 0)
	}

	return response, nil
}

// GetCourseAttendance returns all attendance records across every session in a course.
// Only the course owner (lecturer) can call this.
func (s *CourseService) GetCourseAttendance(ctx context.Context, userID string, courseID string) ([]model.DetailedAttendanceResponse, error) {
	parsedCourseID, err := parseUUID(courseID)
	if err != nil {
		return nil, errors.New("invalid course_id")
	}

	parsedUserID, err := parseUUID(userID)
	if err != nil {
		return nil, err
	}

	member, err := s.courseRepo.GetCourseMember(ctx, db.GetCourseMemberParams{
		CourseID: parsedCourseID,
		UserID:   parsedUserID,
	})
	if err != nil || member.Role != "lecturer" {
		return nil, errors.New("you do not have permission to view attendance for this course")
	}

	records, err := s.attendanceRepo.GetAttendanceByCourse(ctx, parsedCourseID)
	if err != nil {
		return nil, errors.New("could not fetch attendance records")
	}

	var response []model.DetailedAttendanceResponse
	for _, r := range records {
		response = append(response, model.DetailedAttendanceResponse{
			ID:                  r.ID.String(),
			SessionID:           r.SessionID.String(),
			UserID:              r.UserID.String(),
			MarkedAt:            r.MarkedAt.Time.Format("2006-01-02 15:04:05"),
			Method:              r.Method,
			Latitude:            r.Latitude,
			Longitude:           r.Longitude,
			DistanceFromCenter:  r.DistanceFromCenter,
			MockLocationDetected: r.MockLocationDetected,
			ConfidenceScore:     r.ConfidenceScore,
			WeekNumber:          r.WeekNumber,
			DeviceID:            r.DeviceID.String,
			DeviceModel:         r.DeviceModel.String,
			OsVersion:           r.OsVersion.String,
			FirstName:           r.FirstName,
			LastName:            r.LastName,
			MatriculationNumber: r.MatriculationNumber.String,
		})
	}

	return response, nil
}

func (s *CourseService) RemoveStudent(ctx context.Context, ownerID string, courseID string, targetUserID string) error {
	parsedOwnerID, err := parseUUID(ownerID)
	if err != nil {
		return errors.New("invalid owner_id")
	}

	parsedCourseID, err := parseUUID(courseID)
	if err != nil {
		return errors.New("invalid course_id")
	}

	parsedTargetID, err := parseUUID(targetUserID)
	if err != nil {
		return errors.New("invalid target_user_id")
	}

	course, err := s.courseRepo.GetCourseByID(ctx, parsedCourseID)
	if err != nil {
		return errors.New("course not found")
	}

	actor, err := s.courseRepo.GetCourseMember(ctx, db.GetCourseMemberParams{
		CourseID: parsedCourseID,
		UserID:   parsedOwnerID,
	})
	if err != nil || actor.Role != "lecturer" {
		return errors.New("you do not have permission to remove members")
	}

	if parsedTargetID == course.OwnerID {
		return errors.New("the course owner cannot be removed")
	}

	member, err := s.courseRepo.GetCourseMember(ctx, db.GetCourseMemberParams{
		CourseID: parsedCourseID,
		UserID:   parsedTargetID,
	})
	if err != nil {
		return errors.New("user is not a member of this course")
	}

	if member.Role == "lecturer" && parsedTargetID == parsedOwnerID {
		return errors.New("you cannot remove yourself from the course — delete the course instead")
	}

	if err := s.attendanceRepo.DeleteAttendanceRecordsByUserAndCourse(ctx, parsedTargetID, parsedCourseID); err != nil {
		return errors.New("could not remove attendance records")
	}

	if err := s.courseRepo.RemoveCourseMember(ctx, parsedCourseID, parsedTargetID); err != nil {
		return errors.New("could not remove member")
	}

	return nil
}

func (s *CourseService) RotateInviteCode(ctx context.Context, userID string, courseID string) (string, error) {
	parsedCourseID, err := parseUUID(courseID)
	if err != nil {
		return "", errors.New("invalid course_id")
	}

	parsedUserID, err := parseUUID(userID)
	if err != nil {
		return "", err
	}

	actor, err := s.courseRepo.GetCourseMember(ctx, db.GetCourseMemberParams{
		CourseID: parsedCourseID,
		UserID:   parsedUserID,
	})
	if err != nil || actor.Role != "lecturer" {
		return "", errors.New("you do not have permission to rotate invite codes for this course")
	}

	newInviteCode, err := generateInvitationCode()
	if err != nil {
		return "", err
	}

	_, err = s.courseRepo.UpdateCourseInviteCode(ctx, db.UpdateCourseInviteCodeParams{
		ID:         parsedCourseID,
		InviteCode: newInviteCode,
	})
	if err != nil {
		return "", errors.New("could not rotate invite code")
	}

	return newInviteCode, nil
}

func (s *CourseService) UpdateCourseSettings(ctx context.Context, userID string, courseID string, req model.UpdateCourseSettingsRequest) (model.CourseResponse, error) {
	if req.ConfidenceThreshold < 0 || req.ConfidenceThreshold > 1 {
		return model.CourseResponse{}, errors.New("confidence_threshold must be between 0.00 and 1.00")
	}

	parsedUserID, err := parseUUID(userID)
	if err != nil {
		return model.CourseResponse{}, errors.New("invalid user ID")
	}

	parsedCourseID, err := parseUUID(courseID)
	if err != nil {
		return model.CourseResponse{}, errors.New("invalid course ID")
	}

	actor, err := s.courseRepo.GetCourseMember(ctx, db.GetCourseMemberParams{
		CourseID: parsedCourseID,
		UserID:   parsedUserID,
	})
	if err != nil || actor.Role != "lecturer" {
		return model.CourseResponse{}, errors.New("you do not have permission to update settings for this course")
	}

	var numericThreshold pgtype.Numeric
	err = numericThreshold.Scan(fmt.Sprintf("%f", req.ConfidenceThreshold))
	if err != nil {
		return model.CourseResponse{}, errors.New("invalid threshold format")
	}

	course, err := s.courseRepo.UpdateCourseConfidenceThreshold(ctx, db.UpdateCourseConfidenceThresholdParams{
		ConfidenceThreshold: numericThreshold,
		ID:                  parsedCourseID,
	})
	if err != nil {
		return model.CourseResponse{}, errors.New("could not update settings or forbidden (not owner)")
	}

	thresholdVal, _ := course.ConfidenceThreshold.Float64Value()

	return model.CourseResponse{
		ID:                  course.ID.String(),
		OwnerID:             course.OwnerID.String(),
		Title:               course.Title,
		Code:                course.Code,
		Department:          course.Department.String,
		InviteCode:          course.InviteCode,
		ConfidenceThreshold: thresholdVal.Float64,
		CreatedAt:           course.CreatedAt.Time.Format("2006-01-02 15:04:05"),
	}, nil
}