package utils

import "github.com/mojocn/base64Captcha"

var store = base64Captcha.DefaultMemStore

// 生成验证码

func MakeCaptcha() (id, b64s string, answer string, err error) {
	// height: 80, width: 240, noiseCount: 0(噪点), showLineOptions: 2(干扰线), length: 4(4位数)
	// NewDriverDigit 生成数字验证码
	driver := base64Captcha.NewDriverDigit(80, 240, 4, 0.7, 80)

	// 创建验证码实例
	c := base64Captcha.NewCaptcha(driver, store)

	// 生成 (id 是验证码的唯一标识，b64s 是图片的 base64 字符串)
	return c.Generate()
}

//校验验证码

func VerifyCaptcha(id string, answer string) bool {
	return store.Verify(id, answer, true)
}
