package usecase

import (
	"errors"

	"backend/internal/auth"
	"backend/internal/helpers"
	"backend/internal/models"
	"backend/internal/repository"
)

var (
	ErrEmailAlreadyUsed    = errors.New("email already used")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrMissingJWTSecret    = errors.New("jwt secret is not configured")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)

type AuthUsecase struct {
	userRepo     *repository.UserRepository
	tokenService *auth.TokenService
}

func NewAuthUsecase(userRepo *repository.UserRepository, tokenService *auth.TokenService) *AuthUsecase {
	return &AuthUsecase{userRepo: userRepo, tokenService: tokenService}
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

func (u *AuthUsecase) Login(email, password string) (auth.TokenPair, error) {
	user, err := u.userRepo.FindByEmail(email)
	if err != nil {
		return auth.TokenPair{}, err
	}
	if user == nil {
		return auth.TokenPair{}, ErrInvalidCredentials
	}

	if err := helpers.CheckPasswordHash(password, user.Password); err != nil {
		return auth.TokenPair{}, ErrInvalidCredentials
	}

	identity := auth.NewTokenIdentityFromUser(user)
	pair, err := u.tokenService.IssueTokenPair(identity)
	if err != nil {
		if errors.Is(err, auth.ErrMissingSecret) {
			return auth.TokenPair{}, ErrMissingJWTSecret
		}
		return auth.TokenPair{}, err
	}

	return pair, nil
}

func (u *AuthUsecase) Refresh(refreshToken string) (auth.TokenPair, error) {
	if refreshToken == "" {
		return auth.TokenPair{}, ErrInvalidRefreshToken
	}

	identity, err := u.tokenService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return auth.TokenPair{}, ErrInvalidRefreshToken
	}

	u.tokenService.RevokeRefreshToken(refreshToken)
	pair, err := u.tokenService.IssueTokenPair(identity)
	if err != nil {
		if errors.Is(err, auth.ErrMissingSecret) {
			return auth.TokenPair{}, ErrMissingJWTSecret
		}
		return auth.TokenPair{}, err
	}

	return pair, nil
}

func (u *AuthUsecase) Logout(accessJTI, refreshToken string) {
	u.tokenService.BlacklistAccessToken(accessJTI)
	if refreshToken != "" {
		u.tokenService.RevokeRefreshToken(refreshToken)
	}
}
