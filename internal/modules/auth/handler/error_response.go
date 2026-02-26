package handler

import (
	"perfect-pic-server/internal/modules/common/httpx"

	"github.com/gin-gonic/gin"
)

func WriteServiceError(c *gin.Context, err error, fallbackMessage string) {
	httpx.WriteServiceError(c, err, fallbackMessage)
}

//func writeServiceError(c *gin.Context, err error, fallbackMessage string) {
//	httpx.WriteServiceError(c, err, fallbackMessage)
//}
