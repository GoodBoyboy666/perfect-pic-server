package repository

import "perfect-pic-server/internal/model"

type PasskeyStore interface {
	ListPasskeyCredentialsByUserID(userID uint) ([]model.PasskeyCredential, error)
	CountPasskeyCredentialsByUserID(userID uint) (int64, error)
	FindPasskeyCredentialByCredentialID(credentialID string) (*model.PasskeyCredential, error)
	CreatePasskeyCredential(credential *model.PasskeyCredential) error
	UpdatePasskeyCredentialData(userID uint, credentialID string, credentialJSON string) error
	DeletePasskeyCredentialByID(userID uint, passkeyID uint) error
	UpdatePasskeyCredentialNameByID(userID uint, passkeyID uint, name string) error
	DeletePasskeyCredentialsByUserID(userID uint) error
}
