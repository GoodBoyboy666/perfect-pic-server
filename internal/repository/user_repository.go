package repository

import (
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/model"
)

type UserStore interface {
	FindByID(id uint) (*model.User, error)
	FindUnscopedByID(id uint) (*model.User, error)
	FindByUsername(username string) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	Create(user *model.User) error
	Save(user *model.User) error
	UpdateUsernameByID(userID uint, username string) error
	UpdatePasswordByID(userID uint, hashedPassword string) error
	UpdateAvatar(user *model.User, filename string) error
	ClearAvatar(user *model.User) error
	UpdateByID(userID uint, updates map[string]interface{}) error
	FieldExists(field consts.UserField, value string, excludeUserID *uint, includeDeleted bool) (bool, error)
	ListUsers(keyword string, showDeleted bool, order string, offset int, limit int) ([]model.User, int64, error)
	HardDeleteUserWithImages(userID uint) error
	SoftDeleteUser(userID uint, timestamp int64) error
	CountAll() (int64, error)
}
