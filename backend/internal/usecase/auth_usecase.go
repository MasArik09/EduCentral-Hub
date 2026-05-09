package usecase

import (
	"errors"
	"os"
	"time"

	"backend/internal/helpers"
	"backend/internal/models"
	"backend/internal/repository"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrEmailAlreadyUsed   = errors.New("email already used")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrMissingJWTSecret   = errors.New("jwt secret is not configured")
)

type AuthUsecase struct {
	userRepo *repository.UserRepository
}

func NewAuthUsecase(userRepo *repository.UserRepository) *AuthUsecase {
	return &AuthUsecase{userRepo: userRepo}
}

func (u *AuthUsecase) Register(username, email, password string, roleID uint) (*models.User, error) {
	existing, err := u.userRepo.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailAlreadyUsed
	}

	hashed, err := helpers.HashPassword(password)
	if err != nil {
		return nil, err
	}

	if roleID == 0 {
		roleID = 1
	}

	user := &models.User{
		Username: username,
		Email:    email,
		Password: hashed,
		RoleID:   roleID,
	}

	if err := u.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *AuthUsecase) Login(email, password string) (string, error) {
	user, err := u.userRepo.FindByEmail(email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", ErrInvalidCredentials
	}

	if err := helpers.CheckPasswordHash(password, user.Password); err != nil {
		return "", ErrInvalidCredentials
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", ErrMissingJWTSecret
	}

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role.Name,
		"role_id": user.RoleID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}
