package consts

const (
	// CaptchaProviderDisabled 关闭验证码。
	CaptchaProviderDisabled = ""
	// CaptchaProviderImage 图形验证码。
	CaptchaProviderImage = "image"
	// CaptchaProviderTurnstile Cloudflare Turnstile。
	CaptchaProviderTurnstile = "turnstile"
	// CaptchaProviderRecaptcha Google reCAPTCHA。
	CaptchaProviderRecaptcha = "recaptcha"
	// CaptchaProviderHcaptcha hCaptcha。
	CaptchaProviderHcaptcha = "hcaptcha"
	// CaptchaProviderGeetest GeeTest。
	CaptchaProviderGeetest = "geetest"
)
