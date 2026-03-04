package captcha

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const DefaultHcaptchaVerifyURL = "https://hcaptcha.com/siteverify"

type HcaptchaConfig struct {
	SiteKey          string
	SecretKey        string
	VerifyURL        string
	ExpectedHostname string
}

type HcaptchaVerifyResponse struct {
	Success    bool     `json:"success"`
	Hostname   string   `json:"hostname"`
	ErrorCodes []string `json:"error-codes"`
}

func VerifyHcaptchaCaptcha(cfg HcaptchaConfig, token, remoteIP string) (bool, string) {
	if cfg.SiteKey == "" || cfg.SecretKey == "" {
		return false, "验证码配置错误，请联系管理员"
	}
	if strings.TrimSpace(token) == "" {
		return false, "请完成人机验证"
	}

	ok, err := verifyHcaptcha(httpClient, cfg, token, remoteIP)
	if err != nil {
		log.Printf("⚠️ hCaptcha 验证失败: %v", err)
		return false, "人机验证服务不可用，请稍后重试"
	}
	if !ok {
		return false, "人机验证失败，请重试"
	}

	return true, ""
}

func verifyHcaptcha(httpClient *http.Client, cfg HcaptchaConfig, token, remoteIP string) (bool, error) {
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
		return false, fmt.Errorf("hcaptcha verify status code: %d", resp.StatusCode)
	}

	var result HcaptchaVerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	if !result.Success {
		return false, nil
	}

	expectedHostname := strings.TrimSpace(cfg.ExpectedHostname)
	if expectedHostname != "" && !strings.EqualFold(expectedHostname, strings.TrimSpace(result.Hostname)) {
		return false, fmt.Errorf("hcaptcha hostname mismatch: expected %s, got %s", expectedHostname, result.Hostname)
	}

	return true, nil
}
