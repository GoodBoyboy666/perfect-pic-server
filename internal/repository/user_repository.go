package repository

import "perfect-pic-server/internal/model"

type UserField string

const (
	UserFieldUsername UserField = "username"
	UserFieldEmail    UserField = "email"
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
	FieldExists(field UserField, value string, excludeUserID *uint, includeDeleted bool) (bool, error)
	AdminListUsers(keyword string, showDeleted bool, order string, offset int, limit int) ([]model.User, int64, error)
	HardDeleteUserWithImages(userID uint) error
	AdminSoftDeleteUser(userID uint, timestamp int64) error
	CountAll() (int64, error)
}
