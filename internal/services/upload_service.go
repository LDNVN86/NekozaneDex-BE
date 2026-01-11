package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"nekozanedex/internal/config"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type UploadService interface {
	UploadImage(file multipart.File, filename string, folder string) (string, error)
	UploadImageBytes(data []byte, filename string, folder string) (string, error)
	UploadMultipleImages(files []*multipart.FileHeader, folder string) ([]string, error)
	DeleteImage(publicID string) error
}

type uploadService struct {
	cld *cloudinary.Cloudinary
	cfg *config.Config
}

func NewUploadService(cfg *config.Config) (UploadService, error) {
	// Validate config
	if cfg.Cloudinary.CloudName == "" || cfg.Cloudinary.APIKey == "" || cfg.Cloudinary.APISecret == "" {
		return nil, errors.New("cloudinary config is not set")
	}

	// Initialize Cloudinary
	cld, err := cloudinary.NewFromParams(
		cfg.Cloudinary.CloudName,
		cfg.Cloudinary.APIKey,
		cfg.Cloudinary.APISecret,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cloudinary: %w", err)
	}

	return &uploadService{
		cld: cld,
		cfg: cfg,
	}, nil
}

// UploadImage - Upload single image to Cloudinary
func (s *uploadService) UploadImage(file multipart.File, filename string, folder string) (string, error) {
	ctx := context.Background()

	// Get file extension
	ext := strings.ToLower(filepath.Ext(filename))
	if !isValidImageExtension(ext) {
		return "", errors.New("định dạng file không hợp lệ (chỉ chấp nhận jpg, jpeg, png, webp, gif)")
	}

	// Upload to Cloudinary
	uploadResult, err := s.cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder:         folder,
		ResourceType:   "image",
		AllowedFormats: []string{"jpg", "jpeg", "png", "webp", "gif"},
		Transformation: "q_auto,f_auto", // Auto quality and format
	})
	if err != nil {
		return "", fmt.Errorf("upload failed: %w", err)
	}

	// Debug: Log result
	fmt.Printf("[Cloudinary] Upload result - SecureURL: %s, Error: %v\n", uploadResult.SecureURL, uploadResult.Error)
	
	if uploadResult.Error.Message != "" {
		return "", fmt.Errorf("cloudinary error: %s", uploadResult.Error.Message)
	}

	return uploadResult.SecureURL, nil
}

// UploadImageBytes - Upload image bytes to Cloudinary (for processed images)
func (s *uploadService) UploadImageBytes(data []byte, filename string, folder string) (string, error) {
	ctx := context.Background()

	// Get file extension
	ext := strings.ToLower(filepath.Ext(filename))
	if !isValidImageExtension(ext) {
		return "", errors.New("định dạng file không hợp lệ (chỉ chấp nhận jpg, jpeg, png, webp, gif)")
	}

	// Create reader from bytes
	reader := bytes.NewReader(data)

	// Upload to Cloudinary with WebP conversion
	uploadResult, err := s.cld.Upload.Upload(ctx, reader, uploader.UploadParams{
		Folder:         folder,
		ResourceType:   "image",
		Transformation: "f_webp,q_auto", // Force WebP format with auto quality
	})
	if err != nil {
		return "", fmt.Errorf("upload failed: %w", err)
	}

	// Debug: Log result
	fmt.Printf("[Cloudinary] Upload bytes result - SecureURL: %s, Error: %v\n", uploadResult.SecureURL, uploadResult.Error)

	if uploadResult.Error.Message != "" {
		return "", fmt.Errorf("cloudinary error: %s", uploadResult.Error.Message)
	}

	return uploadResult.SecureURL, nil
}

// UploadMultipleImages - Upload multiple images (for manga chapters)
func (s *uploadService) UploadMultipleImages(files []*multipart.FileHeader, folder string) ([]string, error) {
	urls := make([]string, 0, len(files))

	for i, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file %d: %w", i, err)
		}
		defer file.Close()

		url, err := s.UploadImage(file, fileHeader.Filename, folder)
		if err != nil {
			return nil, fmt.Errorf("failed to upload file %d: %w", i, err)
		}

		urls = append(urls, url)
	}

	return urls, nil
}

// DeleteImage - Delete image from Cloudinary
func (s *uploadService) DeleteImage(publicID string) error {
	ctx := context.Background()

	_, err := s.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	return nil
}

// Helper: Check valid image extensions
func isValidImageExtension(ext string) bool {
	validExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
		".gif":  true,
	}
	return validExts[ext]
}
