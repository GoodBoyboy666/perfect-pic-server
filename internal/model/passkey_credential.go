package model

import "time"

type PasskeyCredential struct {
	ID           uint `json:"id" gorm:"primaryKey"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	UserID       uint   `json:"user_id" gorm:"not null;index"`
	CredentialID string `json:"credential_id" gorm:"not null;uniqueIndex;size:255"`
	Credential   string `json:"-" gorm:"type:text;not null"`
	User         User   `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE;" json:"-"`
}
