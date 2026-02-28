package app

import (
	"perfect-pic-server/internal/common"
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
	"testing"
)

func TestPasskeyUseCase_EnsureUserPasskeyCapacity_Conflict(t *testing.T) {
	f := setupAppFixture(t)

	u := model.User{
		Username:      "alice",
		Password:      "x",
		Status:        1,
		Email:         "alice@example.com",
		EmailVerified: true,
	}
	if err := db.DB.Create(&u).Error; err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	for i := 0; i < 10; i++ {
		record := model.PasskeyCredential{
			UserID:       u.ID,
			CredentialID: "cred_cap_" + string(rune('a'+i)),
			Credential:   `{"id":"x"}`,
		}
		if err := db.DB.Create(&record).Error; err != nil {
			t.Fatalf("create passkey failed: %v", err)
		}
	}

	err := f.passkeyUC.ensureUserPasskeyCapacity(u.ID)
	assertServiceErrorCode(t, err, common.ErrorCodeConflict)
}

func TestPasskeyUseCase_BeginPasskeyLogin_Success(t *testing.T) {
	f := setupAppFixture(t)

	sessionID, options, err := f.passkeyUC.BeginPasskeyLogin()
	if err != nil {
		t.Fatalf("BeginPasskeyLogin failed: %v", err)
	}
	if sessionID == "" {
		t.Fatalf("expected non-empty session id")
	}
	if options == nil || options.Response.Challenge.String() == "" {
		t.Fatalf("expected valid assertion options")
	}
}

func TestPasskeyUseCase_FinishPasskeyLogin_InvalidSession(t *testing.T) {
	f := setupAppFixture(t)

	_, err := f.passkeyUC.FinishPasskeyLogin("bad-session", []byte(`{}`))
	assertServiceErrorCode(t, err, common.ErrorCodeValidation)
}
