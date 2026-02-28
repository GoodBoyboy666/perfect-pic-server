package config

import settingsrepo "perfect-pic-server/internal/repository"

type DBConfig struct {
	settingStore settingsrepo.SettingStore
}