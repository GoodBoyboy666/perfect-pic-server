package service

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"

	"golang.org/x/crypto/bcrypt"
)

func TestListSettingsForAdmin_MasksSensitive(t *testing.T) {
	setupTestDB(t)

	_ = db.DB.Create(&model.Setting{Key: "k1", Value: "v1", Sensitive: false}).Error
	_ = db.DB.Create(&model.Setting{Key: "k2", Value: "secret", Sensitive: true}).Error

	settings, err := ListSettingsForAdmin()
	if err != nil {
		t.Fatalf("ListSettingsForAdmin: %v", err)
	}

	m := map[string]string{}
	for _, s := range settings {
		m[s.Key] = s.Value
	}
	if m["k1"] != "v1" {
		t.Fatalf("expected k1=v1, got %q", m["k1"])
	}
	if m["k2"] != "**********" {
		t.Fatalf("expected sensitive masked, got %q", m["k2"])
	}
}

func TestGetServerStatsForAdminAndUserDetail(t *testing.T) {
	setupTestDB(t)

	u := model.User{Username: "alice", Password: "x", Status: 1, Email: "a@example.com"}
	_ = db.DB.Create(&u).Error
	_ = db.DB.Create(&model.Image{Filename: "a.png", Path: "2026/02/13/a.png", Size: 10, Width: 1, Height: 1, MimeType: ".png", UploadedAt: 1, UserID: u.ID}).Error

	stats, err := GetServerStatsForAdmin()
	if err != nil {
		t.Fatalf("GetServerStatsForAdmin: %v", err)
	}
	if stats.ImageCount != 1 || stats.StorageUsage != 10 || stats.UserCount != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}

	got, err := GetUserDetailForAdmin(u.ID)
	if err != nil {
		t.Fatalf("GetUserDetailForAdmin: %v", err)
	}
	if got.Username != "alice" {
		t.Fatalf("unexpected user: %+v", got)
	}
}

func TestUpdateSettingsForAdmin_MaskedSensitiveIsNotOverwritten(t *testing.T) {
	setupTestDB(t)

	_ = db.DB.Create(&model.Setting{Key: "s1", Value: "secret", Sensitive: true}).Error
	_ = db.DB.Create(&model.Setting{Key: "n1", Value: "old", Sensitive: false}).Error

	err := UpdateSettingsForAdmin([]UpdateSettingPayload{
		{Key: "s1", Value: "**********"}, // should be ignored
		{Key: "n1", Value: "**********"}, // should overwrite (not sensitive)
		{Key: "new", Value: "val"},       // upsert
	})
	if err != nil {
		t.Fatalf("UpdateSettingsForAdmin: %v", err)
	}

	var s1 model.Setting
	_ = db.DB.Where("key = ?", "s1").First(&s1).Error
	if s1.Value != "secret" {
		t.Fatalf("expected sensitive value preserved, got %q", s1.Value)
	}
	var n1 model.Setting
	_ = db.DB.Where("key = ?", "n1").First(&n1).Error
	if n1.Value != "**********" {
		t.Fatalf("expected non-sensitive overwritten, got %q", n1.Value)
	}
	var n model.Setting
	_ = db.DB.Where("key = ?", "new").First(&n).Error
	if n.Value != "val" {
		t.Fatalf("expected new=val, got %q", n.Value)
	}
}

