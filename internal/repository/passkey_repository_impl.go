package repository

import (
	"perfect-pic-server/internal/model"

	"gorm.io/gorm"
)

type PasskeyRepository struct {
	db *gorm.DB
}

// ListPasskeyCredentialsByUserID 返回指定用户的全部 Passkey 凭据记录。
func (r *PasskeyRepository) ListPasskeyCredentialsByUserID(userID uint) ([]model.PasskeyCredential, error) {
	var credentials []model.PasskeyCredential
	if err := r.db.Where("user_id = ?", userID).Order("id asc").Find(&credentials).Error; err != nil {
		return nil, err
	}
	return credentials, nil
}

// CountPasskeyCredentialsByUserID 统计指定用户已绑定的 Passkey 数量。
func (r *PasskeyRepository) CountPasskeyCredentialsByUserID(userID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.PasskeyCredential{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// FindPasskeyCredentialByCredentialID 通过 credential_id 查找凭据。
func (r *PasskeyRepository) FindPasskeyCredentialByCredentialID(credentialID string) (*model.PasskeyCredential, error) {
	var credential model.PasskeyCredential
	if err := r.db.Where("credential_id = ?", credentialID).First(&credential).Error; err != nil {
		return nil, err
	}
	return &credential, nil
}

// CreatePasskeyCredential 创建 Passkey 凭据记录。
func (r *PasskeyRepository) CreatePasskeyCredential(credential *model.PasskeyCredential) error {
	return r.db.Create(credential).Error
}

// UpdatePasskeyCredentialData 更新指定凭据的序列化内容（例如登录后 signCount 变化）。
func (r *PasskeyRepository) UpdatePasskeyCredentialData(userID uint, credentialID string, credentialJSON string) error {
	tx := r.db.Model(&model.PasskeyCredential{}).
		Where("user_id = ? AND credential_id = ?", userID, credentialID).
		Update("credential", credentialJSON)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeletePasskeyCredentialByID 删除指定用户下的某条 Passkey 凭据记录。
func (r *PasskeyRepository) DeletePasskeyCredentialByID(userID uint, passkeyID uint) error {
	tx := r.db.Where("user_id = ? AND id = ?", userID, passkeyID).Delete(&model.PasskeyCredential{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdatePasskeyCredentialNameByID 更新指定用户下某条 Passkey 凭据的人类可读名称。
func (r *PasskeyRepository) UpdatePasskeyCredentialNameByID(userID uint, passkeyID uint, name string) error {
	tx := r.db.Model(&model.PasskeyCredential{}).
		Where("user_id = ? AND id = ?", userID, passkeyID).
		Update("name", name)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeletePasskeyCredentialsByUserID 删除指定用户下的全部 Passkey 凭据记录。
func (r *PasskeyRepository) DeletePasskeyCredentialsByUserID(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.PasskeyCredential{}).Error
}
