package handlers

import (
	"bank-app/sessions"
	"database/sql"
	"fmt"
	"net/http"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := sessions.Store.Get(r, "session")

		if err != nil || session.Values["user_id"] == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		r.Header.Set("X-User-ID", fmt.Sprintf("%d", session.Values["user_id"]))

		next(w, r)
	}
}

func AdminMiddleware(db *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := sessions.Store.Get(r, "session")

		if err != nil || session.Values["user_id"] == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		userID := session.Values["user_id"].(int)

		var role string
		err = db.QueryRow("SELECT role FROM users WHERE id = $1", userID).Scan(&role)

		if err != nil || role != "admin" {
			http.Error(w, "Доступ запрещен", http.StatusForbidden)
			return
		}

		r.Header.Set("X-User-ID", fmt.Sprintf("%d", userID))

		next(w, r)
	}
}
