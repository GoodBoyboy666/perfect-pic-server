package service

import (
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type InitPayload struct {
	Username        string
	Password        string
	SiteName        string
	SiteDescription string
}

// IsSystemInitialized 返回系统是否已完成初始化。
func IsSystemInitialized() bool {
	return !GetBool(consts.ConfigAllowInit)
}

// InitializeSystem 执行系统初始化：写入站点设置并创建管理员账号。
func InitializeSystem(payload InitPayload) error {
	passwordHashed, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	settingsToUpdate := map[string]string{
		consts.ConfigSiteName:        payload.SiteName,
		consts.ConfigSiteDescription: payload.SiteDescription,
		consts.ConfigAllowInit:       "false",
	}
	newUser := model.User{
		Username: payload.Username,
		Password: string(passwordHashed),
		Avatar:   "",
		Admin:    true,
	}
	err = repository.System.InitializeSystem(settingsToUpdate, &newUser)
	if err != nil {
		return err
	}

	ClearCache()
	return nil
}
