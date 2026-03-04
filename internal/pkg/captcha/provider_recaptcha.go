package captcha

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const DefaultRecaptchaVerifyURL = "https://www.google.com/recaptcha/api/siteverify"

type RecaptchaConfig struct {
	SiteKey          string
	SecretKey        string
	VerifyURL        string
	ExpectedHostname string
}

type RecaptchaVerifyResponse struct {
	Success    bool     `json:"success"`
	Hostname   string   `json:"hostname"`
	Action     string   `json:"action"`
	Score      float64  `json:"score"`
	ErrorCodes []string `json:"error-codes"`
}

func VerifyRecaptchaCaptcha(cfg RecaptchaConfig, token, remoteIP string) (bool, string) {
	if cfg.SiteKey == "" || cfg.SecretKey == "" {
		return false, "验证码配置错误，请联系管理员"
	}
	if strings.TrimSpace(token) == "" {
		return false, "请完成人机验证"
	}

	ok, err := verifyRecaptcha(httpClient, cfg, token, remoteIP)
	if err != nil {
		log.Printf("⚠️ reCAPTCHA 验证失败: %v", err)
		return false, "人机验证服务不可用，请稍后重试"
	}
	if !ok {
		return false, "人机验证失败，请重试"
	}

	return true, ""
}

func verifyRecaptcha(httpClient *http.Client, cfg RecaptchaConfig, token, remoteIP string) (bool, error) {
	form := url.Values{}
	form.Set("secret", cfg.SecretKey)
	form.Set("response", strings.TrimSpace(token))
	if strings.TrimSpace(remoteIP) != "" {
		form.Set("remoteip", strings.TrimSpace(remoteIP))
	}

	req, err := http.NewRequest(http.MethodPost, cfg.VerifyURL, strings.NewReader(form.Encode()))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("recaptcha verify status code: %d", resp.StatusCode)
	}

	var result RecaptchaVerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	if !result.Success {
		return false, nil
	}

	expectedHostname := strings.TrimSpace(cfg.ExpectedHostname)
	if expectedHostname != "" && !strings.EqualFold(expectedHostname, strings.TrimSpace(result.Hostname)) {
		return false, fmt.Errorf("recaptcha hostname mismatch: expected %s, got %s", expectedHostname, result.Hostname)
	}

	return true, nil
}
