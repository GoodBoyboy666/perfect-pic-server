package service

import (
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/modules/user/repo"
	platformservice "perfect-pic-server/internal/platform/service"
)

type ImageService interface {
	FindUnscopedByUserID(userID uint) ([]model.Image, error)
}

type Service struct {
	*platformservice.AppService
	userStore    repo.UserStore
	imageService ImageService
}

func New(appService *platformservice.AppService, userStore repo.UserStore) *Service {
	return &Service{
		AppService: appService,
		userStore:  userStore,
	}
}

func (s *Service) SetImageService(imageService ImageService) {
	s.imageService = imageService
}
