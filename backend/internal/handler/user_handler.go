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

const maxProfileImageSize = 2 << 20

var allowedProfileExtensions = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".png":  {},
}

type UserHandler struct {
	usecase *usecase.UserUsecase
}

func NewUserHandler(usecase *usecase.UserUsecase) *UserHandler {
	return &UserHandler{usecase: usecase}
}

func (h *UserHandler) UploadProfilePicture(c *gin.Context) {
	userID, _, ok := getAuthClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxProfileImageSize)

	file, err := c.FormFile("image")
	if err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file size exceeds 2MB"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "image file is required"})
		return
	}

	if file.Size > maxProfileImageSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file size exceeds 2MB"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if _, ok := allowedProfileExtensions[ext]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid image type"})
		return
	}

	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	storagePath := filepath.Join("uploads", "profiles", filename)

	if err := c.SaveUploadedFile(file, storagePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store image"})
		return
	}

	publicPath := "/uploads/profiles/" + filename
	if err := h.usecase.UpdateProfilePicture(userID, publicPath); err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile picture"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"profile_picture": publicPath})
}
