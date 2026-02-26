package repo

import (
	"fmt"
	"perfect-pic-server/internal/model"
	"strings"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindUnscopedByID(id uint) (*model.User, error) {
	var user model.User
	if err := r.db.Unscoped().First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) Save(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) UpdateUsernameByID(userID uint, username string) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("username", username).Error
}

func (r *UserRepository) UpdatePasswordByID(userID uint, hashedPassword string) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("password", hashedPassword).Error
}

func (r *UserRepository) UpdateAvatar(user *model.User, filename string) error {
	return r.db.Model(user).Update("avatar", filename).Error
}

func (r *UserRepository) ClearAvatar(user *model.User) error {
	return r.db.Model(user).Select("Avatar").Updates(map[string]interface{}{"avatar": ""}).Error
}

func (r *UserRepository) UpdateByID(userID uint, updates map[string]interface{}) error {
	var user model.User
	if err := r.db.First(&user, userID).Error; err != nil {
		return err
	}
	return r.db.Model(&user).Updates(updates).Error
}

func (r *UserRepository) FieldExists(field UserField, value string, excludeUserID *uint, includeDeleted bool) (bool, error) {
	query := r.db.Model(&model.User{})
	if includeDeleted {
		query = query.Unscoped()
	}
	if excludeUserID != nil {
		query = query.Where("id != ?", *excludeUserID)
	}

	var count int64
	if err := query.Where(string(field)+" = ?", value).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ListPasskeyCredentialsByUserID 返回指定用户的全部 Passkey 凭据记录。
func (r *UserRepository) ListPasskeyCredentialsByUserID(userID uint) ([]model.PasskeyCredential, error) {
	var credentials []model.PasskeyCredential
	if err := r.db.Where("user_id = ?", userID).Order("id asc").Find(&credentials).Error; err != nil {
		return nil, err
	}
	return credentials, nil
}

// CountPasskeyCredentialsByUserID 统计指定用户已绑定的 Passkey 数量。
func (r *UserRepository) CountPasskeyCredentialsByUserID(userID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.PasskeyCredential{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// FindPasskeyCredentialByCredentialID 通过 credential_id 查找凭据。
func (r *UserRepository) FindPasskeyCredentialByCredentialID(credentialID string) (*model.PasskeyCredential, error) {
	var credential model.PasskeyCredential
	if err := r.db.Where("credential_id = ?", credentialID).First(&credential).Error; err != nil {
		return nil, err
	}
	return &credential, nil
}

// CreatePasskeyCredential 创建 Passkey 凭据记录。
func (r *UserRepository) CreatePasskeyCredential(credential *model.PasskeyCredential) error {
	return r.db.Create(credential).Error
}

// UpdatePasskeyCredentialData 更新指定凭据的序列化内容（例如登录后 signCount 变化）。
func (r *UserRepository) UpdatePasskeyCredentialData(userID uint, credentialID string, credentialJSON string) error {
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
func (r *UserRepository) DeletePasskeyCredentialByID(userID uint, passkeyID uint) error {
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
func (r *UserRepository) UpdatePasskeyCredentialNameByID(userID uint, passkeyID uint, name string) error {
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
func (r *UserRepository) DeletePasskeyCredentialsByUserID(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.PasskeyCredential{}).Error
}

func (r *UserRepository) AdminListUsers(
	keyword string,
	showDeleted bool,
	order string,
	offset int,
	limit int,
) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := r.db.Model(&model.User{})
	if showDeleted {
		query = query.Unscoped()
	}
	kw := strings.TrimSpace(keyword)
	if kw != "" {
		query = query.Where("username LIKE ?", "%"+kw+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Order(order).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *UserRepository) HardDeleteUserWithImages(userID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var user model.User
		if err := tx.Unscoped().First(&user, userID).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&model.Image{}).Error; err != nil {
			return err
		}
		return tx.Unscoped().Delete(&user).Error
	})
}

func (r *UserRepository) AdminSoftDeleteUser(userID uint, timestamp int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var user model.User
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}

		newUsername, newEmail := buildSoftDeletedIdentity(user, timestamp)
		if err := tx.Model(&user).Updates(map[string]interface{}{
			"username": newUsername,
			"email":    newEmail,
			"status":   3,
		}).Error; err != nil {
			return err
		}

		return tx.Delete(&user).Error
	})
}

func (r *UserRepository) CountAll() (int64, error) {
	var count int64
	if err := r.db.Model(&model.User{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func buildSoftDeletedIdentity(user model.User, timestamp int64) (string, string) {
	newUsername := fmt.Sprintf("%s_del_%d", user.Username, timestamp)
	newEmail := fmt.Sprintf("del_%d_%s", timestamp, user.Email)
	if len(newEmail) > 255 {
		newEmail = newEmail[:255]
	}
	return newUsername, newEmail
}
