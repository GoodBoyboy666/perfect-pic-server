package repository

import (
	"errors"
	"perfect-pic-server/internal/model"
)

var ErrSystemAlreadyInitialized = errors.New("system already initialized")

type SystemStore interface {
	InitializeSystem(settingValues map[string]string, admin *model.User) error
}
