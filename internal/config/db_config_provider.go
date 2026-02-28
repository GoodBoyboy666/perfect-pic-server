package config

import settingsrepo "perfect-pic-server/internal/repository"

type DBConfig struct {
	settingStore settingsrepo.SettingStore
}

func NewDBConfig(settingStore settingsrepo.SettingStore) *DBConfig {
	return &DBConfig{settingStore: settingStore}
}
