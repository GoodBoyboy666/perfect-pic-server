package service

import (
	"context"
	"errors"
	"fmt"
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/utils"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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

// buildAdminUserListQuery 构建管理员用户列表基础查询。
func buildAdminUserListQuery(params AdminUserListParams) *gorm.DB {
	query := db.DB.Model(&model.User{})
	if params.ShowDeleted {
		query = query.Unscoped()
	}
	if strings.TrimSpace(params.Keyword) != "" {
		query = query.Where("username LIKE ?", "%"+params.Keyword+"%")
	}
	return query
}

// resolveAdminUserSortOrder 解析管理员用户列表排序表达式。
func resolveAdminUserSortOrder(order string) string {
	if order == "asc" {
		return "id asc"
	}
	return "id desc"
}

// validateAdminCreateUserInput 校验管理员创建用户输入是否合法。
func validateAdminCreateUserInput(input AdminCreateUserInput) (string, error) {
	if ok, msg := utils.ValidatePassword(input.Password); !ok {
		return msg, nil
	}
	if ok, msg := utils.ValidateUsername(input.Username); !ok {
		return msg, nil
	}

	usernameTaken, err := IsUsernameTaken(input.Username, nil, true)
	if err != nil {
		return "", err
	}
	if usernameTaken {
		return "用户名已存在", nil
	}

	if input.Email != nil && *input.Email != "" {
		if ok, msg := utils.ValidateEmail(*input.Email); !ok {
			return msg, nil
		}
		emailTaken, err := IsEmailTaken(*input.Email, nil, true)
		if err != nil {
			return "", err
		}
		if emailTaken {
			return "邮箱已被注册", nil
		}
	}

	return "", nil
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
func applyAdminCreateUserOptionals(user *model.User, input AdminCreateUserInput) string {
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
			return "存储配额不能为负数（-1除外）"
		}
	}

	if input.Status != nil {
		if *input.Status == 1 || *input.Status == 2 {
			user.Status = *input.Status
		} else {
			return "无效的用户状态"
		}
	}

	return ""
}

// prepareAdminUsernameUpdate 校验并准备用户名更新字段。
func prepareAdminUsernameUpdate(userID uint, username *string, updates map[string]interface{}) (string, error) {
	if username == nil || *username == "" {
		return "", nil
	}
	if ok, msg := utils.ValidateUsername(*username); !ok {
		return msg, nil
	}

	excludeID := userID
	usernameTaken, err := IsUsernameTaken(*username, &excludeID, true)
	if err != nil {
		return "", err
	}
	if usernameTaken {
		return "该用户名已被其他用户占用", nil
	}

	updates["username"] = *username
	return "", nil
}

// prepareAdminPasswordUpdate 校验并准备密码更新字段。
func prepareAdminPasswordUpdate(password *string, updates map[string]interface{}) (string, error) {
	if password == nil || *password == "" {
		return "", nil
	}
	if ok, msg := utils.ValidatePassword(*password); !ok {
		return msg, nil
	}

	hashedPassword, err := hashPassword(*password)
	if err != nil {
		return "", err
	}

	updates["password"] = hashedPassword
	return "", nil
}

// prepareAdminEmailUpdate 校验并准备邮箱更新字段。
func prepareAdminEmailUpdate(userID uint, email *string, updates map[string]interface{}) (string, error) {
	if email == nil || *email == "" {
		return "", nil
	}
	if ok, msg := utils.ValidateEmail(*email); !ok {
		return msg, nil
	}

	excludeID := userID
	emailTaken, err := IsEmailTaken(*email, &excludeID, true)
	if err != nil {
		return "", err
	}
	if emailTaken {
		return "该邮箱已被其他用户占用", nil
	}

	updates["email"] = *email
	return "", nil
}

// prepareAdminEmailVerifiedUpdate 准备邮箱验证状态更新字段。
func prepareAdminEmailVerifiedUpdate(emailVerified *bool, updates map[string]interface{}) {
	if emailVerified != nil {
		updates["email_verified"] = *emailVerified
	}
}

// prepareAdminStorageQuotaUpdate 校验并准备存储配额更新字段。
func prepareAdminStorageQuotaUpdate(storageQuota *int64, updates map[string]interface{}) string {
	if storageQuota == nil {
		return ""
	}
	if *storageQuota == -1 {
		updates["storage_quota"] = nil
		return ""
	}
	if *storageQuota >= 0 {
		updates["storage_quota"] = *storageQuota
		return ""
	}
	return "存储配额不能为负数（-1除外）"
}

// prepareAdminStatusUpdate 校验并准备用户状态更新字段。
func prepareAdminStatusUpdate(status *int, updates map[string]interface{}) string {
	if status == nil {
		return ""
	}
	if *status == 1 || *status == 2 {
		updates["status"] = *status
		return ""
	}
	return "无效的用户状态"
}

// hardDeleteUserForAdmin 执行管理员硬删除，包含文件与关联记录清理。
func hardDeleteUserForAdmin(userID uint) error {
	if err := DeleteUserFiles(userID); err != nil {
		return err
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		var user model.User
		if err := tx.Unscoped().First(&user, userID).Error; err != nil {
			return err
		}
		// 不依赖数据库外键级联（SQLite 需要 foreign_keys=ON 且旧表可能没有 CASCADE 约束），
		// 这里显式清理关联图片记录，保证硬删除用户后不会残留 images。
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&model.Image{}).Error; err != nil {
			return err
		}
		return tx.Unscoped().Delete(&user).Error
	})
}

// softDeleteUserForAdmin 执行管理员软删除并重写唯一字段。
func softDeleteUserForAdmin(userID uint) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		var user model.User
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}

		newUsername, newEmail := buildSoftDeletedIdentity(user, time.Now().Unix())
		if err := tx.Model(&user).Updates(map[string]interface{}{
			"username": newUsername,
			"email":    newEmail,
			"status":   3,
		}).Error; err != nil {
			return err
		}

		return tx.Delete(&user).Error
	})
}

// buildSoftDeletedIdentity 构造软删除后的用户名与邮箱占位值。
func buildSoftDeletedIdentity(user model.User, timestamp int64) (string, string) {
	newUsername := fmt.Sprintf("%s_del_%d", user.Username, timestamp)
	newEmail := fmt.Sprintf("del_%d_%s", timestamp, user.Email)
	if len(newEmail) > 255 {
		newEmail = newEmail[:255]
	}
	return newUsername, newEmail
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

// isUserFieldTaken 按指定字段检查用户记录是否存在。
func isUserFieldTaken(field string, value string, excludeUserID *uint, includeDeleted bool) (bool, error) {
	query := db.DB.Model(&model.User{})
	if includeDeleted {
		query = query.Unscoped()
	}
	if excludeUserID != nil {
		query = query.Where("id != ?", *excludeUserID)
	}

	var count int64
	if err := query.Where(field+" = ?", value).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
