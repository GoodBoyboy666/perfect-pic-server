package handler

import (
	"net/http"
	"perfect-pic-server/internal/consts"
	moduledto "perfect-pic-server/internal/modules/auth/dto"
	"perfect-pic-server/internal/modules/common/httpx"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Login(c *gin.Context) {
	var req moduledto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if verified, msg := h.authService.VerifyCaptchaChallenge(req.CaptchaID, req.CaptchaAnswer, req.CaptchaToken, c.ClientIP()); !verified {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	token, err := h.authService.LoginUser(req.Username, req.Password)
	if err != nil {
		httpx.WriteServiceError(c, err, "登录失败，请稍后重试")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   token,
		"message": "登录成功",
	})
}

func (h *Handler) Register(c *gin.Context) {
	var req moduledto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误"})
		return
	}

	if verified, msg := h.authService.VerifyCaptchaChallenge(req.CaptchaID, req.CaptchaAnswer, req.CaptchaToken, c.ClientIP()); !verified {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	if err := h.authService.RegisterUser(req.Username, req.Password, req.Email); err != nil {
		httpx.WriteServiceError(c, err, "注册失败，请稍后重试")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "注册成功，请前往邮箱验证"})
}

func (h *Handler) EmailVerify(c *gin.Context) {
	var req moduledto.TokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	tokenString := req.Token

	alreadyVerified, err := h.authService.VerifyEmail(tokenString)
	if err != nil {
		httpx.WriteServiceError(c, err, "验证失败，请稍后重试")
		return
	}

	if alreadyVerified {
		c.JSON(http.StatusOK, gin.H{"message": "邮箱已验证，无需重复验证"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "邮箱验证成功，现在可以登录了"})
}

func (h *Handler) EmailChangeVerify(c *gin.Context) {
	var req moduledto.TokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	tokenString := req.Token

	if err := h.authService.VerifyEmailChange(tokenString); err != nil {
		httpx.WriteServiceError(c, err, "邮箱修改失败，请稍后重试")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "邮箱修改成功"})
}

// RequestPasswordReset 请求重置密码
func (h *Handler) RequestPasswordReset(c *gin.Context) {
	var req moduledto.RequestPasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if verified, msg := h.authService.VerifyCaptchaChallenge(req.CaptchaID, req.CaptchaAnswer, req.CaptchaToken, c.ClientIP()); !verified {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	if err := h.authService.RequestPasswordReset(req.Email); err != nil {
		httpx.WriteServiceError(c, err, "生成重置链接失败，请稍后重试")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "如果该邮箱已注册，重置密码邮件将发送至您的邮箱"})
}

// ResetPassword 执行重置密码
func (h *Handler) ResetPassword(c *gin.Context) {
	var req moduledto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if err := h.authService.ResetPassword(req.Token, req.NewPassword); err != nil {
		httpx.WriteServiceError(c, err, "密码重置失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码重置成功，请使用新密码登录"})
}

func (h *Handler) GetRegisterState(c *gin.Context) {
	initialized := h.authService.IsSystemInitialized()
	allowRegister := initialized && h.authService.GetBool(consts.ConfigAllowRegister)
	c.JSON(http.StatusOK, gin.H{
		"allow_register": allowRegister,
	})
}

// BeginPasskeyLogin 创建 Passkey 登录挑战并返回会话 ID。
func (h *Handler) BeginPasskeyLogin(c *gin.Context) {
	var req moduledto.BeginPasskeyLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if verified, msg := h.authService.VerifyCaptchaChallenge(req.CaptchaID, req.CaptchaAnswer, req.CaptchaToken, c.ClientIP()); !verified {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	sessionID, assertion, err := h.authService.BeginPasskeyLogin()
	if err != nil {
		httpx.WriteServiceError(c, err, "创建 Passkey 登录挑战失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id":        sessionID,
		"assertion_options": assertion,
	})
}

// FinishPasskeyLogin 完成 Passkey 登录校验并返回 JWT。
func (h *Handler) FinishPasskeyLogin(c *gin.Context) {
	var req moduledto.FinishPasskeyLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	token, err := h.authService.FinishPasskeyLogin(req.SessionID, req.Credential)
	if err != nil {
		httpx.WriteServiceError(c, err, "Passkey 登录失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   token,
		"message": "登录成功",
	})
}
