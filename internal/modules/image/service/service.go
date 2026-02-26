package service

import (
	"perfect-pic-server/internal/modules/image/repo"
	platformservice "perfect-pic-server/internal/platform/service"
)

type Service struct {
	*platformservice.AppService
	userStore  repo.UserStore
	imageStore repo.ImageStore
}

func New(appService *platformservice.AppService, userStore repo.UserStore, imageStore repo.ImageStore) *Service {
	return &Service{
		AppService: appService,
		userStore:  userStore,
		imageStore: imageStore,
	}
}
