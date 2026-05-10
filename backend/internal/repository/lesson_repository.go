package repository

import (
	"errors"

	"backend/internal/models"

	"gorm.io/gorm"
)

type LessonRepository struct {
	db *gorm.DB
}

func NewLessonRepository(db *gorm.DB) *LessonRepository {
	return &LessonRepository{db: db}
}

func (r *LessonRepository) CreateLesson(lesson *models.Lesson) error {
	return r.db.Create(lesson).Error
}

func (r *LessonRepository) GetLessonsByCourseID(courseID uint) ([]models.Lesson, error) {
	var lessons []models.Lesson
	if err := r.db.Where("course_id = ?", courseID).Find(&lessons).Error; err != nil {
		return nil, err
	}

	return lessons, nil
}

func (r *LessonRepository) GetLessonByID(lessonID uint) (*models.Lesson, error) {
	var lesson models.Lesson
	if err := r.db.First(&lesson, lessonID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &lesson, nil
}

func (r *LessonRepository) GetLessonWithCourse(lessonID uint) (*models.Lesson, error) {
	var lesson models.Lesson
	if err := r.db.Preload("Course").First(&lesson, lessonID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &lesson, nil
}

func (r *LessonRepository) UpdateLesson(lesson *models.Lesson) error {
	return r.db.Save(lesson).Error
}

func (r *LessonRepository) DeleteLesson(lessonID uint) error {
	return r.db.Delete(&models.Lesson{}, lessonID).Error
}

func (r *LessonRepository) GetCourseByID(courseID uint) (*models.Course, error) {
	var course models.Course
	if err := r.db.First(&course, courseID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &course, nil
}

func (r *LessonRepository) EnrollmentExists(userID, courseID uint) (bool, error) {
	var enrollment models.Enrollment
	if err := r.db.Where("user_id = ? AND course_id = ?", userID, courseID).First(&enrollment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
