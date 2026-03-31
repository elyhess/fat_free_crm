package handler

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/elyhess/fat-free-crm-backend/internal/auth"
)

func avatarRouter(t *testing.T) (*http.ServeMux, *auth.JWTService, string) {
	t.Helper()
	db := testDB(t)
	tmpDir := t.TempDir()

	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	cfg := RouterConfig{DB: db, JWTSecret: "test-secret", JWTExpiryHours: 1, AvatarDir: tmpDir}
	router := NewRouter(cfg)
	mux := http.NewServeMux()
	mux.Handle("/", router)

	// Seed a user
	n := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO users (id, username, email, encrypted_password, password_salt, admin, first_name, last_name, created_at, updated_at) VALUES (1, 'admin', 'admin@test.com', 'x', 'y', true, 'Admin', 'User', ?, ?)", n, n)
	db.Exec("INSERT INTO users (id, username, email, encrypted_password, password_salt, admin, first_name, last_name, created_at, updated_at) VALUES (2, 'demo', 'demo@test.com', 'x', 'y', false, 'Demo', 'User', ?, ?)", n, n)

	return mux, jwtSvc, tmpDir
}

func createTestPNG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

func uploadAvatar(t *testing.T, mux *http.ServeMux, token string, fileData []byte, filename, contentType string) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("avatar", filename)
	if err != nil {
		t.Fatal(err)
	}
	part.Write(fileData)
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/profile/avatar", &body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestUploadAvatar(t *testing.T) {
	mux, jwtSvc, tmpDir := avatarRouter(t)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)

	pngData := createTestPNG()
	rec := uploadAvatar(t, mux, tok, pngData, "photo.png", "image/png")

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["avatar_url"] != "/api/v1/avatars/1" {
		t.Errorf("expected avatar_url /api/v1/avatars/1, got %v", resp["avatar_url"])
	}

	// Verify file on disk
	path := filepath.Join(tmpDir, "uploads/avatars/avatar_1.png")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("expected avatar file on disk")
	}

	// Verify the serve endpoint returns the image
	serveReq := httptest.NewRequest("GET", "/api/v1/avatars/1", nil)
	serveReq.Header.Set("Authorization", "Bearer "+tok)
	serveRec := httptest.NewRecorder()
	mux.ServeHTTP(serveRec, serveReq)

	if serveRec.Code != 200 {
		t.Errorf("expected serve to return 200 after upload, got %d", serveRec.Code)
	}
	if ct := serveRec.Header().Get("Content-Type"); ct != "image/png" {
		t.Errorf("expected Content-Type image/png, got %s", ct)
	}
}

func TestUploadAvatar_ReplaceExisting(t *testing.T) {
	mux, jwtSvc, tmpDir := avatarRouter(t)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)

	// Upload first
	pngData := createTestPNG()
	uploadAvatar(t, mux, tok, pngData, "photo.png", "image/png")

	// Upload again (replace)
	rec := uploadAvatar(t, mux, tok, pngData, "photo2.png", "image/png")
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Only one file should exist
	files, _ := filepath.Glob(filepath.Join(tmpDir, "uploads/avatars/avatar_1.*"))
	if len(files) != 1 {
		t.Errorf("expected exactly 1 avatar file, got %d: %v", len(files), files)
	}
}

func TestUploadAvatar_NoAuth(t *testing.T) {
	mux, _, _ := avatarRouter(t)

	req := httptest.NewRequest("POST", "/api/v1/profile/avatar", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 401 {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestUploadAvatar_InvalidType(t *testing.T) {
	mux, jwtSvc, _ := avatarRouter(t)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	h := make(map[string][]string)
	h["Content-Disposition"] = []string{`form-data; name="avatar"; filename="test.txt"`}
	h["Content-Type"] = []string{"text/plain"}
	part, _ := writer.CreatePart(h)
	part.Write([]byte("not an image"))
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/profile/avatar", &body)
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Errorf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteAvatar(t *testing.T) {
	mux, jwtSvc, tmpDir := avatarRouter(t)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)

	// Upload first
	pngData := createTestPNG()
	uploadAvatar(t, mux, tok, pngData, "photo.png", "image/png")

	// Delete
	req := httptest.NewRequest("DELETE", "/api/v1/profile/avatar", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// File removed
	path := filepath.Join(tmpDir, "uploads/avatars/avatar_1.png")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("expected avatar file to be deleted")
	}
}

func TestServeAvatar(t *testing.T) {
	mux, jwtSvc, _ := avatarRouter(t)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)

	// Upload
	pngData := createTestPNG()
	uploadAvatar(t, mux, tok, pngData, "photo.png", "image/png")

	// Serve (auth not required for serving)
	req := httptest.NewRequest("GET", "/api/v1/avatars/1", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "image/png" {
		t.Errorf("expected Content-Type image/png, got %s", ct)
	}
}

func TestServeAvatar_GravatarFallback(t *testing.T) {
	mux, jwtSvc, _ := avatarRouter(t)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)

	// No avatar uploaded — should redirect to Gravatar
	req := httptest.NewRequest("GET", "/api/v1/avatars/1", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 302 {
		t.Fatalf("expected 302 redirect, got %d: %s", rec.Code, rec.Body.String())
	}

	location := rec.Header().Get("Location")
	if !bytes.Contains([]byte(location), []byte("gravatar.com/avatar/")) {
		t.Errorf("expected Gravatar redirect, got %s", location)
	}
}

func TestProfileIncludesAvatarURL(t *testing.T) {
	mux, jwtSvc, _ := avatarRouter(t)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)

	req := httptest.NewRequest("GET", "/api/v1/profile", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	body := rec.Body.Bytes()
	var resp map[string]interface{}
	json.Unmarshal(body, &resp)

	if resp["avatar_url"] != "/api/v1/avatars/1" {
		t.Errorf("expected avatar_url in profile response, got %v", resp["avatar_url"])
	}
}

func TestServeAvatar_NoUser(t *testing.T) {
	mux, jwtSvc, _ := avatarRouter(t)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)

	req := httptest.NewRequest("GET", "/api/v1/avatars/999", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Errorf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUploadAvatar_MissingFile(t *testing.T) {
	mux, jwtSvc, _ := avatarRouter(t)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/profile/avatar", &body)
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Errorf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestUploadAvatar_FileTooLarge is skipped in CI because generating 5MB+ takes time.
func TestUploadAvatar_FileTooLarge(t *testing.T) {
	mux, jwtSvc, _ := avatarRouter(t)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)

	// Create a body larger than 5MB
	bigData := make([]byte, 6<<20)
	for i := range bigData {
		bigData[i] = 0xFF
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, _ := writer.CreateFormFile("avatar", "big.png")
	io.Copy(part, bytes.NewReader(bigData))
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/profile/avatar", &body)
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Errorf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}
