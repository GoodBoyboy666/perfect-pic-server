package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	commonpkg "perfect-pic-server/internal/common"
	"perfect-pic-server/internal/consts"
	moduledto "perfect-pic-server/internal/dto"
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/pkg/validator"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const userStatusCacheTTL = 1 * time.Minute

// GenerateForgetPasswordToken 生成忘记密码 Token，有效期 15 分钟
func (s *UserService) GenerateForgetPasswordToken(userID uint) (string, error) {
	// 使用 crypto/rand 生成 32 字节的高熵随机字符串 (64字符Hex)
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	ttl := 15 * time.Minute
	tokenKey := s.cache.RedisKey("password_reset", "token", token)
	userKey := s.cache.RedisKey("password_reset", "user", strconv.FormatUint(uint64(userID), 10))
	uidStr := strconv.FormatUint(uint64(userID), 10)
	s.cache.SetIndexed(userKey, tokenKey, uidStr, ttl)
	return token, nil
}

// VerifyForgetPasswordToken 验证忘记密码 Token
func (s *UserService) VerifyForgetPasswordToken(token string) (uint, bool) {
	tokenKey := s.cache.RedisKey("password_reset", "token", token)
	uidStr, ok := s.cache.Get(tokenKey)
	if !ok {
		return 0, false
	}
	uid, parseErr := strconv.ParseUint(uidStr, 10, 64)
	if parseErr != nil || uid > math.MaxUint {
		s.cache.Delete(tokenKey)
		return 0, false
	}
	userKey := s.cache.RedisKey("password_reset", "user", strconv.FormatUint(uid, 10))
	if !s.cache.CompareAndDeletePair(userKey, tokenKey, tokenKey, uidStr) {
		s.cache.Delete(tokenKey)
		return 0, false
	}
	return uint(uid), true
}

// GenerateEmailVerificationToken 生成邮箱验证 Token，有效期 30 分钟。
func (s *UserService) GenerateEmailVerificationToken(userID uint, email string) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	ttl := 30 * time.Minute
	payload, err := json.Marshal(moduledto.EmailVerifyRedisPayload{
		UserID: userID,
		Email:  email,
	})
	if err != nil {
		return "", err
	}

	tokenKey := s.cache.RedisKey("email_verify", "token", token)
	userKey := s.cache.RedisKey("email_verify", "user", strconv.FormatUint(uint64(userID), 10))
	s.cache.SetIndexed(userKey, tokenKey, string(payload), ttl)
	return token, nil
}

// VerifyEmailVerificationToken 验证并消费邮箱验证 Token。
func (s *UserService) VerifyEmailVerificationToken(token string) (uint, string, bool) {
	if token == "" {
		return 0, "", false
	}

	tokenKey := s.cache.RedisKey("email_verify", "token", token)
	raw, ok := s.cache.Get(tokenKey)
	if !ok {
		return 0, "", false
	}

	var payload moduledto.EmailVerifyRedisPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil || payload.UserID == 0 || payload.Email == "" {
		s.cache.Delete(tokenKey)
		return 0, "", false
	}

	userKey := s.cache.RedisKey("email_verify", "user", strconv.FormatUint(uint64(payload.UserID), 10))
	if !s.cache.CompareAndDeletePair(userKey, tokenKey, tokenKey, raw) {
		s.cache.Delete(tokenKey)
		return 0, "", false
	}

	return payload.UserID, payload.Email, true
}

// GenerateEmailChangeToken 生成修改邮箱 Token，有效期 30 分钟。
func (s *UserService) GenerateEmailChangeToken(userID uint, oldEmail, newEmail string) (string, error) {
	// 使用 crypto/rand 生成 32 字节的高熵随机字符串 (64字符Hex)
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	ttl := 30 * time.Minute
	payload, err := json.Marshal(moduledto.EmailChangeRedisPayload{
		UserID:   userID,
		OldEmail: oldEmail,
		NewEmail: newEmail,
	})
	if err != nil {
		return "", err
	}
	tokenKey := s.cache.RedisKey("email_change", "token", token)
	userKey := s.cache.RedisKey("email_change", "user", strconv.FormatUint(uint64(userID), 10))
	s.cache.SetIndexed(userKey, tokenKey, string(payload), ttl)
	return token, nil
}

// VerifyEmailChangeToken 验证并消费修改邮箱 Token。
//
//nolint:gocyclo
func (s *UserService) VerifyEmailChangeToken(token string) (*moduledto.EmailChangeToken, bool) {
	if token == "" {
		return nil, false
	}

	tokenKey := s.cache.RedisKey("email_change", "token", token)
	raw, ok := s.cache.Get(tokenKey)
	if !ok {
		return nil, false
	}

	var payload moduledto.EmailChangeRedisPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil || payload.UserID == 0 {
		s.cache.Delete(tokenKey)
		return nil, false
	}

	userKey := s.cache.RedisKey("email_change", "user", strconv.FormatUint(uint64(payload.UserID), 10))
	if !s.cache.CompareAndDeletePair(userKey, tokenKey, tokenKey, raw) {
		s.cache.Delete(tokenKey)
		return nil, false
	}
	return &moduledto.EmailChangeToken{
		UserID:   payload.UserID,
		Token:    token,
		OldEmail: payload.OldEmail,
		NewEmail: payload.NewEmail,
	}, true
}

