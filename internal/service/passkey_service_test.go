package service

import (
	"encoding/json"
	platformservice "perfect-pic-server/internal/common"
	"testing"

	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"

	"github.com/go-webauthn/webauthn/webauthn"
)

// 测试内容：验证用户可获取自己的 Passkey 列表。
func TestListUserPasskeys_Success(t *testing.T) {
	setupTestDB(t)

	u := model.User{
		Username:      "alice",
		Password:      "x",
		Status:        1,
		Email:         "a@example.com",
		EmailVerified: true,
	}
	if err := db.DB.Create(&u).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	record := model.PasskeyCredential{
		UserID:       u.ID,
		CredentialID: "cred_1",
		Name:         "MacBook Pro",
		Credential:   mustMarshalPasskeyCredentialForTest(t, webauthn.Credential{ID: []byte{1, 2, 3}}),
	}
	if err := db.DB.Create(&record).Error; err != nil {
		t.Fatalf("创建 Passkey 失败: %v", err)
	}

	list, err := testService.ListUserPasskeys(u.ID)
	if err != nil {
		t.Fatalf("ListUserPasskeys 返回错误: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("期望 1 条 Passkey，实际为 %d", len(list))
	}
	if list[0].ID != record.ID || list[0].CredentialID != "cred_1" || list[0].Name != "MacBook Pro" {
		t.Fatalf("非预期 Passkey 记录: %+v", list[0])
	}
	if list[0].CreatedAt == 0 {
		t.Fatalf("期望 created_at 非 0，实际为 %+v", list[0])
	}
}

// 测试内容：验证用户可删除自己的 Passkey。
func TestDeleteUserPasskey_Success(t *testing.T) {
	setupTestDB(t)

	u := model.User{Username: "alice", Password: "x", Status: 1, Email: "a@example.com"}
	if err := db.DB.Create(&u).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	record := model.PasskeyCredential{
		UserID:       u.ID,
		CredentialID: "cred_del",
		Credential:   mustMarshalPasskeyCredentialForTest(t, webauthn.Credential{ID: []byte{9, 9, 9}}),
	}
	if err := db.DB.Create(&record).Error; err != nil {
		t.Fatalf("创建 Passkey 失败: %v", err)
	}

	if err := testService.DeleteUserPasskey(u.ID, record.ID); err != nil {
		t.Fatalf("DeleteUserPasskey 返回错误: %v", err)
	}

	var count int64
	if err := db.DB.Model(&model.PasskeyCredential{}).Where("id = ?", record.ID).Count(&count).Error; err != nil {
		t.Fatalf("查询 Passkey 失败: %v", err)
	}
	if count != 0 {
		t.Fatalf("期望 Passkey 被删除，实际 count=%d", count)
	}
}

// 测试内容：验证删除不存在的 Passkey 返回 not_found。
func TestDeleteUserPasskey_NotFound(t *testing.T) {
	setupTestDB(t)

	err := testService.DeleteUserPasskey(1, 999)
	if err == nil {
		t.Fatalf("期望返回错误")
	}

	serviceErr, ok := platformservice.AsServiceError(err)
	if !ok || serviceErr.Code != platformservice.ErrorCodeNotFound {
		t.Fatalf("期望 not_found 错误，实际为: %#v (%v)", serviceErr, err)
	}
}

// 测试内容：验证用户可更新自己的 Passkey 名称。
func TestUpdateUserPasskeyName_Success(t *testing.T) {
	setupTestDB(t)

	u := model.User{Username: "alice", Password: "x", Status: 1, Email: "a@example.com"}
	if err := db.DB.Create(&u).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	record := model.PasskeyCredential{
		UserID:       u.ID,
		CredentialID: "cred_name_1",
		Name:         "旧名称",
		Credential:   mustMarshalPasskeyCredentialForTest(t, webauthn.Credential{ID: []byte{1, 9, 9}}),
	}
	if err := db.DB.Create(&record).Error; err != nil {
		t.Fatalf("创建 Passkey 失败: %v", err)
	}

	if err := testService.UpdateUserPasskeyName(u.ID, record.ID, "  iPhone  "); err != nil {
		t.Fatalf("UpdateUserPasskeyName 返回错误: %v", err)
	}

	var got model.PasskeyCredential
	if err := db.DB.First(&got, record.ID).Error; err != nil {
		t.Fatalf("查询 Passkey 失败: %v", err)
	}
	if got.Name != "iPhone" {
		t.Fatalf("期望名称为 iPhone，实际为 %q", got.Name)
	}
}

// 测试内容：验证更新 Passkey 名称时名称为空返回校验错误。
func TestUpdateUserPasskeyName_Validation(t *testing.T) {
	setupTestDB(t)

	err := testService.UpdateUserPasskeyName(1, 1, "   ")
	if err == nil {
		t.Fatalf("期望返回错误")
	}

	serviceErr, ok := platformservice.AsServiceError(err)
	if !ok || serviceErr.Code != platformservice.ErrorCodeValidation {
		t.Fatalf("期望 validation 错误，实际为: %#v (%v)", serviceErr, err)
	}
}

// 测试内容：验证更新不存在的 Passkey 名称返回 not_found。
func TestUpdateUserPasskeyName_NotFound(t *testing.T) {
	setupTestDB(t)

	err := testService.UpdateUserPasskeyName(1, 999, "My Passkey")
	if err == nil {
		t.Fatalf("期望返回错误")
	}

	serviceErr, ok := platformservice.AsServiceError(err)
	if !ok || serviceErr.Code != platformservice.ErrorCodeNotFound {
		t.Fatalf("期望 not_found 错误，实际为: %#v (%v)", serviceErr, err)
	}
}

func mustMarshalPasskeyCredentialForTest(t *testing.T, credential webauthn.Credential) string {
	t.Helper()
	raw, err := json.Marshal(credential)
	if err != nil {
		t.Fatalf("marshal credential: %v", err)
	}
	return string(raw)
}
