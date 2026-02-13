package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/service"

	"github.com/gin-gonic/gin"
)

func TestGetCaptcha_DisabledProvider(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaProvider, Value: ""}).Error
	service.ClearCache()

	r := gin.New()
	r.GET("/captcha", GetCaptcha)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/captcha", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp struct {
		Provider string `json:"provider"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Provider != service.CaptchaProviderDisabled {
		t.Fatalf("expected disabled provider, got %q", resp.Provider)
	}
}

func TestGetCaptcha_ProvidersWithPublicConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	r := gin.New()
	r.GET("/captcha", GetCaptcha)

	cases := []struct {
		provider string
		key      string
	}{
		{provider: "turnstile", key: consts.ConfigCaptchaTurnstileSiteKey},
		{provider: "recaptcha", key: consts.ConfigCaptchaRecaptchaSiteKey},
		{provider: "hcaptcha", key: consts.ConfigCaptchaHcaptchaSiteKey},
		{provider: "geetest", key: consts.ConfigCaptchaGeetestCaptchaID},
	}

	for _, tc := range cases {
		_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaProvider, Value: tc.provider}).Error
		_ = db.DB.Save(&model.Setting{Key: tc.key, Value: "pub"}).Error
		service.ClearCache()

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/captcha", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("provider %s expected 200, got %d body=%s", tc.provider, w.Code, w.Body.String())
		}
	}
}

func TestGetCaptchaImage_ImageProvider(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaProvider, Value: "image"}).Error
	service.ClearCache()

	r := gin.New()
	r.GET("/captcha/image", GetCaptchaImage)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/captcha/image", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp struct {
		ID    string `json:"captcha_id"`
		Image string `json:"captcha_image"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ID == "" || resp.Image == "" {
		t.Fatalf("expected captcha_id and captcha_image to be non-empty")
	}
}

func TestGetCaptchaImage_NonImageProviderReturns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	_ = db.DB.Save(&model.Setting{Key: consts.ConfigCaptchaProvider, Value: ""}).Error
	service.ClearCache()

	r := gin.New()
	r.GET("/captcha/image", GetCaptchaImage)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/captcha/image", nil))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
}
