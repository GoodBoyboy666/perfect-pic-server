package service

import "perfect-pic-server/internal/model"

func (s *PasskeyService) DeletePasskeyCredentialsByUserID(userID uint) error {
	return s.passkeyStore.DeletePasskeyCredentialsByUserID(userID)
}

func (s *PasskeyService) CreatePasskeyCredential(credential *model.PasskeyCredential) error {
	return s.passkeyStore.CreatePasskeyCredential(credential)
}

func (s *PasskeyService) UpdatePasskeyCredentialData(userID uint, credentialID string, credentialJSON string) error {
	return s.passkeyStore.UpdatePasskeyCredentialData(userID, credentialID, credentialJSON)
}
