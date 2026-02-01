package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint `json:"id" gorm:"primaryKey"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	Username     string         `json:"username" gorm:"unique;not null"`
	Password     string         `json:"-" gorm:"not null"`
	Admin        bool           `json:"admin" gorm:"not null"`
	Status       int            `json:"status" gorm:"default:1"` // 1: 正常, 2: 封禁, 3: 软删除(停用)
	Avatar       string         `json:"avatar"`
	StorageQuota *int64         `json:"storage_quota"`
	StorageUsed  int64          `json:"storage_used" gorm:"default:0"` // 已用存储空间 (Bytes)
	Photos       []Image        `json:"-"`
}
