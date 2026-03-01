package config

import (
	settingsrepo "perfect-pic-server/internal/repository"
	"sync"
)

type DBConfig struct {
	settingStore  settingsrepo.SettingStore
	settingsCache sync.Map
}

func NewDBConfig(settingStore settingsrepo.SettingStore) *DBConfig {
	return &DBConfig{settingStore: settingStore}
}
