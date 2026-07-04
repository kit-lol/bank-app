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
		userIDStr := r.Header.Get("X-User-ID")
		if userIDStr == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "Ошибка авторизации", http.StatusUnauthorized)
			return
		}

		user, _ := repository.GetUserByID(db, userID)

		transactions, err := repository.GetUserTransactions(db, userID)
		if err == nil {
			for i := range transactions {
				transactions[i].OperationType = utils.TranslateOperationType(transactions[i].OperationType)
			}
		} else {
			transactions = []models.Transaction{} // Чтобы шаблон не падал
		}

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
				err := rows.Scan(&d.ID, &typeID, &d.Amount, &d.Rate, &d.CanDeposit, &d.CanWithdraw)
				if err != nil {
					fmt.Println("Ошибка сканирования вклада:", err)
					continue
				}
				d.TypeName = utils.GetDepositName(typeID)
				deposits = append(deposits, d)
			}
		} else {
			deposits = []DepositView{}
		}

		// Получаем сообщения
		flashMessages := getFlashes(w, r)

		data := struct {
			User         *models.User
			Transactions []models.Transaction
			Deposits     []DepositView
			Flashes      []map[string]string
		}{
			User:         user,
			Transactions: transactions,
			Deposits:     deposits,
			Flashes:      flashMessages,
		}

		tmpl := template.Must(template.ParseFiles("templates/dashboard.html"))
		tmpl.Execute(w, data)
	}
}
