package service

import (
	"os"
	"path/filepath"
	platformservice "perfect-pic-server/internal/common"
	moduledto "perfect-pic-server/internal/dto"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"

	"golang.org/x/crypto/bcrypt"
)

func assertServiceErrorCode(t *testing.T, err error, code platformservice.ErrorCode) *platformservice.ServiceError {
	t.Helper()
	serviceErr, ok := platformservice.AsServiceError(err)
	if !ok {
		t.Fatalf("期望 ServiceError，实际为: %v", err)
	}
	if serviceErr.Code != code {
		t.Fatalf("期望错误码 %q，实际为 %q", code, serviceErr.Code)
	}
	return serviceErr
}

// 测试内容：验证重置密码令牌可生成、可校验且为一次性使用。
func TestGenerateAndVerifyForgetPasswordToken_OneTimeUse(t *testing.T) {
	setupTestDB(t)
	resetPasswordResetStore()

	token, err := testService.GenerateForgetPasswordToken(42)
	if err != nil {
		t.Fatalf("GenerateForgetPasswordToken: %v", err)
	}
	if len(token) != 64 {
		t.Fatalf("期望 64-char hex token，实际为 len=%d token=%q", len(token), token)
	}

	uid, ok := testService.VerifyForgetPasswordToken(token)
	if !ok || uid != 42 {
		t.Fatalf("期望 valid token for uid=42，实际为 uid=%d ok=%v", uid, ok)
	}

	uid2, ok2 := testService.VerifyForgetPasswordToken(token)
	if ok2 || uid2 != 0 {
		t.Fatalf("期望 one-time use token to be 无效 on second use，实际为 uid=%d ok=%v", uid2, ok2)
	}
}

// 测试内容：验证并发场景下同一个重置密码 token 最多只能成功一次。
func TestVerifyForgetPasswordToken_ConcurrentOneTime(t *testing.T) {
	setupTestDB(t)
	resetPasswordResetStore()

	token, err := testService.GenerateForgetPasswordToken(99)
	if err != nil {
		t.Fatalf("GenerateForgetPasswordToken: %v", err)
	}

	var success int32
	const workers = 16
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			if _, ok := testService.VerifyForgetPasswordToken(token); ok {
				atomic.AddInt32(&success, 1)
			}
		}()
	}
	wg.Wait()

	if success != 1 {
		t.Fatalf("期望并发消费仅成功 1 次，实际成功 %d 次", success)
	}
}

// 测试内容：验证删除用户文件会移除头像目录和图片文件记录。
func TestDeleteUserFiles_RemovesAvatarDirAndImages(t *testing.T) {
	setupTestDB(t)

	tmp := t.TempDir()
	oldwd, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("切换工作目录失败: %v", err)
	}
	defer func() { _ = os.Chdir(oldwd) }()

	userID := uint(7)
	_ = db.DB.Create(&model.User{
		ID:       userID,
		Username: "user_7",
		Password: "x",
		Status:   1,
	}).Error

	// 创建包含文件的头像目录。
	avatarDir := filepath.Join("uploads", "avatars", "7")
	if err := os.MkdirAll(avatarDir, 0755); err != nil {
		t.Fatalf("创建头像目录失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(avatarDir, "a.txt"), []byte("x"), 0644); err != nil {
		t.Fatalf("写入头像文件失败: %v", err)
	}

	// 创建图片记录和物理文件。
	imgRel := filepath.ToSlash(filepath.Join("2026", "02", "13", "x.png"))
	imgLocal := filepath.FromSlash(imgRel)
	imgFile := filepath.Join("uploads", "imgs", imgLocal)
	if err := os.MkdirAll(filepath.Dir(imgFile), 0755); err != nil {
		t.Fatalf("创建图片目录失败: %v", err)
	}
	if err := os.WriteFile(imgFile, []byte{0x89, 0x50, 0x4E, 0x47}, 0644); err != nil {
		t.Fatalf("写入图片文件失败: %v", err)
	}

	if err := db.DB.Create(&model.Image{
		Filename:   "x.png",
		Path:       imgRel,
		Size:       4,
		MimeType:   ".png",
		UploadedAt: 1,
		UserID:     userID,
		Width:      1,
		Height:     1,
	}).Error; err != nil {
		t.Fatalf("create image record: %v", err)
	}

	if err := testService.DeleteUserFiles(userID); err != nil {
		t.Fatalf("DeleteUserFiles: %v", err)
	}

	if _, err := os.Stat(avatarDir); !os.IsNotExist(err) {
		t.Fatalf("期望 avatar dir to be removed, stat err=%v", err)
	}
	if _, err := os.Stat(imgFile); !os.IsNotExist(err) {
		t.Fatalf("期望 image file to be removed, stat err=%v", err)
	}
}

