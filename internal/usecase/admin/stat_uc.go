package admin

import (
	platformservice "perfect-pic-server/internal/common"
	moduledto "perfect-pic-server/internal/dto"
	"runtime"
)

// AdminGetServerStats 获取后台仪表盘统计数据。
func (c *StatUseCase) AdminGetServerStats() (*moduledto.ServerStatsResponse, error) {
	imageCount, err := c.imageStore.CountAll()
	if err != nil {
		return nil, platformservice.NewInternalError("统计图片数据失败")
	}

	totalSize, err := c.imageStore.SumAllSize()
	if err != nil {
		return nil, platformservice.NewInternalError("统计图片数据失败")
	}

	userCount, err := c.userStore.CountAll()
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
