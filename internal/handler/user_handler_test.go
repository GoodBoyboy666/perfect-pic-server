package handler

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/testutils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// 测试内容：验证获取当前用户信息接口返回成功。
func TestGetSelfInfo_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	u := model.User{Username: "alice", Password: "x", Status: 1, Email: "a@example.com"}
	_ = db.DB.Create(&u).Error

	r := gin.New()
	r.GET("/me", func(c *gin.Context) { c.Set("id", u.ID); c.Next() }, testHandler.GetSelfInfo)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/me", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际为 %d", w.Code)
	}
}

// 测试内容：验证更新用户名的非法与合法场景。
func TestUpdateSelfUsername_ValidAndInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u := model.User{Username: "alice", Password: string(hashed), Status: 1, Email: "a@example.com"}
	_ = db.DB.Create(&u).Error

	r := gin.New()
	r.PATCH("/username", func(c *gin.Context) { c.Set("id", u.ID); c.Set("admin", true); c.Next() }, testHandler.UpdateSelfUsername)

	bodyBad, _ := json.Marshal(gin.H{"username": "ab"})
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, httptest.NewRequest(http.MethodPatch, "/username", bytes.NewReader(bodyBad)))
	if w1.Code != http.StatusBadRequest {
		t.Fatalf("期望 400，实际为 %d body=%s", w1.Code, w1.Body.String())
	}

	bodyOK, _ := json.Marshal(gin.H{"username": "alice2"})
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httptest.NewRequest(http.MethodPatch, "/username", bytes.NewReader(bodyOK)))
	if w2.Code != http.StatusOK {
		t.Fatalf("期望 200，实际为 %d body=%s", w2.Code, w2.Body.String())
	}
}

// 测试内容：验证更新密码的错误与成功路径。
func TestUpdateSelfPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u := model.User{Username: "alice", Password: string(hashed), Status: 1, Email: "a@example.com"}
	_ = db.DB.Create(&u).Error

	r := gin.New()
	r.PATCH("/password", func(c *gin.Context) { c.Set("id", u.ID); c.Next() }, testHandler.UpdateSelfPassword)

	bodyBad, _ := json.Marshal(gin.H{"old_password": "wrong", "new_password": "abc123456"})
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, httptest.NewRequest(http.MethodPatch, "/password", bytes.NewReader(bodyBad)))
	if w1.Code != http.StatusBadRequest {
		t.Fatalf("期望 400，实际为 %d body=%s", w1.Code, w1.Body.String())
	}

	bodyOK, _ := json.Marshal(gin.H{"old_password": "abc12345", "new_password": "abc123456"})
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httptest.NewRequest(http.MethodPatch, "/password", bytes.NewReader(bodyOK)))
	if w2.Code != http.StatusOK {
		t.Fatalf("期望 200，实际为 %d body=%s", w2.Code, w2.Body.String())
	}
}

// 测试内容：验证更新头像成功并写入头像文件。
func TestUpdateSelfAvatar_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	tmp := t.TempDir()
	oldwd, _ := os.Getwd()
	_ = os.Chdir(tmp)
	defer func() { _ = os.Chdir(oldwd) }()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u := model.User{Username: "alice", Password: string(hashed), Status: 1, Email: "a@example.com"}
	_ = db.DB.Create(&u).Error

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, _ := w.CreateFormFile("file", "a.png")
	_, _ = part.Write(testutils.MinimalPNG())
	_ = w.Close()

	r := gin.New()
	r.PATCH("/avatar", func(c *gin.Context) { c.Set("id", u.ID); c.Next() }, testHandler.UpdateSelfAvatar)

	req := httptest.NewRequest(http.MethodPatch, "/avatar", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("期望 200，实际为 %d body=%s", rec.Code, rec.Body.String())
	}

	// 确保文件已在 uploads/avatars/{uid} 下创建
	userDir := filepath.Join("uploads", "avatars", strconv.FormatUint(uint64(u.ID), 10))
	entries, err := os.ReadDir(userDir)
	if err != nil || len(entries) == 0 {
		t.Fatalf("期望 avatar file created in %q: err=%v entries=%d", userDir, err, len(entries))
	}
}

