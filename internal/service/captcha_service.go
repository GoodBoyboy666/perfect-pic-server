package service

import (
	"perfect-pic-server/internal/consts"
	moduledto "perfect-pic-server/internal/dto"
	"perfect-pic-server/internal/pkg/captcha"
	"strings"
)

// GetCaptchaProviderInfo 获取当前验证码提供方与前端可见配置。
func (s *CaptchaService) GetCaptchaProviderInfo() moduledto.CaptchaProviderResponse {
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
func (s *CaptchaService) VerifyCaptchaChallenge(captchaID, captchaAnswer, captchaToken, remoteIP string) (bool, string) {
	provider := s.loadCaptchaProvider()

	switch provider {
	case consts.CaptchaProviderDisabled:
		return true, ""
	case consts.CaptchaProviderTurnstile:
		return captcha.VerifyTurnstileCaptcha(s.getTurnstileConfig(), captchaToken, remoteIP)
	case consts.CaptchaProviderRecaptcha:
		return captcha.VerifyRecaptchaCaptcha(s.getRecaptchaConfig(), captchaToken, remoteIP)
	case consts.CaptchaProviderHcaptcha:
		return captcha.VerifyHcaptchaCaptcha(s.getHcaptchaConfig(), captchaToken, remoteIP)
	case consts.CaptchaProviderGeetest:
		return captcha.VerifyGeetestCaptcha(s.getGeetestConfig(), captchaToken)
	case consts.CaptchaProviderImage:
		fallthrough
	default:
		return captcha.VerifyImageCaptcha(captchaID, captchaAnswer)
	}
}

// loadCaptchaProvider 读取并标准化验证码提供方。
func (s *CaptchaService) loadCaptchaProvider() string {
	provider := strings.ToLower(strings.TrimSpace(s.dbConfig.GetString(consts.ConfigCaptchaProvider)))
	switch provider {
	case consts.CaptchaProviderDisabled, consts.CaptchaProviderImage, consts.CaptchaProviderTurnstile, consts.CaptchaProviderRecaptcha, consts.CaptchaProviderHcaptcha, consts.CaptchaProviderGeetest:
		return provider
	default:
		return consts.CaptchaProviderImage
	}
}

func (s *CaptchaService) getGeetestConfig() captcha.GeetestConfig {
	verifyURL := strings.TrimSpace(s.dbConfig.GetString(consts.ConfigCaptchaGeetestVerifyURL))
	if verifyURL == "" {
		verifyURL = captcha.DefaultGeetestVerifyURL
	}

	return captcha.GeetestConfig{
		CaptchaID:  strings.TrimSpace(s.dbConfig.GetString(consts.ConfigCaptchaGeetestCaptchaID)),
		CaptchaKey: strings.TrimSpace(s.dbConfig.GetString(consts.ConfigCaptchaGeetestCaptchaKey)),
		VerifyURL:  verifyURL,
	}
}

func (s *CaptchaService) getHcaptchaConfig() captcha.HcaptchaConfig {
	verifyURL := strings.TrimSpace(s.dbConfig.GetString(consts.ConfigCaptchaHcaptchaVerifyURL))
	if verifyURL == "" {
		verifyURL = captcha.DefaultHcaptchaVerifyURL
	}

	return captcha.HcaptchaConfig{
		SiteKey:          strings.TrimSpace(s.dbConfig.GetString(consts.ConfigCaptchaHcaptchaSiteKey)),
		SecretKey:        strings.TrimSpace(s.dbConfig.GetString(consts.ConfigCaptchaHcaptchaSecretKey)),
		VerifyURL:        verifyURL,
		ExpectedHostname: strings.TrimSpace(s.dbConfig.GetString(consts.ConfigCaptchaHcaptchaExpectedHostname)),
	}
}

func (s *CaptchaService) getRecaptchaConfig() captcha.RecaptchaConfig {
	verifyURL := strings.TrimSpace(s.dbConfig.GetString(consts.ConfigCaptchaRecaptchaVerifyURL))
	if verifyURL == "" {
		verifyURL = captcha.DefaultRecaptchaVerifyURL
	}

	return captcha.RecaptchaConfig{
		SiteKey:          strings.TrimSpace(s.dbConfig.GetString(consts.ConfigCaptchaRecaptchaSiteKey)),
		SecretKey:        strings.TrimSpace(s.dbConfig.GetString(consts.ConfigCaptchaRecaptchaSecretKey)),
		VerifyURL:        verifyURL,
		ExpectedHostname: strings.TrimSpace(s.dbConfig.GetString(consts.ConfigCaptchaRecaptchaExpectedHostname)),
	}
}

func (s *CaptchaService) getTurnstileConfig() captcha.TurnstileConfig {
	verifyURL := strings.TrimSpace(s.dbConfig.GetString(consts.ConfigCaptchaTurnstileVerifyURL))
	if verifyURL == "" {
		verifyURL = captcha.DefaultTurnstileVerifyURL
	}

	return captcha.TurnstileConfig{
		SiteKey:          strings.TrimSpace(s.dbConfig.GetString(consts.ConfigCaptchaTurnstileSiteKey)),
		SecretKey:        strings.TrimSpace(s.dbConfig.GetString(consts.ConfigCaptchaTurnstileSecretKey)),
		VerifyURL:        verifyURL,
		ExpectedHostname: strings.TrimSpace(s.dbConfig.GetString(consts.ConfigCaptchaTurnstileExpectedHostname)),
	}
}
