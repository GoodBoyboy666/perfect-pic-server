package service

import "perfect-pic-server/internal/model"

func (s *ImageService) CreateAndIncreaseUserStorage(image *model.Image, userID uint, size int64) error {
	return s.imageStore.CreateAndIncreaseUserStorage(image, userID, size)
}
