package dto

import (
	"encoding/json"
	"time"
)

type ForgetPasswordToken struct {
	UserID    uint
	Token     string
	ExpiresAt time.Time
}

type EmailChangeToken struct {
	UserID    uint
	Token     string
	OldEmail  string
	NewEmail  string
	ExpiresAt time.Time
}

type EmailChangeRedisPayload struct {
	UserID   uint   `json:"user_id"`
	OldEmail string `json:"old_email"`
	NewEmail string `json:"new_email"`
}

type UserProfileResponse struct {
	ID           uint   `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	Avatar       string `json:"avatar"`
	Admin        bool   `json:"admin"`
	StorageQuota *int64 `json:"storage_quota"`
	StorageUsed  int64  `json:"storage_used"`
}

type AdminUserListRequest struct {
	Page        int
	PageSize    int
	Keyword     string
	ShowDeleted bool
	Order       string
}

type AdminUserUpdateRequest struct {
	Username      *string
	Password      *string
	Email         *string
	EmailVerified *bool
	StorageQuota  *int64
	Status        *int
}

type AdminCreateUserRequest struct {
	Username      string
	Password      string
	Email         *string
	EmailVerified *bool
	StorageQuota  *int64
	Status        *int
}

type CreateUserRequest struct {
	Username      string  `json:"username" binding:"required"`
	Password      string  `json:"password" binding:"required"`
	Email         *string `json:"email"`
	EmailVerified *bool   `json:"email_verified"`
	StorageQuota  *int64  `json:"storage_quota"`
	Status        *int    `json:"status"`
}

type UpdateUserRequest struct {
	Username      *string `json:"username"`
	Password      *string `json:"password"`
	Email         *string `json:"email"`
	EmailVerified *bool   `json:"email_verified"`
	StorageQuota  *int64  `json:"storage_quota"`
	Status        *int    `json:"status"`
}

type UpdateSelfUsernameRequest struct {
	Username string `json:"username" binding:"required"`
}

type UpdateSelfPasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type RequestUpdateEmailRequest struct {
	Password string `json:"password" binding:"required"`
	NewEmail string `json:"new_email" binding:"required"`
}

type FinishPasskeyRegistrationRequest struct {
	SessionID  string          `json:"session_id" binding:"required"`
	Credential json.RawMessage `json:"credential" binding:"required"`
}

type UpdatePasskeyNameRequest struct {
	Name string `json:"name" binding:"required"`
}
