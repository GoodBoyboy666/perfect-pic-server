package captcha

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
	"strings"
)

const DefaultGeetestVerifyURL = "https://gcaptcha4.geetest.com/validate"

type GeetestConfig struct {
	CaptchaID  string
	CaptchaKey string
	VerifyURL  string
}

type GeetestVerifyTokenPayload struct {
	LotNumber     string `json:"lot_number"`
	CaptchaOutput string `json:"captcha_output"`
	PassToken     string `json:"pass_token"`
	GenTime       string `json:"gen_time"`
}

type GeetestVerifyResponse struct {
	Result string `json:"result"`
	Reason string `json:"reason"`
}

// VerifyGeetestCaptcha GeeTest 模式下，captcha_token 是 base64 编码的 JSON 字符串。
func VerifyGeetestCaptcha(cfg GeetestConfig, token string) (bool, string) {
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

	var payload GeetestVerifyTokenPayload
	if err := json.Unmarshal(tokenBytes, &payload); err != nil {
		return false, "验证码参数格式错误"
	}

	if payload.LotNumber == "" || payload.CaptchaOutput == "" || payload.PassToken == "" || payload.GenTime == "" {
		return false, "验证码参数不完整"
	}

	ok, err := verifyGeetest(httpClient, cfg, payload)
	if err != nil {
		log.Printf("⚠️ GeeTest 验证失败: %v", err)
		return false, "人机验证服务不可用，请稍后重试"
	}
	if !ok {
		return false, "人机验证失败，请重试"
	}

	return true, ""
}

func verifyGeetest(httpClient *http.Client, cfg GeetestConfig, payload GeetestVerifyTokenPayload) (bool, error) {
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

	resp, err := httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("geetest verify status code: %d", resp.StatusCode)
	}

	var result GeetestVerifyResponse
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
