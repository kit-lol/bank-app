package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		_ = cookie
		next(w, r)
	}
}

func AdminMiddleware(db *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		userID, _ := strconv.Atoi(cookie.Value)

		var role string
		err = db.QueryRow("SELECT role FROM users WHERE id = $1", userID).Scan(&role)

		if err != nil || role != "admin" {
			http.Error(w, "Доступ запрещен", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}
