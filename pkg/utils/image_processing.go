package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png" // Register PNG decoder
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

// ImageConfig holds configuration for image processing
type ImageConfig struct {
	MaxWidth  int
	MaxHeight int
	Quality   int // 1-100 for JPEG
}

// AvatarConfig for user avatars
var AvatarConfig = ImageConfig{
	MaxWidth:  200,
	MaxHeight: 200,
	Quality:   85,
}

// ChapterImageConfig for manga chapter images
var ChapterImageConfig = ImageConfig{
	MaxWidth:  1200,
	MaxHeight: 2000,
	Quality:   80,
}

// ProcessedImage contains the result of image processing
type ProcessedImage struct {
	Data         []byte
	Filename     string
	OriginalSize int64
	NewSize      int64
	Width        int
	Height       int
	Format       string
}

// ProcessImage resizes and optimizes an image
// Note: Cloudinary will auto-convert to WebP when serving (f_auto)
func ProcessImage(file multipart.File, header *multipart.FileHeader, config ImageConfig) (*ProcessedImage, error) {
	// Read original file
	originalData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("không thể đọc file: %w", err)
	}

	// Decode image
	img, format, err := image.Decode(bytes.NewReader(originalData))
	if err != nil {
		return nil, fmt.Errorf("không thể decode ảnh: %w", err)
	}

	// Get original dimensions
	bounds := img.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()

	// Calculate new dimensions while maintaining aspect ratio
	newWidth, newHeight := calculateDimensions(origWidth, origHeight, config.MaxWidth, config.MaxHeight)

	// Resize image
	resizedImg := imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)

	// Encode to JPEG (Cloudinary will convert to WebP via f_auto)
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, resizedImg, &jpeg.Options{Quality: config.Quality})
	if err != nil {
		return nil, fmt.Errorf("không thể encode ảnh: %w", err)
	}

	// Generate new filename
	ext := filepath.Ext(header.Filename)
	baseName := strings.TrimSuffix(header.Filename, ext)
	newFilename := baseName + ".jpg"

	return &ProcessedImage{
		Data:         buf.Bytes(),
		Filename:     newFilename,
		OriginalSize: int64(len(originalData)),
		NewSize:      int64(buf.Len()),
		Width:        newWidth,
		Height:       newHeight,
		Format:       format,
	}, nil
}

// ProcessAvatar processes an avatar image with avatar-specific settings
func ProcessAvatar(file multipart.File, header *multipart.FileHeader) (*ProcessedImage, error) {
	return ProcessImage(file, header, AvatarConfig)
}

// ProcessChapterImage processes a chapter image
func ProcessChapterImage(file multipart.File, header *multipart.FileHeader) (*ProcessedImage, error) {
	return ProcessImage(file, header, ChapterImageConfig)
}

// calculateDimensions calculates new dimensions while maintaining aspect ratio
func calculateDimensions(origWidth, origHeight, maxWidth, maxHeight int) (int, int) {
	// If image is smaller than max, keep original size
	if origWidth <= maxWidth && origHeight <= maxHeight {
		return origWidth, origHeight
	}

	// Calculate scale factor
	widthRatio := float64(maxWidth) / float64(origWidth)
	heightRatio := float64(maxHeight) / float64(origHeight)

	// Use the smaller ratio to ensure image fits within bounds
	ratio := widthRatio
	if heightRatio < widthRatio {
		ratio = heightRatio
	}

	newWidth := int(float64(origWidth) * ratio)
	newHeight := int(float64(origHeight) * ratio)

	return newWidth, newHeight
}
