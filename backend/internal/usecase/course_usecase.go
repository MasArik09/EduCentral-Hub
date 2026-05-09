package usecase

import (
	"errors"
	"strings"

	"backend/internal/models"
	"backend/internal/repository"
)

var (
	ErrUnauthorizedCourseCreation = errors.New("only admin or guru can create courses")
	ErrUnauthorizedEnrollment     = errors.New("only admin or siswa can enroll")
	ErrCourseNotFound             = errors.New("course not found")
	ErrAlreadyEnrolled            = errors.New("student already enrolled")
)

type LessonInput struct {
	Title   string
	Content string
}

type CourseUsecase struct {
	courseRepo *repository.CourseRepository
}

func NewCourseUsecase(courseRepo *repository.CourseRepository) *CourseUsecase {
	return &CourseUsecase{courseRepo: courseRepo}
}

func (u *CourseUsecase) CreateCourse(teacherID uint, role, title, description string, lessons []LessonInput) (*models.Course, error) {
	if !isRoleAllowed(role, []string{"admin", "guru"}) {
		return nil, ErrUnauthorizedCourseCreation
	}

	course := &models.Course{
		Title:       title,
		Description: description,
		TeacherID:   teacherID,
	}

	if err := u.courseRepo.CreateCourse(course); err != nil {
		return nil, err
	}

	if len(lessons) > 0 {
		lessonModels := make([]models.Lesson, 0, len(lessons))
		for _, lesson := range lessons {
			if strings.TrimSpace(lesson.Title) == "" {
				continue
			}
			lessonModels = append(lessonModels, models.Lesson{
				CourseID: course.ID,
				Title:    lesson.Title,
				Content:  lesson.Content,
			})
		}

		if err := u.courseRepo.CreateLessons(lessonModels); err != nil {
			return nil, err
		}
	}

	return u.courseRepo.GetCourseDetails(course.ID)
}

func (u *CourseUsecase) EnrollCourse(studentID uint, role string, courseID uint) error {
	if !isRoleAllowed(role, []string{"admin", "siswa"}) {
		return ErrUnauthorizedEnrollment
	}

	course, err := u.courseRepo.GetCourseByID(courseID)
	if err != nil {
		return err
	}
	if course == nil {
		return ErrCourseNotFound
	}

	exists, err := u.courseRepo.EnrollmentExists(studentID, courseID)
	if err != nil {
		return err
	}
	if exists {
		return ErrAlreadyEnrolled
	}

	return u.courseRepo.CreateEnrollment(studentID, courseID)
}

func (u *CourseUsecase) GetCourseDetails(courseID uint) (*models.Course, error) {
	course, err := u.courseRepo.GetCourseDetails(courseID)
	if err != nil {
		return nil, err
	}
	if course == nil {
		return nil, ErrCourseNotFound
	}

	return course, nil
}

func isRoleAllowed(role string, allowed []string) bool {
	role = strings.TrimSpace(strings.ToLower(role))
	for _, candidate := range allowed {
		if role == candidate {
			return true
		}
	}
	return false
}
