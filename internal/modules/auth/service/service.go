package service

import (
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/model"
	userdto "perfect-pic-server/internal/modules/user/dto"
	platformservice "perfect-pic-server/internal/platform/service"
)

type UserService interface {
	FindByID(id uint) (*model.User, error)
	FindByUsername(username string) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	Create(user *model.User) error
	Save(user *model.User) error
	IsUsernameTaken(username string, excludeUserID *uint, includeDeleted bool) (bool, error)
	IsEmailTaken(email string, excludeUserID *uint, includeDeleted bool) (bool, error)
	GenerateForgetPasswordToken(userID uint) (string, error)
	VerifyForgetPasswordToken(token string) (uint, bool)
	GenerateEmailChangeToken(userID uint, oldEmail, newEmail string) (string, error)
	VerifyEmailChangeToken(token string) (*userdto.EmailChangeToken, bool)
	ListPasskeyCredentialsByUserID(userID uint) ([]model.PasskeyCredential, error)
	CountPasskeyCredentialsByUserID(userID uint) (int64, error)
	FindPasskeyCredentialByCredentialID(credentialID string) (*model.PasskeyCredential, error)
	CreatePasskeyCredential(credential *model.PasskeyCredential) error
	UpdatePasskeyCredentialData(userID uint, credentialID string, credentialJSON string) error
	DeletePasskeyCredentialByID(userID uint, passkeyID uint) error
	UpdatePasskeyCredentialNameByID(userID uint, passkeyID uint, name string) error
}

type Service struct {
	*platformservice.AppService
	userService UserService
}

func New(appService *platformservice.AppService, userService UserService) *Service {
	return &Service{
		AppService:  appService,
		userService: userService,
	}
}

func (s *Service) IsSystemInitialized() bool {
	return !s.GetBool(consts.ConfigAllowInit)
}