// 测试内容：验证邮箱变更 token 为一次性使用，且同一用户新 token 会使旧 token 失效。
func TestEmailChangeToken_OneTimeAndReplaced(t *testing.T) {
	setupTestDB(t)
	resetEmailChangeStore()
	t.Cleanup(resetEmailChangeStore)

	tok1, err := testService.GenerateEmailChangeToken(1, "old@example.com", "new1@example.com")
	if err != nil {
		t.Fatalf("GenerateEmailChangeToken: %v", err)
	}
	if _, ok := testService.VerifyEmailChangeToken(tok1); !ok {
		t.Fatalf("expected token to be valid on first consume")
	}
	if _, ok := testService.VerifyEmailChangeToken(tok1); ok {
		t.Fatalf("expected token to be one-time use")
	}

	tok2, err := testService.GenerateEmailChangeToken(2, "old@example.com", "new2@example.com")
	if err != nil {
		t.Fatalf("GenerateEmailChangeToken: %v", err)
	}
	tok3, err := testService.GenerateEmailChangeToken(2, "old@example.com", "new3@example.com")
	if err != nil {
		t.Fatalf("GenerateEmailChangeToken: %v", err)
	}
	if _, ok := testService.VerifyEmailChangeToken(tok2); ok {
		t.Fatalf("expected older token to be invalid after replaced by a new one")
	}
	item, ok := testService.VerifyEmailChangeToken(tok3)
	if !ok || item == nil || item.NewEmail != "new3@example.com" {
		t.Fatalf("expected newest token to be valid, got item=%+v ok=%v", item, ok)
	}
}

// 测试内容：验证过期的邮箱变更 token 会被拒绝（内存回退分支）。
func TestEmailChangeToken_ExpiredRejected(t *testing.T) {
	setupTestDB(t)
	resetEmailChangeStore()
	t.Cleanup(resetEmailChangeStore)

	const expiredToken = "expired_email_change_token"
	expired := moduledto.EmailChangeToken{
		UserID:    1001,
		Token:     expiredToken,
		OldEmail:  "old@example.com",
		NewEmail:  "new@example.com",
		ExpiresAt: time.Now().Add(-time.Minute),
	}
	testService.userService.emailChangeStore.Store(uint(1001), expiredToken)
	testService.userService.emailChangeTokenStore.Store(expiredToken, expired)

	item, ok := testService.VerifyEmailChangeToken(expiredToken)
	if ok || item != nil {
		t.Fatalf("expected expired token to be rejected, got item=%+v ok=%v", item, ok)
	}
}

// 测试内容：验证并发场景下同一个邮箱变更 token 最多只能成功一次。
func TestVerifyEmailChangeToken_ConcurrentOneTime(t *testing.T) {
	setupTestDB(t)
	resetEmailChangeStore()

	token, err := testService.GenerateEmailChangeToken(101, "old@example.com", "new@example.com")
	if err != nil {
		t.Fatalf("GenerateEmailChangeToken: %v", err)
	}

	var success int32
	const workers = 16
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			if _, ok := testService.VerifyEmailChangeToken(token); ok {
				atomic.AddInt32(&success, 1)
			}
		}()
	}
	wg.Wait()

	if success != 1 {
		t.Fatalf("期望并发消费仅成功 1 次，实际成功 %d 次", success)
	}
}

