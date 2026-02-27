package service

func (s *Service) GetSystemDefaultStorageQuota() int64 {
	return s.GetDefaultStorageQuota()
}
