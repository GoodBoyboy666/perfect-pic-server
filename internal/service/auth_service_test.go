package service

import (
	"testing"
	"time"

	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/utils"

	"golang.org/x/crypto/bcrypt"
)

func TestLoginUser_Success(t *testing.T) {
	setupTestDB(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u := model.User{
		Username:      "alice",
		Password:      string(hashed),
		Admin:         true,
		Status:        1,
		Email:         "alice@example.com",
		EmailVerified: true,
	}
	if err := db.DB.Create(&u).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	token, err := LoginUser("alice", "abc12345")
	if err != nil {
		t.Fatalf("LoginUser error: %v", err)
	}
	claims, err := utils.ParseLoginToken(token)
	if err != nil {
		t.Fatalf("ParseLoginToken error: %v", err)
	}
	if claims.ID != u.ID || claims.Username != "alice" || !claims.Admin {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestLoginUser_WrongPasswordUnauthorized(t *testing.T) {
	setupTestDB(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u := model.User{
		Username: "alice",
		Password: string(hashed),
		Status:   1,
	}
	_ = db.DB.Create(&u).Error

	_, err := LoginUser("alice", "wrongpass1")
	if err == nil {
		t.Fatalf("expected error")
	}
	authErr, ok := AsAuthError(err)
	if !ok || authErr.Code != AuthErrorUnauthorized {
		t.Fatalf("expected unauthorized auth error, got: %#v (%v)", authErr, err)
	}
}

func TestLoginUser_BlockedUnverifiedForbidden(t *testing.T) {
	setupTestDB(t)

	// Enable block-unverified.
	if err := db.DB.Save(&model.Setting{Key: consts.ConfigBlockUnverifiedUsers, Value: "true"}).Error; err != nil {
		t.Fatalf("set setting: %v", err)
	}
	ClearCache()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u := model.User{
		Username:      "alice",
		Password:      string(hashed),
		Status:        1,
		Email:         "alice@example.com",
		EmailVerified: false,
	}
	if err := db.DB.Create(&u).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	_, err := LoginUser("alice", "abc12345")
	if err == nil {
		t.Fatalf("expected error")
	}
	authErr, ok := AsAuthError(err)
	if !ok || authErr.Code != AuthErrorForbidden {
		t.Fatalf("expected forbidden auth error, got: %#v (%v)", authErr, err)
	}
}

func TestRegisterUser_ValidationError(t *testing.T) {
	setupTestDB(t)

	err := RegisterUser("ab", "short", "bad-email")
	if err == nil {
		t.Fatalf("expected error")
	}
	authErr, ok := AsAuthError(err)
	if !ok || authErr.Code != AuthErrorValidation {
		t.Fatalf("expected validation auth error, got: %#v (%v)", authErr, err)
	}
}

func TestRegisterUser_DuplicateUsernameConflict(t *testing.T) {
	setupTestDB(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u := model.User{Username: "alice", Password: string(hashed), Status: 1, Email: "a1@example.com"}
	if err := db.DB.Create(&u).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	err := RegisterUser("alice", "abc12345", "alice2@example.com")
	if err == nil {
		t.Fatalf("expected error")
	}
	authErr, ok := AsAuthError(err)
	if !ok || authErr.Code != AuthErrorConflict {
		t.Fatalf("expected conflict auth error, got: %#v (%v)", authErr, err)
	}
}

func TestVerifyEmail_SetsVerified(t *testing.T) {
	setupTestDB(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u := model.User{Username: "alice", Password: string(hashed), Status: 1, Email: "alice@example.com", EmailVerified: false}
	if err := db.DB.Create(&u).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	token, err := utils.GenerateEmailToken(u.ID, u.Email, 1*time.Hour)
	if err != nil {
		t.Fatalf("GenerateEmailToken: %v", err)
	}

	already, err := VerifyEmail(token)
	if err != nil {
		t.Fatalf("VerifyEmail error: %v", err)
	}
	if already {
		t.Fatalf("expected alreadyVerified=false on first verify")
	}

	var got model.User
	if err := db.DB.First(&got, u.ID).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if !got.EmailVerified {
		t.Fatalf("expected user to be verified")
	}

	already2, err := VerifyEmail(token)
	if err != nil {
		t.Fatalf("VerifyEmail second call error: %v", err)
	}
	if !already2 {
		t.Fatalf("expected alreadyVerified=true on second verify")
	}
}
