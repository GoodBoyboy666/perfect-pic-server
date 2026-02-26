package handler

import (
	"net/http"
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/utils"

	"github.com/gin-gonic/gin"
)

// GetCaptcha 获取验证码
func (h *Handler) GetCaptcha(c *gin.Context) {
	providerInfo := h.authService.GetCaptchaProviderInfo()
	if providerInfo.Provider == consts.CaptchaProviderDisabled {
		c.JSON(http.StatusOK, gin.H{
			"provider": providerInfo.Provider,
		})
		return
	}

	if providerInfo.Provider == consts.CaptchaProviderTurnstile {
		c.JSON(http.StatusOK, gin.H{
			"provider":      providerInfo.Provider,
			"public_config": providerInfo.PublicConfig,
		})
		return
	}

	if providerInfo.Provider == consts.CaptchaProviderRecaptcha {
		c.JSON(http.StatusOK, gin.H{
			"provider":      providerInfo.Provider,
			"public_config": providerInfo.PublicConfig,
		})
		return
	}

	if providerInfo.Provider == consts.CaptchaProviderHcaptcha {
		c.JSON(http.StatusOK, gin.H{
			"provider":      providerInfo.Provider,
			"public_config": providerInfo.PublicConfig,
		})
		return
	}

	if providerInfo.Provider == consts.CaptchaProviderGeetest {
		c.JSON(http.StatusOK, gin.H{
			"provider":      providerInfo.Provider,
			"public_config": providerInfo.PublicConfig,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"provider": providerInfo.Provider,
	})
}

// GetCaptchaImage 获取图形验证码图片
func (h *Handler) GetCaptchaImage(c *gin.Context) {
	providerInfo := h.authService.GetCaptchaProviderInfo()
	if providerInfo.Provider != consts.CaptchaProviderImage {
		c.JSON(http.StatusBadRequest, gin.H{"error": "当前验证码模式非图形验证码"})
		return
	}

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
