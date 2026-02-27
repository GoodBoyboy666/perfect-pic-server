package service

import (
	"errors"
	platformservice "perfect-pic-server/internal/common"
	"perfect-pic-server/internal/consts"
	moduledto "perfect-pic-server/internal/dto"
	"perfect-pic-server/internal/model"
	systemrepo "perfect-pic-server/internal/repository"
	"perfect-pic-server/internal/utils"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// IsSystemInitialized 返回系统是否已完成初始化。
func (s *Service) IsSystemInitialized() bool {
	return !s.GetBool(consts.ConfigAllowInit)
}

// InitializeSystem 执行系统初始化：写入站点设置并创建管理员账号。
func (s *Service) InitializeSystem(payload moduledto.InitRequest) error {
	if s.IsSystemInitialized() {
		return platformservice.NewForbiddenError("已初始化，无法重复初始化")
	}
	if ok, msg := utils.ValidateUsernameAllowReserved(payload.Username); !ok {
		return platformservice.NewValidationError(msg)
	}
	if ok, msg := utils.ValidatePassword(payload.Password); !ok {
		return platformservice.NewValidationError(msg)
	}
	if strings.TrimSpace(payload.SiteName) == "" {
		return platformservice.NewValidationError("站点名称不能为空")
	}
	if strings.TrimSpace(payload.SiteDescription) == "" {
		return platformservice.NewValidationError("站点描述不能为空")
	}

	passwordHashed, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		return platformservice.NewInternalError("初始化失败")
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
	err = s.systemStore.InitializeSystem(settingsToUpdate, &newUser)
	if err != nil {
		if errors.Is(err, systemrepo.ErrSystemAlreadyInitialized) {
			// 其他实例已完成初始化时，立即清理缓存并返回业务态冲突。
			s.ClearCache()
			return platformservice.NewForbiddenError("已初始化，无法重复初始化")
		}
		return platformservice.NewInternalError("初始化失败")
	}

	s.ClearCache()
	return nil
}
