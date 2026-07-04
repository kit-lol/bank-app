package handlers

import (
	"bank-app/logger"
	"bank-app/models"
	"bank-app/repository"
	"bank-app/service"
	"database/sql"
	"html/template"
	"net/http"
	"strconv"

	"go.uber.org/zap"
)

func AdminDashboardHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := repository.GetAllUser(db)
		if err != nil {
			logger.Log.Error("Ошибка получения списка пользователей", zap.Error(err))
			http.Error(w, "Ошибка: "+err.Error(), 500)
			return
		}

		activeCount := 0
		blockedCount := 0
		for _, u := range users {
			if u.IsActive {
				activeCount++
			} else {
				blockedCount++
			}
		}

		logger.Log.Info("Админ просмотрел список пользователей", zap.Int("count", len(users)))

		data := struct {
			Users        []models.User
			Total        int
			ActiveCount  int
			BlockedCount int
		}{
			Users:        users,
			Total:        len(users),
			ActiveCount:  activeCount,
			BlockedCount: blockedCount,
		}

		tmpl := template.Must(template.ParseFiles("templates/admin_dashboard.html"))
		err = tmpl.Execute(w, data)
		if err != nil {
			logger.Log.Error("Ошибка рендера шаблона админки", zap.Error(err))
		}
	}
}

func AdminActionHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDStr := r.FormValue("user_id")
		action := r.FormValue("action")
		amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)
		userID, err := strconv.Atoi(userIDStr)

		if err != nil {
			logger.Log.Error("Некорректный ID пользователя в действии админа", zap.String("userIDStr", userIDStr))
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
			return
		}

		logger.Log.Info("Действие администратора",
			zap.String("action", action),
			zap.Int("targetUserID", userID),
			zap.Float64("amount", amount))

		switch action {
		case "add_funds":
			if amount <= 0 {
				http.Redirect(w, r, "/admin", http.StatusSeeOther)
				return
			}
			service.ValidateAndAdjustBalance(db, userID, amount, true)

		case "withdraw_funds":
			if amount <= 0 {
				http.Redirect(w, r, "/admin", http.StatusSeeOther)
				return
			}
			service.ValidateAndAdjustBalance(db, userID, -amount, true)

		case "toggle_status":
			var isActive bool
			err := db.QueryRow("SELECT is_active FROM users WHERE id = $1", userID).Scan(&isActive)
			if err != nil {
				logger.Log.Error("Ошибка получения статуса пользователя", zap.Error(err))
				break
			}

			newStatus := !isActive
			_, err = db.Exec("UPDATE users SET is_active = $1 WHERE id = $2", newStatus, userID)
			if err != nil {
				logger.Log.Error("Ошибка обновления статуса", zap.Error(err))
			} else {
				logger.Log.Info("Статус пользователя изменен", zap.Int("userID", userID), zap.Bool("newStatus", newStatus))
			}

		case "view_user":
			http.Redirect(w, r, "/admin/user?id="+userIDStr, http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
}

func AdminUserDetailHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDStr := r.URL.Query().Get("id")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "Некорректный ID пользователя", http.StatusBadRequest)
			return
		}

		rows, err := db.Query("SELECT id, user_id, amount, interest_rate, status, created_at FROM deposits WHERE user_id = $1", userID)
		if err != nil {
			logger.Log.Error("Ошибка поиска вкладов пользователя", zap.Error(err), zap.Int("userID", userID))
			http.Error(w, "Ошибка БД", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var deposits []models.Deposit
		for rows.Next() {
			var d models.Deposit
			err := rows.Scan(&d.ID, &d.UserID, &d.Amount, &d.InterestRate, &d.Status, &d.CreatedAt)
			if err != nil {
				continue
			}
			deposits = append(deposits, d)
		}

		transactions, _ := repository.GetUserTransactions(db, userID)

		data := struct {
			UserID       int
			Deposits     []models.Deposit
			Transactions []models.Transaction
		}{
			UserID:       userID,
			Deposits:     deposits,
			Transactions: transactions,
		}

		tmpl := template.Must(template.ParseFiles("templates/admin_user_detail.html"))
		tmpl.Execute(w, data)
	}
}
