package service

import (
	"perfect-pic-server/internal/modules/settings/repo"
	platformservice "perfect-pic-server/internal/platform/service"
)

type Service struct {
	*platformservice.AppService
	settingStore repo.SettingStore
}

func New(appService *platformservice.AppService, settingStore repo.SettingStore) *Service {
	return &Service{
		AppService:   appService,
		settingStore: settingStore,
	}
}
