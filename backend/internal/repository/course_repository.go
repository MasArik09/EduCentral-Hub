package repository

import (
	"errors"

	"backend/internal/models"

	"gorm.io/gorm"
)

type CourseRepository struct {
	db *gorm.DB
}

func NewCourseRepository(db *gorm.DB) *CourseRepository {
	return &CourseRepository{db: db}
}

func (r *CourseRepository) CreateCourse(course *models.Course) error {
	return r.db.Create(course).Error
}

func (r *CourseRepository) GetCourseByID(courseID uint) (*models.Course, error) {
	var course models.Course
	if err := r.db.First(&course, courseID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &course, nil
}

func (r *CourseRepository) GetCourseDetails(courseID uint) (*models.Course, error) {
	var course models.Course
	if err := r.db.Preload("Lessons").Preload("Teacher").Preload("Teacher.Role").Preload("Teacher.Profile").Preload("Students").First(&course, courseID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &course, nil
}

func (r *CourseRepository) UpdateCourse(course *models.Course) error {
	return r.db.Save(course).Error
}

func (r *CourseRepository) DeleteCourse(courseID uint) error {
	return r.db.Delete(&models.Course{}, courseID).Error
}

func (r *CourseRepository) CreateLesson(lesson *models.Lesson) error {
	return r.db.Create(lesson).Error
}

func (r *CourseRepository) CreateLessons(lessons []models.Lesson) error {
	if len(lessons) == 0 {
		return nil
	}
	return r.db.Create(&lessons).Error
}

func (r *CourseRepository) GetLessonByID(lessonID uint) (*models.Lesson, error) {
	var lesson models.Lesson
	if err := r.db.First(&lesson, lessonID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &lesson, nil
}

func (r *CourseRepository) ListLessonsByCourse(courseID uint) ([]models.Lesson, error) {
	var lessons []models.Lesson
	if err := r.db.Where("course_id = ?", courseID).Find(&lessons).Error; err != nil {
		return nil, err
	}

	return lessons, nil
}

func (r *CourseRepository) UpdateLesson(lesson *models.Lesson) error {
	return r.db.Save(lesson).Error
}

func (r *CourseRepository) DeleteLesson(lessonID uint) error {
	return r.db.Delete(&models.Lesson{}, lessonID).Error
}

func (r *CourseRepository) EnrollmentExists(userID, courseID uint) (bool, error) {
	var enrollment models.Enrollment
	if err := r.db.Where("user_id = ? AND course_id = ?", userID, courseID).First(&enrollment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *CourseRepository) CreateEnrollment(userID, courseID uint) error {
	enrollment := &models.Enrollment{
		UserID:   userID,
		CourseID: courseID,
	}
	return r.db.Create(enrollment).Error
}
