package service

import (
	moduledto "perfect-pic-server/internal/modules/system/dto"
	platformservice "perfect-pic-server/internal/platform/service"
	"runtime"
)

// AdminGetServerStats 获取后台仪表盘统计数据。
func (s *Service) AdminGetServerStats() (*moduledto.ServerStatsResponse, error) {
	imageCount, err := s.imageService.CountAll()
	if err != nil {
		return nil, platformservice.NewInternalError("统计图片数据失败")
	}

	totalSize, err := s.imageService.SumAllSize()
	if err != nil {
		return nil, platformservice.NewInternalError("统计图片数据失败")
	}

	userCount, err := s.userService.CountAll()
	if err != nil {
		return nil, platformservice.NewInternalError("统计用户数据失败")
	}

	return &moduledto.ServerStatsResponse{
		ImageCount:   imageCount,
		StorageUsage: totalSize,
		UserCount:    userCount,
		SystemInfo: moduledto.SystemInfoResponse{
			OS:           runtime.GOOS,
			Arch:         runtime.GOARCH,
			GoVersion:    runtime.Version(),
			NumCPU:       runtime.NumCPU(),
			NumGoroutine: runtime.NumGoroutine(),
		},
	}, nil
}
