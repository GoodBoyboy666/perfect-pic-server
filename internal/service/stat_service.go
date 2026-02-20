package service

import (
	"perfect-pic-server/internal/repository"
	"runtime"
)

type SystemInfo struct {
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	GoVersion    string `json:"go_version"`
	NumCPU       int    `json:"num_cpu"`
	NumGoroutine int    `json:"num_goroutine"`
}

type ServerStats struct {
	ImageCount   int64      `json:"image_count"`
	StorageUsage int64      `json:"storage_usage"`
	UserCount    int64      `json:"user_count"`
	SystemInfo   SystemInfo `json:"system_info"`
}

// AdminGetServerStats 获取后台仪表盘统计数据。
func AdminGetServerStats() (*ServerStats, error) {
	imageCount, err := repository.Image.CountAll()
	if err != nil {
		return nil, NewInternalError("统计图片数据失败")
	}

	totalSize, err := repository.Image.SumAllSize()
	if err != nil {
		return nil, NewInternalError("统计图片数据失败")
	}

	userCount, err := repository.User.CountAll()
	if err != nil {
		return nil, NewInternalError("统计用户数据失败")
	}

	return &ServerStats{
		ImageCount:   imageCount,
		StorageUsage: totalSize,
		UserCount:    userCount,
		SystemInfo: SystemInfo{
			OS:           runtime.GOOS,
			Arch:         runtime.GOARCH,
			GoVersion:    runtime.Version(),
			NumCPU:       runtime.NumCPU(),
			NumGoroutine: runtime.NumGoroutine(),
		},
	}, nil
}
