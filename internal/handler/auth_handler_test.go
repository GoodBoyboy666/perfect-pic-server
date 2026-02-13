package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/service"
	"perfect-pic-server/internal/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func TestLoginHandler_SuccessAndUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	// Disable captcha in tests.
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaProvider, Value: ""}).Error
	service.ClearCache()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u := model.User{Username: "alice", Password: string(hashed), Status: 1, Email: "a@example.com", EmailVerified: true}
	_ = db.DB.Create(&u).Error

	r := gin.New()
	r.POST("/login", Login)

	body, _ := json.Marshal(gin.H{"username": "alice", "password": "abc12345"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body)))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	var okResp struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &okResp)
	if okResp.Token == "" {
		t.Fatalf("expected token")
	}
	if _, err := utils.ParseLoginToken(okResp.Token); err != nil {
		t.Fatalf("token parse error: %v", err)
	}

	body2, _ := json.Marshal(gin.H{"username": "alice", "password": "wrongpass1"})
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body2)))
	if w2.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", w2.Code, w2.Body.String())
	}
}

func TestLoginHandler_BindError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	r := gin.New()
	r.POST("/login", Login)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader([]byte("{bad"))))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestRegisterHandler_ForbiddenWhenDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaProvider, Value: ""}).Error
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigAllowRegister, Value: "false"}).Error
	service.ClearCache()

	r := gin.New()
	r.POST("/register", Register)

	body, _ := json.Marshal(gin.H{"username": "alice", "password": "abc12345", "email": "a@example.com"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body)))
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestEmailVerifyHandler_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u := model.User{Username: "alice", Password: string(hashed), Status: 1, Email: "a@example.com", EmailVerified: false}
	_ = db.DB.Create(&u).Error

	token, err := utils.GenerateEmailToken(u.ID, u.Email, time.Hour)
	if err != nil {
		t.Fatalf("GenerateEmailToken: %v", err)
	}

	r := gin.New()
	r.POST("/auth/email-verify", EmailVerify)

	body, _ := json.Marshal(gin.H{"token": token})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/auth/email-verify", bytes.NewReader(body)))
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status %d body=%s", w.Code, w.Body.String())
	}
}

func TestRegisterHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaProvider, Value: ""}).Error
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigAllowRegister, Value: "true"}).Error
	service.ClearCache()

	r := gin.New()
	r.POST("/register", Register)

	body, _ := json.Marshal(gin.H{"username": "alice_1", "password": "abc12345", "email": "a1@example.com"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body)))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestRequestPasswordResetHandler_Always200ForUnknownEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaProvider, Value: ""}).Error
	service.ClearCache()

	r := gin.New()
	r.POST("/auth/password/reset/request", RequestPasswordReset)

	body, _ := json.Marshal(gin.H{"email": "unknown@example.com"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/auth/password/reset/request", bytes.NewReader(body)))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestRequestPasswordResetHandler_BindError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	r := gin.New()
	r.POST("/auth/password/reset/request", RequestPasswordReset)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/auth/password/reset/request", bytes.NewReader([]byte("{bad"))))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestResetPasswordHandler_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u := model.User{Username: "alice", Password: string(hashed), Status: 1, Email: "a@example.com"}
	_ = db.DB.Create(&u).Error

	token, err := service.GenerateForgetPasswordToken(u.ID)
	if err != nil {
		t.Fatalf("GenerateForgetPasswordToken: %v", err)
	}

	r := gin.New()
	r.POST("/auth/password/reset", ResetPassword)

	body, _ := json.Marshal(gin.H{"token": token, "new_password": "abc123456"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/auth/password/reset", bytes.NewReader(body)))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestResetPasswordHandler_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	r := gin.New()
	r.POST("/auth/password/reset", ResetPassword)

	body, _ := json.Marshal(gin.H{"token": "bad", "new_password": "abc123456"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/auth/password/reset", bytes.NewReader(body)))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestEmailChangeVerifyHandler_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u := model.User{Username: "alice", Password: string(hashed), Status: 1, Email: "old@example.com", EmailVerified: true}
	_ = db.DB.Create(&u).Error

	token, err := utils.GenerateEmailChangeToken(u.ID, "old@example.com", "new@example.com", time.Hour)
	if err != nil {
		t.Fatalf("GenerateEmailChangeToken: %v", err)
	}

	r := gin.New()
	r.POST("/auth/email-change-verify", EmailChangeVerify)

	body, _ := json.Marshal(gin.H{"token": token})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/auth/email-change-verify", bytes.NewReader(body)))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestEmailChangeVerifyHandler_BindError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	r := gin.New()
	r.POST("/auth/email-change-verify", EmailChangeVerify)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/auth/email-change-verify", bytes.NewReader([]byte("{bad"))))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGetRegisterStateHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	_ = db.DB.Save(&model.Setting{Key: consts.ConfigAllowRegister, Value: "false"}).Error
	service.ClearCache()

	r := gin.New()
	r.GET("/register", GetRegisterState)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/register", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
}
