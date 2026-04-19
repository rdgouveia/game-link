package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	maxAvatarSize = 5 << 20 // 5 MB
)

var allowedAvatarTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

func (h *Handler) UploadAvatar(c *gin.Context) {
	claims := c.MustGet("claims").(jwt.MapClaims)
	userID := claims["sub"].(string)

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxAvatarSize)

	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "avatar file is required"})
		return
	}

	contentType := fileHeader.Header.Get("Content-Type")
	ext, ok := allowedAvatarTypes[contentType]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported file type, use jpeg, png or webp"})
		return
	}

	if fileHeader.Size > maxAvatarSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file exceeds 5 MB limit"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not read file"})
		return
	}
	defer file.Close()

	// Ignora a extensão enviada pelo cliente — usa a do content-type
	key := fmt.Sprintf("%s/%s/%s%s", h.r2.AvatarDir, userID, uuid.NewString(), ext)
	_ = filepath.Ext(fileHeader.Filename) // intencional: não confiamos no nome do arquivo

	url, err := h.r2.UploadAvatar(c.Request.Context(), key, file, fileHeader.Size, contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
		return
	}

	if err := h.store.UpdateUserAvatar(c.Request.Context(), userID, url); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save avatar url"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"avatarUrl": url})
}

// sanitizeFilename garante que o nome não tenha path traversal.
func sanitizeFilename(name string) string {
	return strings.ReplaceAll(filepath.Base(name), "..", "")
}