// 测试内容：验证默认存储配额读取与配置覆盖逻辑。
func TestGetSystemDefaultStorageQuota(t *testing.T) {
	setupTestDB(t)

	// 当设置缺失时应回退到 DefaultSettings 的默认值。
	if got := testService.GetSystemDefaultStorageQuota(); got <= 0 {
		t.Fatalf("期望 positive default quota，实际为 %d", got)
	}

	// 覆盖为自定义值。
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigDefaultStorageQuota, Value: "123"}).Error
	testService.ClearCache()
	if got := testService.GetSystemDefaultStorageQuota(); got != 123 {
		t.Fatalf("期望 123，实际为 %d", got)
	}

	// 非法值（<=0）应统一回退到默认 1GB。
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigDefaultStorageQuota, Value: "-1"}).Error
	testService.ClearCache()
	if got := testService.GetSystemDefaultStorageQuota(); got != 1073741824 {
		t.Fatalf("期望 fallback 1GB，实际为 %d", got)
	}
}

// 测试内容：验证管理员获取用户详情接口可正确返回用户信息。
func TestGetUserDetailForAdmin(t *testing.T) {
	setupTestDB(t)

	u := model.User{Username: "alice", Password: "x", Status: 1, Email: "a@example.com"}
	_ = db.DB.Create(&u).Error

	got, err := testService.AdminGetUserDetail(u.ID)
	if err != nil {
		t.Fatalf("AdminGetUserDetail: %v", err)
	}
	if got.Username != "alice" {
		t.Fatalf("非预期 user: %+v", got)
	}
}

