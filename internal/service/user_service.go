package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"math"
	commonpkg "perfect-pic-server/internal/common"
	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/consts"
	moduledto "perfect-pic-server/internal/dto"
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/utils"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	// passwordResetStore 存储忘记密码 Token
	// Key: UserID (uint), Value: Token (string)
	passwordResetStore sync.Map
	// passwordResetTokenStore 存储忘记密码 Token 索引
	// Key: Token (string), Value: moduledto.ForgetPasswordToken
	passwordResetTokenStore sync.Map

	// emailChangeStore 存储修改邮箱 Token
	// Key: UserID (uint), Value: Token (string)
	emailChangeStore sync.Map
	// emailChangeTokenStore 存储修改邮箱 Token 索引
	// Key: Token (string), Value: moduledto.EmailChangeToken
	emailChangeTokenStore sync.Map
)

var errRedisTokenCASMismatch = errors.New("redis token cas mismatch")

// GenerateForgetPasswordToken 生成忘记密码 Token，有效期 15 分钟
func (s *UserService) GenerateForgetPasswordToken(userID uint) (string, error) {
	// 使用 crypto/rand 生成 32 字节的高熵随机字符串 (64字符Hex)
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	resetToken := moduledto.ForgetPasswordToken{
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
			if err2 := redisClient.Set(ctx, userKey, token, 15*time.Minute).Err(); err2 == nil {
				return token, nil
			} else {
				log.Printf("⚠️ Redis 写入密码重置用户索引失败，回退内存 token 存储: %v", err2)
			}
			// 避免出现 tokenKey 已写入但 userKey 缺失的不一致状态。
			_ = redisClient.Del(ctx, tokenKey).Err()
		} else {
			log.Printf("⚠️ Redis 写入密码重置 token 失败，回退内存 token 存储: %v", err)
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
func (s *UserService) VerifyForgetPasswordToken(token string) (uint, bool) {
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
		if !errors.Is(err, redis.Nil) {
			log.Printf("⚠️ Redis 读取密码重置 token 失败，回退内存 token 存储: %v", err)
		}
	}

	// LoadAndDelete 保证并发下同一 token 只会被成功消费一次。
	val, ok := passwordResetTokenStore.LoadAndDelete(token)
	if !ok {
		return 0, false
	}

	resetToken, ok := val.(moduledto.ForgetPasswordToken)
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
func (s *UserService) GenerateEmailChangeToken(userID uint, oldEmail, newEmail string) (string, error) {
	// 使用 crypto/rand 生成 32 字节的高熵随机字符串 (64字符Hex)
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	changeToken := moduledto.EmailChangeToken{
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

		payload, err := json.Marshal(moduledto.EmailChangeRedisPayload{
			UserID:   userID,
			OldEmail: oldEmail,
			NewEmail: newEmail,
		})
		if err != nil {
			return "", err
		}

		tokenKey := RedisKey("email_change", "token", token)
		if err := redisClient.Set(ctx, tokenKey, payload, 30*time.Minute).Err(); err == nil {
			if err2 := redisClient.Set(ctx, userKey, token, 30*time.Minute).Err(); err2 == nil {
				return token, nil
			} else {
				log.Printf("⚠️ Redis 写入邮箱修改用户索引失败，回退内存 token 存储: %v", err2)
			}
			// 避免出现 tokenKey 已写入但 userKey 缺失的不一致状态。
			_ = redisClient.Del(ctx, tokenKey).Err()
		} else {
			log.Printf("⚠️ Redis 写入邮箱修改 token 失败，回退内存 token 存储: %v", err)
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
//
//nolint:gocyclo
func (s *UserService) VerifyEmailChangeToken(token string) (*moduledto.EmailChangeToken, bool) {
	if token == "" {
		return nil, false
	}

	if redisClient := GetRedisClient(); redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		tokenKey := RedisKey("email_change", "token", token)
		raw, err := redisClient.Get(ctx, tokenKey).Result()
		if err == nil && raw != "" {
			var payload moduledto.EmailChangeRedisPayload
			if err := json.Unmarshal([]byte(raw), &payload); err != nil || payload.UserID == 0 {
				_ = redisClient.Del(ctx, tokenKey).Err()
				return nil, false
			}

			userKey := RedisKey("email_change", "user", strconv.FormatUint(uint64(payload.UserID), 10))
			casErr := verifyAndConsumeRedisTokenPair(ctx, redisClient, tokenKey, userKey, token, raw)
			if casErr == nil {
				return &moduledto.EmailChangeToken{
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
		if err != nil && !errors.Is(err, redis.Nil) {
			log.Printf("⚠️ Redis 读取邮箱修改 token 失败，回退内存 token 存储: %v", err)
		}
	}

	// LoadAndDelete 保证并发下同一 token 只会被成功消费一次。
	val, ok := emailChangeTokenStore.LoadAndDelete(token)
	if !ok {
		return nil, false
	}

	changeToken, ok := val.(moduledto.EmailChangeToken)
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
	if ok, msg := utils.ValidateUsername(newUsername); !ok {
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

	cfg := config.Get()
	token, err := utils.GenerateLoginToken(userID, newUsername, isAdmin, time.Hour*time.Duration(cfg.JWT.ExpirationHours))
	if err != nil {
		return "", commonpkg.NewInternalError("更新失败")
	}

	return token, nil
}

// UpdatePasswordByOldPassword 使用旧密码校验后更新新密码。
func (s *UserService) UpdatePasswordByOldPassword(userID uint, oldPassword, newPassword string) error {
	if ok, msg := utils.ValidatePassword(newPassword); !ok {
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
