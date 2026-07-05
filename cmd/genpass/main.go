package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Генерация надёжного пароля для БД (24 байта = 32 символа base64)
	dbPassBytes := make([]byte, 24)
	rand.Read(dbPassBytes)
	dbPass := base64.URLEncoding.EncodeToString(dbPassBytes)
	fmt.Println("DB_PASS:", dbPass)

	// Генерация надёжного SESSION_SECRET (32 байта = 44 символа base64)
	sessionBytes := make([]byte, 32)
	rand.Read(sessionBytes)
	sessionSecret := base64.URLEncoding.EncodeToString(sessionBytes)
	fmt.Println("SESSION_SECRET:", sessionSecret)

	// Генерация bcrypt-хеша для нового пароля admin
	newAdminPass := "BankAdmin$2026!Secure"
	hash, _ := bcrypt.GenerateFromPassword([]byte(newAdminPass), 14)
	fmt.Println("ADMIN_PASSWORD:", newAdminPass)
	fmt.Println("ADMIN_HASH:", string(hash))
}
