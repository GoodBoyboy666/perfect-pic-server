package service

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
)

func TestGetCaptchaProviderInfo_PublicConfig(t *testing.T) {
	setupTestDB(t)

	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaProvider, Value: "turnstile"}).Error
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaTurnstileSiteKey, Value: "site"}).Error
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaTurnstileSecretKey, Value: "secret"}).Error
	ClearCache()

	info := GetCaptchaProviderInfo()
	if info.Provider != CaptchaProviderTurnstile {
		t.Fatalf("expected turnstile, got %q", info.Provider)
	}
	if info.PublicConfig["turnstile_site_key"] != "site" {
		t.Fatalf("expected site key in public config")
	}
}

func TestVerifyCaptchaChallenge_ProviderConfigMissing(t *testing.T) {
	setupTestDB(t)

	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaProvider, Value: "hcaptcha"}).Error
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaHcaptchaSiteKey, Value: ""}).Error
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaHcaptchaSecretKey, Value: ""}).Error
	ClearCache()

	ok, msg := VerifyCaptchaChallenge("", "", "token", "1.1.1.1")
	if ok || msg == "" {
		t.Fatalf("expected config error, got ok=%v msg=%q", ok, msg)
	}
}

func TestVerifyCaptchaChallenge_GeetestTokenParseErrors(t *testing.T) {
	setupTestDB(t)

	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaProvider, Value: "geetest"}).Error
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaGeetestCaptchaID, Value: "id"}).Error
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaGeetestCaptchaKey, Value: "key"}).Error
	ClearCache()

	ok, msg := VerifyCaptchaChallenge("", "", "not-base64", "")
	if ok || msg == "" {
		t.Fatalf("expected format error, got ok=%v msg=%q", ok, msg)
	}

	// base64 ok but missing required fields
	b, _ := json.Marshal(map[string]string{"lot_number": "x"})
	token := base64.StdEncoding.EncodeToString(b)
	ok, msg = VerifyCaptchaChallenge("", "", token, "")
	if ok || msg == "" {
		t.Fatalf("expected incomplete error, got ok=%v msg=%q", ok, msg)
	}
}

func TestVerifyCaptchaChallenge_RemoteProvidersViaTestServer(t *testing.T) {
	setupTestDB(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/turnstile":
			_ = json.NewEncoder(w).Encode(turnstileVerifyResponse{Success: true, Hostname: "example.com"})
		case "/recaptcha":
			_ = json.NewEncoder(w).Encode(recaptchaVerifyResponse{Success: true, Hostname: "example.com"})
		case "/hcaptcha":
			_ = json.NewEncoder(w).Encode(hcaptchaVerifyResponse{Success: true, Hostname: "example.com"})
		case "/geetest":
			_ = json.NewEncoder(w).Encode(geetestVerifyResponse{Result: "success"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	oldClient := captchaHTTPClient
	captchaHTTPClient = srv.Client()
	defer func() { captchaHTTPClient = oldClient }()

	// Turnstile: token empty + success + hostname mismatch error path.
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaProvider, Value: "turnstile"}).Error
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaTurnstileSiteKey, Value: "site"}).Error
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaTurnstileSecretKey, Value: "secret"}).Error
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaTurnstileVerifyURL, Value: srv.URL + "/turnstile"}).Error
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaTurnstileExpectedHostname, Value: ""}).Error
	ClearCache()

	ok, msg := VerifyCaptchaChallenge("", "", "", "1.1.1.1")
	if ok || msg == "" {
		t.Fatalf("expected token required, got ok=%v msg=%q", ok, msg)
	}
	ok, msg = VerifyCaptchaChallenge("", "", "token", "1.1.1.1")
	if !ok || msg != "" {
		t.Fatalf("expected success, got ok=%v msg=%q", ok, msg)
	}

	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaTurnstileExpectedHostname, Value: "wrong-host"}).Error
	ClearCache()
	ok, msg = VerifyCaptchaChallenge("", "", "token", "1.1.1.1")
	if ok || msg == "" {
		t.Fatalf("expected failure, got ok=%v msg=%q", ok, msg)
	}

	// reCAPTCHA
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaProvider, Value: "recaptcha"}).Error
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaRecaptchaSiteKey, Value: "site"}).Error
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaRecaptchaSecretKey, Value: "secret"}).Error
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaRecaptchaVerifyURL, Value: srv.URL + "/recaptcha"}).Error
	ClearCache()
	ok, msg = VerifyCaptchaChallenge("", "", "token", "1.1.1.1")
	if !ok || msg != "" {
		t.Fatalf("expected success, got ok=%v msg=%q", ok, msg)
	}

	// hCaptcha
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaProvider, Value: "hcaptcha"}).Error
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaHcaptchaSiteKey, Value: "site"}).Error
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaHcaptchaSecretKey, Value: "secret"}).Error
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaHcaptchaVerifyURL, Value: srv.URL + "/hcaptcha"}).Error
	ClearCache()
	ok, msg = VerifyCaptchaChallenge("", "", "token", "1.1.1.1")
	if !ok || msg != "" {
		t.Fatalf("expected success, got ok=%v msg=%q", ok, msg)
	}

	// GeeTest
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaProvider, Value: "geetest"}).Error
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaGeetestCaptchaID, Value: "id"}).Error
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaGeetestCaptchaKey, Value: "key"}).Error
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaGeetestVerifyURL, Value: srv.URL + "/geetest"}).Error
	ClearCache()

	p := geetestVerifyTokenPayload{
		LotNumber:     "lot",
		CaptchaOutput: "out",
		PassToken:     "pass",
		GenTime:       "time",
	}
	payload, _ := json.Marshal(p)
	tok := base64.StdEncoding.EncodeToString(payload)
	ok, msg = VerifyCaptchaChallenge("", "", tok, "")
	if !ok || msg != "" {
		t.Fatalf("expected success, got ok=%v msg=%q", ok, msg)
	}
}
