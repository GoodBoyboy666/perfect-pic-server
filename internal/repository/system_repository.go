package repository

import "perfect-pic-server/internal/model"

type SystemStore interface {
	InitializeSystem(settingValues map[string]string, admin *model.User) error
}
