package service

import (
	"bank-app/logger"
	"database/sql"

	"go.uber.org/zap"
)

type DepositData struct {
	ID       int
	UserID   int
	Amount   float64
	Interest float64
}

func AccrueInterest(db *sql.DB) {
	logger.Log.Info("Начало процесса начисления процентов...")

	rows, err := db.Query(`
		SELECT id, user_id, amount, (amount * interest_rate / 100 / 365) as interest 
		FROM deposits 
		WHERE status = 'ACTIVE' 
		AND last_accrual < NOW() - INTERVAL '10 minute'`)

	if err != nil {
		logger.Log.Error("Ошибка запроса данных для начисления процентов", zap.Error(err))
		return
	}

	var toUpdate []DepositData
	for rows.Next() {
		var d DepositData
		rows.Scan(&d.ID, &d.UserID, &d.Amount, &d.Interest)
		toUpdate = append(toUpdate, d)
	}
	rows.Close()

	if len(toUpdate) == 0 {
		logger.Log.Info("Нет активных вкладов для начисления процентов")
		return
	}

	successCount := 0
	for _, d := range toUpdate {
		tx, err := db.Begin()
		if err != nil {
			continue
		}

		_, err = tx.Exec("UPDATE deposits SET amount = amount + $1, last_accrual = NOW() WHERE id = $2", d.Interest, d.ID)
		_, err = tx.Exec(`INSERT INTO transactions (user_id, deposit_id, amount, operation_type, created_at) VALUES ($1, $2, $3, 'ACCRUAL', NOW())`, d.UserID, d.ID, d.Interest)

		if err != nil {
			logger.Log.Error("Ошибка при обработке вклада", zap.Error(err), zap.Int("depositID", d.ID))
			tx.Rollback()
		} else {
			tx.Commit()
			successCount++
		}
	}

	logger.Log.Info("Процесс начисления завершен", zap.Int("processed", successCount), zap.Int("total", len(toUpdate)))
}
