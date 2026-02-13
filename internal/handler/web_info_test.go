package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/service"

	"github.com/gin-gonic/gin"
)

func TestGetWebInfo_ReturnsAllowedKeys(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	_ = db.DB.Save(&model.Setting{Key: consts.ConfigSiteName, Value: "MySite"}).Error
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigSiteDescription, Value: "Desc"}).Error
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigSiteLogo, Value: "logo.png"}).Error
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigSiteFavicon, Value: "favicon.ico"}).Error
	service.ClearCache()

	r := gin.New()
	r.GET("/webinfo", GetWebInfo)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/webinfo", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var items []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &items); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	if len(items) != 4 {
		t.Fatalf("expected 4 items, got %d", len(items))
	}
}

func TestWebInfoPrefixHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	r := gin.New()
	r.GET("/image_prefix", GetImagePrefix)
	r.GET("/avatar_prefix", GetAvatarPrefix)
	r.GET("/default_storage_quota", GetDefaultStorageQuota)

	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/image_prefix", nil))
	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w1.Code)
	}

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/avatar_prefix", nil))
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w2.Code)
	}

	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, httptest.NewRequest(http.MethodGet, "/default_storage_quota", nil))
	if w3.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w3.Code)
	}
}
