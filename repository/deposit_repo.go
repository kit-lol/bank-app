package repository

import (
	"bank-app/models"
	"database/sql"
	"fmt"
)

func GetDepositTypes(db *sql.DB) ([]models.DepositType, error) {
	query := `SELECT id, name, interest_rate, min_amount, can_withdraw, can_deposit, description FROM deposit_types`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []models.DepositType
	for rows.Next() {
		var t models.DepositType
		rows.Scan(&t.ID, &t.Name, &t.InterestRate, &t.MinAmount, &t.CanWithdraw, &t.CanDeposit, &t.Description)
		types = append(types, t)
	}
	return types, nil
}

func WithdrawFromDeposit(db *sql.DB, userID int, depositID string, amount float64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var currentAmount float64
	err = tx.QueryRow("SELECT amount FROM deposits WHERE id = $1 AND user_id = $2 FOR UPDATE", depositID, userID).Scan(&currentAmount)
	if err != nil || currentAmount < amount {
		return fmt.Errorf("недостаточно средств")
	}

	// 2. Списываем со вклада и возвращаем на баланс юзера
	_, err = tx.Exec("UPDATE deposits SET amount = amount - $1 WHERE id = $2", amount, depositID)
	_, err = tx.Exec("UPDATE users SET balance = balance + $1 WHERE id = $2", amount, userID)

	// 3. Логируем
	_, err = tx.Exec("INSERT INTO transactions (user_id, deposit_id, amount, operation_type, created_at) VALUES ($1, $2, $3, 'WITHDRAW', NOW())", userID, depositID, amount)

	return tx.Commit()
}

// Закрытие вклада (перенос остатка на баланс)
func CloseDeposit(db *sql.DB, userID int, depositID string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var amount float64
	err = tx.QueryRow("SELECT amount FROM deposits WHERE id = $1 AND user_id = $2 AND status = 'ACTIVE' FOR UPDATE", depositID, userID).Scan(&amount)
	if err != nil {
		return err
	}

	// Переносим деньги и меняем статус
	_, err = tx.Exec("UPDATE users SET balance = balance + $1 WHERE id = $2", amount, userID)
	_, err = tx.Exec("UPDATE deposits SET status = 'CLOSED' WHERE id = $1", depositID)
	_, err = tx.Exec("INSERT INTO transactions (user_id, deposit_id, amount, operation_type, created_at) VALUES ($1, $2, $3, 'CLOSE_DEPOSIT', NOW())", userID, depositID, amount)

	return tx.Commit()
}
