package usecase

import (
	"errors"
	"strings"
	"time"

	"backend/internal/models"
	"backend/internal/repository"
)

var (
	ErrAssignmentUnauthorizedCreate = errors.New("only admin or course teacher can create assignments")
	ErrAssignmentUnauthorizedSubmit = errors.New("only students can submit assignments")
	ErrAssignmentNotFound           = errors.New("assignment not found")
	ErrAssignmentNotEnrolled        = errors.New("student not enrolled in course")
)

type AssignmentUsecase struct {
	repo *repository.AssignmentRepository
}

func NewAssignmentUsecase(repo *repository.AssignmentRepository) *AssignmentUsecase {
	return &AssignmentUsecase{repo: repo}
}

func (u *AssignmentUsecase) CreateAssignment(userID uint, role string, courseID uint, title, description string, dueDate time.Time) (*models.Assignment, error) {
	course, err := u.repo.GetCourseByID(courseID)
	if err != nil {
		return nil, err
	}
	if course == nil {
		return nil, ErrCourseNotFound
	}

	if !isAssignmentCreatorAllowed(role, userID, course.TeacherID) {
		return nil, ErrAssignmentUnauthorizedCreate
	}

	assignment := &models.Assignment{
		CourseID:    courseID,
		Title:       title,
		Description: description,
		DueDate:     dueDate,
	}

	if err := u.repo.CreateAssignment(assignment); err != nil {
		return nil, err
	}

	return assignment, nil
}

func (u *AssignmentUsecase) SubmitAssignment(userID uint, role string, assignmentID uint, fileURL string, submittedAt time.Time) (*models.Submission, error) {
	if !isStudentRole(role) {
		return nil, ErrAssignmentUnauthorizedSubmit
	}

	assignment, err := u.repo.GetAssignmentByID(assignmentID)
	if err != nil {
		return nil, err
	}
	if assignment == nil {
		return nil, ErrAssignmentNotFound
	}

	enrolled, err := u.repo.EnrollmentExists(userID, assignment.CourseID)
	if err != nil {
		return nil, err
	}
	if !enrolled {
		return nil, ErrAssignmentNotEnrolled
	}

	submission := &models.Submission{
		AssignmentID: assignmentID,
		UserID:       userID,
		FileURL:      fileURL,
		SubmittedAt:  submittedAt,
	}

	if err := u.repo.CreateSubmission(submission); err != nil {
		return nil, err
	}

	return submission, nil
}

func isAssignmentCreatorAllowed(role string, userID, teacherID uint) bool {
	return isAdminRole(role) || userID == teacherID
}

func isStudentRole(role string) bool {
	role = strings.ToLower(strings.TrimSpace(role))
	return role == "siswa" || role == "student"
}

func isAdminRole(role string) bool {
	role = strings.ToLower(strings.TrimSpace(role))
	return role == "admin"
}
