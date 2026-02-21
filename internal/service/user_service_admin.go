package service

import (
	"errors"
	"perfect-pic-server/internal/model"
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
func (s *AppService) AdminListUsers(params AdminUserListParams) ([]model.User, int64, error) {
	page, pageSize := normalizeAdminPagination(params.Page, params.PageSize)
	sortOrder := resolveAdminUserSortOrder(params.Order)
	users, total, err := s.repos.User.AdminListUsers(
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
func (s *AppService) AdminGetUserDetail(id uint) (*model.User, error) {
	user, err := s.repos.User.FindUnscopedByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NewNotFoundError("用户不存在")
		}
		return nil, NewInternalError("获取用户详情失败")
	}
	return user, nil
}

// AdminCreateUser 创建后台普通用户。
func (s *AppService) AdminCreateUser(input AdminCreateUserInput) (*model.User, error) {
	if err := s.validateAdminCreateUserInput(input); err != nil {
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

	if err := s.applyAdminCreateUserOptionals(&user, input); err != nil {
		return nil, err
	}

	if err := s.repos.User.Create(&user); err != nil {
		return nil, NewInternalError("创建用户失败")
	}

	return &user, nil
}

// AdminPrepareUserUpdates 校验后台用户更新输入并构建可持久化的 updates。
func (s *AppService) AdminPrepareUserUpdates(userID uint, req AdminUserUpdateInput) (map[string]interface{}, error) {
	updates := make(map[string]interface{})

	if err := s.prepareAdminUsernameUpdate(userID, req.Username, updates); err != nil {
		return nil, err
	}

	if err := s.prepareAdminPasswordUpdate(req.Password, updates); err != nil {
		return nil, err
	}

	if err := s.prepareAdminEmailUpdate(userID, req.Email, updates); err != nil {
		return nil, err
	}

	s.prepareAdminEmailVerifiedUpdate(req.EmailVerified, updates)

	if err := s.prepareAdminStorageQuotaUpdate(req.StorageQuota, updates); err != nil {
		return nil, err
	}

	if err := s.prepareAdminStatusUpdate(req.Status, updates); err != nil {
		return nil, err
	}

	return updates, nil
}

// AdminApplyUserUpdates 将更新字段应用到指定用户。
func (s *AppService) AdminApplyUserUpdates(userID uint, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	if err := s.repos.User.UpdateByID(userID, updates); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NewNotFoundError("用户不存在")
		}
		return NewInternalError("更新用户失败")
	}
	return nil
}

// AdminDeleteUser 删除用户。
// hardDelete=true 时执行彻底删除；否则执行软删除并清理唯一字段占用。
func (s *AppService) AdminDeleteUser(userID uint, hardDelete bool) error {
	if hardDelete {
		if err := s.DeleteUserFiles(userID); err != nil {
			return NewInternalError("删除用户失败")
		}
		// 显式删除该用户所有 Passkey 凭据，避免旧 SQLite 表外键缺失导致级联删除失效。
		if err := s.repos.User.DeletePasskeyCredentialsByUserID(userID); err != nil {
			return NewInternalError("删除用户失败")
		}
		if err := s.repos.User.HardDeleteUserWithImages(userID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return NewNotFoundError("用户不存在")
			}
			return NewInternalError("删除用户失败")
		}
		return nil
	}

	if err := s.repos.User.AdminSoftDeleteUser(userID, time.Now().Unix()); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NewNotFoundError("用户不存在")
		}
		return NewInternalError("删除用户失败")
	}
	return nil
}
