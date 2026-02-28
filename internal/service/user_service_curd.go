package service

import "perfect-pic-server/internal/model"

func (s *UserService) CreateUser(user *model.User) error {
	return s.userStore.Create(user)
}

func (s *UserService) SaveUser(user *model.User) error {
	return s.userStore.Save(user)
}

func (s *UserService) UpdateAvatar(user *model.User, newAvatar string) error {
	return s.userStore.UpdateAvatar(user, newAvatar)
}

func (s *UserService) ClearAvatar(user *model.User) error {
	return s.userStore.ClearAvatar(user)
}

func (s *UserService) HardDeleteUserWithImages(userID uint) error {
	return s.userStore.HardDeleteUserWithImages(userID)
}

func (s *UserService) SoftDeleteUser(userID uint, timestamp int64) error {
	return s.userStore.SoftDeleteUser(userID, timestamp)
}
