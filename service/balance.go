package service

import (
	"database/sql"
	"fmt"
)

func ValidateAndAdjustBalance(db *sql.DB, userID int, amount float64, isAdmin bool) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %w", err)
	}
	defer tx.Rollback()

	var currentBalance float64
	err = tx.QueryRow("SELECT balance FROM users WHERE id = $1 FOR UPDATE", userID).Scan(&currentBalance)
	if err != nil {
		return err
	}

	if !isAdmin && (currentBalance+amount) < 0 {
		return fmt.Errorf("недостаточно средств")
	}

	_, err = tx.Exec("UPDATE users SET balance = balance + $1 WHERE id = $2", amount, userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
