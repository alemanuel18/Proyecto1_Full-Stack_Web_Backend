package models

import "time"

type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Series struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Title       string    `json:"title"`
	Genre       *string   `json:"genre"`        // nullable
	Status      string    `json:"status"`
	Rating      *int      `json:"rating"`       // nullable
	CoverURL    *string   `json:"cover_url"`    // nullable
	Description *string   `json:"description"`  // nullable
	Episodes    *int      `json:"episodes"`     // nullable
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Request bodies

type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateSeriesRequest struct {
	Title       string  `json:"title"`
	Genre       *string `json:"genre"`
	Status      string  `json:"status"`
	Rating      *int    `json:"rating"`
	Description *string `json:"description"`
	Episodes    *int    `json:"episodes"`
}

type UpdateSeriesRequest struct {
	Title       *string `json:"title"`
	Genre       *string `json:"genre"`
	Status      *string `json:"status"`
	Rating      *int    `json:"rating"`
	CoverURL    *string `json:"cover_url"`
	Description *string `json:"description"`
	Episodes    *int    `json:"episodes"`
}

// Response wrappers

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type PaginatedSeries struct {
	Data       []Series `json:"data"`
	Total      int      `json:"total"`
	Page       int      `json:"page"`
	Limit      int      `json:"limit"`
	TotalPages int      `json:"total_pages"`
}