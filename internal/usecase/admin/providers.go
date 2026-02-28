package admin

import (
	"perfect-pic-server/internal/repository"
	"perfect-pic-server/internal/service"
)

type UserManageUseCase struct {
	userService    *service.UserService
	imageService   *service.ImageService
	passkeyService *service.PasskeyService
}

type SettingsUseCase struct {
	emailService *service.EmailService
}

type StatUseCase struct {
	imageStore repository.ImageStore
	userStore  repository.UserStore
}

func NewUserManageUseCase(
	userService *service.UserService,
	imageService *service.ImageService,
	passkeyService *service.PasskeyService,
) *UserManageUseCase {
	return &UserManageUseCase{
		userService:    userService,
		imageService:   imageService,
		passkeyService: passkeyService,
	}
}

func NewSettingsUseCase(emailService *service.EmailService) *SettingsUseCase {
	return &SettingsUseCase{emailService: emailService}
}

func NewStatUseCase(imageStore repository.ImageStore, userStore repository.UserStore) *StatUseCase {
	return &StatUseCase{imageStore: imageStore, userStore: userStore}
}

