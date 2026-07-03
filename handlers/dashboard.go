package handlers

import (
	"bank-app/models"
	"bank-app/repository"
	"bank-app/utils"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

func DashboardHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		userID, _ := strconv.Atoi(cookie.Value)

		user, _ := repository.GetUserByID(db, userID)

		transactions, err := repository.GetUserTransactions(db, userID)
		if err == nil {
			for i := range transactions {
				transactions[i].OperationType = utils.TranslateOperationType(transactions[i].OperationType)
			}
		}

		// 3. Получаем активные вклады
		rows, err := db.Query(`
		SELECT d.id, d.type_id, d.amount, d.interest_rate, t.can_deposit, t.can_withdraw 
		FROM deposits d 
		JOIN deposit_types t ON d.type_id = t.id 
		WHERE d.user_id = $1 AND d.status = 'ACTIVE'`, userID)

		type DepositView struct {
			ID          int
			TypeName    string
			Amount      float64
			Rate        float64
			CanDeposit  bool
			CanWithdraw bool
		}

		var deposits []DepositView

		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var d DepositView
				var typeID int
				// Сканируем вместе с правами из таблицы типов
				err := rows.Scan(&d.ID, &typeID, &d.Amount, &d.Rate, &d.CanDeposit, &d.CanWithdraw)
				if err != nil {
					fmt.Println("Ошибка сканирования вклада:", err)
					continue
				}

				// Название берем из utils, чтобы избежать проблем с кодировкой
				d.TypeName = utils.GetDepositName(typeID)

				deposits = append(deposits, d)
			}
		}

		data := struct {
			User         *models.User
			Transactions []models.Transaction
			Deposits     []DepositView
		}{user, transactions, deposits}

		tmpl := template.Must(template.ParseFiles("templates/dashboard.html"))
		tmpl.Execute(w, data)
	}
}
