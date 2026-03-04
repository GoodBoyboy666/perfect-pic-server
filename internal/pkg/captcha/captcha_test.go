package captcha_test

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/model"
	pkgcaptcha "perfect-pic-server/internal/pkg/captcha"
	"perfect-pic-server/internal/repository"
	"perfect-pic-server/internal/service"
	"perfect-pic-server/internal/testutils"

	"gorm.io/gorm"
)

type captchaFixture struct {
	gdb            *gorm.DB
	dbConfig       *config.DBConfig
	captchaService *service.CaptchaService
}

func setupCaptchaFixture(t *testing.T) *captchaFixture {
	t.Helper()
	config.InitConfig("")

	gdb := testutils.SetupDB(t)
	settingStore := repository.NewSettingRepository(gdb)
	dbConfig := config.NewDBConfig(settingStore)
	dbConfig.ClearCache()

	return &captchaFixture{
		gdb:            gdb,
		dbConfig:       dbConfig,
		captchaService: service.NewCaptchaService(dbConfig),
	}
}

func (f *captchaFixture) setSetting(key, value string) {
	_ = f.gdb.Save(&model.Setting{Key: key, Value: value}).Error
	f.dbConfig.ClearCache()
}

func TestGetCaptchaProviderInfo_DefaultIsImage(t *testing.T) {
	f := setupCaptchaFixture(t)

	info := f.captchaService.GetCaptchaProviderInfo()
	if info.Provider != consts.CaptchaProviderImage {
		t.Fatalf("期望 default provider to be image，实际为 %q", info.Provider)
	}
}

func TestGetCaptchaProviderInfo_UnknownFallsBackToImage(t *testing.T) {
	f := setupCaptchaFixture(t)
	f.setSetting(consts.ConfigCaptchaProvider, "unknown")

	info := f.captchaService.GetCaptchaProviderInfo()
	if info.Provider != consts.CaptchaProviderImage {
		t.Fatalf("期望 fallback to image，实际为 %q", info.Provider)
	}
}

func TestGetCaptchaProviderInfo_Disabled(t *testing.T) {
	f := setupCaptchaFixture(t)
	f.setSetting(consts.ConfigCaptchaProvider, consts.CaptchaProviderDisabled)

	info := f.captchaService.GetCaptchaProviderInfo()
	if info.Provider != consts.CaptchaProviderDisabled {
		t.Fatalf("期望 disabled provider，实际为 %q", info.Provider)
	}
}

func TestGetCaptchaProviderInfo_PublicConfigByProvider(t *testing.T) {
	f := setupCaptchaFixture(t)

	cases := []struct {
		provider string
		key      string
		wantKey  string
	}{
		{provider: consts.CaptchaProviderTurnstile, key: consts.ConfigCaptchaTurnstileSiteKey, wantKey: "turnstile_site_key"},
		{provider: consts.CaptchaProviderRecaptcha, key: consts.ConfigCaptchaRecaptchaSiteKey, wantKey: "recaptcha_site_key"},
		{provider: consts.CaptchaProviderHcaptcha, key: consts.ConfigCaptchaHcaptchaSiteKey, wantKey: "hcaptcha_site_key"},
		{provider: consts.CaptchaProviderGeetest, key: consts.ConfigCaptchaGeetestCaptchaID, wantKey: "geetest_captcha_id"},
	}

	for _, tc := range cases {
		f.setSetting(consts.ConfigCaptchaProvider, tc.provider)
		f.setSetting(tc.key, "pub")

		info := f.captchaService.GetCaptchaProviderInfo()
		if info.Provider != tc.provider {
			t.Fatalf("期望 provider %q，实际为 %q", tc.provider, info.Provider)
		}
		if info.PublicConfig == nil || info.PublicConfig[tc.wantKey] != "pub" {
			t.Fatalf("期望 public config %q=pub，实际为 %#v", tc.wantKey, info.PublicConfig)
		}
	}
}

