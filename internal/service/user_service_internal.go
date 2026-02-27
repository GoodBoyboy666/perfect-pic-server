package service

import (
	"context"
	"errors"
	platformservice "perfect-pic-server/internal/common"
	moduledto "perfect-pic-server/internal/dto"
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/utils"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

// normalizeAdminPagination 归一化管理员分页参数。
func normalizeAdminPagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	return page, pageSize
}

// resolveAdminUserSortOrder 解析管理员用户列表排序表达式。
func resolveAdminUserSortOrder(order string) string {
	if order == "asc" {
		return "id asc"
	}
	return "id desc"
}

// validateAdminCreateUserInput 校验管理员创建用户输入是否合法。
func (s *Service) validateAdminCreateUserInput(input moduledto.AdminCreateUserRequest) error {
	if ok, msg := utils.ValidatePassword(input.Password); !ok {
		return platformservice.NewValidationError(msg)
	}
	// 管理员后台创建用户允许使用保留用户名（与后台修改用户名规则一致）。
	if ok, msg := utils.ValidateUsernameAllowReserved(input.Username); !ok {
		return platformservice.NewValidationError(msg)
	}

	usernameTaken, err := s.IsUsernameTaken(input.Username, nil, true)
	if err != nil {
		return platformservice.NewInternalError("创建用户失败")
	}
	if usernameTaken {
		return platformservice.NewConflictError("用户名已存在")
	}

	if input.Email != nil && *input.Email != "" {
		if ok, msg := utils.ValidateEmail(*input.Email); !ok {
			return platformservice.NewValidationError(msg)
		}
		emailTaken, err := s.IsEmailTaken(*input.Email, nil, true)
		if err != nil {
			return platformservice.NewInternalError("创建用户失败")
		}
		if emailTaken {
			return platformservice.NewConflictError("邮箱已被注册")
		}
	}

	return nil
}

// hashPassword 使用 bcrypt 对密码进行哈希。
func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// applyAdminCreateUserOptionals 将管理员创建用户的可选字段应用到模型。
func (s *Service) applyAdminCreateUserOptionals(user *model.User, input moduledto.AdminCreateUserRequest) error {
	if input.Email != nil {
		user.Email = *input.Email
	}

	if input.EmailVerified != nil {
		user.EmailVerified = *input.EmailVerified
	}

	if input.StorageQuota != nil {
		if *input.StorageQuota == -1 {
			user.StorageQuota = nil
		} else if *input.StorageQuota >= 0 {
			quota := *input.StorageQuota
			user.StorageQuota = &quota
		} else {
			return platformservice.NewValidationError("存储配额不能为负数（-1除外）")
		}
	}

	if input.Status != nil {
		if *input.Status == 1 || *input.Status == 2 {
			user.Status = *input.Status
		} else {
			return platformservice.NewValidationError("无效的用户状态")
		}
	}

	return nil
}

// prepareAdminUsernameUpdate 校验并准备用户名更新字段。
func (s *Service) prepareAdminUsernameUpdate(userID uint, username *string, updates map[string]interface{}) error {
	if username == nil || *username == "" {
		return nil
	}
	if ok, msg := utils.ValidateUsernameAllowReserved(*username); !ok {
		return platformservice.NewValidationError(msg)
	}

	excludeID := userID
	usernameTaken, err := s.IsUsernameTaken(*username, &excludeID, true)
	if err != nil {
		return platformservice.NewInternalError("更新用户失败")
	}
	if usernameTaken {
		return platformservice.NewConflictError("该用户名已被其他用户占用")
	}

	updates["username"] = *username
	return nil
}

// prepareAdminPasswordUpdate 校验并准备密码更新字段。
func (s *Service) prepareAdminPasswordUpdate(password *string, updates map[string]interface{}) error {
	if password == nil || *password == "" {
		return nil
	}
	if ok, msg := utils.ValidatePassword(*password); !ok {
		return platformservice.NewValidationError(msg)
	}

	hashedPassword, err := hashPassword(*password)
	if err != nil {
		return platformservice.NewInternalError("更新用户失败")
	}

	updates["password"] = hashedPassword
	return nil
}

// prepareAdminEmailUpdate 校验并准备邮箱更新字段。
func (s *Service) prepareAdminEmailUpdate(userID uint, email *string, updates map[string]interface{}) error {
	if email == nil || *email == "" {
		return nil
	}
	if ok, msg := utils.ValidateEmail(*email); !ok {
		return platformservice.NewValidationError(msg)
	}

	excludeID := userID
	emailTaken, err := s.IsEmailTaken(*email, &excludeID, true)
	if err != nil {
		return platformservice.NewInternalError("更新用户失败")
	}
	if emailTaken {
		return platformservice.NewConflictError("该邮箱已被其他用户占用")
	}

	updates["email"] = *email
	return nil
}

// prepareAdminEmailVerifiedUpdate 准备邮箱验证状态更新字段。
func (s *Service) prepareAdminEmailVerifiedUpdate(emailVerified *bool, updates map[string]interface{}) {
	if emailVerified != nil {
		updates["email_verified"] = *emailVerified
	}
}

// prepareAdminStorageQuotaUpdate 校验并准备存储配额更新字段。
func (s *Service) prepareAdminStorageQuotaUpdate(storageQuota *int64, updates map[string]interface{}) error {
	if storageQuota == nil {
		return nil
	}
	if *storageQuota == -1 {
		updates["storage_quota"] = nil
		return nil
	}
	if *storageQuota >= 0 {
		updates["storage_quota"] = *storageQuota
		return nil
	}
	return platformservice.NewValidationError("存储配额不能为负数（-1除外）")
}

// prepareAdminStatusUpdate 校验并准备用户状态更新字段。
func (s *Service) prepareAdminStatusUpdate(status *int, updates map[string]interface{}) error {
	if status == nil {
		return nil
	}
	if *status == 1 || *status == 2 {
		updates["status"] = *status
		return nil
	}
	return platformservice.NewValidationError("无效的用户状态")
}

// verifyAndConsumeRedisTokenPair 使用 Redis WATCH/CAS 原子校验并消费 token 对。
func verifyAndConsumeRedisTokenPair(
	ctx context.Context,
	redisClient *redis.Client,
	tokenKey string,
	userKey string,
	expectedUserToken string,
	expectedTokenValue string,
) error {
	var consumed bool
	watchErr := redisClient.Watch(ctx, func(tx *redis.Tx) error {
		currentUserToken, getErr := tx.Get(ctx, userKey).Result()
		if getErr != nil {
			if errors.Is(getErr, redis.Nil) {
				return redis.TxFailedErr
			}
			return getErr
		}
		if currentUserToken != expectedUserToken {
			return redis.TxFailedErr
		}

		currentTokenValue, tokenErr := tx.Get(ctx, tokenKey).Result()
		if tokenErr != nil {
			if errors.Is(tokenErr, redis.Nil) {
				return redis.TxFailedErr
			}
			return tokenErr
		}
		if currentTokenValue != expectedTokenValue {
			return redis.TxFailedErr
		}

		_, pipeErr := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Del(ctx, tokenKey)
			pipe.Del(ctx, userKey)
			return nil
		})
		if pipeErr != nil {
			return pipeErr
		}

		consumed = true
		return nil
	}, userKey, tokenKey)

	if watchErr == nil && consumed {
		return nil
	}
	if errors.Is(watchErr, redis.TxFailedErr) || (!consumed && watchErr == nil) {
		return errRedisTokenCASMismatch
	}
	return watchErr
}
