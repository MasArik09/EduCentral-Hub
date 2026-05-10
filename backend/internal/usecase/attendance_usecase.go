package usecase

import (
	"errors"
	"time"

	"backend/internal/models"
	"backend/internal/repository"
)

var (
	ErrAttendanceAlreadyMarked = errors.New("attendance already marked for today")
	ErrAttendanceNotEnrolled   = errors.New("not enrolled in course")
	ErrAttendanceCourseNotFound = errors.New("course not found")
)

type AttendanceUsecase struct {
	attendanceRepo *repository.AttendanceRepository
}

func NewAttendanceUsecase(attendanceRepo *repository.AttendanceRepository) *AttendanceUsecase {
	return &AttendanceUsecase{attendanceRepo: attendanceRepo}
}

func (u *AttendanceUsecase) MarkAttendance(userID, courseID uint) (*models.Attendance, error) {
	course, err := u.attendanceRepo.GetCourseByID(courseID)
	if err != nil {
		return nil, err
	}
	if course == nil {
		return nil, ErrAttendanceCourseNotFound
	}

	enrolled, err := u.attendanceRepo.EnrollmentExists(userID, courseID)
	if err != nil {
		return nil, err
	}
	if !enrolled {
		return nil, ErrAttendanceNotEnrolled
	}

	today := toDate(time.Now())
	exists, err := u.attendanceRepo.AttendanceExists(userID, courseID, today)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrAttendanceAlreadyMarked
	}

	attendance := &models.Attendance{
		UserID:   userID,
		CourseID: courseID,
		Status:   "Present",
		Date:     today,
	}

	if err := u.attendanceRepo.CreateAttendance(attendance); err != nil {
		return nil, err
	}

	return attendance, nil
}

func (u *AttendanceUsecase) ListMyAttendance(userID, courseID uint) ([]models.Attendance, error) {
	course, err := u.attendanceRepo.GetCourseByID(courseID)
	if err != nil {
		return nil, err
	}
	if course == nil {
		return nil, ErrAttendanceCourseNotFound
	}

	enrolled, err := u.attendanceRepo.EnrollmentExists(userID, courseID)
	if err != nil {
		return nil, err
	}
	if !enrolled {
		return nil, ErrAttendanceNotEnrolled
	}

	return u.attendanceRepo.ListAttendanceByCourseUser(userID, courseID)
}

func toDate(value time.Time) time.Time {
	location := value.Location()
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, location)
}