// 测试内容：验证请求修改邮箱时密码错误返回 403。
func TestRequestUpdateEmailHandler_ForbiddenOnWrongPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u := model.User{Username: "alice", Password: string(hashed), Status: 1, Email: "a@example.com"}
	_ = db.DB.Create(&u).Error

	r := gin.New()
	r.POST("/email", func(c *gin.Context) { c.Set("id", u.ID); c.Next() }, testHandler.RequestUpdateEmail)

	body, _ := json.Marshal(gin.H{"password": "wrong", "new_email": "new@example.com"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/email", bytes.NewReader(body)))
	if w.Code != http.StatusForbidden {
		t.Fatalf("期望 403，实际为 %d body=%s", w.Code, w.Body.String())
	}
}

// 测试内容：验证获取用户图片数量接口返回正确结果。
func TestGetSelfImagesCountHandler_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	u := model.User{Username: "alice", Password: "x", Status: 1, Email: "a@example.com"}
	_ = db.DB.Create(&u).Error
	_ = db.DB.Create(&model.Image{Filename: "a.png", Path: "2026/02/13/a.png", Size: 1, Width: 1, Height: 1, MimeType: ".png", UploadedAt: 1, UserID: u.ID}).Error

	r := gin.New()
	r.GET("/count", func(c *gin.Context) { c.Set("id", u.ID); c.Next() }, testHandler.GetSelfImagesCount)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/count", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际为 %d body=%s", w.Code, w.Body.String())
	}
}

// 测试内容：验证 Passkey 注册开始接口返回挑战。
func TestBeginPasskeyRegistrationHandler_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	u := model.User{Username: "alice", Password: "x", Status: 1, Email: "a@example.com"}
	_ = db.DB.Create(&u).Error

	r := gin.New()
	r.POST("/passkeys/register/start", func(c *gin.Context) { c.Set("id", u.ID); c.Next() }, testHandler.BeginPasskeyRegistration)

	body, _ := json.Marshal(gin.H{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/passkeys/register/start", bytes.NewReader(body)))
	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际为 %d body=%s", w.Code, w.Body.String())
	}
}

