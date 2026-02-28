package admin

import (
	"perfect-pic-server/internal/common"
	"testing"
)

func TestSettingsUseCase_AdminSendTestEmail_InvalidEmail(t *testing.T) {
	f := setupAdminFixture(t)

	err := f.settingsUC.AdminSendTestEmail("bad-email")
	assertServiceErrorCode(t, err, common.ErrorCodeValidation)
}

func TestSettingsUseCase_AdminSendTestEmail_MissingHostInternalError(t *testing.T) {
	f := setupAdminFixture(t)

	err := f.settingsUC.AdminSendTestEmail("a@example.com")
	assertServiceErrorCode(t, err, common.ErrorCodeInternal)
}
