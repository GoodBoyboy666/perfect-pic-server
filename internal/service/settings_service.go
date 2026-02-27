package service

import (
	settingsrepo "perfect-pic-server/internal/repository"
	"strconv"
	"sync"

	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/model"
)

const defaultValueNotFound = "||__NOT_FOUND__||"
const defaultStorageQuotaBytes int64 = 1073741824

var DefaultSettings = []model.Setting{
	{Key: consts.ConfigSiteName, Value: "Perfect Pic", Desc: "网站名称", Category: "常规"},
	{Key: consts.ConfigSiteDescription, Value: "记录与分享完美瞬间", Desc: "网站描述", Category: "常规"},
	{Key: consts.ConfigSiteLogo, Value: "", Desc: "网站Logo URL", Category: "常规"},
	{Key: consts.ConfigSiteFavicon, Value: "", Desc: "网站Favicon URL", Category: "常规"},
	{Key: consts.ConfigBaseURL, Value: "http://localhost", Desc: "网站基础URL", Category: "常规"},
	{Key: consts.ConfigAllowInit, Value: "true", Desc: "允许初始化管理员账号", Category: "安全"},
	{Key: consts.ConfigAllowRegister, Value: "true", Desc: "开放注册", Category: "安全"},
	{Key: consts.ConfigEnableSMTP, Value: "false", Desc: "启用 SMTP 发送邮件", Category: "邮件服务"},
	{Key: consts.ConfigBlockUnverifiedUsers, Value: "false", Desc: "阻止未验证邮箱用户登录", Category: "安全"},
	{Key: consts.ConfigRequireEmailVerification, Value: "false", Desc: "强制要求注册验证邮箱", Category: "安全"},
	{Key: consts.ConfigMaxUploadSize, Value: "10", Desc: "单个文件最大大小 (MB)", Category: "上传"},
	{Key: consts.ConfigAllowFileExtensions, Value: ".jpg,.jpeg,.png,.gif,.webp", Desc: "允许上传的文件扩展名", Category: "上传"},
	{Key: consts.ConfigDefaultStorageQuota, Value: "1073741824", Desc: "默认用户存储配额 (Bytes, 默认为1GB)", Category: "上传"},
	{Key: consts.ConfigRateLimitEnabled, Value: "true", Desc: "开启接口限流", Category: "速率限制"},
	{Key: consts.ConfigRateLimitAuthRPS, Value: "0.5", Desc: "认证接口每秒请求限制 (RPS)", Category: "速率限制"},
	{Key: consts.ConfigRateLimitAuthBurst, Value: "2", Desc: "认证接口突发请求限制", Category: "速率限制"},
	{Key: consts.ConfigRateLimitUploadRPS, Value: "1.0", Desc: "上传接口每秒请求限制 (RPS)", Category: "速率限制"},
	{Key: consts.ConfigRateLimitUploadBurst, Value: "5", Desc: "上传接口突发请求限制", Category: "速率限制"},
	{Key: consts.ConfigEnableSensitiveRateLimit, Value: "true", Desc: "开启敏感操作频率限制", Category: "速率限制"},
	{Key: consts.ConfigRateLimitPasswordResetIntervalSeconds, Value: "120", Desc: "忘记密码请求最小间隔（秒）", Category: "速率限制"},
	{Key: consts.ConfigRateLimitUsernameUpdateIntervalSeconds, Value: "120", Desc: "修改用户名请求最小间隔（秒）", Category: "速率限制"},
	{Key: consts.ConfigRateLimitEmailUpdateIntervalSeconds, Value: "120", Desc: "修改邮箱请求最小间隔（秒）", Category: "速率限制"},
	{Key: consts.ConfigMaxRequestBodySize, Value: "2", Desc: "非文件上传接口最大请求体限制 (MB)", Category: "服务"},
	{Key: consts.ConfigStaticCacheControl, Value: "public, max-age=31536000", Desc: "静态资源缓存设置 (Cache-Control)", Category: "服务"},
	{Key: consts.ConfigCaptchaProvider, Value: "image", Desc: "验证码提供方（空=关闭, image, turnstile, recaptcha, hcaptcha, geetest）", Category: "验证码"},
	{Key: consts.ConfigCaptchaTurnstileSiteKey, Value: "", Desc: "Cloudflare Turnstile Site Key", Category: "验证码"},
	{Key: consts.ConfigCaptchaTurnstileSecretKey, Value: "", Desc: "Cloudflare Turnstile Secret Key", Category: "验证码", Sensitive: true},
	{Key: consts.ConfigCaptchaTurnstileVerifyURL, Value: "", Desc: "Cloudflare Turnstile 校验地址，留空使用官方默认", Category: "验证码"},
	{Key: consts.ConfigCaptchaTurnstileExpectedHostname, Value: "", Desc: "Cloudflare Turnstile 期望回传域名（可选）", Category: "验证码"},
	{Key: consts.ConfigCaptchaRecaptchaSiteKey, Value: "", Desc: "Google reCAPTCHA Site Key", Category: "验证码"},
	{Key: consts.ConfigCaptchaRecaptchaSecretKey, Value: "", Desc: "Google reCAPTCHA Secret Key", Category: "验证码", Sensitive: true},
	{Key: consts.ConfigCaptchaRecaptchaVerifyURL, Value: "", Desc: "Google reCAPTCHA 校验地址，留空使用官方默认", Category: "验证码"},
	{Key: consts.ConfigCaptchaRecaptchaExpectedHostname, Value: "", Desc: "Google reCAPTCHA 期望回传域名（可选）", Category: "验证码"},
	{Key: consts.ConfigCaptchaHcaptchaSiteKey, Value: "", Desc: "hCaptcha Site Key", Category: "验证码"},
	{Key: consts.ConfigCaptchaHcaptchaSecretKey, Value: "", Desc: "hCaptcha Secret Key", Category: "验证码", Sensitive: true},
	{Key: consts.ConfigCaptchaHcaptchaVerifyURL, Value: "", Desc: "hCaptcha 校验地址，留空使用官方默认", Category: "验证码"},
	{Key: consts.ConfigCaptchaHcaptchaExpectedHostname, Value: "", Desc: "hCaptcha 期望回传域名（可选）", Category: "验证码"},
	{Key: consts.ConfigCaptchaGeetestCaptchaID, Value: "", Desc: "GeeTest Captcha ID", Category: "验证码"},
	{Key: consts.ConfigCaptchaGeetestCaptchaKey, Value: "", Desc: "GeeTest Captcha Key", Category: "验证码", Sensitive: true},
	{Key: consts.ConfigCaptchaGeetestVerifyURL, Value: "", Desc: "GeeTest 校验地址，留空使用官方默认", Category: "验证码"},
}

