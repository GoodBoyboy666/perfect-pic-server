package repo

import "perfect-pic-server/internal/model"

type UserStore interface {
	FindByID(id uint) (*model.User, error)
	FindByUsername(username string) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	Create(user *model.User) error
	Save(user *model.User) error
	ListPasskeyCredentialsByUserID(userID uint) ([]model.PasskeyCredential, error)
	CountPasskeyCredentialsByUserID(userID uint) (int64, error)
	FindPasskeyCredentialByCredentialID(credentialID string) (*model.PasskeyCredential, error)
	CreatePasskeyCredential(credential *model.PasskeyCredential) error
	UpdatePasskeyCredentialData(userID uint, credentialID string, credentialJSON string) error
	DeletePasskeyCredentialByID(userID uint, passkeyID uint) error
	UpdatePasskeyCredentialNameByID(userID uint, passkeyID uint, name string) error
}
