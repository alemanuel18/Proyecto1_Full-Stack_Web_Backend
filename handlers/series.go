package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/seriestracker/backend/middleware"
	"github.com/seriestracker/backend/models"
)

type SeriesHandler struct {
	DB *sql.DB
}

func NewSeriesHandler(db *sql.DB) *SeriesHandler {
	return &SeriesHandler{DB: db}
}

// GET /series
// Supports: ?page=1&limit=10&q=breaking&sort=title&order=asc
func (h *SeriesHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)

	// Pagination
	page := queryInt(r, "page", 1)
	limit := queryInt(r, "limit", 20)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Search
	search := strings.TrimSpace(r.URL.Query().Get("q"))

	// Sort
	sortField := r.URL.Query().Get("sort")
	sortOrder := strings.ToUpper(r.URL.Query().Get("order"))

	allowedSorts := map[string]string{
		"title":      "title",
		"rating":     "rating",
		"status":     "status",
		"created_at": "created_at",
		"updated_at": "updated_at",
	}
	col, ok := allowedSorts[sortField]
	if !ok {
		col = "created_at"
	}
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}

	// Build query
	args := []interface{}{userID}
	where := "WHERE user_id = $1"

	if search != "" {
		args = append(args, "%"+search+"%")
		where += fmt.Sprintf(" AND title ILIKE $%d", len(args))
	}

	// Count total
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM series %s", where)
	if err := h.DB.QueryRow(countQuery, args...).Scan(&total); err != nil {
		respondError(w, http.StatusInternalServerError, "could not count series")
		return
	}

	// Fetch page
	args = append(args, limit, offset)
	dataQuery := fmt.Sprintf(
		`SELECT id, user_id, title, genre, status, rating, cover_url, description, episodes, created_at, updated_at
		 FROM series %s
		 ORDER BY %s %s
		 LIMIT $%d OFFSET $%d`,
		where, col, sortOrder, len(args)-1, len(args),
	)

	rows, err := h.DB.Query(dataQuery, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not fetch series")
		return
	}
	defer rows.Close()

	seriesList := []models.Series{}
	for rows.Next() {
		var s models.Series
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.Title, &s.Genre, &s.Status,
			&s.Rating, &s.CoverURL, &s.Description, &s.Episodes,
			&s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			respondError(w, http.StatusInternalServerError, "could not scan series")
			return
		}
		seriesList = append(seriesList, s)
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	respondJSON(w, http.StatusOK, models.PaginatedSeries{
		Data:       seriesList,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

// GET /series/:id
func (h *SeriesHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	id, err := extractID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid series id")
		return
	}

	var s models.Series
	err = h.DB.QueryRow(
		`SELECT id, user_id, title, genre, status, rating, cover_url, description, episodes, created_at, updated_at
		 FROM series WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&s.ID, &s.UserID, &s.Title, &s.Genre, &s.Status,
		&s.Rating, &s.CoverURL, &s.Description, &s.Episodes,
		&s.CreatedAt, &s.UpdatedAt)

	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "series not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not fetch series")
		return
	}

	respondJSON(w, http.StatusOK, s)
}

// POST /series
func (h *SeriesHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)

	var req models.CreateSeriesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.Title) == "" {
		respondError(w, http.StatusBadRequest, "title is required")
		return
	}

	status := req.Status
	if status == "" {
		status = "plan_to_watch"
	}
	validStatuses := map[string]bool{
		"watching": true, "completed": true,
		"dropped": true, "plan_to_watch": true,
	}
	if !validStatuses[status] {
		respondError(w, http.StatusBadRequest, "status must be one of: watching, completed, dropped, plan_to_watch")
		return
	}

	if req.Rating != nil && (*req.Rating < 1 || *req.Rating > 10) {
		respondError(w, http.StatusBadRequest, "rating must be between 1 and 10")
		return
	}

	var s models.Series
	err := h.DB.QueryRow(
		`INSERT INTO series (user_id, title, genre, status, rating, description, episodes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, user_id, title, genre, status, rating, cover_url, description, episodes, created_at, updated_at`,
		userID, req.Title, req.Genre, status, req.Rating, req.Description, req.Episodes,
	).Scan(&s.ID, &s.UserID, &s.Title, &s.Genre, &s.Status,
		&s.Rating, &s.CoverURL, &s.Description, &s.Episodes,
		&s.CreatedAt, &s.UpdatedAt)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not create series")
		return
	}

	respondJSON(w, http.StatusCreated, s)
}

// PUT /series/:id
func (h *SeriesHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	var req models.UpdateSeriesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate status if provided
	if req.Status != nil {
		validStatuses := map[string]bool{
			"watching": true, "completed": true,
			"dropped": true, "plan_to_watch": true,
		}
		if !validStatuses[*req.Status] {
			respondError(w, http.StatusBadRequest, "status must be one of: watching, completed, dropped, plan_to_watch")
			return
		}
	}

	if req.Rating != nil && (*req.Rating < 1 || *req.Rating > 10) {
		respondError(w, http.StatusBadRequest, "rating must be between 1 and 10")
		return
	}

	// Build dynamic SET clause
	setClauses := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argIdx := 1

	addField := func(col string, val interface{}) {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, argIdx))
		args = append(args, val)
		argIdx++
	}

	if req.Title != nil {
		if strings.TrimSpace(*req.Title) == "" {
			respondError(w, http.StatusBadRequest, "title cannot be empty")
			return
		}
		addField("title", *req.Title)
	}
	if req.Genre != nil {
		addField("genre", *req.Genre)
	}
	if req.Status != nil {
		addField("status", *req.Status)
	}
	if req.Rating != nil {
		addField("rating", *req.Rating)
	}
	if req.CoverURL != nil {
		addField("cover_url", *req.CoverURL)
	}
	if req.Description != nil {
		addField("description", *req.Description)
	}
	if req.Episodes != nil {
		addField("episodes", *req.Episodes)
	}

	args = append(args, id, userID)
	query := fmt.Sprintf(
		`UPDATE series SET %s WHERE id = $%d AND user_id = $%d
		 RETURNING id, user_id, title, genre, status, rating, cover_url, description, episodes, created_at, updated_at`,
		strings.Join(setClauses, ", "), argIdx, argIdx+1,
	)

	var s models.Series
	if err := h.DB.QueryRow(query, args...).Scan(
		&s.ID, &s.UserID, &s.Title, &s.Genre, &s.Status,
		&s.Rating, &s.CoverURL, &s.Description, &s.Episodes,
		&s.CreatedAt, &s.UpdatedAt,
	); err != nil {
		respondError(w, http.StatusInternalServerError, "could not update series")
		return
	}

	respondJSON(w, http.StatusOK, s)
}

// DELETE /series/:id
func (h *SeriesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)
	id, err := extractID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid series id")
		return
	}

	result, err := h.DB.Exec(
		`DELETE FROM series WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not delete series")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		respondError(w, http.StatusNotFound, "series not found")
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 — no body on delete
}

// ── helpers ──────────────────────────────────────────────────────────────────

func queryInt(r *http.Request, key string, def int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func extractID(r *http.Request) (int, error) {
	// Works with path like /series/42 — last segment
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	return strconv.Atoi(parts[len(parts)-1])
}