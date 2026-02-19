package service

import (
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
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

	var users []model.User
	var total int64

	query := buildAdminUserListQuery(params)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortOrder := resolveAdminUserSortOrder(params.Order)
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order(sortOrder).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// AdminGetUserDetail 根据用户 ID 获取详情。
func AdminGetUserDetail(id uint) (*model.User, error) {
	var user model.User
	if err := db.DB.Unscoped().First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
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

	if err := db.DB.Create(&user).Error; err != nil {
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

	var user model.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		return err
	}

	return db.DB.Model(&user).Updates(updates).Error
}

// AdminDeleteUser 删除用户。
// hardDelete=true 时执行彻底删除；否则执行软删除并清理唯一字段占用。
func AdminDeleteUser(userID uint, hardDelete bool) error {
	if hardDelete {
		return hardDeleteUserForAdmin(userID)
	}
	return softDeleteUserForAdmin(userID)
}
