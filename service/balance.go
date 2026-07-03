package service

import (
	"database/sql"
	"fmt"
)

func ValidateAndAdjustBalance(db *sql.DB, userID int, amount float64, isAdmin bool) error {
	var currentBalance float64
	err := db.QueryRow("SELECT balance FROM users WHERE id = $1", userID).Scan(&currentBalance)
	if err != nil {
		return err
	}

	if !isAdmin && (currentBalance+amount) < 0 {
		return fmt.Errorf("недостаточно средств")
	}

	_, err = db.Exec("UPDATE users SET balance = balance + $1 WHERE id = $2", amount, userID)
	return err
}
