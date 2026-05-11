package handler

import (
	"errors"
	"net/http"
	"time"

	"backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

type AttendanceHandler struct {
	usecase *usecase.AttendanceUsecase
}

func NewAttendanceHandler(usecase *usecase.AttendanceUsecase) *AttendanceHandler {
	return &AttendanceHandler{usecase: usecase}
}

type markAttendanceRequest struct {
	UserID      uint   `json:"user_id"`
	UserIDs     []uint `json:"user_ids"`
	Status      string `json:"status"`
	SessionDate string `json:"session_date"`
}

func (h *AttendanceHandler) MarkAttendance(c *gin.Context) {
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

	if !isAttendanceRoleAllowed(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized: Access restricted to teaching staff only"})
		return
	}

	var req markAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.Status == "" || req.SessionDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status and session_date are required"})
		return
	}

	sessionDate, err := time.Parse("2006-01-02", req.SessionDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_date must be in YYYY-MM-DD format"})
		return
	}

	studentIDs := collectStudentIDs(req.UserID, req.UserIDs)
	if len(studentIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id or user_ids are required"})
		return
	}

	attendances, err := h.usecase.MarkAttendanceTeacher(userID, role, courseID, studentIDs, req.Status, sessionDate)
	if err != nil {
		if errors.Is(err, usecase.ErrAttendanceCourseNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
			return
		}
		if errors.Is(err, usecase.ErrAttendanceUnauthorized) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized: Access restricted to teaching staff only"})
			return
		}
		if errors.Is(err, usecase.ErrAttendanceNotEnrolled) {
			c.JSON(http.StatusForbidden, gin.H{"error": "not enrolled in course"})
			return
		}
		if errors.Is(err, usecase.ErrAttendanceAlreadyMarked) {
			c.JSON(http.StatusConflict, gin.H{"error": "attendance already marked for session"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark attendance"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"attendance": attendances})
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

func isAttendanceRoleAllowed(role string) bool {
	switch role {
	case "Admin", "Guru", "Asdos":
		return true
	default:
		return false
	}
}

func collectStudentIDs(singleID uint, batchIDs []uint) []uint {
	unique := make(map[uint]struct{})
	if singleID != 0 {
		unique[singleID] = struct{}{}
	}
	for _, id := range batchIDs {
		if id == 0 {
			continue
		}
		unique[id] = struct{}{}
	}

	if len(unique) == 0 {
		return nil
	}

	ids := make([]uint, 0, len(unique))
	for id := range unique {
		ids = append(ids, id)
	}

	return ids
}
