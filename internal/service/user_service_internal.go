package service

import (
	commonpkg "perfect-pic-server/internal/common"
	moduledto "perfect-pic-server/internal/dto"
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/pkg/validator"

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

// validateCreateUserInput 校验通用创建用户输入是否合法。
func (s *UserService) validateCreateUserInput(input moduledto.CreateUserRequest, allowReservedUsername bool) error {
	if ok, msg := validator.ValidatePassword(input.Password); !ok {
		return commonpkg.NewValidationError(msg)
	}

	if allowReservedUsername {
		if ok, msg := validator.ValidateUsernameAllowReserved(input.Username); !ok {
			return commonpkg.NewValidationError(msg)
		}
	} else {
		if ok, msg := validator.ValidateUsername(input.Username); !ok {
			return commonpkg.NewValidationError(msg)
		}
	}

	usernameTaken, err := s.IsUsernameTaken(input.Username, nil, true)
	if err != nil {
		return commonpkg.NewInternalError("创建用户失败")
	}
	if usernameTaken {
		return commonpkg.NewConflictError("用户名已存在")
	}

	if input.Email != nil && *input.Email != "" {
		if ok, msg := validator.ValidateEmail(*input.Email); !ok {
			return commonpkg.NewValidationError(msg)
		}
		emailTaken, err := s.IsEmailTaken(*input.Email, nil, true)
		if err != nil {
			return commonpkg.NewInternalError("创建用户失败")
		}
		if emailTaken {
			return commonpkg.NewConflictError("邮箱已被注册")
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

// applyCreateUserOptionals 将通用创建用户的可选字段应用到模型。
func (s *UserService) applyCreateUserOptionals(user *model.User, input moduledto.CreateUserRequest) error {
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
			return commonpkg.NewValidationError("存储配额不能为负数（-1除外）")
		}
	}

	if input.Status != nil {
		if *input.Status == 1 || *input.Status == 2 {
			user.Status = *input.Status
		} else {
			return commonpkg.NewValidationError("无效的用户状态")
		}
	}

	return nil
}

// prepareUsernameUpdate 校验并准备用户名更新字段。
func (s *UserService) prepareUsernameUpdate(userID uint, username *string, allowReservedUsername bool, updates map[string]interface{}) error {
	if username == nil || *username == "" {
		return nil
	}
	if allowReservedUsername {
		if ok, msg := validator.ValidateUsernameAllowReserved(*username); !ok {
			return commonpkg.NewValidationError(msg)
		}
	} else {
		if ok, msg := validator.ValidateUsername(*username); !ok {
			return commonpkg.NewValidationError(msg)
		}
	}

	excludeID := userID
	usernameTaken, err := s.IsUsernameTaken(*username, &excludeID, true)
	if err != nil {
		return commonpkg.NewInternalError("更新用户失败")
	}
	if usernameTaken {
		return commonpkg.NewConflictError("该用户名已被其他用户占用")
	}

	updates["username"] = *username
	return nil
}

// preparePasswordUpdate 校验并准备密码更新字段。
func (s *UserService) preparePasswordUpdate(password *string, updates map[string]interface{}) error {
	if password == nil || *password == "" {
		return nil
	}
	if ok, msg := validator.ValidatePassword(*password); !ok {
		return commonpkg.NewValidationError(msg)
	}

	hashedPassword, err := hashPassword(*password)
	if err != nil {
		return commonpkg.NewInternalError("更新用户失败")
	}

	updates["password"] = hashedPassword
	return nil
}

// prepareEmailUpdate 校验并准备邮箱更新字段。
func (s *UserService) prepareEmailUpdate(userID uint, email *string, updates map[string]interface{}) error {
	if email == nil || *email == "" {
		return nil
	}
	if ok, msg := validator.ValidateEmail(*email); !ok {
		return commonpkg.NewValidationError(msg)
	}

	excludeID := userID
	emailTaken, err := s.IsEmailTaken(*email, &excludeID, true)
	if err != nil {
		return commonpkg.NewInternalError("更新用户失败")
	}
	if emailTaken {
		return commonpkg.NewConflictError("该邮箱已被其他用户占用")
	}

	updates["email"] = *email
	return nil
}

// prepareEmailVerifiedUpdate 准备邮箱验证状态更新字段。
func (s *UserService) prepareEmailVerifiedUpdate(emailVerified *bool, updates map[string]interface{}) {
	if emailVerified != nil {
		updates["email_verified"] = *emailVerified
	}
}

// prepareStorageQuotaUpdate 校验并准备存储配额更新字段。
func (s *UserService) prepareStorageQuotaUpdate(storageQuota *int64, updates map[string]interface{}) error {
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
	return commonpkg.NewValidationError("存储配额不能为负数（-1除外）")
}

// prepareStatusUpdate 校验并准备用户状态更新字段。
func (s *UserService) prepareStatusUpdate(status *int, updates map[string]interface{}) error {
	if status == nil {
		return nil
	}
	if *status == 1 || *status == 2 {
		updates["status"] = *status
		return nil
	}
	return commonpkg.NewValidationError("无效的用户状态")
}
