package handler

import (
	"net/http"
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/service"

	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	var req struct {
		Username      string `json:"username" binding:"required"`
		Password      string `json:"password" binding:"required"`
		CaptchaID     string `json:"captcha_id"`
		CaptchaAnswer string `json:"captcha_answer"`
		CaptchaToken  string `json:"captcha_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if verified, msg := service.VerifyCaptchaChallenge(req.CaptchaID, req.CaptchaAnswer, req.CaptchaToken, c.ClientIP()); !verified {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	token, err := service.LoginUser(req.Username, req.Password)
	if err != nil {
		writeServiceError(c, err, "登录失败，请稍后重试")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   token,
		"message": "登录成功",
	})
}

func Register(c *gin.Context) {
	var req struct {
		Username      string `json:"username" binding:"required"`
		Password      string `json:"password" binding:"required"`
		Email         string `json:"email" binding:"required"`
		CaptchaID     string `json:"captcha_id"`
		CaptchaAnswer string `json:"captcha_answer"`
		CaptchaToken  string `json:"captcha_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误"})
		return
	}

	if verified, msg := service.VerifyCaptchaChallenge(req.CaptchaID, req.CaptchaAnswer, req.CaptchaToken, c.ClientIP()); !verified {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	if err := service.RegisterUser(req.Username, req.Password, req.Email); err != nil {
		writeServiceError(c, err, "注册失败，请稍后重试")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "注册成功，请前往邮箱验证"})
}

func EmailVerify(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	tokenString := req.Token

	alreadyVerified, err := service.VerifyEmail(tokenString)
	if err != nil {
		writeServiceError(c, err, "验证失败，请稍后重试")
		return
	}

	if alreadyVerified {
		c.JSON(http.StatusOK, gin.H{"message": "邮箱已验证，无需重复验证"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "邮箱验证成功，现在可以登录了"})
}

func EmailChangeVerify(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	tokenString := req.Token

	if err := service.VerifyEmailChange(tokenString); err != nil {
		writeServiceError(c, err, "邮箱修改失败，请稍后重试")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "邮箱修改成功"})
}

// RequestPasswordReset 请求重置密码
func RequestPasswordReset(c *gin.Context) {
	var req struct {
		Email         string `json:"email" binding:"required,email"`
		CaptchaID     string `json:"captcha_id"`
		CaptchaAnswer string `json:"captcha_answer"`
		CaptchaToken  string `json:"captcha_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if verified, msg := service.VerifyCaptchaChallenge(req.CaptchaID, req.CaptchaAnswer, req.CaptchaToken, c.ClientIP()); !verified {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	if err := service.RequestPasswordReset(req.Email); err != nil {
		writeServiceError(c, err, "生成重置链接失败，请稍后重试")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "如果该邮箱已注册，重置密码邮件将发送至您的邮箱"})
}

// ResetPassword 执行重置密码
func ResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if err := service.ResetPassword(req.Token, req.NewPassword); err != nil {
		writeServiceError(c, err, "密码重置失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码重置成功，请使用新密码登录"})
}

func GetRegisterState(c *gin.Context) {
	initialized := service.IsSystemInitialized()
	allowRegister := initialized && service.GetBool(consts.ConfigAllowRegister)
	c.JSON(http.StatusOK, gin.H{
		"allow_register": allowRegister,
	})
}
