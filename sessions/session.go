package sessions

import (
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

var Store *sessions.CookieStore

// InitSession инициализирует хранилище сессий
func InitSession() {
	secretKey := os.Getenv("SESSION_SECRET")
	if secretKey == "" {
		secretKey = "change-this-to-random-secret-key-in-production"
	}

	Store = sessions.NewCookieStore([]byte(secretKey))

	// Настройки безопасности
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
}
