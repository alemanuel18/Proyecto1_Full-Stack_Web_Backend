package routes

import (
	"database/sql"
	"net/http"

	"github.com/seriestracker/backend/handlers"
	"github.com/seriestracker/backend/middleware"
)

func Register(db *sql.DB, jwtSecret, cloudCloud, cloudKey, cloudSecret string) http.Handler {
	mux := http.NewServeMux()

	authHandler   := handlers.NewAuthHandler(db, jwtSecret)
	seriesHandler := handlers.NewSeriesHandler(db)
	uploadHandler := handlers.NewUploadHandler(db, cloudCloud, cloudKey, cloudSecret)

	auth := middleware.Auth(jwtSecret)

	// ── Public ────────────────────────────────────────────────────────────────
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login",    authHandler.Login)

	// ── Protected ─────────────────────────────────────────────────────────────
	mux.Handle("GET /auth/me", auth(http.HandlerFunc(authHandler.Me)))

	mux.Handle("GET /series",          auth(http.HandlerFunc(seriesHandler.List)))
	mux.Handle("POST /series",         auth(http.HandlerFunc(seriesHandler.Create)))
	mux.Handle("GET /series/{id}",     auth(http.HandlerFunc(seriesHandler.Get)))
	mux.Handle("PUT /series/{id}",     auth(http.HandlerFunc(seriesHandler.Update)))
	mux.Handle("DELETE /series/{id}",  auth(http.HandlerFunc(seriesHandler.Delete)))
	mux.Handle("POST /series/{id}/image", auth(http.HandlerFunc(uploadHandler.UploadCover)))

	return middleware.CORS(mux)
}