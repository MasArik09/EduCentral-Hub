package handler

import (
	"errors"
	"net/http"
	"time"

	"backend/internal/auth"
	"backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	usecase *usecase.AuthUsecase
}

func NewAuthHandler(usecase *usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{usecase: usecase}
}

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	RoleID   uint   `json:"role_id"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshResponse struct {
	Token string `json:"token"`
}

const refreshTokenCookiePath = "/"

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username, email, and password are required"})
		return
	}

	user, err := h.usecase.Register(req.Username, req.Email, req.Password, req.RoleID)
	if err != nil {
		if errors.Is(err, usecase.ErrEmailAlreadyUsed) {
			c.JSON(http.StatusConflict, gin.H{"error": "email already used"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "registration failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user": user})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and password are required"})
		return
	}

	pair, err := h.usecase.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		if errors.Is(err, usecase.ErrMissingJWTSecret) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "jwt secret is not configured"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		return
	}

	setRefreshCookie(c, pair.RefreshToken, pair.RefreshExpiresAt)
	setAccessTokenHeader(c, pair.AccessToken)
	// Keep "token" for backward compatibility.
	c.JSON(http.StatusOK, gin.H{"token": pair.AccessToken})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie(auth.RefreshTokenCookieName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token is required"})
		return
	}

	pair, err := h.usecase.Refresh(refreshToken)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidRefreshToken) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
			return
		}
		if errors.Is(err, usecase.ErrMissingJWTSecret) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "jwt secret is not configured"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "refresh failed"})
		return
	}

	setRefreshCookie(c, pair.RefreshToken, pair.RefreshExpiresAt)
	setAccessTokenHeader(c, pair.AccessToken)
	c.JSON(http.StatusOK, refreshResponse{Token: pair.AccessToken})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	refreshToken, _ := c.Cookie(auth.RefreshTokenCookieName)

	accessJTI := ""
	if value, ok := c.Get("jti"); ok {
		if jti, ok := value.(string); ok {
			accessJTI = jti
		}
	}

	h.usecase.Logout(accessJTI, refreshToken)
	clearRefreshCookie(c)
	c.JSON(http.StatusOK, gin.H{"status": "logged out"})
}

func setAccessTokenHeader(c *gin.Context, token string) {
	if token == "" {
		return
	}

	c.Header(auth.AccessTokenHeaderName, token)
	c.Header("Access-Control-Expose-Headers", auth.AccessTokenHeaderName)
}

func setRefreshCookie(c *gin.Context, token string, expiresAt time.Time) {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}

	secure := c.Request.TLS != nil
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(auth.RefreshTokenCookieName, token, maxAge, refreshTokenCookiePath, "", secure, true)
}

func clearRefreshCookie(c *gin.Context) {
	secure := c.Request.TLS != nil
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(auth.RefreshTokenCookieName, "", -1, refreshTokenCookiePath, "", secure, true)
}
