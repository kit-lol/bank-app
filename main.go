package main

import (
	"bank-app/handlers"
	internal_config "bank-app/internal/config"
	"bank-app/logger"
	"bank-app/models"
	"bank-app/service"
	"bank-app/sessions"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func runMigrations(dbURL string) error {
	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		return fmt.Errorf("ошибка инициализации миграций: %w", err)
	}
	defer m.Close()

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("ошибка применения миграций: %w", err)
	}
	return nil
}

func convertDepositTypes(types []internal_config.DepositType) []models.DepositTypeExtended {
	var result []models.DepositTypeExtended
	for _, t := range types {
		result = append(result, models.DepositTypeExtended{
			ID: t.ID, Name: t.Name, Description: t.Description,
			MinAmount: t.MinAmount, InterestRate: t.InterestRate,
			CanDeposit: t.CanDeposit, CanWithdraw: t.CanWithdraw,
			Icon: t.Icon, Badge: t.Badge, Class: t.Class,
		})
	}
	return result
}

func main() {
	// 1. Инициализация логгера
	logger.InitLogger()
	defer logger.Log.Sync() // Сброс буфера при завершении

	logger.Log.Info("🚀 Запуск приложения Bank App...")

	// 2. Загрузка .env
	if err := godotenv.Load(); err != nil {
		logger.Log.Warn("Файл .env не найден, используем системные переменные")
	}

	// 3. Загрузка конфига
	cfg, err := internal_config.LoadConfig("internal/config/config.yaml")
	if err != nil {
		logger.Log.Fatal("Ошибка загрузки конфигурации", zap.Error(err))
	}
	logger.Log.Info("✅ Конфигурация загружена успешно")

	// 4. Миграции
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host,
		cfg.Database.Port, cfg.Database.Name, cfg.Database.SSLMode)

	logger.Log.Info("🔄 Применение миграций БД...")
	if err := runMigrations(dbURL); err != nil {
		logger.Log.Fatal("Ошибка миграций", zap.Error(err))
	}
	logger.Log.Info("✅ База данных готова к работе")

	// 5. Данные вкладов
	depositTypes := convertDepositTypes(cfg.DepositTypes)
	logger.Log.Info("📊 Типы вкладов загружены", zap.Int("count", len(depositTypes)))

	// 6. Сессии
	os.Setenv("SESSION_SECRET", cfg.Session.SecretKey)
	sessions.InitSession()
	logger.Log.Info("🔐 Сессии инициализированы")

	// 7. Подключение к БД
	db, err := sql.Open("postgres", cfg.GetDSN())
	if err != nil {
		logger.Log.Fatal("Ошибка подключения к БД", zap.Error(err))
	}
	defer db.Close()

	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)

	if err = db.Ping(); err != nil {
		logger.Log.Fatal("Не удалось пингануть БД", zap.Error(err))
	}
	logger.Log.Info("✅ Соединение с БД установлено")

	// 8. Фоновая задача
	go func() {
		ticker := time.NewTicker(cfg.Background.InterestAccrualInterval)
		defer ticker.Stop()
		for range ticker.C {
			logger.Log.Info("⏰ Запущено начисление процентов")
			service.AccrueInterest(db)
		}
	}()

	// 9. Маршруты
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/index.html")
	})
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/login", handlers.LoginHandler(db))
	http.HandleFunc("/register", handlers.RegisterHandler(db))
	http.HandleFunc("/logout", handlers.LogoutHandler)

	http.HandleFunc("/dashboard", handlers.AuthMiddleware(handlers.DashboardHandler(db)))
	http.HandleFunc("/deposit", handlers.AuthMiddleware(handlers.DepositHandler(db)))
	http.HandleFunc("/open-deposit", handlers.AuthMiddleware(handlers.OpenDepositPageHandler(db, depositTypes)))
	http.HandleFunc("/create-deposit", handlers.AuthMiddleware(handlers.CreateDepositHandler(db)))
	http.HandleFunc("/deposit-to-existing", handlers.AuthMiddleware(handlers.DepositToExistingHandler(db)))
	http.HandleFunc("/withdraw-from-deposit", handlers.AuthMiddleware(handlers.WithdrawHandler(db)))
	http.HandleFunc("/close-deposit", handlers.AuthMiddleware(handlers.CloseDepositHandler(db)))

	http.HandleFunc("/admin", handlers.AdminMiddleware(db, handlers.AdminDashboardHandler(db)))
	http.HandleFunc("/admin/action", handlers.AdminMiddleware(db, handlers.AdminActionHandler(db)))
	http.HandleFunc("/admin/user", handlers.AdminMiddleware(db, handlers.AdminUserDetailHandler(db)))
	http.HandleFunc("/admin/close-deposit", handlers.AdminMiddleware(db, handlers.CloseDepositHandler(db)))

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Log.Info("🌍 Сервер слушает порт", zap.String("address", addr))

	server := &http.Server{Addr: addr, ReadTimeout: cfg.Server.ReadTimeout, WriteTimeout: cfg.Server.WriteTimeout}
	logger.Log.Fatal("Ошибка сервера", zap.Error(server.ListenAndServe()))
}
