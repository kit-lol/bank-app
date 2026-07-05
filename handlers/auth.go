package handlers

import (
	"bank-app/logger"
	"bank-app/repository"
	"bank-app/sessions"
	"database/sql"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func RegisterHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			http.ServeFile(w, r, "templates/register.html")
			return
		}

		w.Header().Set("Content-Type", "application/json")

		username := r.FormValue("username")
		password := r.FormValue("password")

		sendError := func(msg string, code int) {
			w.WriteHeader(code)
			json.NewEncoder(w).Encode(map[string]string{"error": msg})
		}

		if len(username) < 3 || len(username) > 50 {
			logger.Log.Warn("Ошибка регистрации: некорректный логин", zap.String("username", username))
			sendError("Логин должен быть от 3 до 50 символов", http.StatusBadRequest)
			return
		}

		if len(password) < 8 {
			logger.Log.Warn("Ошибка регистрации: короткий пароль")
			sendError("Пароль должен содержать минимум 8 символов", http.StatusBadRequest)
			return
		}

		hasLetter := false
		hasNumber := false
		for _, char := range password {
			if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
				hasLetter = true
			}
			if char >= '0' && char <= '9' {
				hasNumber = true
			}
		}

		if !hasLetter || !hasNumber {
			sendError("Пароль должен содержать буквы и цифры", http.StatusBadRequest)
			return
		}

		logger.Log.Info("Попытка регистрации", zap.String("username", username))

		err := repository.RegisterUser(db, username, password)
		if err != nil {
			logger.Log.Error("Ошибка БД при регистрации", zap.Error(err), zap.String("username", username))
			sendError("Этот логин уже занят или произошла ошибка БД", http.StatusInternalServerError)
			return
		}

		user, err := repository.GetUserByUsername(db, username)
		if err != nil {
			logger.Log.Error("Ошибка получения пользователя после регистрации", zap.Error(err))
			sendError("Ошибка при создании сессии", http.StatusInternalServerError)
			return
		}

		session, _ := sessions.Store.Get(r, "session")
		session.Values["user_id"] = user.ID
		session.Save(r, w)

		logger.Log.Info("Пользователь зарегистрирован и вошел в систему", zap.Int("userID", user.ID))

		json.NewEncoder(w).Encode(map[string]string{"redirect": "/dashboard"})
	}
}

func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			http.ServeFile(w, r, "templates/login.html")
			return
		}

		w.Header().Set("Content-Type", "application/json")

		username := r.FormValue("username")
		password := r.FormValue("password")

		user, err := repository.GetUserByUsername(db, username)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Неверный логин или пароль"})
			return
		}

		if !user.IsActive {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "Аккаунт заблокирован"})
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Неверный логин или пароль"})
			return
		}

		session, _ := sessions.Store.Get(r, "session")
		session.Values["user_id"] = user.ID
		session.Save(r, w)

		redirectURL := "/dashboard"
		if user.Role == "admin" {
			redirectURL = "/admin"
		}

		json.NewEncoder(w).Encode(map[string]string{"redirect": redirectURL})
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := sessions.Store.Get(r, "session")
	session.Options.MaxAge = -1
	session.Save(r, w)

	logger.Log.Info("Пользователь вышел из системы")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
