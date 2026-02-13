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

func TestRateLimitMiddleware_DisabledAllowsRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	if err := db.DB.Save(&model.Setting{Key: consts.ConfigRateLimitEnabled, Value: "false"}).Error; err != nil {
		t.Fatalf("set setting: %v", err)
	}
	service.ClearCache()

	r := gin.New()
	r.Use(RateLimitMiddleware(consts.ConfigRateLimitAuthRPS, consts.ConfigRateLimitAuthBurst))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req1 := httptest.NewRequest(http.MethodGet, "/x", nil)
	req1.RemoteAddr = "1.2.3.4:1111"
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/x", nil)
	req2.RemoteAddr = "1.2.3.4:1111"
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w2.Code)
	}
}

func TestRateLimitMiddleware_EnabledBlocksBurst(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	// Enable limiter with 1 token burst and no refill (rps=0).
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigRateLimitEnabled, Value: "true"}).Error
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigRateLimitAuthRPS, Value: "0"}).Error
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigRateLimitAuthBurst, Value: "1"}).Error
	service.ClearCache()

	r := gin.New()
	r.Use(RateLimitMiddleware(consts.ConfigRateLimitAuthRPS, consts.ConfigRateLimitAuthBurst))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req1 := httptest.NewRequest(http.MethodGet, "/x", nil)
	req1.RemoteAddr = "1.2.3.4:1111"
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/x", nil)
	req2.RemoteAddr = "1.2.3.4:1111"
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w2.Code)
	}
}
