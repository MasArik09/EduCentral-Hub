package usecase

import (
	"errors"

	"backend/internal/models"
	"backend/internal/repository"
)

var (
	ErrLessonUnauthorizedCreate = errors.New("only admin or course teacher can create lessons")
	ErrLessonUnauthorizedView   = errors.New("not enrolled in course")
	ErrLessonNotFound           = errors.New("lesson not found")
)

type LessonUsecase struct {
	lessonRepo *repository.LessonRepository
}

func NewLessonUsecase(lessonRepo *repository.LessonRepository) *LessonUsecase {
	return &LessonUsecase{lessonRepo: lessonRepo}
}

func (u *LessonUsecase) CreateLesson(userID uint, role string, courseID uint, title, content string) (*models.Lesson, error) {
	course, err := u.lessonRepo.GetCourseByID(courseID)
	if err != nil {
		return nil, err
	}
	if course == nil {
		return nil, ErrCourseNotFound
	}

	if !isLessonCreatorAllowed(role, userID, course.TeacherID) {
		return nil, ErrLessonUnauthorizedCreate
	}

	lesson := &models.Lesson{
		CourseID: courseID,
		Title:    title,
		Content:  content,
	}

	if err := u.lessonRepo.CreateLesson(lesson); err != nil {
		return nil, err
	}

	return lesson, nil
}

func (u *LessonUsecase) ListLessonsByCourse(userID uint, role string, courseID uint) ([]models.Lesson, error) {
	course, err := u.lessonRepo.GetCourseByID(courseID)
	if err != nil {
		return nil, err
	}
	if course == nil {
		return nil, ErrCourseNotFound
	}

	allowed, err := u.canViewLessons(userID, role, course)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrLessonUnauthorizedView
	}

	return u.lessonRepo.GetLessonsByCourseID(courseID)
}

func (u *LessonUsecase) GetLessonDetail(userID uint, role string, lessonID uint) (*models.Lesson, error) {
	lesson, err := u.lessonRepo.GetLessonWithCourse(lessonID)
	if err != nil {
		return nil, err
	}
	if lesson == nil {
		return nil, ErrLessonNotFound
	}

	allowed, err := u.canViewLessons(userID, role, &lesson.Course)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrLessonUnauthorizedView
	}

	return lesson, nil
}

func (u *LessonUsecase) canViewLessons(userID uint, role string, course *models.Course) (bool, error) {
	if course == nil {
		return false, nil
	}

	if isAdminRole(role) || course.TeacherID == userID {
		return true, nil
	}

	enrolled, err := u.lessonRepo.EnrollmentExists(userID, course.ID)
	if err != nil {
		return false, err
	}

	return enrolled, nil
}

func isLessonCreatorAllowed(role string, userID, teacherID uint) bool {
	return isAdminRole(role) || userID == teacherID
}
