package service

import (
	"errors"
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/repository"
	"time"

	"gorm.io/gorm"
)

type AdminUserListParams struct {
	Page        int
	PageSize    int
	Keyword     string
	ShowDeleted bool
	Order       string
}

type AdminUserUpdateInput struct {
	Username      *string
	Password      *string
	Email         *string
	EmailVerified *bool
	StorageQuota  *int64
	Status        *int
}

type AdminCreateUserInput struct {
	Username      string
	Password      string
	Email         *string
	EmailVerified *bool
	StorageQuota  *int64
	Status        *int
}

// AdminListUsers 按分页与筛选条件查询用户列表。
func AdminListUsers(params AdminUserListParams) ([]model.User, int64, error) {
	page, pageSize := normalizeAdminPagination(params.Page, params.PageSize)
	sortOrder := resolveAdminUserSortOrder(params.Order)
	users, total, err := repository.User.AdminListUsers(
		params.Keyword,
		params.ShowDeleted,
		sortOrder,
		(page-1)*pageSize,
		pageSize,
	)
	if err != nil {
		return nil, 0, NewInternalError("获取用户列表失败")
	}
	return users, total, nil
}

// AdminGetUserDetail 根据用户 ID 获取详情。
func AdminGetUserDetail(id uint) (*model.User, error) {
	user, err := repository.User.FindUnscopedByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NewNotFoundError("用户不存在")
		}
		return nil, NewInternalError("获取用户详情失败")
	}
	return user, nil
}

// AdminCreateUser 创建后台普通用户。
func AdminCreateUser(input AdminCreateUserInput) (*model.User, error) {
	if err := validateAdminCreateUserInput(input); err != nil {
		return nil, err
	}

	hashedPassword, err := hashPassword(input.Password)
	if err != nil {
		return nil, NewInternalError("创建用户失败")
	}

	user := model.User{
		Username: input.Username,
		Password: hashedPassword,
		Admin:    false,
		Status:   1,
	}

	if err := applyAdminCreateUserOptionals(&user, input); err != nil {
		return nil, err
	}

	if err := repository.User.Create(&user); err != nil {
		return nil, NewInternalError("创建用户失败")
	}

	return &user, nil
}

// AdminPrepareUserUpdates 校验后台用户更新输入并构建可持久化的 updates。
func AdminPrepareUserUpdates(userID uint, req AdminUserUpdateInput) (map[string]interface{}, error) {
	updates := make(map[string]interface{})

	if err := prepareAdminUsernameUpdate(userID, req.Username, updates); err != nil {
		return nil, err
	}

	if err := prepareAdminPasswordUpdate(req.Password, updates); err != nil {
		return nil, err
	}

	if err := prepareAdminEmailUpdate(userID, req.Email, updates); err != nil {
		return nil, err
	}

	prepareAdminEmailVerifiedUpdate(req.EmailVerified, updates)

	if err := prepareAdminStorageQuotaUpdate(req.StorageQuota, updates); err != nil {
		return nil, err
	}

	if err := prepareAdminStatusUpdate(req.Status, updates); err != nil {
		return nil, err
	}

	return updates, nil
}

// AdminApplyUserUpdates 将更新字段应用到指定用户。
func AdminApplyUserUpdates(userID uint, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	if err := repository.User.UpdateByID(userID, updates); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NewNotFoundError("用户不存在")
		}
		return NewInternalError("更新用户失败")
	}
	return nil
}

// AdminDeleteUser 删除用户。
// hardDelete=true 时执行彻底删除；否则执行软删除并清理唯一字段占用。
func AdminDeleteUser(userID uint, hardDelete bool) error {
	if hardDelete {
		if err := DeleteUserFiles(userID); err != nil {
			return NewInternalError("删除用户失败")
		}
		if err := repository.User.HardDeleteUserWithImages(userID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return NewNotFoundError("用户不存在")
			}
			return NewInternalError("删除用户失败")
		}
		return nil
	}

	if err := repository.User.AdminSoftDeleteUser(userID, time.Now().Unix()); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NewNotFoundError("用户不存在")
		}
		return NewInternalError("删除用户失败")
	}
	return nil
}
