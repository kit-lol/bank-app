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
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// runMigrations применяет SQL-миграции для создания структуры БД
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

// convertDepositTypes преобразует конфиг YAML в модели для использования в handlers
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
	defer logger.Log.Sync()

	logger.Log.Info("🚀 Запуск приложения Bank App...")

	// 2. Загрузка переменных окружения (.env)
	if err := godotenv.Load(); err != nil {
		logger.Log.Warn("Файл .env не найден, используем системные переменные")
	}

	// 3. Загрузка конфигурации из YAML
	cfg, err := internal_config.LoadConfig("internal/config/config.yaml")
	if err != nil {
		logger.Log.Fatal("Ошибка загрузки конфигурации", zap.Error(err))
	}
	logger.Log.Info("✅ Конфигурация загружена успешно")

	// 4. Применение миграций базы данных
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host,
		cfg.Database.Port, cfg.Database.Name, cfg.Database.SSLMode)

	logger.Log.Info("🔄 Применение миграций БД...")
	if err := runMigrations(dbURL); err != nil {
		logger.Log.Fatal("Ошибка миграций", zap.Error(err))
	}
	logger.Log.Info("✅ База данных готова к работе")

	// 5. Подготовка данных о типах вкладов из конфига
	depositTypes := convertDepositTypes(cfg.DepositTypes)
	logger.Log.Info("📊 Типы вкладов загружены", zap.Int("count", len(depositTypes)))

	// 6. Инициализация хранилища сессий (Cookie Store)
	os.Setenv("SESSION_SECRET", cfg.Session.SecretKey)
	sessions.InitSession()
	logger.Log.Info("🔐 Сессии инициализированы")

	// 7. Подключение к PostgreSQL
	db, err := sql.Open("postgres", cfg.GetDSN())
	if err != nil {
		logger.Log.Fatal("Ошибка подключения к БД", zap.Error(err))
	}
	defer db.Close()

	// Настройка пула соединений
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)

	if err = db.Ping(); err != nil {
		logger.Log.Fatal("Не удалось пингануть БД", zap.Error(err))
	}
	logger.Log.Info("✅ Соединение с БД установлено")

	// 8. Запуск фоновой задачи для начисления процентов по вкладам
	go func() {
		ticker := time.NewTicker(cfg.Background.InterestAccrualInterval)
		defer ticker.Stop()
		for range ticker.C {
			logger.Log.Info("⏰ Запущено начисление процентов")
			service.AccrueInterest(db)
		}
	}()

	// 9. Настройка маршрутизации через ServeMux
	mux := http.NewServeMux()

	// Главная страница
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "templates/index.html")
		} else {
			http.NotFound(w, r)
		}
	})

	// Раздача статических файлов с принудительными MIME-типами
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".css") {
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		} else if strings.HasSuffix(r.URL.Path, ".js") {
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		}
		fs.ServeHTTP(w, r)
	})))

	// Маршруты аутентификации
	mux.HandleFunc("/login", handlers.LoginHandler(db))
	mux.HandleFunc("/register", handlers.RegisterHandler(db))
	mux.HandleFunc("/logout", handlers.LogoutHandler)

	// Защищенные маршруты пользователей
	mux.HandleFunc("/dashboard", handlers.AuthMiddleware(handlers.DashboardHandler(db)))
	mux.HandleFunc("/deposit", handlers.AuthMiddleware(handlers.DepositHandler(db)))
	mux.HandleFunc("/open-deposit", handlers.AuthMiddleware(handlers.OpenDepositPageHandler(db, depositTypes)))
	mux.HandleFunc("/create-deposit", handlers.AuthMiddleware(handlers.CreateDepositHandler(db)))
	mux.HandleFunc("/deposit-to-existing", handlers.AuthMiddleware(handlers.DepositToExistingHandler(db)))
	mux.HandleFunc("/withdraw-from-deposit", handlers.AuthMiddleware(handlers.WithdrawHandler(db)))
	mux.HandleFunc("/close-deposit", handlers.AuthMiddleware(handlers.CloseDepositHandler(db)))

	// Защищенные маршруты администратора
	mux.HandleFunc("/admin", handlers.AdminMiddleware(db, handlers.AdminDashboardHandler(db)))
	mux.HandleFunc("/admin/action", handlers.AdminMiddleware(db, handlers.AdminActionHandler(db)))
	mux.HandleFunc("/admin/user", handlers.AdminMiddleware(db, handlers.AdminUserDetailHandler(db)))
	mux.HandleFunc("/admin/close-deposit", handlers.AdminMiddleware(db, handlers.CloseDepositHandler(db)))

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Log.Info("🌍 Сервер слушает порт", zap.String("address", addr))

	// Security headers middleware
	secureHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		mux.ServeHTTP(w, r)
	})

	server := &http.Server{
		Addr:         addr,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		Handler:      secureHandler,
	}

	logger.Log.Fatal("Ошибка сервера", zap.Error(server.ListenAndServe()))
}
