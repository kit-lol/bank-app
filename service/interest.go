package service

import (
	"database/sql"
	"fmt"
)

// Структура для хранения данных о вкладе перед обновлением
type DepositData struct {
	ID       int
	UserID   int
	Amount   float64
	Interest float64
}

func AccrueInterest(db *sql.DB) {
	// 1. Выбираем данные в память (отдельный запрос)
	rows, err := db.Query(`
		SELECT id, user_id, amount, (amount * interest_rate / 100 / 365) as interest 
		FROM deposits 
		WHERE status = 'ACTIVE' 
		AND last_accrual < NOW() - INTERVAL '10 minute'`)

	if err != nil {
		fmt.Printf("Ошибка запроса данных: %v\n", err)
		return
	}

	var toUpdate []DepositData
	for rows.Next() {
		var d DepositData
		rows.Scan(&d.ID, &d.UserID, &d.Amount, &d.Interest)
		toUpdate = append(toUpdate, d)
	}
	rows.Close() // ЗАКРЫВАЕМ rows ДО начала обновлений

	// 2. Теперь выполняем обновления
	if len(toUpdate) == 0 {
		return
	}

	for _, d := range toUpdate {
		tx, err := db.Begin()
		if err != nil {
			continue
		}

		// Обновляем вклад
		_, err = tx.Exec("UPDATE deposits SET amount = amount + $1, last_accrual = NOW() WHERE id = $2", d.Interest, d.ID)

		// Записываем в историю
		_, err = tx.Exec(`
			INSERT INTO transactions (user_id, deposit_id, amount, operation_type, created_at) 
			VALUES ($1, $2, $3, 'ACCRUAL', NOW())`, d.UserID, d.ID, d.Interest)

		if err != nil {
			fmt.Printf("Ошибка при обработке вклада %d: %v\n", d.ID, err)
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}

	fmt.Printf("Успешно начислено процентов по %d вкладам.\n", len(toUpdate))
}
