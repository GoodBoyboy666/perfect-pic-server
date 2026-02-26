package repo

import "perfect-pic-server/internal/model"

type ListImagesParams struct {
	UserID      *uint
	Username    string
	Filename    string
	ID          *uint
	Offset      int
	Limit       int
	PreloadUser bool
}

type ImageStore interface {
	CreateAndIncreaseUserStorage(image *model.Image, userID uint, size int64) error
	DeleteAndDecreaseUserStorage(image *model.Image) error
	BatchDeleteAndDecreaseUserStorage(imageIDs []uint, userSizeMap map[uint]int64) error
	ListUserImages(userID uint, filename string, id *uint, offset int, limit int) ([]model.Image, int64, error)
	CountByUserID(userID uint) (int64, error)
	FindByIDAndUserID(imageID uint, userID uint) (*model.Image, error)
	FindByIDsAndUserID(ids []uint, userID uint) ([]model.Image, error)
	AdminListImages(username string, filename string, userID *uint, id *uint, offset int, limit int) ([]model.Image, int64, error)
	FindByID(id uint) (*model.Image, error)
	FindByIDs(ids []uint) ([]model.Image, error)
	FindUnscopedByUserID(userID uint) ([]model.Image, error)
	CountAll() (int64, error)
	SumAllSize() (int64, error)
}