// GetUserStatus 获取用户状态，优先从缓存读取，未命中时回源数据库并回写缓存。
func (s *UserService) GetUserStatus(userID uint) (int, error) {
	statusKey := ""
	if s.cache != nil {
		statusKey = s.cache.RedisKey("auth", "user_status", strconv.FormatUint(uint64(userID), 10))
		if cachedStatus, ok := s.cache.Get(statusKey); ok {
			if parsedStatus, err := strconv.Atoi(cachedStatus); err == nil {
				return parsedStatus, nil
			}
			s.cache.Delete(statusKey)
		}
	}

	user, err := s.userStore.FindByID(userID)
	if err != nil {
		return 0, err
	}

	if s.cache != nil {
		s.cache.Set(statusKey, strconv.Itoa(user.Status), userStatusCacheTTL)
	}

	return user.Status, nil
}

// ClearUserStatusCache 清除指定用户的状态缓存。
func (s *UserService) ClearUserStatusCache(userID uint) {
	if s.cache == nil {
		return
	}
	s.cache.Delete(s.cache.RedisKey("auth", "user_status", strconv.FormatUint(uint64(userID), 10)))
}

// GetSystemDefaultStorageQuota 获取系统默认存储配额
func (s *UserService) GetSystemDefaultStorageQuota() int64 {
	return s.dbConfig.GetDefaultStorageQuota()
}

// IsUsernameTaken 检查用户名是否已被占用。
// excludeUserID 用于更新场景下排除当前用户；includeDeleted 为 true 时会包含软删除用户。
func (s *UserService) IsUsernameTaken(username string, excludeUserID *uint, includeDeleted bool) (bool, error) {
	return s.userStore.FieldExists(consts.UserFieldUsername, username, excludeUserID, includeDeleted)
}

// IsEmailTaken 检查邮箱是否已被占用。
// excludeUserID 用于更新场景下排除当前用户；includeDeleted 为 true 时会包含软删除用户。
func (s *UserService) IsEmailTaken(email string, excludeUserID *uint, includeDeleted bool) (bool, error) {
	return s.userStore.FieldExists(consts.UserFieldEmail, email, excludeUserID, includeDeleted)
}

// GetUserByID 按用户 ID 获取用户模型。
func (s *UserService) GetUserByID(userID uint) (*model.User, error) {
	user, err := s.userStore.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, commonpkg.NewNotFoundError("用户不存在")
		}
		return nil, commonpkg.NewInternalError("获取用户信息失败")
	}
	return user, nil
}

// GetUserProfile 获取用户个人资料。
func (s *UserService) GetUserProfile(userID uint) (*moduledto.UserProfileResponse, error) {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, commonpkg.NewNotFoundError("用户不存在")
	}

	return &moduledto.UserProfileResponse{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email,
		Avatar:       user.Avatar,
		Admin:        user.Admin,
		StorageQuota: user.StorageQuota,
		StorageUsed:  user.StorageUsed,
	}, nil
}

// UpdateUsernameAndGenerateToken 更新用户名并签发新登录令牌。
func (s *UserService) UpdateUsernameAndGenerateToken(userID uint, newUsername string, isAdmin bool) (string, error) {
	// Profile 路径统一禁止保留用户名；管理员后台修改用户名走 AdminPrepareUserUpdates（允许保留词）。
	if ok, msg := validator.ValidateUsername(newUsername); !ok {
		return "", commonpkg.NewValidationError(msg)
	}

	excludeID := userID
	usernameTaken, err := s.IsUsernameTaken(newUsername, &excludeID, true)
	if err != nil {
		return "", commonpkg.NewInternalError("更新失败")
	}
	if usernameTaken {
		return "", commonpkg.NewConflictError("用户名已存在")
	}

	if err := s.userStore.UpdateUsernameByID(userID, newUsername); err != nil {
		return "", commonpkg.NewInternalError("更新失败")
	}

	token, err := s.jwt.GenerateLoginToken(userID, newUsername, isAdmin)
	if err != nil {
		return "", commonpkg.NewInternalError("更新失败")
	}

	return token, nil
}

// UpdatePasswordByOldPassword 使用旧密码校验后更新新密码。
func (s *UserService) UpdatePasswordByOldPassword(userID uint, oldPassword, newPassword string) error {
	if ok, msg := validator.ValidatePassword(newPassword); !ok {
		return commonpkg.NewValidationError(msg)
	}

	user, err := s.userStore.FindByID(userID)
	if err != nil {
		return commonpkg.NewNotFoundError("用户不存在")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return commonpkg.NewValidationError("旧密码错误")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return commonpkg.NewInternalError("更新失败")
	}

	if err := s.userStore.UpdatePasswordByID(userID, string(hashedPassword)); err != nil {
		return commonpkg.NewInternalError("更新失败")
	}

	return nil
}
