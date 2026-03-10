package service

func (s *PasskeyService) DeletePasskeyCredentialsByUserID(userID uint) error {
	return s.passkeyStore.DeletePasskeyCredentialsByUserID(userID)
}
