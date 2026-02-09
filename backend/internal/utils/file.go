package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dsi_interna_sys/internal/config"
)

// UploadFile handles file upload with validation
func UploadFile(file multipart.File, header *multipart.FileHeader, subdir string) (string, error) {
	cfg := config.Loaded

	// Validate file size
	if header.Size > cfg.Upload.MaxSize {
		return "", fmt.Errorf("file size exceeds maximum allowed size of %d bytes", cfg.Upload.MaxSize)
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !isAllowedExtension(ext, cfg.Upload.AllowedExts) {
		return "", fmt.Errorf("file type %s is not allowed. Allowed types: %v", ext, cfg.Upload.AllowedExts)
	}

	// Create upload directory if not exists
	uploadPath := filepath.Join(cfg.Upload.Dir, subdir)
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate unique filename
	filename := generateUniqueFilename(header.Filename)
	filePath := filepath.Join(uploadPath, filename)

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	// Return relative path from upload directory
	relativePath := filepath.Join(subdir, filename)
	return relativePath, nil
}

// DeleteFile deletes a file from upload directory
func DeleteFile(relativePath string) error {
	cfg := config.Loaded
	fullPath := filepath.Join(cfg.Upload.Dir, relativePath)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil // File doesn't exist, consider it deleted
	}

	return os.Remove(fullPath)
}

// GetFileExtension returns file extension from filename
func GetFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	return strings.ToLower(strings.TrimPrefix(ext, "."))
}

// isAllowedExtension checks if file extension is allowed
func isAllowedExtension(ext string, allowedExts []string) bool {
	for _, allowed := range allowedExts {
		if ext == allowed || ext == strings.TrimPrefix(allowed, ".") {
			return true
		}
	}
	return false
}

// generateUniqueFilename generates a unique filename with timestamp
func generateUniqueFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	nameWithoutExt := strings.TrimSuffix(originalFilename, ext)

	// Clean filename (remove special characters)
	nameWithoutExt = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, nameWithoutExt)

	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s_%d%s", nameWithoutExt, timestamp, ext)
}

// ValidateFileType validates file type based on extension
func ValidateFileType(filename string, allowedTypes []string) error {
	ext := GetFileExtension(filename)

	for _, allowed := range allowedTypes {
		if ext == allowed {
			return nil
		}
	}

	return fmt.Errorf("file type .%s is not allowed. Allowed types: %v", ext, allowedTypes)
}
