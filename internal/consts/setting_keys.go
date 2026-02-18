package consts

const (

	// ConfigSiteName 网站名称
	ConfigSiteName = "site_name"

	// ConfigSiteDescription 网站描述
	ConfigSiteDescription = "site_description"

	// ConfigSiteLogo 网站Logo URL
	ConfigSiteLogo = "site_logo"

	// ConfigSiteFavicon 网站Favicon URL
	ConfigSiteFavicon = "site_favicon"

	// ConfigBaseURL 网站基础URL (例如 http://localhost:8080)
	ConfigBaseURL = "base_url"

	// ConfigAllowInit 是否允许初始化管理员账号 (true/false)
	ConfigAllowInit = "allow_init"

	// ConfigAllowRegister 是否开放注册 (true/false)
	ConfigAllowRegister = "allow_register"

	// ConfigEnableSMTP 是否启用SMTP发送邮件 (true/false)
	ConfigEnableSMTP = "enable_smtp"

	// ConfigBlockUnverifiedUsers 是否阻止未验证邮箱用户登录 (true/false)
	ConfigBlockUnverifiedUsers = "block_unverified_users"

	// ConfigRequireEmailVerification 注册是否强制要求验证邮箱 (true/false)
	ConfigRequireEmailVerification = "require_email_verification"

	// ConfigMaxUploadSize 图片最大上传限制 (MB)
	ConfigMaxUploadSize = "max_upload_size"

	// ConfigAllowFileExtensions 允许上传的文件扩展名 (逗号分隔)
	ConfigAllowFileExtensions = "allow_file_extensions"

	// ConfigDefaultStorageQuota 默认存储配额 (字节)
	ConfigDefaultStorageQuota = "default_storage_quota"

	// ConfigRateLimitEnabled 是否开启限流
	ConfigRateLimitEnabled = "rate_limit_enabled"

	// ConfigRateLimitAuthRPS 认证接口限流 RPS
	ConfigRateLimitAuthRPS = "rate_limit_auth_rps"

	// ConfigRateLimitAuthBurst 认证接口限流 Burst
	ConfigRateLimitAuthBurst = "rate_limit_auth_burst"

	// ConfigRateLimitUploadRPS 上传接口限流 RPS
	ConfigRateLimitUploadRPS = "rate_limit_upload_rps"

	// ConfigRateLimitUploadBurst 上传接口限流 Burst
	ConfigRateLimitUploadBurst = "rate_limit_upload_burst"

	// ConfigEnableSensitiveRateLimit 是否开启敏感操作（忘记密码、修改用户名、修改邮箱）频率限制
	ConfigEnableSensitiveRateLimit = "enable_sensitive_rate_limit"

	// ConfigRateLimitPasswordResetIntervalSeconds 忘记密码请求最小间隔（秒）
	ConfigRateLimitPasswordResetIntervalSeconds = "rate_limit_password_reset_interval_seconds"

	// ConfigRateLimitUsernameUpdateIntervalSeconds 修改用户名请求最小间隔（秒）
	ConfigRateLimitUsernameUpdateIntervalSeconds = "rate_limit_username_update_interval_seconds"

	// ConfigRateLimitEmailUpdateIntervalSeconds 修改邮箱请求最小间隔（秒）
	ConfigRateLimitEmailUpdateIntervalSeconds = "rate_limit_email_update_interval_seconds"

	// ConfigMaxRequestBodySize 最大API请求体大小 (MB, 排除文件上传)
	ConfigMaxRequestBodySize = "max_request_body_size"

	// ConfigStaticCacheControl 静态资源缓存设置 (Cache-Control header value)
	ConfigStaticCacheControl = "static_cache_control"

	// ConfigTrustedProxies 可信代理列表 (逗号分隔，留空表示不信任代理头)
	ConfigTrustedProxies = "trusted_proxies"

	// ConfigCaptchaProvider 验证码提供方 (image, turnstile, recaptcha, hcaptcha, geetest)
	ConfigCaptchaProvider = "captcha_provider"

	// ConfigCaptchaTurnstileSiteKey Turnstile Site Key
	ConfigCaptchaTurnstileSiteKey = "captcha_turnstile_site_key"

	// ConfigCaptchaTurnstileSecretKey Turnstile Secret Key
	ConfigCaptchaTurnstileSecretKey = "captcha_turnstile_secret_key"

	// ConfigCaptchaTurnstileVerifyURL Turnstile 验证接口地址
	ConfigCaptchaTurnstileVerifyURL = "captcha_turnstile_verify_url"

	// ConfigCaptchaTurnstileExpectedHostname Turnstile 预期回传主机名
	ConfigCaptchaTurnstileExpectedHostname = "captcha_turnstile_expected_hostname"

	// ConfigCaptchaRecaptchaSiteKey reCAPTCHA Site Key
	ConfigCaptchaRecaptchaSiteKey = "captcha_recaptcha_site_key"

	// ConfigCaptchaRecaptchaSecretKey reCAPTCHA Secret Key
	ConfigCaptchaRecaptchaSecretKey = "captcha_recaptcha_secret_key"

	// ConfigCaptchaRecaptchaVerifyURL reCAPTCHA 验证接口地址
	ConfigCaptchaRecaptchaVerifyURL = "captcha_recaptcha_verify_url"

	// ConfigCaptchaRecaptchaExpectedHostname reCAPTCHA 预期回传主机名
	ConfigCaptchaRecaptchaExpectedHostname = "captcha_recaptcha_expected_hostname"

	// ConfigCaptchaHcaptchaSiteKey hCaptcha Site Key
	ConfigCaptchaHcaptchaSiteKey = "captcha_hcaptcha_site_key"

	// ConfigCaptchaHcaptchaSecretKey hCaptcha Secret Key
	ConfigCaptchaHcaptchaSecretKey = "captcha_hcaptcha_secret_key"

	// ConfigCaptchaHcaptchaVerifyURL hCaptcha 验证接口地址
	ConfigCaptchaHcaptchaVerifyURL = "captcha_hcaptcha_verify_url"

	// ConfigCaptchaHcaptchaExpectedHostname hCaptcha 预期回传主机名
	ConfigCaptchaHcaptchaExpectedHostname = "captcha_hcaptcha_expected_hostname"

	// ConfigCaptchaGeetestCaptchaID GeeTest Captcha ID
	ConfigCaptchaGeetestCaptchaID = "captcha_geetest_captcha_id"

	// ConfigCaptchaGeetestCaptchaKey GeeTest Captcha Key
	ConfigCaptchaGeetestCaptchaKey = "captcha_geetest_captcha_key"

	// ConfigCaptchaGeetestVerifyURL GeeTest 验证接口地址
	ConfigCaptchaGeetestVerifyURL = "captcha_geetest_verify_url"
)
