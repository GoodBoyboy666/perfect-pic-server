package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"

	"github.com/gin-gonic/gin"
)

func TestImageManageHandlers_ListAndDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	tmp := t.TempDir()
	oldwd, _ := os.Getwd()
	_ = os.Chdir(tmp)
	defer func() { _ = os.Chdir(oldwd) }()

	u := model.User{Username: "alice", Password: "x", Status: 1, Email: "a@example.com"}
	_ = db.DB.Create(&u).Error

	// Create physical image file + record.
	imgRel := "2026/02/13/a.png"
	imgFile := filepath.Join("uploads", "imgs", filepath.FromSlash(imgRel))
	_ = os.MkdirAll(filepath.Dir(imgFile), 0755)
	_ = os.WriteFile(imgFile, []byte{0x89, 0x50, 0x4E, 0x47}, 0644)

	img := model.Image{Filename: "a.png", Path: imgRel, Size: 4, Width: 1, Height: 1, MimeType: ".png", UploadedAt: 1, UserID: u.ID}
	_ = db.DB.Create(&img).Error

	r := gin.New()
	r.GET("/images", GetImageList)
	r.DELETE("/images/:id", DeleteImage)
	r.DELETE("/images/batch", BatchDeleteImages)

	// List
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/images?page=1&page_size=10&username=ali", nil))
	if w1.Code != http.StatusOK {
		t.Fatalf("list expected 200, got %d body=%s", w1.Code, w1.Body.String())
	}

	// Delete single
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httptest.NewRequest(http.MethodDelete, "/images/1", nil))
	if w2.Code != http.StatusOK {
		t.Fatalf("delete expected 200, got %d body=%s", w2.Code, w2.Body.String())
	}
	if _, err := os.Stat(imgFile); !os.IsNotExist(err) {
		t.Fatalf("expected file removed, err=%v", err)
	}

	// Recreate two images for batch delete
	imgFile1 := filepath.Join("uploads", "imgs", "2026", "02", "13", "b.png")
	imgFile2 := filepath.Join("uploads", "imgs", "2026", "02", "13", "c.png")
	_ = os.WriteFile(imgFile1, []byte{0x89, 0x50, 0x4E, 0x47}, 0644)
	_ = os.WriteFile(imgFile2, []byte{0x89, 0x50, 0x4E, 0x47}, 0644)
	imgB := model.Image{Filename: "b.png", Path: "2026/02/13/b.png", Size: 4, Width: 1, Height: 1, MimeType: ".png", UploadedAt: 1, UserID: u.ID}
	imgC := model.Image{Filename: "c.png", Path: "2026/02/13/c.png", Size: 4, Width: 1, Height: 1, MimeType: ".png", UploadedAt: 1, UserID: u.ID}
	_ = db.DB.Create(&imgB).Error
	_ = db.DB.Create(&imgC).Error

	body, _ := json.Marshal(gin.H{"ids": []uint{imgB.ID, imgC.ID}})
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, httptest.NewRequest(http.MethodDelete, "/images/batch", bytes.NewReader(body)))
	if w3.Code != http.StatusOK {
		t.Fatalf("batch delete expected 200, got %d body=%s", w3.Code, w3.Body.String())
	}
	if _, err := os.Stat(imgFile1); !os.IsNotExist(err) {
		t.Fatalf("expected file1 removed, err=%v", err)
	}
	if _, err := os.Stat(imgFile2); !os.IsNotExist(err) {
		t.Fatalf("expected file2 removed, err=%v", err)
	}
}

func TestBatchDeleteImagesHandler_Errors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	r := gin.New()
	r.DELETE("/images/batch", BatchDeleteImages)

	// bind error
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, httptest.NewRequest(http.MethodDelete, "/images/batch", bytes.NewReader([]byte("{bad"))))
	if w1.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w1.Code, w1.Body.String())
	}

	// empty list
	body, _ := json.Marshal(gin.H{"ids": []uint{}})
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httptest.NewRequest(http.MethodDelete, "/images/batch", bytes.NewReader(body)))
	if w2.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w2.Code, w2.Body.String())
	}

	// too many
	ids := make([]uint, 51)
	for i := range ids {
		ids[i] = uint(i + 1)
	}
	body2, _ := json.Marshal(gin.H{"ids": ids})
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, httptest.NewRequest(http.MethodDelete, "/images/batch", bytes.NewReader(body2)))
	if w3.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w3.Code, w3.Body.String())
	}

	// not found
	body3, _ := json.Marshal(gin.H{"ids": []uint{1}})
	w4 := httptest.NewRecorder()
	r.ServeHTTP(w4, httptest.NewRequest(http.MethodDelete, "/images/batch", bytes.NewReader(body3)))
	if w4.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", w4.Code, w4.Body.String())
	}
}
