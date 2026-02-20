package service

import (
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/repository"
	"perfect-pic-server/internal/utils"
	"strings"

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
	if IsSystemInitialized() {
		return NewForbiddenError("已初始化，无法重复初始化")
	}
	if ok, msg := utils.ValidateUsername(payload.Username); !ok {
		return NewValidationError(msg)
	}
	if ok, msg := utils.ValidatePassword(payload.Password); !ok {
		return NewValidationError(msg)
	}
	if strings.TrimSpace(payload.SiteName) == "" {
		return NewValidationError("站点名称不能为空")
	}
	if strings.TrimSpace(payload.SiteDescription) == "" {
		return NewValidationError("站点描述不能为空")
	}

	passwordHashed, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		return NewInternalError("初始化失败")
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
		return NewInternalError("初始化失败")
	}

	ClearCache()
	return nil
}
