package handler

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

const maxAssignmentUploadSize = 20 << 20

var allowedAssignmentExtensions = map[string]struct{}{
	".pdf":  {},
	".txt":  {},
	".docx": {},
	".zip":  {},
}

type AssignmentHandler struct {
	usecase *usecase.AssignmentUsecase
}

func NewAssignmentHandler(usecase *usecase.AssignmentUsecase) *AssignmentHandler {
	return &AssignmentHandler{usecase: usecase}
}

type createAssignmentRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	DueDate     string `json:"due_date"`
}

func (h *AssignmentHandler) CreateAssignment(c *gin.Context) {
	courseID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course id"})
		return
	}

	var req createAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if strings.TrimSpace(req.Title) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}

	if strings.TrimSpace(req.DueDate) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "due_date is required"})
		return
	}

	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "due_date must be in YYYY-MM-DD format"})
		return
	}

	userID, role, ok := getAuthClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
		return
	}

	assignment, err := h.usecase.CreateAssignment(userID, role, courseID, req.Title, req.Description, dueDate)
	if err != nil {
		if errors.Is(err, usecase.ErrCourseNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
			return
		}
		if errors.Is(err, usecase.ErrAssignmentUnauthorizedCreate) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized: Access restricted to teaching staff only"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create assignment"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"assignment": assignment})
}

func (h *AssignmentHandler) SubmitAssignment(c *gin.Context) {
	assignmentID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assignment id"})
		return
	}

	userID, role, ok := getAuthClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxAssignmentUploadSize)

	file, err := c.FormFile("file")
	if err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file size exceeds 20MB"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	if file.Size > maxAssignmentUploadSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file size exceeds 20MB"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if _, ok := allowedAssignmentExtensions[ext]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file type"})
		return
	}

	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	storagePath := filepath.Join("uploads", "assignments", filename)

	if err := c.SaveUploadedFile(file, storagePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store file"})
		return
	}

	publicPath := "/uploads/assignments/" + filename
	submission, err := h.usecase.SubmitAssignment(userID, role, assignmentID, publicPath, time.Now())
	if err != nil {
		if errors.Is(err, usecase.ErrAssignmentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "assignment not found"})
			return
		}
		if errors.Is(err, usecase.ErrAssignmentUnauthorizedSubmit) {
			c.JSON(http.StatusForbidden, gin.H{"error": "only students can submit assignments"})
			return
		}
		if errors.Is(err, usecase.ErrAssignmentNotEnrolled) {
			c.JSON(http.StatusForbidden, gin.H{"error": "not enrolled in course"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to submit assignment"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"submission": submission})
}
