package app

import (
	"perfect-pic-server/internal/common/httpx"
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/utils"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func TestAuthUseCase_LoginUser_WrongPasswordUnauthorized(t *testing.T) {
	f := setupAppFixture(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u := model.User{
		Username:      "alice",
		Password:      string(hashed),
		Status:        1,
		Email:         "alice@example.com",
		EmailVerified: true,
	}
	if err := db.DB.Create(&u).Error; err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	_, err := f.authUC.LoginUser("alice", "wrongpass")
	assertAuthErrorCode(t, err, httpx.AuthErrorUnauthorized)
}

func TestAuthUseCase_RegisterUser_ForbiddenWhenNotInitialized(t *testing.T) {
	f := setupAppFixture(t)

	err := f.authUC.RegisterUser("alice_1", "abc12345", "alice@example.com")
	assertAuthErrorCode(t, err, httpx.AuthErrorForbidden)
}

func TestAuthUseCase_RegisterUser_SuccessCreatesUser(t *testing.T) {
	f := setupAppFixture(t)
	f.initializeSystem(t)

	if err := f.authUC.RegisterUser("alice_1", "abc12345", "alice@example.com"); err != nil {
		t.Fatalf("RegisterUser failed: %v", err)
	}

	var got model.User
	if err := db.DB.Where("username = ?", "alice_1").First(&got).Error; err != nil {
		t.Fatalf("load created user failed: %v", err)
	}
	if got.EmailVerified {
		t.Fatalf("expected email_verified=false, got true")
	}
	if bcrypt.CompareHashAndPassword([]byte(got.Password), []byte("abc12345")) != nil {
		t.Fatalf("expected stored password hash")
	}
}

func TestAuthUseCase_VerifyEmail_SetsVerifiedAndAlreadyVerified(t *testing.T) {
	f := setupAppFixture(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u := model.User{
		Username:      "alice",
		Password:      string(hashed),
		Status:        1,
		Email:         "alice@example.com",
		EmailVerified: false,
	}
	if err := db.DB.Create(&u).Error; err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	token, err := utils.GenerateEmailToken(u.ID, u.Email, time.Hour)
	if err != nil {
		t.Fatalf("GenerateEmailToken failed: %v", err)
	}

	already, err := f.authUC.VerifyEmail(token)
	if err != nil {
		t.Fatalf("VerifyEmail failed: %v", err)
	}
	if already {
		t.Fatalf("expected already=false on first verify")
	}

	var got model.User
	if err := db.DB.First(&got, u.ID).Error; err != nil {
		t.Fatalf("reload user failed: %v", err)
	}
	if !got.EmailVerified {
		t.Fatalf("expected user email verified")
	}

	already2, err := f.authUC.VerifyEmail(token)
	if err != nil {
		t.Fatalf("VerifyEmail second call failed: %v", err)
	}
	if !already2 {
		t.Fatalf("expected already=true on second verify")
	}
}

func TestAuthUseCase_LoginUser_Success(t *testing.T) {
	f := setupAppFixture(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u := model.User{
		Username:      "alice",
		Password:      string(hashed),
		Status:        1,
		Email:         "alice@example.com",
		EmailVerified: true,
	}
	if err := db.DB.Create(&u).Error; err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	token, err := f.authUC.LoginUser("alice", "abc12345")
	if err != nil {
		t.Fatalf("LoginUser failed: %v", err)
	}
	claims, err := utils.ParseLoginToken(token)
	if err != nil {
		t.Fatalf("ParseLoginToken failed: %v", err)
	}
	if claims.ID != u.ID || claims.Username != "alice" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestAuthUseCase_RegisterUser_DisabledAndConflicts(t *testing.T) {
	f := setupAppFixture(t)
	f.initializeSystem(t)

	if err := db.DB.Save(&model.Setting{Key: consts.ConfigAllowRegister, Value: "false"}).Error; err != nil {
		t.Fatalf("set allow_register=false failed: %v", err)
	}
	f.dbConfig.ClearCache()
	err := f.authUC.RegisterUser("alice_1", "abc12345", "alice@example.com")
	assertAuthErrorCode(t, err, httpx.AuthErrorForbidden)

	if err := db.DB.Save(&model.Setting{Key: consts.ConfigAllowRegister, Value: "true"}).Error; err != nil {
		t.Fatalf("set allow_register=true failed: %v", err)
	}
	f.dbConfig.ClearCache()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	exist := model.User{
		Username: "alice_1",
		Password: string(hashed),
		Status:   1,
		Email:    "taken@example.com",
	}
	if err := db.DB.Create(&exist).Error; err != nil {
		t.Fatalf("create existing user failed: %v", err)
	}

	err = f.authUC.RegisterUser("alice_1", "abc12345", "new@example.com")
	assertAuthErrorCode(t, err, httpx.AuthErrorConflict)

	err = f.authUC.RegisterUser("alice_2", "abc12345", "taken@example.com")
	assertAuthErrorCode(t, err, httpx.AuthErrorConflict)
}

func TestAuthUseCase_VerifyEmail_InvalidTokenValidation(t *testing.T) {
	f := setupAppFixture(t)

	_, err := f.authUC.VerifyEmail("bad-token")
	assertAuthErrorCode(t, err, httpx.AuthErrorValidation)
}

func TestAuthUseCase_VerifyEmailChange_Success(t *testing.T) {
	f := setupAppFixture(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u := model.User{
		Username:      "alice",
		Password:      string(hashed),
		Status:        1,
		Email:         "old@example.com",
		EmailVerified: true,
	}
	if err := db.DB.Create(&u).Error; err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	token, err := f.userService.GenerateEmailChangeToken(u.ID, "old@example.com", "new@example.com")
	if err != nil {
		t.Fatalf("GenerateEmailChangeToken failed: %v", err)
	}

	if err := f.authUC.VerifyEmailChange(token); err != nil {
		t.Fatalf("VerifyEmailChange failed: %v", err)
	}

	var got model.User
	if err := db.DB.First(&got, u.ID).Error; err != nil {
		t.Fatalf("reload user failed: %v", err)
	}
	if got.Email != "new@example.com" || !got.EmailVerified {
		t.Fatalf("unexpected user after email change: %+v", got)
	}
}

func TestAuthUseCase_VerifyEmailChange_Conflict(t *testing.T) {
	f := setupAppFixture(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u1 := model.User{
		Username:      "alice",
		Password:      string(hashed),
		Status:        1,
		Email:         "old@example.com",
		EmailVerified: true,
	}
	u2 := model.User{
		Username:      "bob",
		Password:      string(hashed),
		Status:        1,
		Email:         "taken@example.com",
		EmailVerified: true,
	}
	if err := db.DB.Create(&u1).Error; err != nil {
		t.Fatalf("create user1 failed: %v", err)
	}
	if err := db.DB.Create(&u2).Error; err != nil {
		t.Fatalf("create user2 failed: %v", err)
	}

	token, err := f.userService.GenerateEmailChangeToken(u1.ID, "old@example.com", "taken@example.com")
	if err != nil {
		t.Fatalf("GenerateEmailChangeToken failed: %v", err)
	}

	err = f.authUC.VerifyEmailChange(token)
	assertAuthErrorCode(t, err, httpx.AuthErrorConflict)
}

func TestAuthUseCase_RequestPasswordReset_Branches(t *testing.T) {
	f := setupAppFixture(t)

	if err := f.authUC.RequestPasswordReset("unknown@example.com"); err != nil {
		t.Fatalf("unknown email should return nil, got: %v", err)
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	banned := model.User{
		Username: "banned",
		Password: string(hashed),
		Status:   2,
		Email:    "banned@example.com",
	}
	active := model.User{
		Username: "alice",
		Password: string(hashed),
		Status:   1,
		Email:    "alice@example.com",
	}
	if err := db.DB.Create(&banned).Error; err != nil {
		t.Fatalf("create banned user failed: %v", err)
	}
	if err := db.DB.Create(&active).Error; err != nil {
		t.Fatalf("create active user failed: %v", err)
	}

	err := f.authUC.RequestPasswordReset("banned@example.com")
	assertAuthErrorCode(t, err, httpx.AuthErrorForbidden)

	if err := f.authUC.RequestPasswordReset("alice@example.com"); err != nil {
		t.Fatalf("active user reset request failed: %v", err)
	}
}

func TestAuthUseCase_ResetPassword_SuccessAndValidation(t *testing.T) {
	f := setupAppFixture(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u := model.User{
		Username:      "alice",
		Password:      string(hashed),
		Status:        1,
		Email:         "alice@example.com",
		EmailVerified: false,
	}
	if err := db.DB.Create(&u).Error; err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	err := f.authUC.ResetPassword("bad-token", "short")
	assertAuthErrorCode(t, err, httpx.AuthErrorValidation)

	err = f.authUC.ResetPassword("bad-token", "abc123456")
	assertAuthErrorCode(t, err, httpx.AuthErrorValidation)

	token, err := f.userService.GenerateForgetPasswordToken(u.ID)
	if err != nil {
		t.Fatalf("GenerateForgetPasswordToken failed: %v", err)
	}

	if err := f.authUC.ResetPassword(token, "abc123456"); err != nil {
		t.Fatalf("ResetPassword failed: %v", err)
	}

	var got model.User
	if err := db.DB.First(&got, u.ID).Error; err != nil {
		t.Fatalf("reload user failed: %v", err)
	}
	if !got.EmailVerified {
		t.Fatalf("expected email verified after reset")
	}
	if bcrypt.CompareHashAndPassword([]byte(got.Password), []byte("abc123456")) != nil {
		t.Fatalf("expected password updated")
	}
}

func TestAuthUseCase_VerifyEmail_EmailMismatchValidation(t *testing.T) {
	f := setupAppFixture(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u := model.User{
		Username:      "alice",
		Password:      string(hashed),
		Status:        1,
		Email:         "alice@example.com",
		EmailVerified: false,
	}
	if err := db.DB.Create(&u).Error; err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	token, err := utils.GenerateEmailToken(u.ID, "other@example.com", time.Hour)
	if err != nil {
		t.Fatalf("GenerateEmailToken failed: %v", err)
	}
	_, err = f.authUC.VerifyEmail(token)
	assertAuthErrorCode(t, err, httpx.AuthErrorValidation)
}
