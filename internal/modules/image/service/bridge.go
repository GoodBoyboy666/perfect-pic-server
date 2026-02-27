package service

import "perfect-pic-server/internal/model"

func (s *Service) FindUnscopedByUserID(userID uint) ([]model.Image, error) {
	return s.imageStore.FindUnscopedByUserID(userID)
}

func (s *Service) CountAll() (int64, error) {
	return s.imageStore.CountAll()
}

func (s *Service) SumAllSize() (int64, error) {
	return s.imageStore.SumAllSize()
}
