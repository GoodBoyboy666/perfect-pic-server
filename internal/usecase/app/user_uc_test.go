package app

import (
	"perfect-pic-server/internal/common"
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestUserUseCase_RequestEmailChange_ValidationBranches(t *testing.T) {
	f := setupAppFixture(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u := model.User{
		Username: "alice",
		Password: string(hashed),
		Status:   1,
		Email:    "alice@example.com",
	}
	if err := db.DB.Create(&u).Error; err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	err := f.userUC.RequestEmailChange(u.ID, "abc12345", "bad-email")
	assertServiceErrorCode(t, err, common.ErrorCodeValidation)

	err = f.userUC.RequestEmailChange(u.ID, "wrong", "new@example.com")
	assertServiceErrorCode(t, err, common.ErrorCodeForbidden)

	err = f.userUC.RequestEmailChange(u.ID, "abc12345", "alice@example.com")
	assertServiceErrorCode(t, err, common.ErrorCodeValidation)
}

func TestUserUseCase_RequestEmailChange_Success(t *testing.T) {
	f := setupAppFixture(t)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("abc12345"), bcrypt.DefaultCost)
	u := model.User{
		Username: "alice",
		Password: string(hashed),
		Status:   1,
		Email:    "alice@example.com",
	}
	if err := db.DB.Create(&u).Error; err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	if err := f.userUC.RequestEmailChange(u.ID, "abc12345", "new@example.com"); err != nil {
		t.Fatalf("RequestEmailChange failed: %v", err)
	}
}
