package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"perfect-pic-server/internal/service"

	"github.com/gin-gonic/gin"
)

func TestInitHandlers_StateAndInit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	// Ensure default settings exist so InitializeSystem updates rows.
	service.InitializeSettings()

	r := gin.New()
	r.GET("/init", GetInitState)
	r.POST("/init", Init)

	// Initially not initialized.
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/init", nil))
	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w1.Code)
	}

	body, _ := json.Marshal(gin.H{
		"username":         "admin_1",
		"password":         "abc12345",
		"site_name":        "MySite",
		"site_description": "Desc",
	})
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httptest.NewRequest(http.MethodPost, "/init", bytes.NewReader(body)))
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w2.Code, w2.Body.String())
	}

	// Second init should be forbidden.
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, httptest.NewRequest(http.MethodPost, "/init", bytes.NewReader(body)))
	if w3.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", w3.Code, w3.Body.String())
	}
}
