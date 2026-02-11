package handler

import (
	"net/http"
	"perfect-pic-server/internal/service"
	"perfect-pic-server/internal/utils"

	"github.com/gin-gonic/gin"
)

// GetCaptcha 获取验证码
func GetCaptcha(c *gin.Context) {
	providerInfo := service.GetCaptchaProviderInfo()
	if providerInfo.Provider == service.CaptchaProviderDisabled {
		c.JSON(http.StatusOK, gin.H{
			"provider": providerInfo.Provider,
		})
		return
	}

	if providerInfo.Provider == service.CaptchaProviderTurnstile {
		c.JSON(http.StatusOK, gin.H{
			"provider":      providerInfo.Provider,
			"public_config": providerInfo.PublicConfig,
		})
		return
	}

	if providerInfo.Provider == service.CaptchaProviderRecaptcha {
		c.JSON(http.StatusOK, gin.H{
			"provider":      providerInfo.Provider,
			"public_config": providerInfo.PublicConfig,
		})
		return
	}

	if providerInfo.Provider == service.CaptchaProviderHcaptcha {
		c.JSON(http.StatusOK, gin.H{
			"provider":      providerInfo.Provider,
			"public_config": providerInfo.PublicConfig,
		})
		return
	}

	if providerInfo.Provider == service.CaptchaProviderGeetest {
		c.JSON(http.StatusOK, gin.H{
			"provider":      providerInfo.Provider,
			"public_config": providerInfo.PublicConfig,
		})
		return
	}

	id, b64s, _, err := utils.MakeCaptcha()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "验证码生成失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"provider":      providerInfo.Provider,
		"captcha_id":    id,
		"captcha_image": b64s,
	})
}
