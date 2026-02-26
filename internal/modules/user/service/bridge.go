package service

import "perfect-pic-server/internal/model"

// FindByID 提供跨模块用户查询能力。
func (s *Service) FindByID(id uint) (*model.User, error) {
	return s.userStore.FindByID(id)
}

// FindByUsername 提供按用户名查询能力。
func (s *Service) FindByUsername(username string) (*model.User, error) {
	return s.userStore.FindByUsername(username)
}

// FindByEmail 提供按邮箱查询能力。
func (s *Service) FindByEmail(email string) (*model.User, error) {
	return s.userStore.FindByEmail(email)
}

// Create 提供用户创建能力。
func (s *Service) Create(user *model.User) error {
	return s.userStore.Create(user)
}

// Save 提供用户保存能力。
func (s *Service) Save(user *model.User) error {
	return s.userStore.Save(user)
}

// UpdateAvatar 提供用户头像更新能力。
func (s *Service) UpdateAvatar(user *model.User, filename string) error {
	return s.userStore.UpdateAvatar(user, filename)
}

// ClearAvatar 提供用户头像清空能力。
func (s *Service) ClearAvatar(user *model.User) error {
	return s.userStore.ClearAvatar(user)
}

// CountAll 提供用户总量统计能力。
func (s *Service) CountAll() (int64, error) {
	return s.userStore.CountAll()
}

// ListPasskeyCredentialsByUserID 提供 Passkey 列表查询能力。
func (s *Service) ListPasskeyCredentialsByUserID(userID uint) ([]model.PasskeyCredential, error) {
	return s.userStore.ListPasskeyCredentialsByUserID(userID)
}

// CountPasskeyCredentialsByUserID 提供 Passkey 数量统计能力。
func (s *Service) CountPasskeyCredentialsByUserID(userID uint) (int64, error) {
	return s.userStore.CountPasskeyCredentialsByUserID(userID)
}

// FindPasskeyCredentialByCredentialID 提供按凭据 ID 查询 Passkey 能力。
func (s *Service) FindPasskeyCredentialByCredentialID(credentialID string) (*model.PasskeyCredential, error) {
	return s.userStore.FindPasskeyCredentialByCredentialID(credentialID)
}

// CreatePasskeyCredential 提供 Passkey 创建能力。
func (s *Service) CreatePasskeyCredential(credential *model.PasskeyCredential) error {
	return s.userStore.CreatePasskeyCredential(credential)
}

// UpdatePasskeyCredentialData 提供 Passkey 凭据更新能力。
func (s *Service) UpdatePasskeyCredentialData(userID uint, credentialID string, credentialJSON string) error {
	return s.userStore.UpdatePasskeyCredentialData(userID, credentialID, credentialJSON)
}

// DeletePasskeyCredentialByID 提供 Passkey 删除能力。
func (s *Service) DeletePasskeyCredentialByID(userID uint, passkeyID uint) error {
	return s.userStore.DeletePasskeyCredentialByID(userID, passkeyID)
}

// UpdatePasskeyCredentialNameByID 提供 Passkey 命名更新能力。
func (s *Service) UpdatePasskeyCredentialNameByID(userID uint, passkeyID uint, name string) error {
	return s.userStore.UpdatePasskeyCredentialNameByID(userID, passkeyID, name)
}
