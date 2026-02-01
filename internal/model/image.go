package model

type Image struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	Filename   string `json:"filename" gorm:"not null;unique"`
	Path       string `json:"path" gorm:"not null;unique"`
	Size       int64  `json:"size" gorm:"not null"`
	Width      int    `json:"width" gorm:"not null"`
	Height     int    `json:"height" gorm:"not null"`
	MimeType   string `json:"mime_type" gorm:"not null"`
	UploadedAt int64  `json:"uploaded_at" gorm:"not null;index"`
	UserID     uint   `json:"user_id" gorm:"not null;index"`
	User       User   `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE;" json:"-"`
}
