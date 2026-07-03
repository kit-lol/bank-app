package handlers

import (
	"bank-app/repository"
	"database/sql"
	"fmt"
	"net/http"
	"time"

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

		if len(password) < 8 {
			http.Error(w, "Пароль слишком короткий (минимум 8 символов)", http.StatusBadRequest)
			return
		}

		err := repository.RegisterUser(db, username, password)
		if err != nil {
			http.Error(w, "Ошибка при регистрации: "+err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)

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
			http.Error(w, "Пользователь не найден", http.StatusUnauthorized)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
		if err != nil {
			http.Error(w, "Неверный пароль", http.StatusUnauthorized)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    fmt.Sprintf("%d", user.ID),
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
			Path:     "/",
		})

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}

}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Создаем куку с тем же именем, но с временем истечения в прошлом
	cookie := &http.Cookie{
		Name:    "session_token",
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0), // Устанавливаем время в 1970 год
		MaxAge:  -1,              // Удаляем куку сразу
	}

	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
