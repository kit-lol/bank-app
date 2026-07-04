package repository

import (
	"bank-app/models"
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func RegisterUser(db *sql.DB, username, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return err
	}

	query := `INSERT INTO users (username, password_hash) VALUES ($1, $2)`
	_, err = db.Exec(query, username, string(hashedPassword))
	return err
}

func GetUserByUsername(db *sql.DB, username string) (*models.User, error) {
	query := `SELECT id, username, password_hash, role, is_active FROM users WHERE username = $1`
	row := db.QueryRow(query, username)

	var u models.User
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.IsActive)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func GetUserByID(db *sql.DB, id int) (*models.User, error) {
	query := `SELECT id, username, balance, role FROM users WHERE id = $1`
	var u models.User
	err := db.QueryRow(query, id).Scan(&u.ID, &u.Username, &u.Balance, &u.Role)
	return &u, err
}

func AddTransaction(db *sql.DB, userID int, depositID *int, amount float64, opType string) error {
	query := `INSERT INTO transactions (user_id, deposit_id, amount, operation_type) VALUES ($1, $2, $3, $4)`
	_, err := db.Exec(query, userID, depositID, amount, opType)
	return err
}

func GetUserTransactions(db *sql.DB, userID int) ([]models.Transaction, error) {
	query := `SELECT id, amount, operation_type, created_at FROM transactions WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var t models.Transaction
		err := rows.Scan(&t.ID, &t.Amount, &t.OperationType, &t.CreatedAt)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}
	return transactions, nil
}

func UpdateBalance(db *sql.DB, userID int, amount float64) error {
	query := `UPDATE users SET balance = balance + $1 WHERE id = $2`
	res, err := db.Exec(query, amount, userID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("пользователь не найден")
	}
	return nil
}

func GetAllUser(db *sql.DB) ([]models.User, error) {
	rows, err := db.Query("SELECT id, username, balance, is_active FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		rows.Scan(&u.ID, &u.Username, &u.Balance, &u.IsActive)
		if err != nil {
			fmt.Println("Ошибка сканирования:", err)
			continue
		}
		users = append(users, u)
	}
	return users, nil
}

func SetUserStatus(db *sql.DB, userID int, active bool) error {
	_, err := db.Exec("UPDATE users SET is_active = $1 WHERE id = $2", active, userID)
	return err
}

func AdjustBalance(db *sql.DB, userID int, amount float64) error {
	_, err := db.Exec("UPDATE users SET balance = balance + $1 WHERE id = $2", amount, userID)
	return err
}
