package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

type CourseHandler struct {
	usecase *usecase.CourseUsecase
}

func NewCourseHandler(usecase *usecase.CourseUsecase) *CourseHandler {
	return &CourseHandler{usecase: usecase}
}

type createCourseRequest struct {
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Lessons     []lessonInput `json:"lessons"`
}

type lessonInput struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (h *CourseHandler) CreateCourse(c *gin.Context) {
	var req createCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}

	userID, role, ok := getAuthClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
		return
	}

	lessons := make([]usecase.LessonInput, 0, len(req.Lessons))
	for _, lesson := range req.Lessons {
		lessons = append(lessons, usecase.LessonInput{
			Title:   lesson.Title,
			Content: lesson.Content,
		})
	}

	course, err := h.usecase.CreateCourse(userID, role, req.Title, req.Description, lessons)
	if err != nil {
		if errors.Is(err, usecase.ErrUnauthorizedCourseCreation) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized: Access restricted to teaching staff only"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create course"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"course": course})
}

func (h *CourseHandler) EnrollCourse(c *gin.Context) {
	courseID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course id"})
		return
	}

	userID, role, ok := getAuthClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
		return
	}

	if !isEnrollmentRoleAllowed(c, role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admin or siswa can enroll"})
		return
	}

	if err := h.usecase.EnrollCourse(userID, role, courseID); err != nil {
		if errors.Is(err, usecase.ErrUnauthorizedEnrollment) {
			c.JSON(http.StatusForbidden, gin.H{"error": "only admin or siswa can enroll"})
			return
		}
		if errors.Is(err, usecase.ErrCourseNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
			return
		}
		if errors.Is(err, usecase.ErrAlreadyEnrolled) {
			c.JSON(http.StatusConflict, gin.H{"error": "already enrolled"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enroll"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "enrolled"})
}

func (h *CourseHandler) GetCourseDetails(c *gin.Context) {
	courseID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course id"})
		return
	}

	course, err := h.usecase.GetCourseDetails(courseID)
	if err != nil {
		if errors.Is(err, usecase.ErrCourseNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch course"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"course": course})
}

func getAuthClaims(c *gin.Context) (uint, string, bool) {
	value, ok := c.Get("user_id")
	if !ok {
		return 0, "", false
	}

	userID, ok := value.(uint)
	if !ok {
		return 0, "", false
	}

	roleValue, ok := c.Get("role")
	if !ok {
		return 0, "", false
	}

	role, ok := roleValue.(string)
	if !ok {
		return 0, "", false
	}

	return userID, role, true
}

func parseUintParam(c *gin.Context, key string) (uint, error) {
	value := c.Param(key)
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, err
	}

	return uint(parsed), nil
}

func isEnrollmentRoleAllowed(c *gin.Context, role string) bool {
	roleIDValue, ok := c.Get("role_id")
	if ok {
		roleID, ok := roleIDValue.(uint)
		if ok {
			return roleID == 1 || roleID == 3
		}
	}

	role = strings.ToLower(strings.TrimSpace(role))
	return role == "admin" || role == "siswa"
}
