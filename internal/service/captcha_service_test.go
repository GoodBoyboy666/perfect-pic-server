package service

import (
	"testing"

	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/utils"
)

func TestGetCaptchaProviderInfo_DefaultIsImage(t *testing.T) {
	setupTestDB(t)
	ClearCache()

	info := GetCaptchaProviderInfo()
	if info.Provider != CaptchaProviderImage {
		t.Fatalf("expected default provider to be image, got %q", info.Provider)
	}
}

func TestGetCaptchaProviderInfo_UnknownFallsBackToImage(t *testing.T) {
	setupTestDB(t)

	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaProvider, Value: "unknown"}).Error
	ClearCache()

	info := GetCaptchaProviderInfo()
	if info.Provider != CaptchaProviderImage {
		t.Fatalf("expected fallback to image, got %q", info.Provider)
	}
}

func TestGetCaptchaProviderInfo_Disabled(t *testing.T) {
	setupTestDB(t)

	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaProvider, Value: ""}).Error
	ClearCache()

	info := GetCaptchaProviderInfo()
	if info.Provider != CaptchaProviderDisabled {
		t.Fatalf("expected disabled provider, got %q", info.Provider)
	}
}

func TestGetCaptchaProviderInfo_PublicConfigByProvider(t *testing.T) {
	setupTestDB(t)

	cases := []struct {
		provider string
		key      string
		wantKey  string
	}{
		{provider: CaptchaProviderTurnstile, key: consts.ConfigCaptchaTurnstileSiteKey, wantKey: "turnstile_site_key"},
		{provider: CaptchaProviderRecaptcha, key: consts.ConfigCaptchaRecaptchaSiteKey, wantKey: "recaptcha_site_key"},
		{provider: CaptchaProviderHcaptcha, key: consts.ConfigCaptchaHcaptchaSiteKey, wantKey: "hcaptcha_site_key"},
		{provider: CaptchaProviderGeetest, key: consts.ConfigCaptchaGeetestCaptchaID, wantKey: "geetest_captcha_id"},
	}

	for _, tc := range cases {
		_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaProvider, Value: tc.provider}).Error
		_ = db.DB.Save(&model.Setting{Key: tc.key, Value: "pub"}).Error
		ClearCache()

		info := GetCaptchaProviderInfo()
		if info.Provider != tc.provider {
			t.Fatalf("expected provider %q, got %q", tc.provider, info.Provider)
		}
		if info.PublicConfig == nil || info.PublicConfig[tc.wantKey] != "pub" {
			t.Fatalf("expected public config %q=pub, got %#v", tc.wantKey, info.PublicConfig)
		}
	}
}

func TestVerifyCaptchaChallenge_DisabledProviderAlwaysOK(t *testing.T) {
	setupTestDB(t)

	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaProvider, Value: ""}).Error
	ClearCache()

	ok, msg := VerifyCaptchaChallenge("", "", "", "1.2.3.4")
	if !ok || msg != "" {
		t.Fatalf("expected ok for disabled provider, got ok=%v msg=%q", ok, msg)
	}
}

func TestVerifyCaptchaChallenge_ImageProviderValidates(t *testing.T) {
	setupTestDB(t)

	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaProvider, Value: "image"}).Error
	ClearCache()

	ok, msg := VerifyCaptchaChallenge("", "", "", "1.2.3.4")
	if ok || msg == "" {
		t.Fatalf("expected failure for empty captcha fields, got ok=%v msg=%q", ok, msg)
	}

	id, _, answer, err := utils.MakeCaptcha()
	if err != nil {
		t.Fatalf("MakeCaptcha: %v", err)
	}

	ok2, msg2 := VerifyCaptchaChallenge(id, answer, "", "1.2.3.4")
	if !ok2 || msg2 != "" {
		t.Fatalf("expected success for valid captcha, got ok=%v msg=%q", ok2, msg2)
	}
}