// 测试内容：验证 Passkey 达到上限时注册开始接口返回 409。
func TestBeginPasskeyRegistrationHandler_ConflictWhenLimitExceeded(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	u := model.User{Username: "alice", Password: "x", Status: 1, Email: "a@example.com"}
	_ = db.DB.Create(&u).Error
	for i := 0; i < 10; i++ {
		_ = db.DB.Create(&model.PasskeyCredential{
			UserID:       u.ID,
			CredentialID: "cred_limit_" + string(rune('a'+i)),
			Credential:   `{}`,
		}).Error
	}

	r := gin.New()
	r.POST("/passkeys/register/start", func(c *gin.Context) { c.Set("id", u.ID); c.Next() }, testHandler.BeginPasskeyRegistration)

	body, _ := json.Marshal(gin.H{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/passkeys/register/start", bytes.NewReader(body)))
	if w.Code != http.StatusConflict {
		t.Fatalf("期望 409，实际为 %d body=%s", w.Code, w.Body.String())
	}
}

// 测试内容：验证 Passkey 注册完成接口在无效会话时返回 400。
func TestFinishPasskeyRegistrationHandler_InvalidSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	u := model.User{Username: "alice", Password: "x", Status: 1, Email: "a@example.com"}
	_ = db.DB.Create(&u).Error

	r := gin.New()
	r.POST("/passkeys/register/finish", func(c *gin.Context) { c.Set("id", u.ID); c.Next() }, testHandler.FinishPasskeyRegistration)

	body, _ := json.Marshal(gin.H{
		"session_id": "bad-session",
		"credential": gin.H{},
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/passkeys/register/finish", bytes.NewReader(body)))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望 400，实际为 %d body=%s", w.Code, w.Body.String())
	}
}

// 测试内容：验证用户获取 Passkey 列表接口返回成功。
func TestListSelfPasskeysHandler_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	u := model.User{Username: "alice", Password: "x", Status: 1, Email: "a@example.com"}
	_ = db.DB.Create(&u).Error
	_ = db.DB.Create(&model.PasskeyCredential{
		UserID:       u.ID,
		CredentialID: "cred_1",
		Name:         "Office Key",
		Credential:   `{}`,
	}).Error

	r := gin.New()
	r.GET("/passkeys", func(c *gin.Context) { c.Set("id", u.ID); c.Next() }, testHandler.ListSelfPasskeys)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/passkeys", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际为 %d body=%s", w.Code, w.Body.String())
	}

	var resp struct {
		List []map[string]any `json:"list"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	if len(resp.List) != 1 {
		t.Fatalf("期望 1 条记录，实际为 %d", len(resp.List))
	}
	item := resp.List[0]
	if item["id"] == nil || item["credential_id"] == nil || item["created_at"] == nil || item["name"] == nil {
		t.Fatalf("缺少必要字段，实际为 %+v", item)
	}
	if _, exists := item["updated_at"]; exists {
		t.Fatalf("不应返回 updated_at 字段，实际为 %+v", item)
	}
}

// 测试内容：验证用户删除自己的 Passkey 接口返回成功。
func TestDeleteSelfPasskeyHandler_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	u := model.User{Username: "alice", Password: "x", Status: 1, Email: "a@example.com"}
	_ = db.DB.Create(&u).Error
	pk := model.PasskeyCredential{
		UserID:       u.ID,
		CredentialID: "cred_del",
		Credential:   `{}`,
	}
	_ = db.DB.Create(&pk).Error

	r := gin.New()
	r.DELETE("/passkeys/:id", func(c *gin.Context) { c.Set("id", u.ID); c.Next() }, testHandler.DeleteSelfPasskey)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/passkeys/"+strconv.FormatUint(uint64(pk.ID), 10), nil))
	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际为 %d body=%s", w.Code, w.Body.String())
	}
}

// 测试内容：验证删除不存在的 Passkey 返回 404。
func TestDeleteSelfPasskeyHandler_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	u := model.User{Username: "alice", Password: "x", Status: 1, Email: "a@example.com"}
	_ = db.DB.Create(&u).Error

	r := gin.New()
	r.DELETE("/passkeys/:id", func(c *gin.Context) { c.Set("id", u.ID); c.Next() }, testHandler.DeleteSelfPasskey)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/passkeys/999", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("期望 404，实际为 %d body=%s", w.Code, w.Body.String())
	}
}

// 测试内容：验证用户可修改自己的 Passkey 名称。
func TestUpdateSelfPasskeyNameHandler_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	u := model.User{Username: "alice", Password: "x", Status: 1, Email: "a@example.com"}
	_ = db.DB.Create(&u).Error
	pk := model.PasskeyCredential{
		UserID:       u.ID,
		CredentialID: "cred_name_1",
		Name:         "旧名称",
		Credential:   `{}`,
	}
	_ = db.DB.Create(&pk).Error

	r := gin.New()
	r.PATCH("/passkeys/:id/name", func(c *gin.Context) { c.Set("id", u.ID); c.Next() }, testHandler.UpdateSelfPasskeyName)

	body, _ := json.Marshal(gin.H{"name": "My iPhone"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/passkeys/"+strconv.FormatUint(uint64(pk.ID), 10)+"/name", bytes.NewReader(body)))
	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际为 %d body=%s", w.Code, w.Body.String())
	}

	var got model.PasskeyCredential
	if err := db.DB.First(&got, pk.ID).Error; err != nil {
		t.Fatalf("查询 Passkey 失败: %v", err)
	}
	if got.Name != "My iPhone" {
		t.Fatalf("期望名称被更新，实际为 %q", got.Name)
	}
}

// 测试内容：验证修改 Passkey 名称时参数错误返回 400。
func TestUpdateSelfPasskeyNameHandler_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	u := model.User{Username: "alice", Password: "x", Status: 1, Email: "a@example.com"}
	_ = db.DB.Create(&u).Error

	r := gin.New()
	r.PATCH("/passkeys/:id/name", func(c *gin.Context) { c.Set("id", u.ID); c.Next() }, testHandler.UpdateSelfPasskeyName)

	body, _ := json.Marshal(gin.H{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/passkeys/1/name", bytes.NewReader(body)))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望 400，实际为 %d body=%s", w.Code, w.Body.String())
	}
}
