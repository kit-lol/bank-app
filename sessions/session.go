package sessions

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

var Store *sessions.CookieStore

// InitSession инициализирует хранилище сессий
func InitSession() {
	secretKey := os.Getenv("SESSION_SECRET")
	if secretKey == "" {
		log.Fatal("SESSION_SECRET не задан. Установите переменную окружения SESSION_SECRET.")
	}

	Store = sessions.NewCookieStore([]byte(secretKey))

	// Настройки безопасности
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		Secure:   true, // Требует HTTPS; для локальной разработки можно временно false
		SameSite: http.SameSiteLaxMode,
	}
}
