package service

import (
	"perfect-pic-server/internal/modules/system/repo"
	platformservice "perfect-pic-server/internal/platform/service"
)

type UserService interface {
	CountAll() (int64, error)
}

type ImageService interface {
	CountAll() (int64, error)
	SumAllSize() (int64, error)
}

type Service struct {
	*platformservice.AppService
	systemStore  repo.SystemStore
	userService  UserService
	imageService ImageService
}

func New(
	appService *platformservice.AppService,
	systemStore repo.SystemStore,
	userService UserService,
	imageService ImageService,
) *Service {
	return &Service{
		AppService:   appService,
		systemStore:  systemStore,
		userService:  userService,
		imageService: imageService,
	}
}
