package auth

import (
	"net/http"
	"rr/web/internal/models"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

type Service struct{}

func NewService(clientID, clientSecret, callbackURL, sessionSecret string) *Service {
	// Initialize the session store
	store := sessions.NewCookieStore([]byte(sessionSecret))
	store.MaxAge(86400 * 30) // 30 days
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = false // Set to true in production with HTTPS

	gothic.Store = store

	// Set up Gothic session store and provider
	provider := google.New(clientID, clientSecret, callbackURL,
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/userinfo.profile",
	)

	// Set provider name consistently
	provider.SetName("google")

	goth.UseProviders(provider)

	// Configure Gothic to always use Google provider
	gothic.GetProviderName = func(req *http.Request) (string, error) {
		return "google", nil
	}

	return &Service{}
}

func (s *Service) BeginAuth(w http.ResponseWriter, r *http.Request) {
	// Set the provider name and other state information before beginning auth
	r.URL.Query().Set("provider", "google")
	gothic.BeginAuthHandler(w, r)
}

func (s *Service) CompleteAuth(w http.ResponseWriter, r *http.Request) (*models.User, error) {
	gothUser, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		return nil, err
	}

	return &models.User{
		Email:   gothUser.Email,
		Name:    gothUser.Name,
		Picture: gothUser.AvatarURL,
	}, nil
}

func (s *Service) Logout(w http.ResponseWriter, r *http.Request) error {
	return gothic.Logout(w, r)
}
