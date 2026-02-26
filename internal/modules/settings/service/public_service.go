package service

import "perfect-pic-server/internal/consts"

func (s *Service) GetSystemDefaultStorageQuota() int64 {
	quota := s.GetInt64(consts.ConfigDefaultStorageQuota)
	if quota <= 0 {
		return 1073741824
	}
	return quota
}
