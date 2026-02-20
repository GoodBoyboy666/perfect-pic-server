package repository

import (
	"fmt"
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
	"strings"

	"gorm.io/gorm"
)

type UserRepository struct{}

type UserField string

const (
	UserFieldUsername UserField = "username"
	UserFieldEmail    UserField = "email"
)

var User = &UserRepository{}

func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	if err := db.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindUnscopedByID(id uint) (*model.User, error) {
	var user model.User
	if err := db.DB.Unscoped().First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Create(user *model.User) error {
	return db.DB.Create(user).Error
}

func (r *UserRepository) Save(user *model.User) error {
	return db.DB.Save(user).Error
}

func (r *UserRepository) UpdateUsernameByID(userID uint, username string) error {
	return db.DB.Model(&model.User{}).Where("id = ?", userID).Update("username", username).Error
}

func (r *UserRepository) UpdatePasswordByID(userID uint, hashedPassword string) error {
	return db.DB.Model(&model.User{}).Where("id = ?", userID).Update("password", hashedPassword).Error
}

func (r *UserRepository) UpdateAvatar(user *model.User, filename string) error {
	return db.DB.Model(user).Update("avatar", filename).Error
}

func (r *UserRepository) ClearAvatar(user *model.User) error {
	return db.DB.Model(user).Select("Avatar").Updates(map[string]interface{}{"avatar": ""}).Error
}

func (r *UserRepository) UpdateByID(userID uint, updates map[string]interface{}) error {
	var user model.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		return err
	}
	return db.DB.Model(&user).Updates(updates).Error
}

func (r *UserRepository) FieldExists(field UserField, value string, excludeUserID *uint, includeDeleted bool) (bool, error) {
	query := db.DB.Model(&model.User{})
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

func (r *UserRepository) AdminListUsers(
	keyword string,
	showDeleted bool,
	order string,
	offset int,
	limit int,
) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := db.DB.Model(&model.User{})
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
	return db.DB.Transaction(func(tx *gorm.DB) error {
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
	return db.DB.Transaction(func(tx *gorm.DB) error {
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
	if err := db.DB.Model(&model.User{}).Count(&count).Error; err != nil {
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
