package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/repository"
	"perfect-pic-server/internal/utils"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
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

type emailChangeRedisPayload struct {
	UserID   uint   `json:"user_id"`
	OldEmail string `json:"old_email"`
	NewEmail string `json:"new_email"`
}

type UserProfile struct {
	ID           uint   `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	Avatar       string `json:"avatar"`
	Admin        bool   `json:"admin"`
	StorageQuota *int64 `json:"storage_quota"`
	StorageUsed  int64  `json:"storage_used"`
}

var (
	// passwordResetStore 存储忘记密码 Token
	// Key: UserID (uint), Value: Token (string)
	passwordResetStore sync.Map
	// passwordResetTokenStore 存储忘记密码 Token 索引
	// Key: Token (string), Value: ForgetPasswordToken
	passwordResetTokenStore sync.Map

	// emailChangeStore 存储修改邮箱 Token
	// Key: UserID (uint), Value: Token (string)
	emailChangeStore sync.Map
	// emailChangeTokenStore 存储修改邮箱 Token 索引
	// Key: Token (string), Value: EmailChangeToken
	emailChangeTokenStore sync.Map
)

var errRedisTokenCASMismatch = errors.New("redis token cas mismatch")

// GenerateForgetPasswordToken 生成忘记密码 Token，有效期 15 分钟
func GenerateForgetPasswordToken(userID uint) (string, error) {
	// 使用 crypto/rand 生成 32 字节的高熵随机字符串 (64字符Hex)
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	resetToken := ForgetPasswordToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	if redisClient := GetRedisClient(); redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		// 保证一个用户只有一个有效 token
		userKey := RedisKey("password_reset", "user", strconv.FormatUint(uint64(userID), 10))
		if oldToken, err := redisClient.Get(ctx, userKey).Result(); err == nil && oldToken != "" {
			oldTokenKey := RedisKey("password_reset", "token", oldToken)
			_ = redisClient.Del(ctx, oldTokenKey).Err()
		}

		tokenKey := RedisKey("password_reset", "token", token)
		if err := redisClient.Set(ctx, tokenKey, strconv.FormatUint(uint64(userID), 10), 15*time.Minute).Err(); err == nil {
			if err := redisClient.Set(ctx, userKey, token, 15*time.Minute).Err(); err == nil {
				return token, nil
			}
			// 避免出现 tokenKey 已写入但 userKey 缺失的不一致状态。
			_ = redisClient.Del(ctx, tokenKey).Err()
		}
	}

	// 存储（覆盖之前的）
	if prev, ok := passwordResetStore.Load(userID); ok {
		if prevToken, ok2 := prev.(string); ok2 && prevToken != "" {
			passwordResetTokenStore.Delete(prevToken)
		}
	}
	passwordResetStore.Store(userID, token)
	passwordResetTokenStore.Store(token, resetToken)
	return token, nil
}

// VerifyForgetPasswordToken 验证忘记密码 Token
func VerifyForgetPasswordToken(token string) (uint, bool) {
	if redisClient := GetRedisClient(); redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		tokenKey := RedisKey("password_reset", "token", token)
		uidStr, err := redisClient.Get(ctx, tokenKey).Result()
		if err == nil {
			uid, parseErr := strconv.ParseUint(uidStr, 10, 64)
			if parseErr == nil {
				// Ensure the parsed UID fits into the platform-dependent uint type.
				if uid > math.MaxUint {
					_ = redisClient.Del(ctx, tokenKey).Err()
					return 0, false
				}
				userKey := RedisKey("password_reset", "user", strconv.FormatUint(uid, 10))
				casErr := verifyAndConsumeRedisTokenPair(ctx, redisClient, tokenKey, userKey, token, uidStr)
				if casErr == nil {
					return uint(uid), true
				}

				// 比对失败或并发竞争时，仅清理当前 tokenKey，避免误删新 token 对应的 userKey。
				if errors.Is(casErr, errRedisTokenCASMismatch) {
					_ = redisClient.Del(ctx, tokenKey).Err()
					return 0, false
				}

				return 0, false
			}
			_ = redisClient.Del(ctx, tokenKey).Err()
			return 0, false
		}
	}

	// LoadAndDelete 保证并发下同一 token 只会被成功消费一次。
	val, ok := passwordResetTokenStore.LoadAndDelete(token)
	if !ok {
		return 0, false
	}

	resetToken, ok := val.(ForgetPasswordToken)
	if !ok {
		return 0, false
	}

	// 仅当 user->token 映射仍指向当前 token 时再删除，避免误删更新后的新 token 映射。
	if current, ok := passwordResetStore.Load(resetToken.UserID); ok {
		if currentToken, ok2 := current.(string); ok2 && currentToken == token {
			passwordResetStore.Delete(resetToken.UserID)
		}
	}

	if time.Now().After(resetToken.ExpiresAt) {
		return 0, false
	}
	return resetToken.UserID, true
}

// GenerateEmailChangeToken 生成修改邮箱 Token，有效期 30 分钟。
func GenerateEmailChangeToken(userID uint, oldEmail, newEmail string) (string, error) {
	// 使用 crypto/rand 生成 32 字节的高熵随机字符串 (64字符Hex)
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	changeToken := EmailChangeToken{
		UserID:    userID,
		Token:     token,
		OldEmail:  oldEmail,
		NewEmail:  newEmail,
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}

	if redisClient := GetRedisClient(); redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		// 保证一个用户只有一个有效 token
		userKey := RedisKey("email_change", "user", strconv.FormatUint(uint64(userID), 10))
		if oldToken, err := redisClient.Get(ctx, userKey).Result(); err == nil && oldToken != "" {
			oldTokenKey := RedisKey("email_change", "token", oldToken)
			_ = redisClient.Del(ctx, oldTokenKey).Err()
		}

		payload, err := json.Marshal(emailChangeRedisPayload{
			UserID:   userID,
			OldEmail: oldEmail,
			NewEmail: newEmail,
		})
		if err != nil {
			return "", err
		}

		tokenKey := RedisKey("email_change", "token", token)
		if err := redisClient.Set(ctx, tokenKey, payload, 30*time.Minute).Err(); err == nil {
			if err := redisClient.Set(ctx, userKey, token, 30*time.Minute).Err(); err == nil {
				return token, nil
			}
			// 避免出现 tokenKey 已写入但 userKey 缺失的不一致状态。
			_ = redisClient.Del(ctx, tokenKey).Err()
		}
	}

	// 存储（覆盖之前的）
	if prev, ok := emailChangeStore.Load(userID); ok {
		if prevToken, ok2 := prev.(string); ok2 && prevToken != "" {
			emailChangeTokenStore.Delete(prevToken)
		}
	}
	emailChangeStore.Store(userID, token)
	emailChangeTokenStore.Store(token, changeToken)
	return token, nil
}

// VerifyEmailChangeToken 验证并消费修改邮箱 Token。
func VerifyEmailChangeToken(token string) (*EmailChangeToken, bool) {
	if token == "" {
		return nil, false
	}

	if redisClient := GetRedisClient(); redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		tokenKey := RedisKey("email_change", "token", token)
		raw, err := redisClient.Get(ctx, tokenKey).Result()
		if err == nil && raw != "" {
			var payload emailChangeRedisPayload
			if err := json.Unmarshal([]byte(raw), &payload); err != nil || payload.UserID == 0 {
				_ = redisClient.Del(ctx, tokenKey).Err()
				return nil, false
			}

			userKey := RedisKey("email_change", "user", strconv.FormatUint(uint64(payload.UserID), 10))
			casErr := verifyAndConsumeRedisTokenPair(ctx, redisClient, tokenKey, userKey, token, raw)
			if casErr == nil {
				return &EmailChangeToken{
					UserID:   payload.UserID,
					OldEmail: payload.OldEmail,
					NewEmail: payload.NewEmail,
				}, true
			}

			// 比对失败或并发竞争时，仅清理当前 tokenKey，避免误删新 token 对应的 userKey。
			if errors.Is(casErr, errRedisTokenCASMismatch) {
				_ = redisClient.Del(ctx, tokenKey).Err()
				return nil, false
			}
			return nil, false
		}
	}

	// LoadAndDelete 保证并发下同一 token 只会被成功消费一次。
	val, ok := emailChangeTokenStore.LoadAndDelete(token)
	if !ok {
		return nil, false
	}

	changeToken, ok := val.(EmailChangeToken)
	if !ok {
		return nil, false
	}

	// 仅当 user->token 映射仍指向当前 token 时再删除，避免误删更新后的新 token 映射。
	if current, ok := emailChangeStore.Load(changeToken.UserID); ok {
		if currentToken, ok2 := current.(string); ok2 && currentToken == token {
			emailChangeStore.Delete(changeToken.UserID)
		}
	}

	if time.Now().After(changeToken.ExpiresAt) {
		return nil, false
	}
	return &changeToken, true
}

// GetSystemDefaultStorageQuota 获取系统默认存储配额
func GetSystemDefaultStorageQuota() int64 {
	quota := GetInt64(consts.ConfigDefaultStorageQuota)
	if quota == 0 {
		return 1073741824 // 兜底 1GB
	}
	return quota
}

// DeleteUserFiles 删除指定用户的所有关联文件（头像、上传的照片）
// 此函数只负责删除物理文件，不处理数据库记录的清理
func DeleteUserFiles(userID uint) error {
	cfg := config.Get()

	// 1. 删除头像目录
	// 头像存储结构: data/avatars/{userID}/filename
	avatarRoot := cfg.Upload.AvatarPath
	if avatarRoot == "" {
		avatarRoot = "uploads/avatars"
	}
	avatarRootAbs, err := filepath.Abs(avatarRoot)
	if err != nil {
		return fmt.Errorf("failed to resolve avatar root: %w", err)
	}
	// 先校验头像根目录节点本身，避免根目录直接是符号链接。
	if err := utils.EnsurePathNotSymlink(avatarRootAbs); err != nil {
		return fmt.Errorf("avatar root symlink risk: %w", err)
	}

	userAvatarDir, err := utils.SecureJoin(avatarRootAbs, fmt.Sprintf("%d", userID))
	if err != nil {
		return fmt.Errorf("failed to build avatar dir: %w", err)
	}
	// 在执行 RemoveAll 前再做一次链路检查，确保目标目录链路未被并发替换为符号链接。
	if err := utils.EnsureNoSymlinkBetween(avatarRootAbs, userAvatarDir); err != nil {
		return fmt.Errorf("avatar dir symlink risk: %w", err)
	}

	// RemoveAll 删除路径及其包含的任何子项。如果路径不存在，RemoveAll 返回 nil（无错误）。
	if err := os.RemoveAll(userAvatarDir); err != nil {
		// 记录日志或打印错误，但不中断后续操作
		log.Printf("Warning: Failed to delete avatar directory for user %d: %v\n", userID, err)
	}

	// 2. 查找并删除用户上传的所有图片
	// Unscoped() 确保即使是软删除的图片也能被查出来删除文件
	images, err := repository.Image.FindUnscopedByUserID(userID)
	if err != nil {
		return fmt.Errorf("failed to retrieve user images: %w", err)
	}

	uploadRoot := cfg.Upload.Path
	if uploadRoot == "" {
		uploadRoot = "uploads/imgs"
	}
	uploadRootAbs, err := filepath.Abs(uploadRoot)
	if err != nil {
		return fmt.Errorf("failed to resolve upload root: %w", err)
	}
	// 先校验上传根目录节点本身，避免根目录直接是符号链接。
	if err := utils.EnsurePathNotSymlink(uploadRootAbs); err != nil {
		return fmt.Errorf("upload root symlink risk: %w", err)
	}

	for _, img := range images {
		// 转换路径分隔符以适配当前系统 (DB中存储的是 web 格式 '/')
		localPath := filepath.FromSlash(img.Path)
		fullPath, secureErr := utils.SecureJoin(uploadRootAbs, localPath)
		if secureErr != nil {
			log.Printf("Warning: Skip unsafe image path for user %d (%s): %v\n", userID, img.Path, secureErr)
			continue
		}

		if err := os.Remove(fullPath); err != nil {
			if !os.IsNotExist(err) {
				log.Printf("Warning: Failed to delete image file %s: %v\n", fullPath, err)
			}
		}
	}

	return nil
}

// IsUsernameTaken 检查用户名是否已被占用。
// excludeUserID 用于更新场景下排除当前用户；includeDeleted 为 true 时会包含软删除用户。
func IsUsernameTaken(username string, excludeUserID *uint, includeDeleted bool) (bool, error) {
	return repository.User.FieldExists(repository.UserFieldUsername, username, excludeUserID, includeDeleted)
}

// IsEmailTaken 检查邮箱是否已被占用。
// excludeUserID 用于更新场景下排除当前用户；includeDeleted 为 true 时会包含软删除用户。
func IsEmailTaken(email string, excludeUserID *uint, includeDeleted bool) (bool, error) {
	return repository.User.FieldExists(repository.UserFieldEmail, email, excludeUserID, includeDeleted)
}

// GetUserByID 按用户 ID 获取用户模型。
func GetUserByID(userID uint) (*model.User, error) {
	return repository.User.FindByID(userID)
}

// GetUserProfile 获取用户个人资料。
func GetUserProfile(userID uint) (*UserProfile, error) {
	user, err := GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	return &UserProfile{
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
func UpdateUsernameAndGenerateToken(userID uint, newUsername string, isAdmin bool) (string, string, error) {
	if ok, msg := utils.ValidateUsername(newUsername); !ok {
		return "", msg, nil
	}

	excludeID := userID
	usernameTaken, err := IsUsernameTaken(newUsername, &excludeID, true)
	if err != nil {
		return "", "", err
	}
	if usernameTaken {
		return "", "用户名已存在", nil
	}

	if err := repository.User.UpdateUsernameByID(userID, newUsername); err != nil {
		return "", "", err
	}

	cfg := config.Get()
	token, err := utils.GenerateLoginToken(userID, newUsername, isAdmin, time.Hour*time.Duration(cfg.JWT.ExpirationHours))
	if err != nil {
		return "", "", err
	}

	return token, "", nil
}

// UpdatePasswordByOldPassword 使用旧密码校验后更新新密码。
func UpdatePasswordByOldPassword(userID uint, oldPassword, newPassword string) (string, error) {
	if ok, msg := utils.ValidatePassword(newPassword); !ok {
		return msg, nil
	}

	user, err := repository.User.FindByID(userID)
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return "旧密码错误", nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	if err := repository.User.UpdatePasswordByID(userID, string(hashedPassword)); err != nil {
		return "", err
	}

	return "", nil
}

// RequestEmailChange 发起邮箱修改流程并异步发送验证邮件。
func RequestEmailChange(userID uint, password, newEmail string) (string, error) {
	if ok, msg := utils.ValidateEmail(newEmail); !ok {
		return msg, nil
	}

	user, err := repository.User.FindByID(userID)
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "密码错误", nil
	}

	if user.Email == newEmail {
		return "新邮箱不能与当前邮箱相同", nil
	}

	emailTaken, err := IsEmailTaken(newEmail, nil, true)
	if err != nil {
		return "", err
	}
	if emailTaken {
		return "该邮箱已被使用", nil
	}

	token, err := GenerateEmailChangeToken(user.ID, user.Email, newEmail)
	if err != nil {
		return "", err
	}

	baseURL := GetString(consts.ConfigBaseURL)
	if baseURL == "" {
		baseURL = "http://localhost"
	}
	if len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}
	verifyURL := fmt.Sprintf("%s/auth/email-change-verify?token=%s", baseURL, token)

	if shouldSendEmail() {
		go func() {
			_ = SendEmailChangeVerification(newEmail, user.Username, user.Email, newEmail, verifyURL)
		}()
	}

	return "", nil
}
