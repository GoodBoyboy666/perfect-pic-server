package service

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	moduledto "perfect-pic-server/internal/dto"
	"testing"

	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/testutils"
)

// 测试内容：验证图片文件校验在合法图片时返回通过。
func TestValidateImageFile_OK(t *testing.T) {
	setupTestDB(t)

	fh := mustFileHeader(t, "a.png", testutils.MinimalPNG())
	ok, ext, err := testService.ValidateImageFile(fh)
	if !ok || err != nil {
		t.Fatalf("期望 ok，实际为 ok=%v ext=%q err=%v", ok, ext, err)
	}
	if ext != ".png" {
		t.Fatalf("期望 .png ext，实际为 %q", ext)
	}
}

// 测试内容：验证不支持的文件扩展名会被拒绝。
func TestValidateImageFile_RejectsUnsupportedExt(t *testing.T) {
	setupTestDB(t)

	fh := mustFileHeader(t, "a.exe", testutils.MinimalPNG())
	ok, ext, err := testService.ValidateImageFile(fh)
	if ok || err == nil {
		t.Fatalf("期望 failure，实际为 ok=%v ext=%q err=%v", ok, ext, err)
	}
	if ext != ".exe" {
		t.Fatalf("期望 ext to be .exe，实际为 %q", ext)
	}
}

// 测试内容：验证删除图片会移除文件并同步更新存储占用。
func TestDeleteImage_RemovesFileAndUpdatesStorage(t *testing.T) {
	setupTestDB(t)

	tmp := t.TempDir()
	oldwd, _ := os.Getwd()
	_ = os.Chdir(tmp)
	defer func() { _ = os.Chdir(oldwd) }()

	u := model.User{Username: "alice", Password: "x", Status: 1, Email: "a@example.com", StorageUsed: 4}
	_ = db.DB.Create(&u).Error

	imgRel := "2026/02/13/a.png"
	full := filepath.Join("uploads", "imgs", filepath.FromSlash(imgRel))
	_ = os.MkdirAll(filepath.Dir(full), 0755)
	_ = os.WriteFile(full, []byte{0x89, 0x50, 0x4E, 0x47}, 0644)

	img := model.Image{Filename: "a.png", Path: imgRel, Size: 4, Width: 1, Height: 1, MimeType: ".png", UploadedAt: 1, UserID: u.ID}
	_ = db.DB.Create(&img).Error

	if err := testService.DeleteImage(&img); err != nil {
		t.Fatalf("DeleteImage: %v", err)
	}
	if _, err := os.Stat(full); !os.IsNotExist(err) {
		t.Fatalf("期望 file deleted, err=%v", err)
	}

	var count int64
	_ = db.DB.Model(&model.Image{}).Count(&count).Error
	if count != 0 {
		t.Fatalf("期望 image record deleted")
	}

	var got model.User
	_ = db.DB.First(&got, u.ID).Error
	if got.StorageUsed != 0 {
		t.Fatalf("期望 storage_used 0，实际为 %d", got.StorageUsed)
	}
}

// 测试内容：验证批量删除图片会移除文件并更新存储占用。
func TestBatchDeleteImages_RemovesFilesAndUpdatesStorage(t *testing.T) {
	setupTestDB(t)

	tmp := t.TempDir()
	oldwd, _ := os.Getwd()
	_ = os.Chdir(tmp)
	defer func() { _ = os.Chdir(oldwd) }()

	u := model.User{Username: "alice", Password: "x", Status: 1, Email: "a@example.com", StorageUsed: 8}
	_ = db.DB.Create(&u).Error

	img1 := model.Image{Filename: "a.png", Path: "2026/02/13/a.png", Size: 4, Width: 1, Height: 1, MimeType: ".png", UploadedAt: 1, UserID: u.ID}
	img2 := model.Image{Filename: "b.png", Path: "2026/02/13/b.png", Size: 4, Width: 1, Height: 1, MimeType: ".png", UploadedAt: 1, UserID: u.ID}
	_ = db.DB.Create(&img1).Error
	_ = db.DB.Create(&img2).Error

	full1 := filepath.Join("uploads", "imgs", filepath.FromSlash(img1.Path))
	full2 := filepath.Join("uploads", "imgs", filepath.FromSlash(img2.Path))
	_ = os.MkdirAll(filepath.Dir(full1), 0755)
	_ = os.WriteFile(full1, []byte{0x89, 0x50, 0x4E, 0x47}, 0644)
	_ = os.WriteFile(full2, []byte{0x89, 0x50, 0x4E, 0x47}, 0644)

	if err := testService.BatchDeleteImages([]model.Image{img1, img2}); err != nil {
		t.Fatalf("BatchDeleteImages: %v", err)
	}
	if _, err := os.Stat(full1); !os.IsNotExist(err) {
		t.Fatalf("期望 file1 deleted, err=%v", err)
	}
	if _, err := os.Stat(full2); !os.IsNotExist(err) {
		t.Fatalf("期望 file2 deleted, err=%v", err)
	}

	var got model.User
	_ = db.DB.First(&got, u.ID).Error
	if got.StorageUsed != 0 {
		t.Fatalf("期望 storage_used 0，实际为 %d", got.StorageUsed)
	}
}

