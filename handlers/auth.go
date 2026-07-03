package handlers

import (
	"bank-app/logger"
	"bank-app/repository"
	"bank-app/sessions"
	"database/sql"
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

		username := r.FormValue("username")
		password := r.FormValue("password")

		logger.Log.Info("Попытка регистрации", zap.String("username", username))

		if len(password) < 8 {
			logger.Log.Warn("Регистрация отклонена: короткий пароль", zap.String("username", username))
			http.Error(w, "Пароль слишком короткий (минимум 8 символов)", http.StatusBadRequest)
			return
		}

		err := repository.RegisterUser(db, username, password)
		if err != nil {
			logger.Log.Error("Ошибка БД при регистрации", zap.Error(err), zap.String("username", username))
			http.Error(w, "Ошибка при регистрации", http.StatusInternalServerError)
			return
		}

		user, err := repository.GetUserByUsername(db, username)
		if err != nil {
			logger.Log.Error("Ошибка получения пользователя после регистрации", zap.Error(err))
			http.Error(w, "Ошибка при входе после регистрации", http.StatusInternalServerError)
			return
		}

		session, _ := sessions.Store.Get(r, "session")
		session.Values["user_id"] = user.ID
		session.Save(r, w)

		logger.Log.Info("Пользователь зарегистрирован и вошел в систему", zap.Int("userID", user.ID))
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			http.ServeFile(w, r, "templates/login.html")
			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")

		user, err := repository.GetUserByUsername(db, username)
		if err != nil {
			logger.Log.Warn("Неудачная попытка входа: пользователь не найден", zap.String("username", username))
			http.Error(w, "Неверный логин или пароль", http.StatusUnauthorized)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
		if err != nil {
			logger.Log.Warn("Неудачная попытка входа: неверный пароль", zap.String("username", username))
			http.Error(w, "Неверный логин или пароль", http.StatusUnauthorized)
			return
		}

		// Создаем сессию
		session, _ := sessions.Store.Get(r, "session")
		session.Values["user_id"] = user.ID
		session.Save(r, w)

		logger.Log.Info("Успешный вход в систему", zap.Int("userID", user.ID), zap.String("username", username), zap.String("role", user.Role))

		// === НОВОЕ: Перенаправление в зависимости от роли ===
		if user.Role == "admin" {
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		}
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := sessions.Store.Get(r, "session")
	session.Options.MaxAge = -1
	session.Save(r, w)

	logger.Log.Info("Пользователь вышел из системы")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
