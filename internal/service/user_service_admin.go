package service

import (
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/repository"
	"time"
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
	return repository.User.AdminListUsers(
		params.Keyword,
		params.ShowDeleted,
		sortOrder,
		(page-1)*pageSize,
		pageSize,
	)
}

// AdminGetUserDetail 根据用户 ID 获取详情。
func AdminGetUserDetail(id uint) (*model.User, error) {
	return repository.User.FindUnscopedByID(id)
}

// AdminCreateUser 创建后台普通用户。
func AdminCreateUser(input AdminCreateUserInput) (*model.User, string, error) {
	if msg, err := validateAdminCreateUserInput(input); err != nil || msg != "" {
		return nil, msg, err
	}

	hashedPassword, err := hashPassword(input.Password)
	if err != nil {
		return nil, "", err
	}

	user := model.User{
		Username: input.Username,
		Password: hashedPassword,
		Admin:    false,
		Status:   1,
	}

	if msg := applyAdminCreateUserOptionals(&user, input); msg != "" {
		return nil, msg, nil
	}

	if err := repository.User.Create(&user); err != nil {
		return nil, "", err
	}

	return &user, "", nil
}

// AdminPrepareUserUpdates 校验后台用户更新输入并构建可持久化的 updates。
func AdminPrepareUserUpdates(userID uint, req AdminUserUpdateInput) (map[string]interface{}, string, error) {
	updates := make(map[string]interface{})

	if msg, err := prepareAdminUsernameUpdate(userID, req.Username, updates); err != nil || msg != "" {
		return nil, msg, err
	}

	if msg, err := prepareAdminPasswordUpdate(req.Password, updates); err != nil || msg != "" {
		return nil, msg, err
	}

	if msg, err := prepareAdminEmailUpdate(userID, req.Email, updates); err != nil || msg != "" {
		return nil, msg, err
	}

	prepareAdminEmailVerifiedUpdate(req.EmailVerified, updates)

	if msg := prepareAdminStorageQuotaUpdate(req.StorageQuota, updates); msg != "" {
		return nil, msg, nil
	}

	if msg := prepareAdminStatusUpdate(req.Status, updates); msg != "" {
		return nil, msg, nil
	}

	return updates, "", nil
}

// AdminApplyUserUpdates 将更新字段应用到指定用户。
func AdminApplyUserUpdates(userID uint, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	return repository.User.UpdateByID(userID, updates)
}

// AdminDeleteUser 删除用户。
// hardDelete=true 时执行彻底删除；否则执行软删除并清理唯一字段占用。
func AdminDeleteUser(userID uint, hardDelete bool) error {
	if hardDelete {
		if err := DeleteUserFiles(userID); err != nil {
			return err
		}
		return repository.User.HardDeleteUserWithImages(userID)
	}
	return repository.User.AdminSoftDeleteUser(userID, time.Now().Unix())
}
