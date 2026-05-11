package handler

import (
	"errors"
	"net/http"

	"backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

type LessonHandler struct {
	usecase *usecase.LessonUsecase
}

func NewLessonHandler(usecase *usecase.LessonUsecase) *LessonHandler {
	return &LessonHandler{usecase: usecase}
}

type createLessonRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (h *LessonHandler) CreateLesson(c *gin.Context) {
	courseID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course id"})
		return
	}

	var req createLessonRequest
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

	lesson, err := h.usecase.CreateLesson(userID, role, courseID, req.Title, req.Content)
	if err != nil {
		if errors.Is(err, usecase.ErrCourseNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
			return
		}
		if errors.Is(err, usecase.ErrLessonUnauthorizedCreate) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized: Access restricted to teaching staff only"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create lesson"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"lesson": lesson})
}

func (h *LessonHandler) ListLessonsByCourse(c *gin.Context) {
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

	lessons, err := h.usecase.ListLessonsByCourse(userID, role, courseID)
	if err != nil {
		if errors.Is(err, usecase.ErrCourseNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
			return
		}
		if errors.Is(err, usecase.ErrLessonUnauthorizedView) {
			c.JSON(http.StatusForbidden, gin.H{"error": "not enrolled in course"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch lessons"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"lessons": lessons})
}

func (h *LessonHandler) GetLessonDetail(c *gin.Context) {
	lessonID, err := parseUintParam(c, "lesson_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lesson id"})
		return
	}

	userID, role, ok := getAuthClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
		return
	}

	lesson, err := h.usecase.GetLessonDetail(userID, role, lessonID)
	if err != nil {
		if errors.Is(err, usecase.ErrLessonNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "lesson not found"})
			return
		}
		if errors.Is(err, usecase.ErrLessonUnauthorizedView) {
			c.JSON(http.StatusForbidden, gin.H{"error": "not enrolled in course"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch lesson"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"lesson": lesson})
}
