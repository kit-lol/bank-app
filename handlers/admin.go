package handlers

import (
	"bank-app/models"
	"bank-app/repository"
	"bank-app/service"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

func AdminDashboardHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := repository.GetAllUser(db)
		if err != nil {
			http.Error(w, "Ошибка: "+err.Error(), 500)
			return
		}

		// ОТЛАДКА: Выведем в консоль, что именно мы передаем
		fmt.Printf("DEBUG: Отправляем в шаблон %d пользователей\n", len(users))
		for i, u := range users {
			fmt.Printf("User %d: ID=%d, Name=%s\n", i, u.ID, u.Username)
		}

		tmpl := template.Must(template.ParseFiles("templates/admin_dashboard.html"))
		err = tmpl.Execute(w, users) // Передаем слайс []models.User
		if err != nil {
			fmt.Println("Ошибка шаблона:", err)
		}
	}
}

func AdminActionHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDStr := r.FormValue("user_id")
		action := r.FormValue("action")
		amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)
		userID, _ := strconv.Atoi(userIDStr)

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

		// 1. Получаем все вклады (убрали WHERE status = 'ACTIVE', чтобы увидеть все)
		rows, err := db.Query("SELECT id, user_id, currency_id, amount, interest_rate, status, created_at FROM deposits WHERE user_id = $1", userID)
		if err != nil {
			fmt.Printf("Ошибка при поиске вкладов для пользователя %d: %v\n", userID, err)
			http.Error(w, "Ошибка БД", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var deposits []models.Deposit
		for rows.Next() {
			var d models.Deposit
			err := rows.Scan(&d.ID, &d.UserID, &d.CurrencyID, &d.Amount, &d.InterestRate, &d.Status, &d.CreatedAt)
			if err != nil {
				fmt.Printf("Ошибка сканирования вклада: %v\n", err)
				continue
			}
			deposits = append(deposits, d)
		}

		fmt.Printf("DEBUG: Для пользователя %d найдено %d вкладов\n", userID, len(deposits))

		// 2. Получаем транзакции
		transactions, _ := repository.GetUserTransactions(db, userID)

		// 3. Формируем данные для шаблона
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
