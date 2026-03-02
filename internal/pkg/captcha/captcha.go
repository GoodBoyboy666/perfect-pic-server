package captcha

import (
	"net/http"
	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/consts"
	moduledto "perfect-pic-server/internal/dto"
	"strings"
	"time"
)

// Captcha 封装验证码 provider 读取与校验流程。
type Captcha struct {
	dbConfig    *config.DBConfig
	httpClient  *http.Client
	defaultHTTP *http.Client
}

// NewCaptcha 创建验证码服务实例。
func NewCaptcha(dbConfig *config.DBConfig) *Captcha {
	defaultHTTP := &http.Client{Timeout: 5 * time.Second}
	return &Captcha{
		dbConfig:    dbConfig,
		httpClient:  defaultHTTP,
		defaultHTTP: defaultHTTP,
	}
}

// SetHTTPClient 覆盖验证码远程校验 HTTP 客户端；传 nil 会恢复默认客户端。
func (s *Captcha) SetHTTPClient(client *http.Client) {
	if client == nil {
		s.httpClient = s.defaultHTTP
		return
	}
	s.httpClient = client
}

// GetProviderInfo 获取当前验证码提供方与前端可见配置。
func (s *Captcha) GetProviderInfo() moduledto.CaptchaProviderResponse {
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

// GetCaptchaProviderInfo 兼容旧调用名。
func (s *Captcha) GetCaptchaProviderInfo() moduledto.CaptchaProviderResponse {
	return s.GetProviderInfo()
}

// VerifyChallenge 按当前 provider 校验验证码。
func (s *Captcha) VerifyChallenge(captchaID, captchaAnswer, captchaToken, remoteIP string) (bool, string) {
	provider := s.loadCaptchaProvider()

	switch provider {
	case consts.CaptchaProviderDisabled:
		return true, ""
	case consts.CaptchaProviderTurnstile:
		return verifyTurnstileCaptcha(s.getTurnstileConfig(), s.httpClient, captchaToken, remoteIP)
	case consts.CaptchaProviderRecaptcha:
		return verifyRecaptchaCaptcha(s.getRecaptchaConfig(), s.httpClient, captchaToken, remoteIP)
	case consts.CaptchaProviderHcaptcha:
		return verifyHcaptchaCaptcha(s.getHcaptchaConfig(), s.httpClient, captchaToken, remoteIP)
	case consts.CaptchaProviderGeetest:
		return verifyGeetestCaptcha(s.getGeetestConfig(), s.httpClient, captchaToken)
	case consts.CaptchaProviderImage:
		fallthrough
	default:
		return verifyImageCaptcha(captchaID, captchaAnswer)
	}
}

// VerifyCaptchaChallenge 兼容旧调用名。
func (s *Captcha) VerifyCaptchaChallenge(captchaID, captchaAnswer, captchaToken, remoteIP string) (bool, string) {
	return s.VerifyChallenge(captchaID, captchaAnswer, captchaToken, remoteIP)
}

// loadCaptchaProvider 读取并标准化验证码提供方。
func (s *Captcha) loadCaptchaProvider() string {
	provider := strings.ToLower(strings.TrimSpace(s.dbConfig.GetString(consts.ConfigCaptchaProvider)))
	switch provider {
	case consts.CaptchaProviderDisabled, consts.CaptchaProviderImage, consts.CaptchaProviderTurnstile, consts.CaptchaProviderRecaptcha, consts.CaptchaProviderHcaptcha, consts.CaptchaProviderGeetest:
		return provider
	default:
		return consts.CaptchaProviderImage
	}
}
