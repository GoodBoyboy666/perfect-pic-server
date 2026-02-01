package handler

import (
	"net/http"
	"perfect-pic-server/internal/utils"

	"github.com/gin-gonic/gin"
)

// GetCaptcha 获取验证码
func GetCaptcha(c *gin.Context) {
	id, b64s, _, err := utils.MakeCaptcha()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "验证码生成失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"captcha_id":    id,
		"captcha_image": b64s,
	})
}
