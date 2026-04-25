package handlers

import (
	"bytes"
	"crypto/sha1"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/seriestracker/backend/middleware"
)

type UploadHandler struct {
	DB              *sql.DB
	CloudinaryCloud string
	CloudinaryKey   string
	CloudinarySecret string
}

func NewUploadHandler(db *sql.DB, cloud, key, secret string) *UploadHandler {
	return &UploadHandler{
		DB:               db,
		CloudinaryCloud:  cloud,
		CloudinaryKey:    key,
		CloudinarySecret: secret,
	}
}

// POST /series/:id/image
// Accepts multipart/form-data with field "image" (max 1MB as per requirements)
func (h *UploadHandler) UploadCover(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	id, err := extractID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid series id")
		return
	}

	// Check ownership
	var exists bool
	if err := h.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM series WHERE id=$1 AND user_id=$2)`, id, userID).Scan(&exists); err != nil || !exists {
		respondError(w, http.StatusNotFound, "series not found")
		return
	}

	// Limit upload size to 2MB (a bit above the 1MB requirement for safety)
	r.Body = http.MaxBytesReader(w, r.Body, 2<<20)
	if err := r.ParseMultipartForm(2 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "image too large (max 1MB)")
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		respondError(w, http.StatusBadRequest, "missing 'image' field in form")
		return
	}
	defer file.Close()

	// Validate content type
	contentType := header.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
		respondError(w, http.StatusBadRequest, "only jpeg, png, gif and webp images are allowed")
		return
	}

	// Read file bytes
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not read image")
		return
	}

	// Upload to Cloudinary
	imageURL, err := h.uploadToCloudinary(fileBytes, contentType, fmt.Sprintf("series_%d", id))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not upload image: "+err.Error())
		return
	}

	// Update series cover_url
	var updatedURL string
	err = h.DB.QueryRow(
		`UPDATE series SET cover_url = $1, updated_at = NOW()
		 WHERE id = $2 AND user_id = $3
		 RETURNING cover_url`,
		imageURL, id, userID,
	).Scan(&updatedURL)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not update series cover")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"cover_url": updatedURL,
		"message":   "image uploaded successfully",
	})
}

// uploadToCloudinary uploads raw bytes and returns the secure URL
func (h *UploadHandler) uploadToCloudinary(fileBytes []byte, contentType, publicID string) (string, error) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	// Build signature: SHA1("public_id=X&timestamp=Y" + secret)
	sigPayload := fmt.Sprintf("public_id=%s&timestamp=%s%s", publicID, timestamp, h.CloudinarySecret)
	sig := fmt.Sprintf("%x", sha1.Sum([]byte(sigPayload)))

	// Build multipart body
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add file
	ext := extensionFromMIME(contentType)
	part, err := writer.CreateFormFile("file", "image"+ext)
	if err != nil {
		return "", err
	}
	if _, err := part.Write(fileBytes); err != nil {
		return "", err
	}

	// Add fields
	fields := map[string]string{
		"api_key":   h.CloudinaryKey,
		"timestamp": timestamp,
		"public_id": publicID,
		"signature": sig,
		"folder":    "series_tracker",
	}
	for k, v := range fields {
		_ = writer.WriteField(k, v)
	}
	writer.Close()

	// POST to Cloudinary
	uploadURL := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", h.CloudinaryCloud)
	resp, err := http.Post(uploadURL, writer.FormDataContentType(), &body)
	if err != nil {
		return "", fmt.Errorf("cloudinary request failed: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("could not parse cloudinary response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		errMsg, _ := result["error"].(map[string]interface{})
		msg := "unknown cloudinary error"
		if errMsg != nil {
			msg = fmt.Sprintf("%v", errMsg["message"])
		}
		return "", fmt.Errorf("cloudinary error: %s", msg)
	}

	secureURL, ok := result["secure_url"].(string)
	if !ok {
		return "", fmt.Errorf("no secure_url in cloudinary response")
	}

	return secureURL, nil
}

func isValidImageType(ct string) bool {
	allowed := []string{"image/jpeg", "image/png", "image/gif", "image/webp"}
	for _, a := range allowed {
		if strings.HasPrefix(ct, a) {
			return true
		}
	}
	return false
}

func extensionFromMIME(mime string) string {
	switch {
	case strings.Contains(mime, "png"):
		return ".png"
	case strings.Contains(mime, "gif"):
		return ".gif"
	case strings.Contains(mime, "webp"):
		return ".webp"
	default:
		return ".jpg"
	}
}

// Ensure url package is used
var _ = url.Values{}