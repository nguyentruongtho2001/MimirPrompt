package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	imagesDir string
}

func NewUploadHandler(imagesDir string) *UploadHandler {
	return &UploadHandler{imagesDir: imagesDir}
}

// UploadImage handles image file uploads
func (h *UploadHandler) UploadImage(c *gin.Context) {
	// Get the file from form
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
		return
	}
	defer file.Close()

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File must be an image"})
		return
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		// Default to jpg if no extension
		switch contentType {
		case "image/png":
			ext = ".png"
		case "image/gif":
			ext = ".gif"
		case "image/webp":
			ext = ".webp"
		default:
			ext = ".jpg"
		}
	}

	filename := fmt.Sprintf("upload_%d%s", time.Now().UnixNano(), ext)
	filePath := filepath.Join(h.imagesDir, filename)

	// Create the images directory if it doesn't exist
	if err := os.MkdirAll(h.imagesDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Create the destination file
	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	defer dst.Close()

	// Copy the uploaded file to the destination
	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Return the URL path for the uploaded image
	imageURL := "/images/" + filename

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"url":      imageURL,
		"filename": filename,
	})
}
