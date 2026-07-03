package main

import (
	"bank-app/config"
	"bank-app/handlers"
	"bank-app/models"
	"bank-app/service"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var depositTypes []models.DepositTypeExtended

func main() {
	depositTypes, err := config.LoadDepositConfig("config/deposits.json")
	if err != nil {
		log.Fatal("Ошибка загрузки конфига вкладов:", err)
	}

	log.Printf("✅ Загружено %d типов вкладов", len(depositTypes))

	// 1. Загрузка конфигурации
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используем переменные окружения")
	}

	// 2. Подключение к БД
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_NAME"))

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 3. Запуск начисления процентов в фоновом режиме
	go func() {
		fmt.Println("Фоновая задача запущена")
		ticker := time.NewTicker(10 * time.Minute)
		for range ticker.C {
			service.AccrueInterest(db)
		}
	}()

	// 4. Маршрутизация (Routes)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/index.html")
	})

	// Раздача статики (стили и js)
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Пользовательские маршруты
	http.HandleFunc("/login", handlers.LoginHandler(db))
	http.HandleFunc("/register", handlers.RegisterHandler(db))
	http.HandleFunc("/logout", handlers.LogoutHandler)
	http.HandleFunc("/dashboard", handlers.AuthMiddleware(handlers.DashboardHandler(db)))
	http.HandleFunc("/deposit", handlers.AuthMiddleware(handlers.DepositHandler(db)))

	// Вклады
	http.HandleFunc("/open-deposit", handlers.AuthMiddleware(handlers.OpenDepositPageHandler(db)))
	http.HandleFunc("/create-deposit", handlers.AuthMiddleware(handlers.CreateDepositHandler(db)))
	http.HandleFunc("/deposit-to-existing", handlers.AuthMiddleware(handlers.DepositToExistingHandler(db)))
	http.HandleFunc("/withdraw-from-deposit", handlers.AuthMiddleware(handlers.WithdrawHandler(db)))
	http.HandleFunc("/close-deposit", handlers.AuthMiddleware(handlers.CloseDepositHandler(db)))

	// Админские маршруты
	http.HandleFunc("/admin", handlers.AdminMiddleware(db, handlers.AdminDashboardHandler(db)))
	http.HandleFunc("/admin/action", handlers.AdminMiddleware(db, handlers.AdminActionHandler(db)))
	http.HandleFunc("/admin/user", handlers.AdminMiddleware(db, handlers.AdminUserDetailHandler(db)))
	http.HandleFunc("/admin/close-deposit", handlers.AdminMiddleware(db, handlers.CloseDepositHandler(db)))

	log.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
