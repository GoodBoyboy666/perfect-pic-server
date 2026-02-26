package httpx

import (
	"net/http"
	"perfect-pic-server/internal/platform/service"

	"github.com/gin-gonic/gin"
)

// WriteServiceError writes a standardized HTTP error response for service-layer errors.
func WriteServiceError(c *gin.Context, err error, fallbackMessage string) {
	if serviceErr, ok := service.AsServiceError(err); ok {
		c.JSON(serviceErrorStatus(serviceErr.Code), gin.H{"error": serviceErr.Message})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": fallbackMessage})
}

func serviceErrorStatus(code service.ErrorCode) int {
	switch code {
	case service.ErrorCodeValidation:
		return http.StatusBadRequest
	case service.ErrorCodeUnauthorized:
		return http.StatusUnauthorized
	case service.ErrorCodeForbidden:
		return http.StatusForbidden
	case service.ErrorCodeConflict:
		return http.StatusConflict
	case service.ErrorCodeNotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
