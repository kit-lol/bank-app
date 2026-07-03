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
		logger.Log.Info("Админ просмотрел список пользователей", zap.Int("count", len(users)))

		tmpl := template.Must(template.ParseFiles("templates/admin_dashboard.html"))
		err = tmpl.Execute(w, users)
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
		userID, _ := strconv.Atoi(userIDStr)

		logger.Log.Info("Действие администратора", zap.String("action", action), zap.Int("targetUserID", userID), zap.Float64("amount", amount))

		switch action {
		case "add_funds":
			service.ValidateAndAdjustBalance(db, userID, amount, true)
		case "withdraw_funds":
			service.ValidateAndAdjustBalance(db, userID, -amount, true)
		case "toggle_status":
			db.Exec("UPDATE users SET is_active = NOT is_active WHERE id = $1", userID)
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