type Service struct {
	userStore     settingsrepo.UserStore
	imageStore    settingsrepo.ImageStore
	settingStore  settingsrepo.SettingStore
	systemStore   settingsrepo.SystemStore
	settingsCache sync.Map
}

// AppService is kept as a compatibility alias for middleware/main usage.
type AppService = Service

func New(settingStore settingsrepo.SettingStore) *Service {
	return &Service{settingStore: settingStore}
}

func NewAppService(
	userStore settingsrepo.UserStore,
	imageStore settingsrepo.ImageStore,
	settingStore settingsrepo.SettingStore,
	systemStore settingsrepo.SystemStore,
) *Service {
	return &Service{
		userStore:    userStore,
		imageStore:   imageStore,
		settingStore: settingStore,
		systemStore:  systemStore,
	}
}

func (s *Service) ClearCache() {
	s.settingsCache.Range(func(key, value interface{}) bool {
		s.settingsCache.Delete(key)
		return true
	})
}

func (s *Service) InitializeSettings() error {
	if err := s.settingStore.InitializeDefaults(DefaultSettings); err != nil {
		return err
	}

	allowedKeys := make([]string, 0, len(DefaultSettings))
	for _, def := range DefaultSettings {
		allowedKeys = append(allowedKeys, def.Key)
	}
	if err := s.settingStore.DeleteNotInKeys(allowedKeys); err != nil {
		return err
	}

	s.ClearCache()
	return nil
}

func (s *Service) GetString(key string) string {
	if val, ok := s.settingsCache.Load(key); ok {
		strVal, ok := val.(string)
		if !ok {
			s.settingsCache.Delete(key)
		} else {
			if strVal == defaultValueNotFound {
				return ""
			}
			return strVal
		}
	}

	setting, err := s.settingStore.FindByKey(key)
	if err != nil {
		for _, def := range DefaultSettings {
			if def.Key == key {
				newSetting := def
				_ = s.settingStore.Create(&newSetting)

				s.settingsCache.Store(key, newSetting.Value)
				return newSetting.Value
			}
		}

		s.settingsCache.Store(key, defaultValueNotFound)
		return ""
	}
	s.settingsCache.Store(key, setting.Value)

	return setting.Value
}

func (s *Service) GetInt(key string) int {
	valStr := s.GetString(key)
	if valStr == "" {
		return 0
	}

	val, err := strconv.Atoi(valStr)
	if err != nil {
		return 0
	}
	return val
}

func (s *Service) GetInt64(key string) int64 {
	valStr := s.GetString(key)
	if valStr == "" {
		return 0
	}

	val, err := strconv.ParseInt(valStr, 10, 64)
	if err != nil {
		return 0
	}
	return val
}

func (s *Service) GetFloat64(key string) float64 {
	valStr := s.GetString(key)
	if valStr == "" {
		return 0
	}

	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return 0
	}
	return val
}

func (s *Service) GetBool(key string) bool {
	valStr := s.GetString(key)
	if valStr == "" {
		return false
	}

	val, err := strconv.ParseBool(valStr)
	if err != nil {
		return false
	}
	return val
}

// GetDefaultStorageQuota 获取默认存储配额（字节）。
// 当配置值非法（<= 0）时回退到 1GB。
func (s *Service) GetDefaultStorageQuota() int64 {
	quota := s.GetInt64(consts.ConfigDefaultStorageQuota)
	if quota <= 0 {
		return defaultStorageQuotaBytes
	}
	return quota
}
