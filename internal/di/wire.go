//go:build wireinject
// +build wireinject

package di

import (
	"perfect-pic-server/internal/modules"
	imagerepo "perfect-pic-server/internal/modules/image/repo"
	settingsrepo "perfect-pic-server/internal/modules/settings/repo"
	systemrepo "perfect-pic-server/internal/modules/system/repo"
	userrepo "perfect-pic-server/internal/modules/user/repo"
	"perfect-pic-server/internal/platform/service"
	"perfect-pic-server/internal/router"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func InitializeApplication(gormDB *gorm.DB) (*Application, error) {
	wire.Build(
		userrepo.NewUserRepository,
		imagerepo.NewImageRepository,
		settingsrepo.NewSettingRepository,
		systemrepo.NewSystemRepository,
		service.NewAppService,
		modules.New,
		router.NewRouter,
		NewApplication,
	)
	return nil, nil
}
