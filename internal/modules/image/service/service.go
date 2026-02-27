package service

import (
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/modules/image/repo"
	platformservice "perfect-pic-server/internal/platform/service"
)

type UserService interface {
	FindByID(id uint) (*model.User, error)
	UpdateAvatar(user *model.User, filename string) error
	ClearAvatar(user *model.User) error
}

type Service struct {
	*platformservice.AppService
	userService UserService
	imageStore  repo.ImageStore
}

func New(appService *platformservice.AppService, userService UserService, imageStore repo.ImageStore) *Service {
	return &Service{
		AppService:  appService,
		userService: userService,
		imageStore:  imageStore,
	}
}
