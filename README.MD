# Architecture Overview:

The application is containerized using Docker and uses PostgreSQL as its database
It uses the Templ templating engine for HTML rendering
Authentication is handled via Google OAuth
The project follows a clean architecture pattern with separated concerns

# Main Components:

## Entry Point (cmd/server/main.go):


Initializes all core services (database, auth, handlers)
Sets up a HTTP server with proper timeouts
Implements graceful shutdown


## Authentication (internal/auth/auth.go):


Handles Google OAuth flow using the goth library
Manages user sessions with gorilla/sessions
Provides login/logout functionality


## Database (internal/db/db.go):


Manages PostgreSQL connection using pgx
Creates necessary tables on startup
Handles user data persistence and record retrieval


## Configuration (internal/config/config.go):


Loads environment variables
Manages application configuration (database URL, OAuth credentials, etc.)

## Application Flow:

## User Access Flow:

CopyUser visits / → Middleware checks auth → 
If not authenticated → Redirect to /login →
Google OAuth flow → Callback to /callback →
Store user info → Redirect to homepage

## Protected Page Flow:

CopyAuthenticated user visits / →
Load records from database →
Render homepage with user data and records

## Data Display:


The homepage (hello.templ) shows:

User information (name, email, avatar)
A table of records if authenticated
Login/Logout buttons




## Authentication Flow:

Copy/login → Google OAuth → /callback →
Store user in database → Set auth cookie →
Redirect to homepage
Environment Setup:

## Requires environment variables for:

DATABASE_URL
GOOGLE_CLIENT_ID
GOOGLE_CLIENT_SECRET
SESSION_SECRET
HOST (optional, defaults to http://localhost:3000)



## Docker Setup:

Multi-stage build for smaller final image
Uses Alpine Linux for base image
Runs as non-root user for security
Exposes port 3000
