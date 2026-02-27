package service

import (
	"errors"
	platformservice "perfect-pic-server/internal/common"
	moduledto "perfect-pic-server/internal/dto"
	"perfect-pic-server/internal/model"
	"time"

	"gorm.io/gorm"
)

// AdminListUsers 按分页与筛选条件查询用户列表。
func (s *Service) AdminListUsers(params moduledto.AdminUserListRequest) ([]model.User, int64, error) {
	page, pageSize := normalizeAdminPagination(params.Page, params.PageSize)
	sortOrder := resolveAdminUserSortOrder(params.Order)
	users, total, err := s.userStore.AdminListUsers(
		params.Keyword,
		params.ShowDeleted,
		sortOrder,
		(page-1)*pageSize,
		pageSize,
	)
	if err != nil {
		return nil, 0, platformservice.NewInternalError("获取用户列表失败")
	}
	return users, total, nil
}

// AdminGetUserDetail 根据用户 ID 获取详情。
func (s *Service) AdminGetUserDetail(id uint) (*model.User, error) {
	user, err := s.userStore.FindUnscopedByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, platformservice.NewNotFoundError("用户不存在")
		}
		return nil, platformservice.NewInternalError("获取用户详情失败")
	}
	return user, nil
}

// AdminCreateUser 创建后台普通用户。
func (s *Service) AdminCreateUser(input moduledto.AdminCreateUserRequest) (*model.User, error) {
	if err := s.validateAdminCreateUserInput(input); err != nil {
		return nil, err
	}

	hashedPassword, err := hashPassword(input.Password)
	if err != nil {
		return nil, platformservice.NewInternalError("创建用户失败")
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

	if err := s.userStore.Create(&user); err != nil {
		return nil, platformservice.NewInternalError("创建用户失败")
	}

	return &user, nil
}

// AdminPrepareUserUpdates 校验后台用户更新输入并构建可持久化的 updates。
func (s *Service) AdminPrepareUserUpdates(userID uint, req moduledto.AdminUserUpdateRequest) (map[string]interface{}, error) {
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
func (s *Service) AdminApplyUserUpdates(userID uint, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	if err := s.userStore.UpdateByID(userID, updates); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return platformservice.NewNotFoundError("用户不存在")
		}
		return platformservice.NewInternalError("更新用户失败")
	}
	return nil
}

// AdminDeleteUser 删除用户。
// hardDelete=true 时执行彻底删除；否则执行软删除并清理唯一字段占用。
func (s *Service) AdminDeleteUser(userID uint, hardDelete bool) error {
	if hardDelete {
		if err := s.DeleteUserFiles(userID); err != nil {
			return platformservice.NewInternalError("删除用户失败")
		}
		// 显式删除该用户所有 Passkey 凭据，避免旧 SQLite 表外键缺失导致级联删除失效。
		if err := s.userStore.DeletePasskeyCredentialsByUserID(userID); err != nil {
			return platformservice.NewInternalError("删除用户失败")
		}
		if err := s.userStore.HardDeleteUserWithImages(userID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return platformservice.NewNotFoundError("用户不存在")
			}
			return platformservice.NewInternalError("删除用户失败")
		}
		return nil
	}

	if err := s.userStore.AdminSoftDeleteUser(userID, time.Now().Unix()); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return platformservice.NewNotFoundError("用户不存在")
		}
		return platformservice.NewInternalError("删除用户失败")
	}
	return nil
}
