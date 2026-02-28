package app

import (
	"perfect-pic-server/internal/common/httpx"
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
