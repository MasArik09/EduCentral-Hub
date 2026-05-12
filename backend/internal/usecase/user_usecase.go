package usecase

import (
	"errors"

	"backend/internal/repository"

	"gorm.io/gorm"
)

var ErrUserNotFound = errors.New("user not found")

type UserUsecase struct {
	userRepo *repository.UserRepository
}

func NewUserUsecase(userRepo *repository.UserRepository) *UserUsecase {
	return &UserUsecase{userRepo: userRepo}
}

func (u *UserUsecase) UpdateProfilePicture(userID uint, path string) error {
	if err := u.userRepo.UpdateProfilePicture(userID, path); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	return nil
}
