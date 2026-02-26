package modules

import (
	"perfect-pic-server/internal/modules/auth"
	"perfect-pic-server/internal/modules/image"
	imagerepo "perfect-pic-server/internal/modules/image/repo"
	"perfect-pic-server/internal/modules/settings"
	settingsrepo "perfect-pic-server/internal/modules/settings/repo"
	"perfect-pic-server/internal/modules/system"
	systemrepo "perfect-pic-server/internal/modules/system/repo"
	"perfect-pic-server/internal/modules/user"
	userhandler "perfect-pic-server/internal/modules/user/handler"
	userrepo "perfect-pic-server/internal/modules/user/repo"
	platformservice "perfect-pic-server/internal/platform/service"
)

type AppModules struct {
	Auth     *auth.Module
	User     *user.Module
	Image    *image.Module
	Settings *settings.Module
	System   *system.Module
}

func New(
	appService *platformservice.AppService,
	userStore userrepo.UserStore,
	imageStore imagerepo.ImageStore,
	settingStore settingsrepo.SettingStore,
	systemStore systemrepo.SystemStore,
) *AppModules {
	userModule := user.New(appService, userStore, imageStore)
	imageModule := image.New(appService, userStore, imageStore)
	authModule := auth.New(appService, userStore, userModule.Service)
	userModule.Handler = userhandler.New(userModule.Service, authModule.Service, imageModule.Service)

	return &AppModules{
		Auth:     authModule,
		User:     userModule,
		Image:    imageModule,
		Settings: settings.New(appService, settingStore),
		System:   system.New(appService, systemStore, userStore, imageStore),
	}
}
