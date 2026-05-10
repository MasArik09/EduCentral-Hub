package repository

import (
	"errors"
	"time"

	"backend/internal/models"

	"gorm.io/gorm"
)

type AttendanceRepository struct {
	db *gorm.DB
}

func NewAttendanceRepository(db *gorm.DB) *AttendanceRepository {
	return &AttendanceRepository{db: db}
}

func (r *AttendanceRepository) CreateAttendance(attendance *models.Attendance) error {
	return r.db.Create(attendance).Error
}

func (r *AttendanceRepository) AttendanceExists(userID, courseID uint, date time.Time) (bool, error) {
	var attendance models.Attendance
	if err := r.db.Where("user_id = ? AND course_id = ? AND date = ?", userID, courseID, date).First(&attendance).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *AttendanceRepository) ListAttendanceByCourseUser(userID, courseID uint) ([]models.Attendance, error) {
	var attendances []models.Attendance
	if err := r.db.Where("user_id = ? AND course_id = ?", userID, courseID).Order("date desc").Find(&attendances).Error; err != nil {
		return nil, err
	}

	return attendances, nil
}

func (r *AttendanceRepository) GetCourseByID(courseID uint) (*models.Course, error) {
	var course models.Course
	if err := r.db.First(&course, courseID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &course, nil
}

func (r *AttendanceRepository) EnrollmentExists(userID, courseID uint) (bool, error) {
	var enrollment models.Enrollment
	if err := r.db.Where("user_id = ? AND course_id = ?", userID, courseID).First(&enrollment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