// 测试内容：验证分页参数在缺省与显式值下的归一化结果。
func TestNormalizePagination(t *testing.T) {
	p, ps := normalizePagination(0, 0)
	if p != 1 || ps != 10 {
		t.Fatalf("期望 defaults 1/10，实际为 %d/%d", p, ps)
	}
	p, ps = normalizePagination(2, 5)
	if p != 2 || ps != 5 {
		t.Fatalf("期望 2/5，实际为 %d/%d", p, ps)
	}
}

// 测试内容：验证用户图片列表支持按文件名过滤并返回正确分页信息。
func TestListUserImages_FiltersAndPaging(t *testing.T) {
	setupTestDB(t)

	u := model.User{Username: "u1", Password: "x", Status: 1, Email: "u1@example.com"}
	if err := db.DB.Create(&u).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	img1 := model.Image{Filename: "cat.png", Path: "2026/02/13/cat.png", Size: 1, Width: 1, Height: 1, MimeType: ".png", UploadedAt: 1, UserID: u.ID}
	img2 := model.Image{Filename: "dog.png", Path: "2026/02/13/dog.png", Size: 1, Width: 1, Height: 1, MimeType: ".png", UploadedAt: 2, UserID: u.ID}
	_ = db.DB.Create(&img1).Error
	_ = db.DB.Create(&img2).Error

	list, total, page, pageSize, err := testService.ListUserImages(moduledto.UserImageListRequest{
		PaginationRequest: moduledto.PaginationRequest{Page: 1, PageSize: 10},
		UserID:            u.ID,
		Filename:          "cat",
	})
	if err != nil {
		t.Fatalf("ListUserImages: %v", err)
	}
	if total != 1 || page != 1 || pageSize != 10 || len(list) != 1 {
		t.Fatalf("非预期 result: total=%d page=%d pageSize=%d len=%d", total, page, pageSize, len(list))
	}
	if list[0].Filename != "cat.png" {
		t.Fatalf("期望 cat.png，实际为 %q", list[0].Filename)
	}
}

// 测试内容：验证只能获取当前用户拥有的图片记录。
func TestGetUserOwnedImage(t *testing.T) {
	setupTestDB(t)

	u := model.User{Username: "u1", Password: "x", Status: 1, Email: "u1@example.com"}
	_ = db.DB.Create(&u).Error
	img := model.Image{Filename: "a.png", Path: "2026/02/13/a.png", Size: 1, Width: 1, Height: 1, MimeType: ".png", UploadedAt: 1, UserID: u.ID}
	_ = db.DB.Create(&img).Error

	got, err := testService.GetUserOwnedImage(img.ID, u.ID)
	if err != nil {
		t.Fatalf("GetUserOwnedImage: %v", err)
	}
	if got.ID != img.ID {
		t.Fatalf("期望 image id %d，实际为 %d", img.ID, got.ID)
	}

	_, err = testService.GetUserOwnedImage(img.ID, u.ID+1)
	if err == nil {
		t.Fatalf("期望返回错误 for non-owned image")
	}
}