func TestVerifyCaptchaChallenge_DisabledProviderAlwaysOK(t *testing.T) {
	f := setupCaptchaFixture(t)
	f.setSetting(consts.ConfigCaptchaProvider, consts.CaptchaProviderDisabled)

	ok, msg := f.captchaService.VerifyCaptchaChallenge("", "", "", "1.2.3.4")
	if !ok || msg != "" {
		t.Fatalf("期望 ok for disabled provider，实际为 ok=%v msg=%q", ok, msg)
	}
}

func TestVerifyCaptchaChallenge_ImageProviderValidates(t *testing.T) {
	f := setupCaptchaFixture(t)
	f.setSetting(consts.ConfigCaptchaProvider, consts.CaptchaProviderImage)

	ok, msg := f.captchaService.VerifyCaptchaChallenge("", "", "", "1.2.3.4")
	if ok || msg == "" {
		t.Fatalf("期望 failure for empty captcha fields，实际为 ok=%v msg=%q", ok, msg)
	}

	id, _, answer, err := pkgcaptcha.MakeCaptcha()
	if err != nil {
		t.Fatalf("MakeCaptcha: %v", err)
	}

	ok2, msg2 := f.captchaService.VerifyCaptchaChallenge(id, answer, "", "1.2.3.4")
	if !ok2 || msg2 != "" {
		t.Fatalf("期望 success for valid captcha，实际为 ok=%v msg=%q", ok2, msg2)
	}
}

func TestGetCaptchaProviderInfo_PublicConfig(t *testing.T) {
	f := setupCaptchaFixture(t)
	f.setSetting(consts.ConfigCaptchaProvider, consts.CaptchaProviderTurnstile)
	f.setSetting(consts.ConfigCaptchaTurnstileSiteKey, "site")
	f.setSetting(consts.ConfigCaptchaTurnstileSecretKey, "secret")

	info := f.captchaService.GetCaptchaProviderInfo()
	if info.Provider != consts.CaptchaProviderTurnstile {
		t.Fatalf("期望 turnstile，实际为 %q", info.Provider)
	}
	if info.PublicConfig["turnstile_site_key"] != "site" {
		t.Fatalf("期望 site key in public config")
	}
}

func TestVerifyCaptchaChallenge_ProviderConfigMissing(t *testing.T) {
	f := setupCaptchaFixture(t)
	f.setSetting(consts.ConfigCaptchaProvider, consts.CaptchaProviderHcaptcha)
	f.setSetting(consts.ConfigCaptchaHcaptchaSiteKey, "")
	f.setSetting(consts.ConfigCaptchaHcaptchaSecretKey, "")

	ok, msg := f.captchaService.VerifyCaptchaChallenge("", "", "token", "1.1.1.1")
	if ok || msg == "" {
		t.Fatalf("期望 config 错误，实际为 ok=%v msg=%q", ok, msg)
	}
}

func TestVerifyCaptchaChallenge_GeetestTokenParseErrors(t *testing.T) {
	f := setupCaptchaFixture(t)
	f.setSetting(consts.ConfigCaptchaProvider, consts.CaptchaProviderGeetest)
	f.setSetting(consts.ConfigCaptchaGeetestCaptchaID, "id")
	f.setSetting(consts.ConfigCaptchaGeetestCaptchaKey, "key")

	ok, msg := f.captchaService.VerifyCaptchaChallenge("", "", "not-base64", "")
	if ok || msg == "" {
		t.Fatalf("期望 format 错误，实际为 ok=%v msg=%q", ok, msg)
	}

	b, _ := json.Marshal(map[string]string{"lot_number": "x"})
	token := base64.StdEncoding.EncodeToString(b)
	ok, msg = f.captchaService.VerifyCaptchaChallenge("", "", token, "")
	if ok || msg == "" {
		t.Fatalf("期望 incomplete 错误，实际为 ok=%v msg=%q", ok, msg)
	}
}

