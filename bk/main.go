// main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/a-h/templ"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Record struct {
	ID    int32
	Name  string
	Value float32
}

type UserInfo struct {
	Email         string
	VerifiedEmail bool
	Name          string
	Picture       string
}

func main() {
	// Try to load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found - using environment variables")
	}

	// OAuth2 config
	oauth2Config := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:3000/callback",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	// Database connection
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer conn.Close(context.Background())

	// Create tables
	_, err = conn.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS playing_with_neon(
			id SERIAL PRIMARY KEY, 
			name TEXT NOT NULL, 
			value REAL
		);

		CREATE TABLE IF NOT EXISTS users(
			id SERIAL PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			name TEXT,
			picture TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		log.Fatal("Error creating tables:", err)
	}

	// Insert sample data
	_, err = conn.Exec(context.Background(), `
		INSERT INTO playing_with_neon(name, value) 
		SELECT LEFT(md5(i::TEXT), 10), random() 
		FROM generate_series(1, 10) s(i)
		ON CONFLICT DO NOTHING;
	`)
	if err != nil {
		log.Fatal("Error inserting sample data:", err)
	}

	// Authentication middleware
	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie("auth_token")
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	// Login handler
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		url := oauth2Config.AuthCodeURL("state")
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	})

	// Callback handler
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		token, err := oauth2Config.Exchange(r.Context(), code)
		if err != nil {
			http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
			return
		}

		client := oauth2Config.Client(r.Context(), token)
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil {
			http.Error(w, "Failed to get user info", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var userInfo UserInfo
		if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
			http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
			return
		}

		// Store or update user in database
		_, err = conn.Exec(r.Context(), `
			INSERT INTO users (email, name, picture)
			VALUES ($1, $2, $3)
			ON CONFLICT (email) 
			DO UPDATE SET name = $2, picture = $3
		`, userInfo.Email, userInfo.Name, userInfo.Picture)
		if err != nil {
			http.Error(w, "Failed to store user", http.StatusInternalServerError)
			return
		}

		// Set auth cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    token.AccessToken,
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
			Path:     "/",
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// Protected home page
	http.Handle("/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		records := []Record{}
		
		rows, err := conn.Query(r.Context(), "SELECT id, name, value FROM playing_with_neon")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var record Record
			if err := rows.Scan(&record.ID, &record.Name, &record.Value); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			records = append(records, record)
		}

		// Get user info from cookie
		cookie, _ := r.Cookie("auth_token")
		client := oauth2Config.Client(r.Context(), &oauth2.Token{
			AccessToken: cookie.Value,
		})
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil {
			http.Error(w, "Failed to get user info", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var userInfo UserInfo
		if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
			http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
			return
		}

		component := hello(userInfo.Name, records, userInfo)
		templ.Handler(component).ServeHTTP(w, r)
	})))

	fmt.Println("Listening on :3000")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatal("Error starting server:", err)
	}
}
