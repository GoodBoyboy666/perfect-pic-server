package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"perfect-pic-server/internal/consts"
	moduledto "perfect-pic-server/internal/dto"
	"perfect-pic-server/internal/utils"
	"strings"
	"time"
)

const (
	// 默认 Turnstile 校验地址。
	defaultTurnstileVerifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"
	// 默认 reCAPTCHA 校验地址。
	defaultRecaptchaVerifyURL = "https://www.google.com/recaptcha/api/siteverify"
	// 默认 hCaptcha 校验地址。
	defaultHcaptchaVerifyURL = "https://hcaptcha.com/siteverify"
	// 默认 GeeTest 校验地址。
	defaultGeetestVerifyURL = "https://gcaptcha4.geetest.com/validate"
)

// captchaHTTPClient 验证码服务端校验统一 HTTP 客户端。
var captchaHTTPClient = &http.Client{Timeout: 5 * time.Second}

// turnstileConfig Turnstile 专属配置。
type turnstileConfig struct {
	SiteKey          string
	SecretKey        string
	VerifyURL        string
	ExpectedHostname string
}

// recaptchaConfig reCAPTCHA 专属配置。
type recaptchaConfig struct {
	SiteKey          string
	SecretKey        string
	VerifyURL        string
	ExpectedHostname string
}

// hcaptchaConfig hCaptcha 专属配置。
type hcaptchaConfig struct {
	SiteKey          string
	SecretKey        string
	VerifyURL        string
	ExpectedHostname string
}

// geetestConfig GeeTest 专属配置。
type geetestConfig struct {
	CaptchaID  string
	CaptchaKey string
	VerifyURL  string
}

// GetCaptchaProviderInfo 获取当前验证码提供方与前端可见配置。
func (s *Service) GetCaptchaProviderInfo() moduledto.CaptchaProviderResponse {
	provider := s.loadCaptchaProvider()

	switch provider {
	case consts.CaptchaProviderDisabled:
		return moduledto.CaptchaProviderResponse{Provider: consts.CaptchaProviderDisabled}
	case consts.CaptchaProviderTurnstile:
		cfg := s.getTurnstileConfig()
		info := moduledto.CaptchaProviderResponse{Provider: consts.CaptchaProviderTurnstile}
		if cfg.SiteKey != "" {
			info.PublicConfig = map[string]string{"turnstile_site_key": cfg.SiteKey}
		}
		return info
	case consts.CaptchaProviderRecaptcha:
		cfg := s.getRecaptchaConfig()
		info := moduledto.CaptchaProviderResponse{Provider: consts.CaptchaProviderRecaptcha}
		if cfg.SiteKey != "" {
			info.PublicConfig = map[string]string{"recaptcha_site_key": cfg.SiteKey}
		}
		return info
	case consts.CaptchaProviderHcaptcha:
		cfg := s.getHcaptchaConfig()
		info := moduledto.CaptchaProviderResponse{Provider: consts.CaptchaProviderHcaptcha}
		if cfg.SiteKey != "" {
			info.PublicConfig = map[string]string{"hcaptcha_site_key": cfg.SiteKey}
		}
		return info
	case consts.CaptchaProviderGeetest:
		cfg := s.getGeetestConfig()
		info := moduledto.CaptchaProviderResponse{Provider: consts.CaptchaProviderGeetest}
		if cfg.CaptchaID != "" {
			info.PublicConfig = map[string]string{"geetest_captcha_id": cfg.CaptchaID}
		}
		return info
	case consts.CaptchaProviderImage:
		fallthrough
	default:
		return moduledto.CaptchaProviderResponse{Provider: consts.CaptchaProviderImage}
	}
}

// VerifyCaptchaChallenge 按当前 provider 校验验证码。
func (s *Service) VerifyCaptchaChallenge(captchaID, captchaAnswer, captchaToken, remoteIP string) (bool, string) {
	provider := s.loadCaptchaProvider()

	switch provider {
	case consts.CaptchaProviderDisabled:
		return true, ""
	case consts.CaptchaProviderTurnstile:
		return verifyTurnstileCaptcha(s.getTurnstileConfig(), captchaToken, remoteIP)
	case consts.CaptchaProviderRecaptcha:
		return verifyRecaptchaCaptcha(s.getRecaptchaConfig(), captchaToken, remoteIP)
	case consts.CaptchaProviderHcaptcha:
		return verifyHcaptchaCaptcha(s.getHcaptchaConfig(), captchaToken, remoteIP)
	case consts.CaptchaProviderGeetest:
		return verifyGeetestCaptcha(s.getGeetestConfig(), captchaToken)
	case consts.CaptchaProviderImage:
		fallthrough
	default:
		return verifyImageCaptcha(captchaID, captchaAnswer)
	}
}

// loadCaptchaProvider 读取并标准化验证码提供方。
func (s *Service) loadCaptchaProvider() string {
	provider := strings.ToLower(strings.TrimSpace(s.GetString(consts.ConfigCaptchaProvider)))
	switch provider {
	case consts.CaptchaProviderDisabled, consts.CaptchaProviderImage, consts.CaptchaProviderTurnstile, consts.CaptchaProviderRecaptcha, consts.CaptchaProviderHcaptcha, consts.CaptchaProviderGeetest:
		return provider
	default:
		return consts.CaptchaProviderImage
	}
}

