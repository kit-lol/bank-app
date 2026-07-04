package handlers

import (
	"bank-app/logger"
	"bank-app/models"
	"bank-app/repository"
	"bank-app/sessions"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"go.uber.org/zap"
)

// --- ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ДЛЯ СООБЩЕНИЙ ---

func setFlash(w http.ResponseWriter, r *http.Request, message string, msgType string) {
	session, _ := sessions.Store.Get(r, "session")
	session.AddFlash(map[string]string{"message": message, "type": msgType})
	session.Save(r, w)
}

func getFlashes(w http.ResponseWriter, r *http.Request) []map[string]string {
	session, _ := sessions.Store.Get(r, "session")
	var flashes []map[string]string
	for _, f := range session.Flashes() {
		if flash, ok := f.(map[string]string); ok {
			flashes = append(flashes, flash)
		}
	}
	session.Save(r, w) // Очищаем флеш после чтения
	return flashes
}

// --- ОСНОВНЫЕ ХЕНДЛЕРЫ ---

func DepositHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		userIDStr := r.Header.Get("X-User-ID")
		userID, _ := strconv.Atoi(userIDStr)

		amountStr := r.FormValue("amount")
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil || amount <= 0 {
			setFlash(w, r, "Некорректная сумма пополнения", "error")
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		logger.Log.Info("Пополнение баланса", zap.Int("userID", userID), zap.Float64("amount", amount))

		err = repository.UpdateBalance(db, userID, amount)
		if err != nil {
			logger.Log.Error("Ошибка обновления баланса", zap.Error(err))
			setFlash(w, r, "Ошибка при пополнении баланса", "error")
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		err = repository.AddTransaction(db, userID, nil, amount, "DEPOSIT")
		if err != nil {
			logger.Log.Error("Ошибка записи транзакции", zap.Error(err))
		}

		setFlash(w, r, fmt.Sprintf("Баланс успешно пополнен на %.2f ₽", amount), "success")
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

func OpenDepositPageHandler(db *sql.DB, depositTypes []models.DepositTypeExtended) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl := template.Must(template.ParseFiles("templates/open_deposit.html"))
		tmpl.Execute(w, depositTypes)
	}
}

func CreateDepositHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
			return
		}

		userIDStr := r.Header.Get("X-User-ID")
		userID, _ := strconv.Atoi(userIDStr)

		typeID, _ := strconv.Atoi(r.FormValue("type_id"))
		amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)

		if err != nil || amount <= 0 {
			setFlash(w, r, "Некорректная сумма вклада", "error")
			http.Redirect(w, r, "/open-deposit", http.StatusSeeOther)
			return
		}

		if amount > 10000000 {
			setFlash(w, r, "Сумма вклада превышает лимит (10 млн)", "error")
			http.Redirect(w, r, "/open-deposit", http.StatusSeeOther)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			setFlash(w, r, "Ошибка сервера при создании вклада", "error")
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
		defer tx.Rollback()

		var balance float64
		var minAmount float64
		var interestRate float64

		err = tx.QueryRow("SELECT balance FROM users WHERE id = $1 FOR UPDATE", userID).Scan(&balance)
		if err != nil {
			setFlash(w, r, "Ошибка получения данных пользователя", "error")
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		err = tx.QueryRow("SELECT min_amount, interest_rate FROM deposit_types WHERE id = $1", typeID).Scan(&minAmount, &interestRate)
		if err != nil {
			setFlash(w, r, "Ошибка данных о типе вклада", "error")
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		if amount < minAmount {
			setFlash(w, r, fmt.Sprintf("Минимальная сумма для этого вклада: %.2f ₽", minAmount), "error")
			http.Redirect(w, r, "/open-deposit", http.StatusSeeOther)
			return
		}
		if balance < amount {
			setFlash(w, r, "Недостаточно средств на балансе", "error")
			http.Redirect(w, r, "/open-deposit", http.StatusSeeOther)
			return
		}

		_, err = tx.Exec("UPDATE users SET balance = balance - $1 WHERE id = $2", amount, userID)
		_, err = tx.Exec(`INSERT INTO deposits (user_id, type_id, amount, interest_rate, status, created_at) VALUES ($1, $2, $3, $4, 'ACTIVE', NOW())`,
			userID, typeID, amount, interestRate)

		if err != nil {
			setFlash(w, r, "Ошибка создания вклада в БД", "error")
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		tx.Commit()
		setFlash(w, r, "Вклад успешно открыт!", "success")
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

func DepositToExistingHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDStr := r.Header.Get("X-User-ID")
		userID, _ := strconv.Atoi(userIDStr)

		depositID := r.FormValue("deposit_id")
		amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)

		tx, _ := db.Begin()
		defer tx.Rollback()

		_, err := tx.Exec("UPDATE users SET balance = balance - $1 WHERE id = $2 AND balance >= $1", amount, userID)
		if err != nil {
			setFlash(w, r, "Недостаточно средств на балансе для пополнения", "error")
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		_, err = tx.Exec("UPDATE deposits SET amount = amount + $1 WHERE id = $2 AND user_id = $3", amount, depositID, userID)
		if err != nil {
			setFlash(w, r, "Ошибка пополнения вклада", "error")
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		_, err = tx.Exec(`INSERT INTO transactions (user_id, deposit_id, amount, operation_type, created_at) VALUES ($1, $2, $3, 'DEPOSIT_TO_EXISTING', NOW())`, userID, depositID, amount)

		tx.Commit()
		setFlash(w, r, fmt.Sprintf("Вклад успешно пополнен на %.2f ₽", amount), "success")
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

func WithdrawHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDStr := r.Header.Get("X-User-ID")
		userID, _ := strconv.Atoi(userIDStr)

		depositID, _ := strconv.Atoi(r.FormValue("deposit_id"))
		amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)

		err := repository.WithdrawFromDeposit(db, userID, strconv.Itoa(depositID), amount)
		if err != nil {
			setFlash(w, r, err.Error(), "error")
		} else {
			setFlash(w, r, fmt.Sprintf("Со вклада успешно снято %.2f ₽", amount), "success")
		}

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

func CloseDepositHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		depositID := r.FormValue("deposit_id")
		userIDStr := r.Header.Get("X-User-ID")
		if userIDStr == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		userID, _ := strconv.Atoi(userIDStr)

		tx, err := db.Begin()
		if err != nil {
			setFlash(w, r, "Ошибка сервера", "error")
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
		defer tx.Rollback()

		var amount float64
		err = tx.QueryRow("SELECT amount FROM deposits WHERE id = $1 AND user_id = $2", depositID, userID).Scan(&amount)
		if err != nil {
			setFlash(w, r, "Вклад не найден", "error")
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		_, err = tx.Exec("UPDATE users SET balance = balance + $1 WHERE id = $2", amount, userID)
		if err != nil {
			setFlash(w, r, "Ошибка возврата средств", "error")
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		_, err = tx.Exec(`INSERT INTO transactions (user_id, deposit_id, amount, operation_type, created_at) 
                          VALUES ($1, $2, $3, 'CLOSE_DEPOSIT', NOW())`, userID, depositID, amount)

		_, err = tx.Exec("DELETE FROM deposits WHERE id = $1", depositID)

		if err != nil {
			setFlash(w, r, "Ошибка удаления вклада", "error")
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		tx.Commit()
		setFlash(w, r, "Вклад закрыт, средства возвращены на баланс", "success")
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}
