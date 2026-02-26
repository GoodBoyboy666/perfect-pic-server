package service

import (
	settingsrepo "perfect-pic-server/internal/modules/settings/repo"
	settingsruntime "perfect-pic-server/internal/modules/settings/runtime"
)

type AppService struct {
	*settingsruntime.Service
}

func NewAppService(settingStore settingsrepo.SettingStore) *AppService {
	return &AppService{
		Service: settingsruntime.New(settingStore),
	}
}
