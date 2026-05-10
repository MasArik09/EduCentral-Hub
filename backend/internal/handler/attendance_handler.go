package handler

import (
	"errors"
	"net/http"

	"backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

type AttendanceHandler struct {
	usecase *usecase.AttendanceUsecase
}

func NewAttendanceHandler(usecase *usecase.AttendanceUsecase) *AttendanceHandler {
	return &AttendanceHandler{usecase: usecase}
}

func (h *AttendanceHandler) MarkAttendance(c *gin.Context) {
	courseID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course id"})
		return
	}

	userID, _, ok := getAuthClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
		return
	}

	attendance, err := h.usecase.MarkAttendance(userID, courseID)
	if err != nil {
		if errors.Is(err, usecase.ErrAttendanceCourseNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
			return
		}
		if errors.Is(err, usecase.ErrAttendanceNotEnrolled) {
			c.JSON(http.StatusForbidden, gin.H{"error": "not enrolled in course"})
			return
		}
		if errors.Is(err, usecase.ErrAttendanceAlreadyMarked) {
			c.JSON(http.StatusConflict, gin.H{"error": "attendance already marked for today"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark attendance"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"attendance": attendance})
}

func (h *AttendanceHandler) ListMyAttendance(c *gin.Context) {
	courseID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course id"})
		return
	}

	userID, _, ok := getAuthClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
		return
	}

	attendances, err := h.usecase.ListMyAttendance(userID, courseID)
	if err != nil {
		if errors.Is(err, usecase.ErrAttendanceCourseNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
			return
		}
		if errors.Is(err, usecase.ErrAttendanceNotEnrolled) {
			c.JSON(http.StatusForbidden, gin.H{"error": "not enrolled in course"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch attendance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"attendance": attendances})
}
