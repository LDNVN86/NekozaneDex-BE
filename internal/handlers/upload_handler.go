package handlers

import (
	"fmt"
	"path/filepath"
	"strings"

	"nekozanedex/internal/services"
	"nekozanedex/pkg/response"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	uploadService services.UploadService
}

func NewUploadHandler(uploadService services.UploadService) *UploadHandler {
	return &UploadHandler{uploadService: uploadService}
}

// UploadSingleImage godoc
// @Summary Upload single image
// @Tags Upload
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param image formance file true "Image file"
// @Param folder formance string false "Folder name"
// @Success 200 {object} response.Response
// @Router /api/admin/upload [post]
func (h *UploadHandler) UploadSingleImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		response.BadRequest(c, "Không tìm thấy file ảnh")
		return
	}
	defer file.Close()

	// Validate file size (max 10MB)
	if header.Size > 10*1024*1024 {
		response.BadRequest(c, "File quá lớn (tối đa 10MB)")
		return
	}

	// Get folder from form (default: "manga")
	folder := c.DefaultPostForm("folder", "manga")

	// Upload
	url, err := h.uploadService.UploadImage(file, header.Filename, folder)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Oke(c, gin.H{
		"url":      url,
		"filename": header.Filename,
		"size":     header.Size,
	})
}

// UploadChapterImages godoc
// @Summary Upload multiple images for a chapter
// @Tags Upload
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param images formData file true "Image files (multiple)"
// @Param story_slug formData string true "Story slug"
// @Param chapter_number formData string true "Chapter number"
// @Success 200 {object} response.Response
// @Router /api/admin/upload/chapter [post]
func (h *UploadHandler) UploadChapterImages(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		response.BadRequest(c, "Không đọc được form data")
		return
	}

	files := form.File["images"]
	if len(files) == 0 {
		response.BadRequest(c, "Không tìm thấy file ảnh")
		return
	}

	// Validate total size (max 100MB for a chapter)
	var totalSize int64
	for _, f := range files {
		totalSize += f.Size
	}
	if totalSize > 100*1024*1024 {
		response.BadRequest(c, "Tổng dung lượng quá lớn (tối đa 100MB)")
		return
	}

	// Get folder structure
	storySlug := c.PostForm("story_slug")
	chapterNumber := c.PostForm("chapter_number")
	if storySlug == "" || chapterNumber == "" {
		response.BadRequest(c, "Thiếu story_slug hoặc chapter_number")
		return
	}

	folder := fmt.Sprintf("manga/%s/chapter-%s", sanitizeSlug(storySlug), chapterNumber)

	// Upload all images
	urls, err := h.uploadService.UploadMultipleImages(files, folder)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Oke(c, gin.H{
		"urls":       urls,
		"count":      len(urls),
		"total_size": totalSize,
		"folder":     folder,
	})
}

// DeleteImage godoc
// @Summary Delete image from Cloudinary
// @Tags Upload
// @Security BearerAuth
// @Produce json
// @Param public_id query string true "Cloudinary public ID"
// @Success 200 {object} response.Response
// @Router /api/admin/upload [delete]
func (h *UploadHandler) DeleteImage(c *gin.Context) {
	publicID := c.Query("public_id")
	if publicID == "" {
		response.BadRequest(c, "Thiếu public_id")
		return
	}

	if err := h.uploadService.DeleteImage(publicID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Oke(c, gin.H{"message": "Xóa ảnh thành công"})
}

// Helper: Sanitize slug for folder name
func sanitizeSlug(slug string) string {
	// Remove special characters, keep alphanumeric and hyphens
	slug = strings.ToLower(slug)
	slug = strings.ReplaceAll(slug, " ", "-")
	
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	
	return filepath.Clean(result.String())
}
