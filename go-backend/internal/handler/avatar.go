package handler

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/middleware"
	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

const (
	maxAvatarSize = 5 << 20 // 5 MB
	avatarSubdir  = "uploads/avatars"
)

var allowedImageTypes = map[string]bool{
	"image/png":  true,
	"image/jpeg": true,
	"image/gif":  true,
}

// AvatarHandler provides avatar upload, serve, and delete endpoints.
type AvatarHandler struct {
	db      *gorm.DB
	baseDir string // root directory for file storage (e.g. "." or an absolute path)
}

func NewAvatarHandler(db *gorm.DB, baseDir string) *AvatarHandler {
	return &AvatarHandler{db: db, baseDir: baseDir}
}

func (h *AvatarHandler) avatarDir() string {
	return filepath.Join(h.baseDir, avatarSubdir)
}

// UploadAvatar handles POST /profile/avatar.
// Accepts multipart form with "avatar" file field.
func (h *AvatarHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxAvatarSize+1024) // slight overhead for multipart headers
	if err := r.ParseMultipartForm(maxAvatarSize); err != nil {
		writeError(w, http.StatusBadRequest, "file too large (max 5MB)")
		return
	}

	file, _, err := r.FormFile("avatar")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing avatar file")
		return
	}
	defer file.Close()

	// Read file into memory to detect content type and check size
	data, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read file")
		return
	}

	if len(data) > maxAvatarSize {
		writeError(w, http.StatusBadRequest, "file too large (max 5MB)")
		return
	}

	// Detect content type from file bytes
	contentType := http.DetectContentType(data)
	if !allowedImageTypes[contentType] {
		writeError(w, http.StatusBadRequest, "unsupported image type (use PNG, JPEG, or GIF)")
		return
	}

	// Determine file extension from content type
	ext := ".jpg"
	switch contentType {
	case "image/png":
		ext = ".png"
	case "image/gif":
		ext = ".gif"
	}

	// Create uploads directory
	dir := h.avatarDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create upload directory")
		return
	}

	// Write file to disk: avatar_{userID}{ext}
	filename := fmt.Sprintf("avatar_%d%s", claims.UserID, ext)
	destPath := filepath.Join(dir, filename)

	// Remove any existing avatar files for this user (different extensions)
	h.removeAvatarFiles(claims.UserID)

	if err := os.WriteFile(destPath, data, 0o644); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save file")
		return
	}
	written := int64(len(data))

	// Upsert avatar record in DB
	now := time.Now().UTC()
	var avatar model.Avatar
	result := h.db.Where("entity_type = ? AND entity_id = ?", "User", claims.UserID).First(&avatar)
	if result.Error != nil {
		// Create new
		avatar = model.Avatar{
			UserID:           claims.UserID,
			EntityType:       "User",
			EntityID:         claims.UserID,
			ImageFileName:    &filename,
			ImageContentType: &contentType,
			ImageFileSize:    &written,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		h.db.Create(&avatar)
	} else {
		// Update existing
		h.db.Model(&avatar).Updates(map[string]interface{}{
			"image_file_name":    filename,
			"image_content_type": contentType,
			"image_file_size":    written,
			"updated_at":         now,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"avatar_url": fmt.Sprintf("/api/v1/avatars/%d", claims.UserID),
	})
}

// DeleteAvatar handles DELETE /profile/avatar.
func (h *AvatarHandler) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	h.removeAvatarFiles(claims.UserID)
	h.db.Where("entity_type = ? AND entity_id = ?", "User", claims.UserID).Delete(&model.Avatar{})

	writeJSON(w, http.StatusOK, map[string]string{"status": "avatar removed"})
}

// ServeAvatar handles GET /avatars/{user_id}.
// Serves the avatar image file, or redirects to Gravatar if none exists.
func (h *AvatarHandler) ServeAvatar(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.PathValue("user_id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	var avatar model.Avatar
	if err := h.db.Where("entity_type = ? AND entity_id = ?", "User", userID).First(&avatar).Error; err != nil {
		// No avatar — redirect to Gravatar
		h.serveGravatar(w, r, userID)
		return
	}

	if avatar.ImageFileName == nil {
		h.serveGravatar(w, r, userID)
		return
	}

	filePath := filepath.Join(h.avatarDir(), *avatar.ImageFileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		h.serveGravatar(w, r, userID)
		return
	}

	if avatar.ImageContentType != nil {
		w.Header().Set("Content-Type", *avatar.ImageContentType)
	}
	w.Header().Set("Cache-Control", "public, max-age=3600")
	http.ServeFile(w, r, filePath)
}

func (h *AvatarHandler) serveGravatar(w http.ResponseWriter, r *http.Request, userID int64) {
	var user model.User
	if err := h.db.First(&user, userID).Error; err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	hash := md5.Sum([]byte(strings.ToLower(strings.TrimSpace(user.Email))))
	size := r.URL.Query().Get("size")
	if size == "" {
		size = "75"
	}
	gravatarURL := fmt.Sprintf("https://www.gravatar.com/avatar/%x?s=%s&d=mm", hash, size)
	http.Redirect(w, r, gravatarURL, http.StatusFound)
}

// removeAvatarFiles deletes all avatar files for a user from disk.
func (h *AvatarHandler) removeAvatarFiles(userID int64) {
	for _, ext := range []string{".png", ".jpg", ".gif"} {
		path := filepath.Join(h.avatarDir(), fmt.Sprintf("avatar_%d%s", userID, ext))
		os.Remove(path)
	}
}

// GravatarURL returns the Gravatar URL for a user email.
func GravatarURL(email string, size int) string {
	hash := md5.Sum([]byte(strings.ToLower(strings.TrimSpace(email))))
	return fmt.Sprintf("https://www.gravatar.com/avatar/%x?s=%d&d=mm", hash, size)
}