func TestVerifyCaptchaChallenge_RemoteProvidersViaTestServer(t *testing.T) {
	f := setupCaptchaFixture(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/turnstile":
			_ = json.NewEncoder(w).Encode(pkgcaptcha.TurnstileVerifyResponse{Success: true, Hostname: "example.com"})
		case "/recaptcha":
			_ = json.NewEncoder(w).Encode(pkgcaptcha.RecaptchaVerifyResponse{Success: true, Hostname: "example.com"})
		case "/hcaptcha":
			_ = json.NewEncoder(w).Encode(pkgcaptcha.HcaptchaVerifyResponse{Success: true, Hostname: "example.com"})
		case "/geetest":
			_ = json.NewEncoder(w).Encode(pkgcaptcha.GeetestVerifyResponse{Result: "success"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	f.setSetting(consts.ConfigCaptchaProvider, consts.CaptchaProviderTurnstile)
	f.setSetting(consts.ConfigCaptchaTurnstileSiteKey, "site")
	f.setSetting(consts.ConfigCaptchaTurnstileSecretKey, "secret")
	f.setSetting(consts.ConfigCaptchaTurnstileVerifyURL, srv.URL+"/turnstile")
	f.setSetting(consts.ConfigCaptchaTurnstileExpectedHostname, "")

	ok, msg := f.captchaService.VerifyCaptchaChallenge("", "", "", "1.1.1.1")
	if ok || msg == "" {
		t.Fatalf("期望得到 token required，实际为 ok=%v msg=%q", ok, msg)
	}
	ok, msg = f.captchaService.VerifyCaptchaChallenge("", "", "token", "1.1.1.1")
	if !ok || msg != "" {
		t.Fatalf("期望 success，实际为 ok=%v msg=%q", ok, msg)
	}

	f.setSetting(consts.ConfigCaptchaTurnstileExpectedHostname, "wrong-host")
	ok, msg = f.captchaService.VerifyCaptchaChallenge("", "", "token", "1.1.1.1")
	if ok || msg == "" {
		t.Fatalf("期望 failure，实际为 ok=%v msg=%q", ok, msg)
	}

	f.setSetting(consts.ConfigCaptchaProvider, consts.CaptchaProviderRecaptcha)
	f.setSetting(consts.ConfigCaptchaRecaptchaSiteKey, "site")
	f.setSetting(consts.ConfigCaptchaRecaptchaSecretKey, "secret")
	f.setSetting(consts.ConfigCaptchaRecaptchaVerifyURL, srv.URL+"/recaptcha")
	ok, msg = f.captchaService.VerifyCaptchaChallenge("", "", "token", "1.1.1.1")
	if !ok || msg != "" {
		t.Fatalf("期望 success，实际为 ok=%v msg=%q", ok, msg)
	}

	f.setSetting(consts.ConfigCaptchaProvider, consts.CaptchaProviderHcaptcha)
	f.setSetting(consts.ConfigCaptchaHcaptchaSiteKey, "site")
	f.setSetting(consts.ConfigCaptchaHcaptchaSecretKey, "secret")
	f.setSetting(consts.ConfigCaptchaHcaptchaVerifyURL, srv.URL+"/hcaptcha")
	ok, msg = f.captchaService.VerifyCaptchaChallenge("", "", "token", "1.1.1.1")
	if !ok || msg != "" {
		t.Fatalf("期望 success，实际为 ok=%v msg=%q", ok, msg)
	}

	f.setSetting(consts.ConfigCaptchaProvider, consts.CaptchaProviderGeetest)
	f.setSetting(consts.ConfigCaptchaGeetestCaptchaID, "id")
	f.setSetting(consts.ConfigCaptchaGeetestCaptchaKey, "key")
	f.setSetting(consts.ConfigCaptchaGeetestVerifyURL, srv.URL+"/geetest")

	p := pkgcaptcha.GeetestVerifyTokenPayload{
		LotNumber:     "lot",
		CaptchaOutput: "out",
		PassToken:     "pass",
		GenTime:       "time",
	}
	payload, _ := json.Marshal(p)
	tok := base64.StdEncoding.EncodeToString(payload)
	ok, msg = f.captchaService.VerifyCaptchaChallenge("", "", tok, "")
	if !ok || msg != "" {
		t.Fatalf("期望 success，实际为 ok=%v msg=%q", ok, msg)
	}
}