func TestListUsersForAdmin_FilterAndShowDeleted(t *testing.T) {
	setupTestDB(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u1 := model.User{Username: "alice", Password: string(hashed), Status: 1, Email: "a@example.com"}
	u2 := model.User{Username: "bob", Password: string(hashed), Status: 1, Email: "b@example.com"}
	_ = db.DB.Create(&u1).Error
	_ = db.DB.Create(&u2).Error
	_ = db.DB.Delete(&u2).Error

	users, total, err := ListUsersForAdmin(AdminUserListParams{Page: 1, PageSize: 10, Keyword: "ali"})
	if err != nil {
		t.Fatalf("ListUsersForAdmin: %v", err)
	}
	if total != 1 || len(users) != 1 || users[0].Username != "alice" {
		t.Fatalf("unexpected: total=%d len=%d users=%v", total, len(users), users)
	}

	users2, total2, err := ListUsersForAdmin(AdminUserListParams{Page: 1, PageSize: 10, ShowDeleted: true})
	if err != nil {
		t.Fatalf("ListUsersForAdmin(showDeleted): %v", err)
	}
	if total2 != 2 || len(users2) != 2 {
		t.Fatalf("expected 2 users with deleted, got total=%d len=%d", total2, len(users2))
	}
}

func TestCreateUserForAdmin_Validates(t *testing.T) {
	setupTestDB(t)

	user, msg, err := CreateUserForAdmin(AdminCreateUserInput{Username: "ab", Password: "abc12345"})
	if err != nil || msg == "" || user != nil {
		t.Fatalf("expected username validation msg, got user=%v msg=%q err=%v", user, msg, err)
	}

	user, msg, err = CreateUserForAdmin(AdminCreateUserInput{Username: "alice", Password: "short"})
	if err != nil || msg == "" || user != nil {
		t.Fatalf("expected password validation msg, got user=%v msg=%q err=%v", user, msg, err)
	}
}

func TestCreateUserForAdmin_SuccessOptions(t *testing.T) {
	setupTestDB(t)

	email := "a@example.com"
	emailVerified := true
	quota := int64(100)
	status := 2

	user, msg, err := CreateUserForAdmin(AdminCreateUserInput{
		Username:      "alice_1",
		Password:      "abc12345",
		Email:         &email,
		EmailVerified: &emailVerified,
		StorageQuota:  &quota,
		Status:        &status,
	})
	if err != nil || msg != "" || user == nil {
		t.Fatalf("expected success, got user=%v msg=%q err=%v", user, msg, err)
	}
	if user.Email != email || user.EmailVerified != true || user.Status != 2 {
		t.Fatalf("unexpected created user: %+v", user)
	}
	if user.StorageQuota == nil || *user.StorageQuota != 100 {
		t.Fatalf("unexpected quota: %+v", user.StorageQuota)
	}

	// StorageQuota=-1 should set nil.
	q2 := int64(-1)
	user2, msg, err := CreateUserForAdmin(AdminCreateUserInput{
		Username:     "alice_2",
		Password:     "abc12345",
		StorageQuota: &q2,
	})
	if err != nil || msg != "" || user2 == nil {
		t.Fatalf("expected success, got user=%v msg=%q err=%v", user2, msg, err)
	}
	if user2.StorageQuota != nil {
		t.Fatalf("expected nil quota for -1")
	}
}

func TestPrepareAndApplyUserUpdatesForAdmin(t *testing.T) {
	setupTestDB(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u := model.User{Username: "alice", Password: string(hashed), Status: 1, Email: "a@example.com"}
	_ = db.DB.Create(&u).Error

	newName := "alice2"
	newStatus := 2
	updates, msg, err := PrepareUserUpdatesForAdmin(u.ID, AdminUserUpdateInput{Username: &newName, Status: &newStatus})
	if err != nil || msg != "" {
		t.Fatalf("PrepareUserUpdatesForAdmin: msg=%q err=%v", msg, err)
	}
	if err := ApplyUserUpdatesForAdmin(u.ID, updates); err != nil {
		t.Fatalf("ApplyUserUpdatesForAdmin: %v", err)
	}

	var got model.User
	_ = db.DB.First(&got, u.ID).Error
	if got.Username != "alice2" || got.Status != 2 {
		t.Fatalf("unexpected user after update: %+v", got)
	}
}

func TestPrepareUserUpdatesForAdmin_MoreBranches(t *testing.T) {
	setupTestDB(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u := model.User{Username: "alice", Password: string(hashed), Status: 1, Email: "a@example.com"}
	_ = db.DB.Create(&u).Error

	// invalid status
	badStatus := 9
	_, msg, err := PrepareUserUpdatesForAdmin(u.ID, AdminUserUpdateInput{Status: &badStatus})
	if err != nil || msg == "" {
		t.Fatalf("expected status validation msg, got msg=%q err=%v", msg, err)
	}

	// invalid quota
	badQuota := int64(-2)
	_, msg, err = PrepareUserUpdatesForAdmin(u.ID, AdminUserUpdateInput{StorageQuota: &badQuota})
	if err != nil || msg == "" {
		t.Fatalf("expected quota validation msg, got msg=%q err=%v", msg, err)
	}

	// email taken
	u2 := model.User{Username: "bobby", Password: string(hashed), Status: 1, Email: "taken@example.com"}
	_ = db.DB.Create(&u2).Error
	newEmail := "taken@example.com"
	_, msg, err = PrepareUserUpdatesForAdmin(u.ID, AdminUserUpdateInput{Email: &newEmail})
	if err != nil || msg == "" {
		t.Fatalf("expected email taken msg, got msg=%q err=%v", msg, err)
	}

	// update password + clear quota (-1)
	newPass := "abc123456"
	clearQuota := int64(-1)
	updates, msg, err := PrepareUserUpdatesForAdmin(u.ID, AdminUserUpdateInput{Password: &newPass, StorageQuota: &clearQuota})
	if err != nil || msg != "" {
		t.Fatalf("expected success, got msg=%q err=%v", msg, err)
	}
	if err := ApplyUserUpdatesForAdmin(u.ID, updates); err != nil {
		t.Fatalf("apply: %v", err)
	}

	var got model.User
	_ = db.DB.First(&got, u.ID).Error
	if got.StorageQuota != nil {
		t.Fatalf("expected quota cleared, got %+v", got.StorageQuota)
	}
	if bcrypt.CompareHashAndPassword([]byte(got.Password), []byte(newPass)) != nil {
		t.Fatalf("expected password updated")
	}
}

func TestDeleteUserForAdmin_SoftDelete(t *testing.T) {
	setupTestDB(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u := model.User{Username: "alice", Password: string(hashed), Status: 1, Email: "a@example.com"}
	_ = db.DB.Create(&u).Error

	if err := DeleteUserForAdmin(u.ID, false); err != nil {
		t.Fatalf("DeleteUserForAdmin: %v", err)
	}

	var got model.User
	if err := db.DB.Unscoped().First(&got, u.ID).Error; err != nil {
		t.Fatalf("load deleted user: %v", err)
	}
	if got.Status != 3 {
		t.Fatalf("expected status 3, got %d", got.Status)
	}
	if got.Username == "alice" || got.Email == "a@example.com" {
		t.Fatalf("expected unique fields to be rewritten, got username=%q email=%q", got.Username, got.Email)
	}
}

func TestDeleteUserForAdmin_HardDeleteAlsoDeletesFiles(t *testing.T) {
	setupTestDB(t)

	tmp := t.TempDir()
	oldwd, _ := os.Getwd()
	_ = os.Chdir(tmp)
	defer func() { _ = os.Chdir(oldwd) }()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u := model.User{Username: "alice", Password: string(hashed), Status: 1, Email: "a@example.com"}
	_ = db.DB.Create(&u).Error

	// Create avatar file under uploads/avatars/{uid}/a.txt
	realAvatarDir := filepath.Join("uploads", "avatars", strconv.FormatUint(uint64(u.ID), 10))
	_ = os.MkdirAll(realAvatarDir, 0755)
	_ = os.WriteFile(filepath.Join(realAvatarDir, "a.txt"), []byte("x"), 0644)

	// Create image record + physical file under uploads/imgs/{path}
	imgRel := "2026/02/13/a.png"
	imgLocal := filepath.FromSlash(imgRel)
	imgFile := filepath.Join("uploads", "imgs", imgLocal)
	_ = os.MkdirAll(filepath.Dir(imgFile), 0755)
	_ = os.WriteFile(imgFile, []byte{0x89, 0x50, 0x4E, 0x47}, 0644)
	_ = db.DB.Create(&model.Image{Filename: "a.png", Path: imgRel, Size: 4, Width: 1, Height: 1, MimeType: ".png", UploadedAt: 1, UserID: u.ID}).Error

	if err := DeleteUserForAdmin(u.ID, true); err != nil {
		t.Fatalf("hard delete: %v", err)
	}

	if err := db.DB.Unscoped().First(&model.User{}, u.ID).Error; err == nil {
		t.Fatalf("expected user to be hard-deleted")
	}
	if _, err := os.Stat(realAvatarDir); !os.IsNotExist(err) {
		t.Fatalf("expected avatar dir deleted, err=%v", err)
	}
	if _, err := os.Stat(imgFile); !os.IsNotExist(err) {
		t.Fatalf("expected image file deleted, err=%v", err)
	}
}
