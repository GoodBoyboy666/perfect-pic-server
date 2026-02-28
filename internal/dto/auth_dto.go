package dto

import "encoding/json"

type LoginRequest struct {
	Username      string `json:"username" binding:"required"`
	Password      string `json:"password" binding:"required"`
	CaptchaID     string `json:"captcha_id"`
	CaptchaAnswer string `json:"captcha_answer"`
	CaptchaToken  string `json:"captcha_token"`
}

type RegisterRequest struct {
	Username      string `json:"username" binding:"required"`
	Password      string `json:"password" binding:"required"`
	Email         string `json:"email" binding:"required"`
	CaptchaID     string `json:"captcha_id"`
	CaptchaAnswer string `json:"captcha_answer"`
	CaptchaToken  string `json:"captcha_token"`
}

type TokenRequest struct {
	Token string `json:"token" binding:"required"`
}

type RequestPasswordResetRequest struct {
	Email         string `json:"email" binding:"required,email"`
	CaptchaID     string `json:"captcha_id"`
	CaptchaAnswer string `json:"captcha_answer"`
	CaptchaToken  string `json:"captcha_token"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type BeginPasskeyLoginRequest struct {
	CaptchaID     string `json:"captcha_id"`
	CaptchaAnswer string `json:"captcha_answer"`
	CaptchaToken  string `json:"captcha_token"`
}

type FinishPasskeyLoginRequest struct {
	SessionID  string          `json:"session_id" binding:"required"`
	Credential json.RawMessage `json:"credential" binding:"required"`
}

type CaptchaProviderResponse struct {
	Provider     string            `json:"provider"`
	PublicConfig map[string]string `json:"public_config,omitempty"`
}

type UserPasskeyResponse struct {
	ID           uint   `json:"id"`
	CredentialID string `json:"credential_id"`
	Name         string `json:"name"`
	CreatedAt    int64  `json:"created_at"`
}

type TurnstileVerifyResponse struct {
	Success  bool   `json:"success"`
	Hostname string `json:"hostname"`
}

type RecaptchaVerifyResponse struct {
	Success    bool     `json:"success"`
	Hostname   string   `json:"hostname"`
	Action     string   `json:"action"`
	Score      float64  `json:"score"`
	ErrorCodes []string `json:"error-codes"`
}

type HcaptchaVerifyResponse struct {
	Success    bool     `json:"success"`
	Hostname   string   `json:"hostname"`
	ErrorCodes []string `json:"error-codes"`
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