// 测试内容：验证管理员用户列表支持关键词过滤与包含已删除用户。
func TestListUsersForAdmin_FilterAndShowDeleted(t *testing.T) {
	setupTestDB(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u1 := model.User{Username: "alice", Password: string(hashed), Status: 1, Email: "a@example.com"}
	u2 := model.User{Username: "bob", Password: string(hashed), Status: 1, Email: "b@example.com"}
	_ = db.DB.Create(&u1).Error
	_ = db.DB.Create(&u2).Error
	_ = db.DB.Delete(&u2).Error

	users, total, err := testService.AdminListUsers(moduledto.AdminUserListRequest{Page: 1, PageSize: 10, Keyword: "ali"})
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if total != 1 || len(users) != 1 || users[0].Username != "alice" {
		t.Fatalf("非预期: total=%d len=%d users=%v", total, len(users), users)
	}

	users2, total2, err := testService.AdminListUsers(moduledto.AdminUserListRequest{Page: 1, PageSize: 10, ShowDeleted: true})
	if err != nil {
		t.Fatalf("testService.ListUsers(showDeleted): %v", err)
	}
	if total2 != 2 || len(users2) != 2 {
		t.Fatalf("期望 2 users with deleted，实际为 total=%d len=%d", total2, len(users2))
	}
}

// 测试内容：验证管理员创建用户时的用户名与密码校验。
func TestCreateUserForAdmin_Validates(t *testing.T) {
	setupTestDB(t)

	user, err := testService.AdminCreateUser(moduledto.AdminCreateUserRequest{Username: "ab", Password: "abc12345"})
	if user != nil {
		t.Fatalf("期望 user=nil，实际为 %v", user)
	}
	assertServiceErrorCode(t, err, platformservice.ErrorCodeValidation)

	user, err = testService.AdminCreateUser(moduledto.AdminCreateUserRequest{Username: "alice", Password: "short"})
	if user != nil {
		t.Fatalf("期望 user=nil，实际为 %v", user)
	}
	assertServiceErrorCode(t, err, platformservice.ErrorCodeValidation)
}

// 测试内容：验证管理员创建用户时允许使用保留用户名。
func TestCreateUserForAdmin_AllowsReservedUsername(t *testing.T) {
	setupTestDB(t)

	user, err := testService.AdminCreateUser(moduledto.AdminCreateUserRequest{
		Username: "admin",
		Password: "abc12345",
	})
	if err != nil || user == nil {
		t.Fatalf("期望 success，实际为 user=%v err=%v", user, err)
	}
	if user.Username != "admin" {
		t.Fatalf("非预期用户名: %+v", user)
	}
}

// 测试内容：验证管理员创建用户可选字段与配额清空逻辑。
func TestCreateUserForAdmin_SuccessOptions(t *testing.T) {
	setupTestDB(t)

	email := "a@example.com"
	emailVerified := true
	quota := int64(100)
	status := 2

	user, err := testService.AdminCreateUser(moduledto.AdminCreateUserRequest{
		Username:      "alice_1",
		Password:      "abc12345",
		Email:         &email,
		EmailVerified: &emailVerified,
		StorageQuota:  &quota,
		Status:        &status,
	})
	if err != nil || user == nil {
		t.Fatalf("期望 success，实际为 user=%v err=%v", user, err)
	}
	if user.Email != email || user.EmailVerified != true || user.Status != 2 {
		t.Fatalf("非预期 created user: %+v", user)
	}
	if user.StorageQuota == nil || *user.StorageQuota != 100 {
		t.Fatalf("非预期 quota: %+v", user.StorageQuota)
	}

	// StorageQuota=-1 应设置为 nil。
	q2 := int64(-1)
	user2, err := testService.AdminCreateUser(moduledto.AdminCreateUserRequest{
		Username:     "alice_2",
		Password:     "abc12345",
		StorageQuota: &q2,
	})
	if err != nil || user2 == nil {
		t.Fatalf("期望 success，实际为 user=%v err=%v", user2, err)
	}
	if user2.StorageQuota != nil {
		t.Fatalf("期望为 nil quota for -1")
	}
}

// 测试内容：验证管理员更新用户信息的准备与应用流程。
func TestPrepareAndApplyUserUpdatesForAdmin(t *testing.T) {
	setupTestDB(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u := model.User{Username: "alice", Password: string(hashed), Status: 1, Email: "a@example.com"}
	_ = db.DB.Create(&u).Error

	newName := "alice2"
	newStatus := 2
	updates, err := testService.AdminPrepareUserUpdates(u.ID, moduledto.AdminUserUpdateRequest{Username: &newName, Status: &newStatus})
	if err != nil {
		t.Fatalf("AdminPrepareUserUpdates: err=%v", err)
	}
	if err := testService.AdminApplyUserUpdates(u.ID, updates); err != nil {
		t.Fatalf("AdminApplyUserUpdates: %v", err)
	}

	var got model.User
	_ = db.DB.First(&got, u.ID).Error
	if got.Username != "alice2" || got.Status != 2 {
		t.Fatalf("非预期 user after update: %+v", got)
	}
}

// 测试内容：验证管理员后台修改用户名时允许使用保留用户名。
func TestPrepareAndApplyUserUpdatesForAdmin_AllowsReservedUsername(t *testing.T) {
	setupTestDB(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u := model.User{Username: "alice", Password: string(hashed), Status: 1, Email: "a@example.com"}
	_ = db.DB.Create(&u).Error

	newName := "admin"
	updates, err := testService.AdminPrepareUserUpdates(u.ID, moduledto.AdminUserUpdateRequest{Username: &newName})
	if err != nil {
		t.Fatalf("AdminPrepareUserUpdates: err=%v", err)
	}
	if err := testService.AdminApplyUserUpdates(u.ID, updates); err != nil {
		t.Fatalf("AdminApplyUserUpdates: %v", err)
	}

	var got model.User
	_ = db.DB.First(&got, u.ID).Error
	if got.Username != "admin" {
		t.Fatalf("非预期 user after update: %+v", got)
	}
}

// 测试内容：验证管理员更新用户的异常分支与密码/配额更新路径。
func TestPrepareUserUpdatesForAdmin_MoreBranches(t *testing.T) {
	setupTestDB(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u := model.User{Username: "alice", Password: string(hashed), Status: 1, Email: "a@example.com"}
	_ = db.DB.Create(&u).Error

	// 无效状态
	badStatus := 9
	_, err := testService.AdminPrepareUserUpdates(u.ID, moduledto.AdminUserUpdateRequest{Status: &badStatus})
	assertServiceErrorCode(t, err, platformservice.ErrorCodeValidation)

	// 无效配额
	badQuota := int64(-2)
	_, err = testService.AdminPrepareUserUpdates(u.ID, moduledto.AdminUserUpdateRequest{StorageQuota: &badQuota})
	assertServiceErrorCode(t, err, platformservice.ErrorCodeValidation)

	// 邮箱已被占用
	u2 := model.User{Username: "bobby", Password: string(hashed), Status: 1, Email: "taken@example.com"}
	_ = db.DB.Create(&u2).Error
	newEmail := "taken@example.com"
	_, err = testService.AdminPrepareUserUpdates(u.ID, moduledto.AdminUserUpdateRequest{Email: &newEmail})
	assertServiceErrorCode(t, err, platformservice.ErrorCodeConflict)

	// 更新密码并清空配额（-1）
	newPass := "abc123456"
	clearQuota := int64(-1)
	updates, err := testService.AdminPrepareUserUpdates(u.ID, moduledto.AdminUserUpdateRequest{Password: &newPass, StorageQuota: &clearQuota})
	if err != nil {
		t.Fatalf("期望 success，实际为 err=%v", err)
	}
	if err := testService.AdminApplyUserUpdates(u.ID, updates); err != nil {
		t.Fatalf("apply: %v", err)
	}

	var got model.User
	_ = db.DB.First(&got, u.ID).Error
	if got.StorageQuota != nil {
		t.Fatalf("期望 quota cleared，实际为 %+v", got.StorageQuota)
	}
	if bcrypt.CompareHashAndPassword([]byte(got.Password), []byte(newPass)) != nil {
		t.Fatalf("期望 password updated")
	}
}

// 测试内容：验证用户名占用检测在是否包含已删除用户时的差异。
func TestIsUsernameTaken_IncludeDeleted(t *testing.T) {
	setupTestDB(t)

	u := model.User{Username: "alice", Password: "x", Status: 1, Email: "a@example.com"}
	if err := db.DB.Create(&u).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	if err := db.DB.Delete(&u).Error; err != nil {
		t.Fatalf("soft delete user: %v", err)
	}

	taken, err := testService.IsUsernameTaken("alice", nil, false)
	if err != nil {
		t.Fatalf("IsUsernameTaken: %v", err)
	}
	if taken {
		t.Fatalf("期望 username not taken when exclude deleted")
	}

	taken2, err := testService.IsUsernameTaken("alice", nil, true)
	if err != nil {
		t.Fatalf("testService.IsUsernameTaken(includeDeleted): %v", err)
	}
	if !taken2 {
		t.Fatalf("期望 username taken when include deleted")
	}
}

// 测试内容：验证邮箱占用检测支持排除指定用户 ID。
func TestIsEmailTaken_ExcludeUserID(t *testing.T) {
	setupTestDB(t)

	u1 := model.User{Username: "a1", Password: "x", Status: 1, Email: "x@example.com"}
	u2 := model.User{Username: "a2", Password: "x", Status: 1, Email: "y@example.com"}
	if err := db.DB.Create(&u1).Error; err != nil {
		t.Fatalf("创建用户1失败: %v", err)
	}
	if err := db.DB.Create(&u2).Error; err != nil {
		t.Fatalf("创建用户2失败: %v", err)
	}

	exclude := u1.ID
	taken, err := testService.IsEmailTaken("x@example.com", &exclude, true)
	if err != nil {
		t.Fatalf("IsEmailTaken: %v", err)
	}
	if taken {
		t.Fatalf("期望 email not taken when excluding matching user")
	}
}

// 测试内容：验证用户资料查询返回完整字段。
func TestGetUserProfile(t *testing.T) {
	setupTestDB(t)

	q := int64(123)
	u := model.User{
		Username:     "alice",
		Password:     "x",
		Status:       1,
		Email:        "alice@example.com",
		Admin:        true,
		Avatar:       "a.png",
		StorageQuota: &q,
		StorageUsed:  10,
	}
	_ = db.DB.Create(&u).Error

	profile, err := testService.GetUserProfile(u.ID)
	if err != nil {
		t.Fatalf("GetUserProfile: %v", err)
	}
	if profile.Username != "alice" || profile.Admin != true || profile.Avatar != "a.png" || profile.StorageUsed != 10 {
		t.Fatalf("非预期 profile: %+v", profile)
	}
}

// 测试内容：验证用户名更新的校验、冲突提示与成功后 token 生成。
func TestUpdateUsernameAndGenerateToken(t *testing.T) {
	setupTestDB(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u := model.User{Username: "alice", Password: string(hashed), Status: 1, Email: "a@example.com"}
	_ = db.DB.Create(&u).Error

	token, err := testService.UpdateUsernameAndGenerateToken(u.ID, "ab", false)
	if token != "" {
		t.Fatalf("期望 token 为空，实际为 %q", token)
	}
	assertServiceErrorCode(t, err, platformservice.ErrorCodeValidation)

	u2 := model.User{Username: "bobby", Password: string(hashed), Status: 1, Email: "b@example.com"}
	_ = db.DB.Create(&u2).Error
	_, err = testService.UpdateUsernameAndGenerateToken(u.ID, "bobby", false)
	if serviceErr := assertServiceErrorCode(t, err, platformservice.ErrorCodeConflict); serviceErr.Message != "用户名已存在" {
		t.Fatalf("期望 conflict message，实际为 %q", serviceErr.Message)
	}

	_, err = testService.UpdateUsernameAndGenerateToken(u.ID, "admin", false)
	assertServiceErrorCode(t, err, platformservice.ErrorCodeValidation)

	_, err = testService.UpdateUsernameAndGenerateToken(u.ID, "root", true)
	assertServiceErrorCode(t, err, platformservice.ErrorCodeValidation)

}

// 测试内容：验证通过旧密码更新密码的校验与成功路径。
func TestUpdatePasswordByOldPassword(t *testing.T) {
	setupTestDB(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u := model.User{Username: "alice", Password: string(hashed), Status: 1, Email: "a@example.com"}
	_ = db.DB.Create(&u).Error

	err := testService.UpdatePasswordByOldPassword(u.ID, "abc12345", "short")
	assertServiceErrorCode(t, err, platformservice.ErrorCodeValidation)

	err = testService.UpdatePasswordByOldPassword(u.ID, "wrong", "abc123456")
	if serviceErr := assertServiceErrorCode(t, err, platformservice.ErrorCodeValidation); serviceErr.Message != "旧密码错误" {
		t.Fatalf("期望 old password message，实际为 %q", serviceErr.Message)
	}

	err = testService.UpdatePasswordByOldPassword(u.ID, "abc12345", "abc123456")
	if err != nil {
		t.Fatalf("期望 success，实际为 err=%v", err)
	}

	var got model.User
	_ = db.DB.First(&got, u.ID).Error
	if bcrypt.CompareHashAndPassword([]byte(got.Password), []byte("abc123456")) != nil {
		t.Fatalf("期望 password to be updated")
	}
}
