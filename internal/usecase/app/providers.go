package app

import (
	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/repository"
	"perfect-pic-server/internal/service"
)

type AppUseCase struct {
	Auth    *AuthUseCase
	User    *UserUseCase
	Image   *ImageUseCase
	Passkey *PasskeyUseCase
}
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

func NewAuthUseCase(
	authService service.AuthService,
	userStore repository.UserStore,
	userService service.UserService,
	emailService service.EmailService,
	initService service.InitService,
	dbConfig *config.DBConfig,
) AuthUseCase {
	return AuthUseCase{
		authService:  authService,
		userStore:    userStore,
		userService:  userService,
		emailService: emailService,
		initService:  initService,
		dbConfig:     *dbConfig,
	}
}

func NewUserUseCase(
	userService service.UserService,
	userStore repository.UserStore,
	emailService service.EmailService,
	dbConfig *config.DBConfig,
) UserUseCase {
	return UserUseCase{
		userService:  userService,
		userStore:    userStore,
		emailService: emailService,
		dbConfig:     *dbConfig,
	}
}

func NewImageUseCase(
	imageService service.ImageService,
	userService service.UserService,
	userStore repository.UserStore,
	dbConfig *config.DBConfig,
) ImageUseCase {
	return ImageUseCase{
		imageService: imageService,
		userService:  userService,
		userStore:    userStore,
		dbConfig:     *dbConfig,
	}
}

func NewPasskeyUseCase(
	passkeyService service.PasskeyService,
	passkeyStore repository.PasskeyStore,
	authService service.AuthService,
	userStore repository.UserStore,
) PasskeyUseCase {
	return PasskeyUseCase{
		passkeyService: passkeyService,
		passkeyStore:   passkeyStore,
		authService:    authService,
		userStore:      userStore,
	}
}

func NewAppUseCase(
	authUseCase *AuthUseCase,
	userUseCase *UserUseCase,
	imageUseCase *ImageUseCase,
	passkeyUseCase *PasskeyUseCase,
) AppUseCase {
	return AppUseCase{
		Auth:    authUseCase,
		User:    userUseCase,
		Image:   imageUseCase,
		Passkey: passkeyUseCase,
	}
}
