package repository

import (
	"errors"

	"backend/internal/models"

	"gorm.io/gorm"
)

type AssignmentRepository struct {
	db *gorm.DB
}

func NewAssignmentRepository(db *gorm.DB) *AssignmentRepository {
	return &AssignmentRepository{db: db}
}

func (r *AssignmentRepository) CreateAssignment(assignment *models.Assignment) error {
	return r.db.Create(assignment).Error
}

func (r *AssignmentRepository) GetAssignmentByID(assignmentID uint) (*models.Assignment, error) {
	var assignment models.Assignment
	if err := r.db.First(&assignment, assignmentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &assignment, nil
}

func (r *AssignmentRepository) CreateSubmission(submission *models.Submission) error {
	return r.db.Create(submission).Error
}

func (r *AssignmentRepository) GetCourseByID(courseID uint) (*models.Course, error) {
	var course models.Course
	if err := r.db.First(&course, courseID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &course, nil
}

func (r *AssignmentRepository) EnrollmentExists(userID, courseID uint) (bool, error) {
	var enrollment models.Enrollment
	if err := r.db.Where("user_id = ? AND course_id = ?", userID, courseID).First(&enrollment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
