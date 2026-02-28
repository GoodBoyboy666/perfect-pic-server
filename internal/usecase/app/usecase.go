package app

import (
	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/repository"
	"perfect-pic-server/internal/service"
)

type AuthUseCase struct {
	authService  service.AuthService
	userStore    repository.UserStore
	userService  service.UserService
	emailService service.EmailService
	initService  service.InitService
	dbConfig     config.DBConfig
}

type UserUseCase struct {
	userService  service.UserService
	userStore    repository.UserStore
	emailService service.EmailService
	dbConfig     config.DBConfig
}

type ImageUseCase struct {
	imageService service.ImageService
	userService  service.UserService
	userStore    repository.UserStore
	dbConfig     config.DBConfig
}

type PasskeyUseCase struct {
	passkeyService service.PasskeyService
	passkeyStore   repository.PasskeyStore
	authService    service.AuthService
	userStore      repository.UserStore
}
