package service

import (
	"perfect-pic-server/internal/modules/system/repo"
	platformservice "perfect-pic-server/internal/platform/service"
)

type Service struct {
	*platformservice.AppService
	systemStore repo.SystemStore
	userStore   repo.UserStore
	imageStore  repo.ImageStore
}

func New(
	appService *platformservice.AppService,
	systemStore repo.SystemStore,
	userStore repo.UserStore,
	imageStore repo.ImageStore,
) *Service {
	return &Service{
		AppService:  appService,
		systemStore: systemStore,
		userStore:   userStore,
		imageStore:  imageStore,
	}
}
