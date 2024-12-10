package handler

import (
	"context"
	"log"
	"net/http"
	"time"

	"rr/web/internal/db"
	"rr/web/internal/models"
	"rr/web/internal/templates"
	"rr/web/internal/auth"
)

type Handler struct {
	db     *db.Database
	auth   *auth.Service
	logger *log.Logger
}

func New(db *db.Database, authService *auth.Service, logger *log.Logger) *Handler {
	return &Handler{
		db:     db,
		auth:   authService,
		logger: logger,
	}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	// Auth routes
	mux.HandleFunc("/login", h.handleAuth)      // Changed from /auth/google to /login
	mux.HandleFunc("/callback", h.handleCallback)  // Changed from /auth/callback to /callback
	mux.HandleFunc("/logout", h.handleLogout)

	// Protected routes
	mux.Handle("/", h.requireAuth(h.handleHome))

	return mux
}

func (h *Handler) handleAuth(w http.ResponseWriter, r *http.Request) {
	h.auth.BeginAuth(w, r)
}

func (h *Handler) handleCallback(w http.ResponseWriter, r *http.Request) {
	user, err := h.auth.CompleteAuth(w, r)
	if err != nil {
		h.logger.Printf("Auth error: %v", err)
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}

	if err := h.db.UpsertUser(r.Context(), user); err != nil {
		h.logger.Printf("Database error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    user.Email,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	// Clear the auth cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	})

	// Call the auth service's logout
	if err := h.auth.Logout(w, r); err != nil {
		h.logger.Printf("Logout error: %v", err)
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *Handler) handleHome(w http.ResponseWriter, r *http.Request) {
	records, err := h.db.GetRecords(r.Context())
	if err != nil {
		h.logger.Printf("Database error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	user := r.Context().Value("user").(*models.User)

	component := templates.Hello(user.Name, records, user)
	if err := component.Render(r.Context(), w); err != nil {
		h.logger.Printf("Render error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) requireAuth(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        cookie, err := r.Cookie("auth_token")
        if err != nil {
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }

        // Load full user data from database
        user, err := h.db.GetUserByEmail(r.Context(), cookie.Value)
        if err != nil {
            h.logger.Printf("Error loading user data: %v", err)
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }

        ctx := context.WithValue(r.Context(), "user", user)
        next.ServeHTTP(w, r.WithContext(ctx))
    }
}
