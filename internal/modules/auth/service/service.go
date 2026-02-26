package service

import (
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/modules/auth/repo"
	userdto "perfect-pic-server/internal/modules/user/dto"
	userservice "perfect-pic-server/internal/modules/user/service"
	platformservice "perfect-pic-server/internal/platform/service"
)

type Service struct {
	*platformservice.AppService
	userStore   repo.UserStore
	userService *userservice.Service
}

func New(appService *platformservice.AppService, userStore repo.UserStore, userService *userservice.Service) *Service {
	return &Service{
		AppService:  appService,
		userStore:   userStore,
		userService: userService,
	}
}

func (s *Service) IsUsernameTaken(username string, excludeUserID *uint, includeDeleted bool) (bool, error) {
	if s.userService == nil {
		return false, platformservice.NewInternalError("服务未初始化")
	}
	return s.userService.IsUsernameTaken(username, excludeUserID, includeDeleted)
}

func (s *Service) IsEmailTaken(email string, excludeUserID *uint, includeDeleted bool) (bool, error) {
	if s.userService == nil {
		return false, platformservice.NewInternalError("服务未初始化")
	}
	return s.userService.IsEmailTaken(email, excludeUserID, includeDeleted)
}

func (s *Service) GenerateForgetPasswordToken(userID uint) (string, error) {
	if s.userService == nil {
		return "", platformservice.NewInternalError("服务未初始化")
	}
	return s.userService.GenerateForgetPasswordToken(userID)
}

func (s *Service) VerifyForgetPasswordToken(token string) (uint, bool) {
	if s.userService == nil {
		return 0, false
	}
	return s.userService.VerifyForgetPasswordToken(token)
}

func (s *Service) GenerateEmailChangeToken(userID uint, oldEmail, newEmail string) (string, error) {
	if s.userService == nil {
		return "", platformservice.NewInternalError("服务未初始化")
	}
	return s.userService.GenerateEmailChangeToken(userID, oldEmail, newEmail)
}

func (s *Service) VerifyEmailChangeToken(token string) (*userdto.EmailChangeToken, bool) {
	if s.userService == nil {
		return nil, false
	}
	return s.userService.VerifyEmailChangeToken(token)
}

func (s *Service) IsSystemInitialized() bool {
	return !s.GetBool(consts.ConfigAllowInit)
}