// getTurnstileConfig 读取 Turnstile 配置。
func (s *Service) getTurnstileConfig() turnstileConfig {
	verifyURL := strings.TrimSpace(s.GetString(consts.ConfigCaptchaTurnstileVerifyURL))
	if verifyURL == "" {
		verifyURL = defaultTurnstileVerifyURL
	}

	return turnstileConfig{
		SiteKey:          strings.TrimSpace(s.GetString(consts.ConfigCaptchaTurnstileSiteKey)),
		SecretKey:        strings.TrimSpace(s.GetString(consts.ConfigCaptchaTurnstileSecretKey)),
		VerifyURL:        verifyURL,
		ExpectedHostname: strings.TrimSpace(s.GetString(consts.ConfigCaptchaTurnstileExpectedHostname)),
	}
}

// getRecaptchaConfig 读取 reCAPTCHA 配置。
func (s *Service) getRecaptchaConfig() recaptchaConfig {
	verifyURL := strings.TrimSpace(s.GetString(consts.ConfigCaptchaRecaptchaVerifyURL))
	if verifyURL == "" {
		verifyURL = defaultRecaptchaVerifyURL
	}

	return recaptchaConfig{
		SiteKey:          strings.TrimSpace(s.GetString(consts.ConfigCaptchaRecaptchaSiteKey)),
		SecretKey:        strings.TrimSpace(s.GetString(consts.ConfigCaptchaRecaptchaSecretKey)),
		VerifyURL:        verifyURL,
		ExpectedHostname: strings.TrimSpace(s.GetString(consts.ConfigCaptchaRecaptchaExpectedHostname)),
	}
}

// getHcaptchaConfig 读取 hCaptcha 配置。
func (s *Service) getHcaptchaConfig() hcaptchaConfig {
	verifyURL := strings.TrimSpace(s.GetString(consts.ConfigCaptchaHcaptchaVerifyURL))
	if verifyURL == "" {
		verifyURL = defaultHcaptchaVerifyURL
	}

	return hcaptchaConfig{
		SiteKey:          strings.TrimSpace(s.GetString(consts.ConfigCaptchaHcaptchaSiteKey)),
		SecretKey:        strings.TrimSpace(s.GetString(consts.ConfigCaptchaHcaptchaSecretKey)),
		VerifyURL:        verifyURL,
		ExpectedHostname: strings.TrimSpace(s.GetString(consts.ConfigCaptchaHcaptchaExpectedHostname)),
	}
}

// getGeetestConfig 读取 GeeTest 配置。
func (s *Service) getGeetestConfig() geetestConfig {
	verifyURL := strings.TrimSpace(s.GetString(consts.ConfigCaptchaGeetestVerifyURL))
	if verifyURL == "" {
		verifyURL = defaultGeetestVerifyURL
	}

	return geetestConfig{
		CaptchaID:  strings.TrimSpace(s.GetString(consts.ConfigCaptchaGeetestCaptchaID)),
		CaptchaKey: strings.TrimSpace(s.GetString(consts.ConfigCaptchaGeetestCaptchaKey)),
		VerifyURL:  verifyURL,
	}
}

// verifyImageCaptcha 校验图形验证码。
func verifyImageCaptcha(captchaID, captchaAnswer string) (bool, string) {
	if strings.TrimSpace(captchaID) == "" || strings.TrimSpace(captchaAnswer) == "" {
		return false, "验证码不能为空"
	}

	if !utils.VerifyCaptcha(captchaID, captchaAnswer) {
		return false, "验证码错误或已过期"
	}

	return true, ""
}

// verifyTurnstileCaptcha 校验 Turnstile challenge token。
func verifyTurnstileCaptcha(cfg turnstileConfig, token, remoteIP string) (bool, string) {
	if cfg.SiteKey == "" || cfg.SecretKey == "" {
		return false, "验证码配置错误，请联系管理员"
	}
	if strings.TrimSpace(token) == "" {
		return false, "请完成人机验证"
	}

	ok, err := verifyTurnstile(cfg, token, remoteIP)
	if err != nil {
		log.Printf("⚠️ Turnstile 验证失败: %v", err)
		return false, "人机验证服务不可用，请稍后重试"
	}
	if !ok {
		return false, "人机验证失败，请重试"
	}

	return true, ""
}

// verifyRecaptchaCaptcha 校验 reCAPTCHA challenge token。
func verifyRecaptchaCaptcha(cfg recaptchaConfig, token, remoteIP string) (bool, string) {
	if cfg.SiteKey == "" || cfg.SecretKey == "" {
		return false, "验证码配置错误，请联系管理员"
	}
	if strings.TrimSpace(token) == "" {
		return false, "请完成人机验证"
	}

	ok, err := verifyRecaptcha(cfg, token, remoteIP)
	if err != nil {
		log.Printf("⚠️ reCAPTCHA 验证失败: %v", err)
		return false, "人机验证服务不可用，请稍后重试"
	}
	if !ok {
		return false, "人机验证失败，请重试"
	}

	return true, ""
}

