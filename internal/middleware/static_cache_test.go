package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/service"

	"github.com/gin-gonic/gin"
)

func TestStaticCacheMiddleware_SetsCacheControl(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	if err := db.DB.Save(&model.Setting{Key: consts.ConfigStaticCacheControl, Value: "public, max-age=60"}).Error; err != nil {
		t.Fatalf("set setting: %v", err)
	}
	service.ClearCache()

	r := gin.New()
	r.Use(StaticCacheMiddleware())
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))

	if got := w.Header().Get("Cache-Control"); got != "public, max-age=60" {
		t.Fatalf("Cache-Control = %q", got)
	}
}
