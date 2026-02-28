package admin

import (
	"perfect-pic-server/internal/repository"
	"perfect-pic-server/internal/service"
)

type UserManageUseCase struct {
	userStore repository.UserStore
	userService service.UserService
	imageService service.ImageService
	passkeyService service.PasskeyService
}

type SettingsUseCase struct {
	emailService service.EmailService
}

type StatUseCase struct {
	imageStore repository.ImageStore
	userStore  repository.UserStore
}