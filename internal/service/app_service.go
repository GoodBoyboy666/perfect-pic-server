package service

import (
	"perfect-pic-server/internal/repository"
	"sync"
)

type AppService struct {
	repos         *repository.Repositories
	settingsCache sync.Map
}

func NewAppService(repos *repository.Repositories) *AppService {
	return &AppService{repos: repos}
}
