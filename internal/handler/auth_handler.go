package handler

import (
	"net/http"
	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/service"
	"perfect-pic-server/internal/utils"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func Login(c *gin.Context) {
	cfg := config.Get()
	var req struct {
		Username      string `json:"username" binding:"required"`
		Password      string `json:"password" binding:"required"`
		CaptchaID     string `json:"captcha_id" binding:"required"`
		CaptchaAnswer string `json:"captcha_answer" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if !utils.VerifyCaptcha(req.CaptchaID, req.CaptchaAnswer) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "验证码错误或已过期"})
		return
	}

	var user model.User
	result := db.DB.Where("username = ?", req.Username).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	if user.Status == 2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "该账号已被封禁"})
		return
	}
	if user.Status == 3 {
		c.JSON(http.StatusForbidden, gin.H{"error": "该账号已停用"})
		return
	}

	// 签发 Token
	token, _ := utils.GenerateToken(user.ID, user.Username, user.Admin, time.Hour*time.Duration(cfg.JWT.ExpirationHours))

	c.JSON(http.StatusOK, gin.H{
		"token":   token,
		"message": "登录成功",
	})
}

func Register(c *gin.Context) {
	var req struct {
		Username      string `json:"username" binding:"required"`
		Password      string `json:"password" binding:"required"`
		CaptchaID     string `json:"captcha_id" binding:"required"`
		CaptchaAnswer string `json:"captcha_answer" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误"})
		return
	}

	if ok, msg := utils.ValidatePassword(req.Password); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	if ok, msg := utils.ValidateUsername(req.Username); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	if !utils.VerifyCaptcha(req.CaptchaID, req.CaptchaAnswer) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "验证码错误或已过期"})
		return
	}

	allowRegister := service.GetBool(consts.ConfigAllowRegister)
	if !allowRegister {
		c.JSON(http.StatusForbidden, gin.H{"error": "注册功能已关闭"})
		return
	}

	var existingUser model.User
	result := db.DB.Where("username = ?", req.Username).First(&existingUser)
	if result.RowsAffected > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	newUser := model.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Admin:    false,
		Avatar:   "", // 默认头像
	}

	if err := db.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "注册成功，请登录"})
}

func GetRegisterState(c *gin.Context) {
	allowRegister := service.GetBool(consts.ConfigAllowRegister)
	c.JSON(http.StatusOK, gin.H{
		"allow_register": allowRegister,
	})
}