// 测试内容：验证图片数量统计及按 ID 获取（用户/管理员）接口的正确性。
func TestGetUserImageCountAndBatchGetters(t *testing.T) {
	setupTestDB(t)

	u := model.User{Username: "u1", Password: "x", Status: 1, Email: "u1@example.com"}
	_ = db.DB.Create(&u).Error

	img1 := model.Image{Filename: "a.png", Path: "2026/02/13/a.png", Size: 1, Width: 1, Height: 1, MimeType: ".png", UploadedAt: 1, UserID: u.ID}
	img2 := model.Image{Filename: "b.png", Path: "2026/02/13/b.png", Size: 1, Width: 1, Height: 1, MimeType: ".png", UploadedAt: 1, UserID: u.ID}
	_ = db.DB.Create(&img1).Error
	_ = db.DB.Create(&img2).Error

	cnt, err := testService.GetUserImageCount(u.ID)
	if err != nil {
		t.Fatalf("GetUserImageCount: %v", err)
	}
	if cnt != 2 {
		t.Fatalf("期望 2，实际为 %d", cnt)
	}

	images, err := testService.GetImagesByIDsForUser([]uint{img1.ID, img2.ID}, u.ID)
	if err != nil || len(images) != 2 {
		t.Fatalf("GetImagesByIDsForUser: err=%v len=%d", err, len(images))
	}

	got, err := testService.AdminGetImageByID(img1.ID)
	if err != nil {
		t.Fatalf("AdminGetImageByID: %v", err)
	}
	if got.ID != img1.ID {
		t.Fatalf("非预期 image id")
	}

	images2, err := testService.AdminGetImagesByIDs([]uint{img1.ID, img2.ID})
	if err != nil || len(images2) != 2 {
		t.Fatalf("AdminGetImagesByIDs: err=%v len=%d", err, len(images2))
	}
}

// 测试内容：验证管理员图片列表支持按用户名、文件名与用户 ID 过滤并预加载用户信息。
func TestListImagesForAdmin_Filters(t *testing.T) {
	setupTestDB(t)

	u1 := model.User{Username: "alice", Password: "x", Status: 1, Email: "a@example.com"}
	u2 := model.User{Username: "bob", Password: "x", Status: 1, Email: "b@example.com"}
	_ = db.DB.Create(&u1).Error
	_ = db.DB.Create(&u2).Error

	_ = db.DB.Create(&model.Image{Filename: "a.png", Path: "2026/02/13/a.png", Size: 1, Width: 1, Height: 1, MimeType: ".png", UploadedAt: 1, UserID: u1.ID}).Error
	_ = db.DB.Create(&model.Image{Filename: "b.png", Path: "2026/02/13/b.png", Size: 1, Width: 1, Height: 1, MimeType: ".png", UploadedAt: 1, UserID: u2.ID}).Error

	images, total, _, _, err := testService.AdminListImages(moduledto.AdminImageListRequest{
		PaginationRequest: moduledto.PaginationRequest{Page: 1, PageSize: 10},
		Username:          "ali",
	})
	if err != nil {
		t.Fatalf("AdminListImages: %v", err)
	}
	if total != 1 || len(images) != 1 {
		t.Fatalf("期望 1 image，实际为 total=%d len=%d", total, len(images))
	}
	if images[0].User.Username != "alice" {
		t.Fatalf("期望 preload user alice，实际为 %q", images[0].User.Username)
	}

	images2, total2, _, _, err := testService.AdminListImages(moduledto.AdminImageListRequest{
		PaginationRequest: moduledto.PaginationRequest{Page: 1, PageSize: 10},
		Filename:          "a.",
		UserID:            &u1.ID,
	})
	if err != nil {
		t.Fatalf("testService.AdminListImages(by filename/user_id): %v", err)
	}
	if total2 != 1 || len(images2) != 1 {
		t.Fatalf("期望 1 image，实际为 total=%d len=%d", total2, len(images2))
	}
	if images2[0].UserID != u1.ID || images2[0].Filename != "a.png" {
		t.Fatalf("非预期过滤结果: user_id=%d filename=%s", images2[0].UserID, images2[0].Filename)
	}
}

func mustFileHeader(t *testing.T, filename string, content []byte) *multipart.FileHeader {
	t.Helper()

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("写入分段失败: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("关闭 writer 失败: %v", err)
	}

	req := httptest.NewRequest("POST", "http://example/upload", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	if err := req.ParseMultipartForm(int64(len(content)) + 1024); err != nil {
		t.Fatalf("ParseMultipartForm: %v", err)
	}

	fhs := req.MultipartForm.File["file"]
	if len(fhs) != 1 {
		t.Fatalf("期望 1 file header，实际为 %d", len(fhs))
	}
	return fhs[0]
}
