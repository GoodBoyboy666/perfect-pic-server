package service

import (
	"testing"

	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
)

// 测试内容：验证管理员统计数据查询的正确性。
func TestGetServerStatsForAdmin(t *testing.T) {
	setupTestDB(t)

	u := model.User{Username: "alice", Password: "x", Status: 1, Email: "a@example.com"}
	_ = db.DB.Create(&u).Error
	_ = db.DB.Create(&model.Image{
		Filename:   "a.png",
		Path:       "2026/02/13/a.png",
		Size:       10,
		Width:      1,
		Height:     1,
		MimeType:   ".png",
		UploadedAt: 1,
		UserID:     u.ID,
	}).Error

	stats, err := AdminGetServerStats()
	if err != nil {
		t.Fatalf("AdminGetServerStats: %v", err)
	}
	if stats.ImageCount != 1 || stats.StorageUsage != 10 || stats.UserCount != 1 {
		t.Fatalf("非预期 stats: %+v", stats)
	}
}
