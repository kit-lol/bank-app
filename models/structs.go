package models

import (
	"encoding/json"
	"time"
)

type User struct {
	ID           int     `json:"id"`
	Username     string  `json:"username"`
	PasswordHash string  `json:"password_hash"`
	Balance      float64 `json:"balance"`
	Role         string  `json:"role"`
	IsActive     bool    `json:"is_active"`
}

type Deposit struct {
	ID           int             `json:"id"`
	UserID       int             `json:"user_id"`
	Amount       float64         `json:"amount"`
	InterestRate float64         `json:"interest_rate"`
	Conditions   json.RawMessage `json:"conditions"`
	Status       string          `json:"status"`
	CreatedAt    time.Time       `json:"created_at"`
	LastAccrual  time.Time       `json:"last_accrual"`
}

type DepositType struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	InterestRate float64 `json:"interest_rate"`
	MinAmount    float64 `json:"min_amount"`
	CanWithdraw  bool    `json:"can_withdraw"`
	CanDeposit   bool    `json:"can_deposit"`
	Description  string  `json:"description"`
}

type DepositConfig struct {
	Deposits []DepositTypeExtended `json:"deposits"`
}

type DepositTypeExtended struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	InterestRate float64 `json:"interest_rate"`
	MinAmount    float64 `json:"min_amount"`
	CanDeposit   bool    `json:"can_deposit"`
	CanWithdraw  bool    `json:"can_withdraw"`
	Class        string  `json:"class"`
	Icon         string  `json:"icon"`
	Badge        string  `json:"badge"`
}

type Transaction struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	DepositID     *int      `json:"deposit_id"`
	Amount        float64   `json:"amount"`
	OperationType string    `json:"operation_type"`
	CreatedAt     time.Time `json:"created_at"`
}
