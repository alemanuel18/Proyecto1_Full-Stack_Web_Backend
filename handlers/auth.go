package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/seriestracker/backend/middleware"
	"github.com/seriestracker/backend/models"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	DB        *sql.DB
	JWTSecret string
}

func NewAuthHandler(db *sql.DB, jwtSecret string) *AuthHandler {
	return &AuthHandler{DB: db, JWTSecret: jwtSecret}
}

// POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate
	if req.Email == "" || req.Username == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "email, username and password are required")
		return
	}
	if len(req.Password) < 8 {
		respondError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not hash password")
		return
	}

	// Insert user
	var user models.User
	err = h.DB.QueryRow(
		`INSERT INTO users (email, username, password_hash)
		 VALUES ($1, $2, $3)
		 RETURNING id, email, username, created_at, updated_at`,
		req.Email, req.Username, string(hash),
	).Scan(&user.ID, &user.Email, &user.Username, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		// Check for unique constraint violations
		if isUniqueViolation(err) {
			respondError(w, http.StatusConflict, "email or username already taken")
			return
		}
		respondError(w, http.StatusInternalServerError, "could not create user")
		return
	}

	token, err := h.generateToken(user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not generate token")
		return
	}

	respondJSON(w, http.StatusCreated, models.AuthResponse{Token: token, User: user})
}

// POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	var user models.User
	err := h.DB.QueryRow(
		`SELECT id, email, username, password_hash, created_at, updated_at
		 FROM users WHERE email = $1`,
		req.Email,
	).Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		respondError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not retrieve user")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		respondError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	token, err := h.generateToken(user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not generate token")
		return
	}

	respondJSON(w, http.StatusOK, models.AuthResponse{Token: token, User: user})
}

// GET /auth/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)

	var user models.User
	err := h.DB.QueryRow(
		`SELECT id, email, username, created_at, updated_at FROM users WHERE id = $1`,
		userID,
	).Scan(&user.ID, &user.Email, &user.Username, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not retrieve user")
		return
	}

	respondJSON(w, http.StatusOK, user)
}

func (h *AuthHandler) generateToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(), // 7 days
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.JWTSecret))
}

// isUniqueViolation checks if a postgres error is a unique constraint violation (code 23505)
func isUniqueViolation(err error) bool {
	return err != nil && (contains(err.Error(), "23505") || contains(err.Error(), "unique"))
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}