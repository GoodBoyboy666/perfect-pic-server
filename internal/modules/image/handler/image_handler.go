package handler

import (
	"log"
	"net/http"
	"perfect-pic-server/internal/modules/common/httpx"
	moduledto "perfect-pic-server/internal/modules/image/dto"
	platformservice "perfect-pic-server/internal/platform/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) UploadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择文件"})
		return
	}

	// 从 JWT 中间件获取用户ID
	userID, exists := c.Get("id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未获取到用户信息"})
		return
	}
	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的用户ID类型"})
		return
	}

	imageRecord, url, err := h.imageService.ProcessImageUpload(file, uid)
	if err != nil {
		if _, ok := platformservice.AsServiceError(err); !ok {
			log.Printf("Upload failed: %v", err)
		}
		httpx.WriteServiceError(c, err, "上传失败，请稍后重试")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg": "上传成功",
		"url": url,
		"id":  imageRecord.ID,
	})
}

func (h *Handler) GetMyImages(c *gin.Context) {
	userID, _ := c.Get("id")

	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")
	filename := c.Query("filename")
	idStr := c.Query("id")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的用户ID类型"})
		return
	}

	var imageID *uint
	if idStr != "" {
		parsed, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || parsed == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id 参数错误"})
			return
		}
		id := uint(parsed)
		imageID = &id
	}

	images, total, page, pageSize, err := h.imageService.ListUserImages(moduledto.UserImageListRequest{
		PaginationRequest: moduledto.PaginationRequest{Page: page, PageSize: pageSize},
		UserID:            uid,
		Filename:          filename,
		ID:                imageID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取图片列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":      images,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// DeleteMyImage 用户删除自己的图片
func (h *Handler) DeleteMyImage(c *gin.Context) {
	userID, _ := c.Get("id")
	idParam := c.Param("id")
	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的用户ID类型"})
		return
	}

	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 参数错误"})
		return
	}

	image, err := h.imageService.GetUserOwnedImage(uint(id), uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "图片不存在或无权删除"})
		return
	}

	if err := h.imageService.DeleteImage(image); err != nil {
		httpx.WriteServiceError(c, err, "删除失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// BatchDeleteMyImages 批量删除用户自己的图片
func (h *Handler) BatchDeleteMyImages(c *gin.Context) {
	userID, _ := c.Get("id")
	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的用户ID类型"})
		return
	}

	var req moduledto.BatchDeleteImagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择要删除的图片"})
		return
	}

	if len(req.IDs) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "一次最多只能删除 50 张图片"})
		return
	}

	images, err := h.imageService.GetImagesByIDsForUser(req.IDs, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查找图片失败"})
		return
	}

	if len(images) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到指定图片或无权删除"})
		return
	}

	if err := h.imageService.BatchDeleteImages(images); err != nil {
		httpx.WriteServiceError(c, err, "删除失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功", "deleted_count": len(images)})
}