// verifyHcaptchaCaptcha 校验 hCaptcha challenge token。
func verifyHcaptchaCaptcha(cfg hcaptchaConfig, token, remoteIP string) (bool, string) {
	if cfg.SiteKey == "" || cfg.SecretKey == "" {
		return false, "验证码配置错误，请联系管理员"
	}
	if strings.TrimSpace(token) == "" {
		return false, "请完成人机验证"
	}

	ok, err := verifyHcaptcha(cfg, token, remoteIP)
	if err != nil {
		log.Printf("⚠️ hCaptcha 验证失败: %v", err)
		return false, "人机验证服务不可用，请稍后重试"
	}
	if !ok {
		return false, "人机验证失败，请重试"
	}

	return true, ""
}

// verifyGeetestCaptcha 校验 GeeTest challenge token。
// GeeTest 模式下，captcha_token 是 JSON 字符串
func verifyGeetestCaptcha(cfg geetestConfig, token string) (bool, string) {
	if cfg.CaptchaID == "" || cfg.CaptchaKey == "" {
		return false, "验证码配置错误，请联系管理员"
	}
	if strings.TrimSpace(token) == "" {
		return false, "请完成人机验证"
	}

	tokenBytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return false, "验证码参数格式错误"
	}

	var payload moduledto.GeetestVerifyTokenPayload
	if err := json.Unmarshal(tokenBytes, &payload); err != nil {
		return false, "验证码参数格式错误"
	}

	if payload.LotNumber == "" || payload.CaptchaOutput == "" || payload.PassToken == "" || payload.GenTime == "" {
		return false, "验证码参数不完整"
	}

	ok, err := verifyGeetest(cfg, payload)
	if err != nil {
		log.Printf("⚠️ GeeTest 验证失败: %v", err)
		return false, "人机验证服务不可用，请稍后重试"
	}
	if !ok {
		return false, "人机验证失败，请重试"
	}

	return true, ""
}

// verifyTurnstile 请求 Turnstile 服务端验证接口。
func verifyTurnstile(cfg turnstileConfig, token, remoteIP string) (bool, error) {
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

	resp, err := captchaHTTPClient.Do(req)
	if err != nil {
		return false, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("turnstile verify status code: %d", resp.StatusCode)
	}

	var result moduledto.TurnstileVerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	if !result.Success {
		return false, nil
	}

	expectedHostname := strings.TrimSpace(cfg.ExpectedHostname)
	if expectedHostname != "" && !strings.EqualFold(expectedHostname, strings.TrimSpace(result.Hostname)) {
		return false, fmt.Errorf("turnstile hostname mismatch: expected %s, got %s", expectedHostname, result.Hostname)
	}

	return true, nil
}

// verifyRecaptcha 请求 reCAPTCHA 服务端验证接口。
func verifyRecaptcha(cfg recaptchaConfig, token, remoteIP string) (bool, error) {
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

	resp, err := captchaHTTPClient.Do(req)
	if err != nil {
		return false, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("recaptcha verify status code: %d", resp.StatusCode)
	}

	var result moduledto.RecaptchaVerifyResponse
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

// verifyHcaptcha 请求 hCaptcha 服务端验证接口。
func verifyHcaptcha(cfg hcaptchaConfig, token, remoteIP string) (bool, error) {
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

	resp, err := captchaHTTPClient.Do(req)
	if err != nil {
		return false, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("hcaptcha verify status code: %d", resp.StatusCode)
	}

	var result moduledto.HcaptchaVerifyResponse
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

// verifyGeetest 请求 GeeTest 服务端验证接口。
func verifyGeetest(cfg geetestConfig, payload moduledto.GeetestVerifyTokenPayload) (bool, error) {
	form := url.Values{}
	form.Set("captcha_id", cfg.CaptchaID)
	form.Set("lot_number", payload.LotNumber)
	form.Set("captcha_output", payload.CaptchaOutput)
	form.Set("pass_token", payload.PassToken)
	form.Set("gen_time", payload.GenTime)
	form.Set("sign_token", buildGeetestSignToken(payload.LotNumber, cfg.CaptchaKey))

	req, err := http.NewRequest(http.MethodPost, cfg.VerifyURL, strings.NewReader(form.Encode()))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := captchaHTTPClient.Do(req)
	if err != nil {
		return false, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("geetest verify status code: %d", resp.StatusCode)
	}

	var result moduledto.GeetestVerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	return strings.EqualFold(strings.TrimSpace(result.Result), "success"), nil
}

// buildGeetestSignToken 生成 GeeTest v4 所需 sign_token。
func buildGeetestSignToken(lotNumber, captchaKey string) string {
	mac := hmac.New(sha256.New, []byte(captchaKey))
	_, _ = mac.Write([]byte(lotNumber))
	return hex.EncodeToString(mac.Sum(nil))
}
