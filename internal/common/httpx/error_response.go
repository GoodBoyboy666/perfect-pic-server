package httpx

import (
	"net/http"
	"perfect-pic-server/internal/common"

	"github.com/gin-gonic/gin"
)

// WriteServiceError writes a standardized HTTP error response for service-layer errors.
func WriteServiceError(c *gin.Context, err error, fallbackMessage string) {
	if serviceErr, ok := common.AsServiceError(err); ok {
		c.JSON(serviceErrorStatus(serviceErr.Code), gin.H{"error": serviceErr.Message})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": fallbackMessage})
}

func serviceErrorStatus(code common.ErrorCode) int {
	switch code {
	case common.ErrorCodeValidation:
		return http.StatusBadRequest
	case common.ErrorCodeUnauthorized:
		return http.StatusUnauthorized
	case common.ErrorCodeForbidden:
		return http.StatusForbidden
	case common.ErrorCodeConflict:
		return http.StatusConflict
	case common.ErrorCodeNotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
