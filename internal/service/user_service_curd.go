package service

import (
	"perfect-pic-server/internal/model"
)

func (s *UserService) SaveUser(user *model.User) error {
	if err := s.userStore.Save(user); err != nil {
		return err
	}
	s.ClearUserAuthCache(user.ID)
	s.ClearUserStatusCache(user.ID)
	return nil
}

func (s *UserService) UpdateAvatar(user *model.User, newAvatar string) error {
	if err := s.userStore.UpdateAvatar(user, newAvatar); err != nil {
		return err
	}
	return nil
}

func (s *UserService) ClearAvatar(user *model.User) error {
	if err := s.userStore.ClearAvatar(user); err != nil {
		return err
	}
	return nil
}

func (s *UserService) HardDeleteUserWithImages(userID uint) error {
	if err := s.userStore.HardDeleteUserWithImages(userID); err != nil {
		return err
	}
	s.ClearUserAuthCache(userID)
	s.ClearUserStatusCache(userID)
	return nil
}

func (s *UserService) SoftDeleteUser(userID uint, timestamp int64) error {
	if err := s.userStore.SoftDeleteUser(userID, timestamp); err != nil {
		return err
	}
	s.ClearUserAuthCache(userID)
	s.ClearUserStatusCache(userID)
	return nil
}
