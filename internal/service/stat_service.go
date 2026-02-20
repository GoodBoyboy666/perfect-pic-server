package service

import (
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
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
	var imageCount int64
	var totalSize int64
	var userCount int64

	if err := db.DB.Model(&model.Image{}).Count(&imageCount).Error; err != nil {
		return nil, err
	}

	if err := db.DB.Model(&model.Image{}).Select("COALESCE(SUM(size), 0)").Scan(&totalSize).Error; err != nil {
		return nil, err
	}

	if err := db.DB.Model(&model.User{}).Count(&userCount).Error; err != nil {
		return nil, err
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
