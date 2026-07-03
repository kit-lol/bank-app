package sessions

import (
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

var Store *sessions.CookieStore

// InitSession инициализирует хранилище сессий
func InitSession() {
	// Получаем секретный ключ из переменной окружения
	secretKey := os.Getenv("SESSION_SECRET")
	if secretKey == "" {
		// ВНИМАНИЕ: Для продакшена ОБЯЗАТЕЛЬНО установи SESSION_SECRET в .env!
		secretKey = "change-this-to-random-secret-key-in-production"
	}

	Store = sessions.NewCookieStore([]byte(secretKey))

	// Настройки безопасности
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400,                // 24 часа в секундах
		HttpOnly: true,                 // Нельзя получить через JavaScript
		Secure:   false,                // Поставь true если используешь HTTPS
		SameSite: http.SameSiteLaxMode, // Защита от CSRF
	}
}
