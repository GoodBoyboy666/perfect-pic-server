package app

import (
	"encoding/json"
	"perfect-pic-server/internal/common"
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
	"testing"

	"github.com/go-webauthn/webauthn/webauthn"
)

func TestPasskeyWebAuthnUser_Methods(t *testing.T) {
	u := &passkeyWebAuthnUser{
		userID:      1,
		username:    "alice",
		id:          []byte("1"),
		credentials: []webauthn.Credential{{ID: []byte{1, 2}}},
	}

	if string(u.WebAuthnID()) != "1" {
		t.Fatalf("unexpected WebAuthnID")
	}
	if u.WebAuthnName() != "alice" || u.WebAuthnDisplayName() != "alice" {
		t.Fatalf("unexpected WebAuthn name/display")
	}
	if len(u.WebAuthnCredentials()) != 1 {
		t.Fatalf("unexpected WebAuthn credentials length")
	}
}

func TestPasskeyUseCase_BeginPasskeyRegistration_Success(t *testing.T) {
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

	sessionID, options, err := f.passkeyUC.BeginPasskeyRegistration(u.ID)
	if err != nil {
		t.Fatalf("BeginPasskeyRegistration failed: %v", err)
	}
	if sessionID == "" {
		t.Fatalf("expected non-empty session id")
	}
	if options == nil || options.Response.Challenge.String() == "" {
		t.Fatalf("expected valid creation options")
	}
}

func TestPasskeyUseCase_FinishPasskeyRegistration_InvalidCredential(t *testing.T) {
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

	sessionID, _, err := f.passkeyUC.BeginPasskeyRegistration(u.ID)
	if err != nil {
		t.Fatalf("BeginPasskeyRegistration failed: %v", err)
	}

	err = f.passkeyUC.FinishPasskeyRegistration(u.ID, sessionID, []byte(`{}`))
	assertServiceErrorCode(t, err, common.ErrorCodeValidation)
}

func TestPasskeyUseCase_FinishPasskeyRegistration_InvalidSession(t *testing.T) {
	f := setupAppFixture(t)

	err := f.passkeyUC.FinishPasskeyRegistration(1, "bad-session", []byte(`{}`))
	assertServiceErrorCode(t, err, common.ErrorCodeValidation)
}

func TestPasskeyUseCase_LoadPasskeyWebAuthnUser_LoginMode(t *testing.T) {
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

	rawCred, err := json.Marshal(webauthn.Credential{ID: []byte{1, 2, 3}})
	if err != nil {
		t.Fatalf("marshal credential failed: %v", err)
	}
	record := model.PasskeyCredential{
		UserID:       u.ID,
		CredentialID: "cred_login_mode",
		Credential:   string(rawCred),
	}
	if err := db.DB.Create(&record).Error; err != nil {
		t.Fatalf("create passkey record failed: %v", err)
	}

	got, err := f.passkeyUC.loadPasskeyWebAuthnUser(u.ID, passkeyWebAuthnUserLoadModeLogin)
	if err != nil {
		t.Fatalf("loadPasskeyWebAuthnUser failed: %v", err)
	}
	if got.userID != u.ID {
		t.Fatalf("expected user id %d, got %d", u.ID, got.userID)
	}
	if got.username != "" {
		t.Fatalf("expected empty username in login mode, got %q", got.username)
	}
	if len(got.credentials) != 1 {
		t.Fatalf("expected 1 credential, got %d", len(got.credentials))
	}
}

func TestPasskeyUseCase_LoadPasskeyWebAuthnUser_InvalidMode(t *testing.T) {
	f := setupAppFixture(t)

	_, err := f.passkeyUC.loadPasskeyWebAuthnUser(1, passkeyWebAuthnUserLoadMode("bad_mode"))
	assertServiceErrorCode(t, err, common.ErrorCodeInternal)
}

func TestPasskeyUseCase_LoadPasskeyWebAuthnUser_RegistrationUserNotFound(t *testing.T) {
	f := setupAppFixture(t)

	_, err := f.passkeyUC.loadPasskeyWebAuthnUser(999, passkeyWebAuthnUserLoadModeRegistration)
	assertServiceErrorCode(t, err, common.ErrorCodeNotFound)
}

func TestPasskeyUseCase_FinishPasskeyLogin_InvalidCredentialWithSession(t *testing.T) {
	f := setupAppFixture(t)

	sessionID, _, err := f.passkeyUC.BeginPasskeyLogin()
	if err != nil {
		t.Fatalf("BeginPasskeyLogin failed: %v", err)
	}

	_, err = f.passkeyUC.FinishPasskeyLogin(sessionID, []byte(`{}`))
	assertServiceErrorCode(t, err, common.ErrorCodeUnauthorized)
}
