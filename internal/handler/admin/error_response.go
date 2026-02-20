package admin

import (
	basehandler "perfect-pic-server/internal/handler"

	"github.com/gin-gonic/gin"
)

func writeServiceError(c *gin.Context, err error, fallbackMessage string) {
	basehandler.WriteServiceError(c, err, fallbackMessage)
}
