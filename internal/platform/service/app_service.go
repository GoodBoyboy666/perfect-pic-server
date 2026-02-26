package service

import (
	settingsrepo "perfect-pic-server/internal/modules/settings/repo"
	"sync"
)

type AppService struct {
	settingStore  settingsrepo.SettingStore
	settingsCache sync.Map
}

func NewAppService(settingStore settingsrepo.SettingStore) *AppService {
	return &AppService{settingStore: settingStore}
}
