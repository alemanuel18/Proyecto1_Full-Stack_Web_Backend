package routes

import (
	"database/sql"
	"net/http"

	"github.com/seriestracker/backend/handlers"
	"github.com/seriestracker/backend/middleware"
)

func Register(db *sql.DB, jwtSecret, cloudCloud, cloudKey, cloudSecret string) http.Handler {
	mux := http.NewServeMux()

	authHandler := handlers.NewAuthHandler(db, jwtSecret)
	seriesHandler := handlers.NewSeriesHandler(db)
	uploadHandler := handlers.NewUploadHandler(db, cloudCloud, cloudKey, cloudSecret)

	authMiddleware := middleware.Auth(jwtSecret)

	// ── Public routes ─────────────────────────────────────────────────────────
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)

	// ── Protected routes ──────────────────────────────────────────────────────
	// Auth
	mux.Handle("GET /auth/me", authMiddleware(http.HandlerFunc(authHandler.Me)))

	// Series CRUD
	mux.Handle("GET /series", authMiddleware(http.HandlerFunc(seriesHandler.List)))
	mux.Handle("GET /series/{id}", authMiddleware(http.HandlerFunc(seriesHandler.Get)))
	mux.Handle("POST /series", authMiddleware(http.HandlerFunc(seriesHandler.Create)))
	mux.Handle("PUT /series/{id}", authMiddleware(http.HandlerFunc(seriesHandler.Update)))
	mux.Handle("DELETE /series/{id}", authMiddleware(http.HandlerFunc(seriesHandler.Delete)))

	// Image upload
	mux.Handle("POST /series/{id}/image", authMiddleware(http.HandlerFunc(uploadHandler.UploadCover)))

	// Wrap everything in CORS middleware
	return middleware.CORS(mux)
}