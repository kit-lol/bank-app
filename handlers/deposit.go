package handlers

import (
	"bank-app/logger"
	"bank-app/models"
	"bank-app/repository"
	"database/sql"
	"html/template"
	"net/http"
	"strconv"

	"go.uber.org/zap"
)

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
			logger.Log.Warn("Некорректная сумма пополнения", zap.String("amount", amountStr))
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		logger.Log.Info("Пополнение баланса", zap.Int("userID", userID), zap.Float64("amount", amount))

		err = repository.UpdateBalance(db, userID, amount)
		if err != nil {
			logger.Log.Error("Ошибка обновления баланса", zap.Error(err), zap.Int("userID", userID))
			http.Error(w, "Ошибка БД", http.StatusInternalServerError)
			return
		}

		err = repository.AddTransaction(db, userID, nil, amount, "DEPOSIT")
		if err != nil {
			logger.Log.Error("Ошибка записи транзакции", zap.Error(err))
			http.Error(w, "Ошибка записи истории", http.StatusInternalServerError)
			return
		}

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

		logger.Log.Info("Запрос на открытие вклада",
			zap.Int("userID", userID),
			zap.Float64("amount", amount),
			zap.Int("typeID", typeID))

		if err != nil || amount <= 0 {
			http.Error(w, "Некорректная сумма", http.StatusBadRequest)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			logger.Log.Error("Не удалось начать транзакцию", zap.Error(err))
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		var balance float64
		var minAmount float64
		var interestRate float64

		err = tx.QueryRow("SELECT balance FROM users WHERE id = $1 FOR UPDATE", userID).Scan(&balance)
		err = tx.QueryRow("SELECT min_amount, interest_rate FROM deposit_types WHERE id = $1", typeID).Scan(&minAmount, &interestRate)

		if err != nil {
			logger.Log.Error("Ошибка получения данных для вклада", zap.Error(err))
			http.Error(w, "Ошибка оформления", http.StatusBadRequest)
			return
		}

		if amount < minAmount {
			logger.Log.Warn("Сумма меньше минимальной", zap.Float64("amount", amount), zap.Float64("min", minAmount))
			http.Error(w, "Сумма меньше минимальной", http.StatusBadRequest)
			return
		}
		if balance < amount {
			logger.Log.Warn("Недостаточно средств для вклада", zap.Float64("balance", balance), zap.Float64("amount", amount))
			http.Error(w, "Недостаточно средств на балансе", http.StatusBadRequest)
			return
		}

		_, err = tx.Exec("UPDATE users SET balance = balance - $1 WHERE id = $2", amount, userID)
		_, err = tx.Exec(`INSERT INTO deposits (user_id, type_id, amount, interest_rate, status, created_at) VALUES ($1, $2, $3, $4, 'ACTIVE', NOW())`,
			userID, typeID, amount, interestRate)

		if err != nil {
			logger.Log.Error("Ошибка создания вклада в БД", zap.Error(err))
			http.Error(w, "Ошибка оформления", http.StatusInternalServerError)
			return
		}

		tx.Commit()
		logger.Log.Info("Вклад успешно открыт", zap.Int("userID", userID), zap.Float64("newBalance", balance-amount))
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

func DepositToExistingHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDStr := r.Header.Get("X-User-ID")
		userID, _ := strconv.Atoi(userIDStr)

		depositID := r.FormValue("deposit_id")
		amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)

		logger.Log.Info("Пополнение существующего вклада", zap.Int("userID", userID), zap.String("depositID", depositID), zap.Float64("amount", amount))

		tx, _ := db.Begin()
		defer tx.Rollback()

		_, err := tx.Exec("UPDATE users SET balance = balance - $1 WHERE id = $2 AND balance >= $1", amount, userID)
		if err != nil {
			logger.Log.Warn("Недостаточно средств для пополнения вклада", zap.Error(err))
			http.Error(w, "Недостаточно средств", http.StatusBadRequest)
			return
		}

		_, err = tx.Exec("UPDATE deposits SET amount = amount + $1 WHERE id = $2 AND user_id = $3", amount, depositID, userID)
		_, err = tx.Exec(`INSERT INTO transactions (user_id, deposit_id, amount, operation_type, created_at) VALUES ($1, $2, $3, 'DEPOSIT_TO_EXISTING', NOW())`, userID, depositID, amount)

		if err != nil {
			logger.Log.Error("Ошибка при пополнении вклада", zap.Error(err))
			http.Error(w, "Ошибка операции", http.StatusInternalServerError)
			return
		}

		tx.Commit()
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

func WithdrawHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDStr := r.Header.Get("X-User-ID")
		userID, _ := strconv.Atoi(userIDStr)

		depositID, _ := strconv.Atoi(r.FormValue("deposit_id"))
		amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)

		logger.Log.Info("Запрос на снятие средств со вклада", zap.Int("userID", userID), zap.Int("depositID", depositID), zap.Float64("amount", amount))

		err := repository.WithdrawFromDeposit(db, userID, strconv.Itoa(depositID), amount)
		if err != nil {
			logger.Log.Warn("Ошибка снятия средств", zap.Error(err))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

func CloseDepositHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		depositID := r.FormValue("deposit_id")

		// Получаем UserID из сессии (через header, установленный middleware)
		userIDStr := r.Header.Get("X-User-ID")
		if userIDStr == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		userID, _ := strconv.Atoi(userIDStr)

		logger.Log.Info("Запрос на закрытие вклада", zap.String("depositID", depositID), zap.Int("userID", userID))

		// Начинаем транзакцию, чтобы всё прошло безопасно
		tx, err := db.Begin()
		if err != nil {
			logger.Log.Error("Не удалось начать транзакцию при закрытии вклада", zap.Error(err))
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// 1. Узнаем сумму вклада перед удалением
		var amount float64
		err = tx.QueryRow("SELECT amount FROM deposits WHERE id = $1 AND user_id = $2", depositID, userID).Scan(&amount)
		if err != nil {
			logger.Log.Error("Вклад не найден или ошибка БД", zap.Error(err))
			http.Error(w, "Вклад не найден", http.StatusBadRequest)
			return
		}

		// 2. Возвращаем деньги на баланс пользователя
		_, err = tx.Exec("UPDATE users SET balance = balance + $1 WHERE id = $2", amount, userID)
		if err != nil {
			logger.Log.Error("Ошибка возврата средств на баланс", zap.Error(err))
			http.Error(w, "Ошибка возврата средств", http.StatusInternalServerError)
			return
		}

		// 3. Записываем операцию в историю транзакций
		_, err = tx.Exec(`INSERT INTO transactions (user_id, deposit_id, amount, operation_type, created_at) 
                          VALUES ($1, $2, $3, 'CLOSE_DEPOSIT', NOW())`, userID, depositID, amount)

		// 4. Удаляем вклад (или меняем статус на CLOSED, если хочешь хранить историю)
		_, err = tx.Exec("DELETE FROM deposits WHERE id = $1", depositID)

		if err != nil {
			logger.Log.Error("Ошибка удаления вклада", zap.Error(err))
			http.Error(w, "Ошибка удаления вклада", http.StatusInternalServerError)
			return
		}

		// Если всё успешно — сохраняем изменения
		tx.Commit()

		logger.Log.Info("Вклад успешно закрыт, средства возвращены", zap.Float64("amount", amount))

		if r.Header.Get("Referer") != "" {
			http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		}
	}
}
