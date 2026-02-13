package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestIntervalRateMiddleware_BlocksSecondRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	r := gin.New()
	r.POST("/x", IntervalRateMiddleware(10*time.Second), func(c *gin.Context) { c.Status(http.StatusOK) })

	req1 := httptest.NewRequest(http.MethodPost, "/x", bytes.NewReader([]byte("a")))
	req1.RemoteAddr = "1.2.3.4:1111"
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w1.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/x", bytes.NewReader([]byte("a")))
	req2.RemoteAddr = "1.2.3.4:1111"
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w2.Code)
	}
}
