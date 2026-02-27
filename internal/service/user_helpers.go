package service

import (
	"perfect-pic-server/internal/model"
)

func (s *Service) findUserByID(id uint) (*model.User, error) {
	return s.userStore.FindByID(id)
}

func (s *Service) findUserByUsername(username string) (*model.User, error) {
	return s.userStore.FindByUsername(username)
}

func (s *Service) findUserByEmail(email string) (*model.User, error) {
	return s.userStore.FindByEmail(email)
}

func (s *Service) createUser(user *model.User) error {
	return s.userStore.Create(user)
}

func (s *Service) saveUser(user *model.User) error {
	return s.userStore.Save(user)
}

func (s *Service) findPasskeyCredentialByCredentialID(credentialID string) (*model.PasskeyCredential, error) {
	return s.userStore.FindPasskeyCredentialByCredentialID(credentialID)
}

func (s *Service) createPasskeyCredential(credential *model.PasskeyCredential) error {
	return s.userStore.CreatePasskeyCredential(credential)
}

func (s *Service) listPasskeyCredentialsByUserID(userID uint) ([]model.PasskeyCredential, error) {
	return s.userStore.ListPasskeyCredentialsByUserID(userID)
}

func (s *Service) deletePasskeyCredentialByID(userID uint, passkeyID uint) error {
	return s.userStore.DeletePasskeyCredentialByID(userID, passkeyID)
}

func (s *Service) updatePasskeyCredentialNameByID(userID uint, passkeyID uint, name string) error {
	return s.userStore.UpdatePasskeyCredentialNameByID(userID, passkeyID, name)
}

func (s *Service) updatePasskeyCredentialData(userID uint, credentialID string, credentialJSON string) error {
	return s.userStore.UpdatePasskeyCredentialData(userID, credentialID, credentialJSON)
}

func (s *Service) countPasskeyCredentialsByUserID(userID uint) (int64, error) {
	return s.userStore.CountPasskeyCredentialsByUserID(userID)
}
