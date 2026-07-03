package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config - основная структура конфигурации
type Config struct {
	Server       ServerConfig     `yaml:"server"`
	Database     DatabaseConfig   `yaml:"database"`
	Session      SessionConfig    `yaml:"session"`
	Background   BackgroundConfig `yaml:"background"`
	DepositTypes []DepositType    `yaml:"deposit_types"`
	Logging      LoggingConfig    `yaml:"logging"`
}

// ServerConfig - настройки сервера
type ServerConfig struct {
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

// DatabaseConfig - настройки базы данных
type DatabaseConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	Name         string `yaml:"name"`
	SSLMode      string `yaml:"ssl_mode"`
	MaxOpenConns int    `yaml:"max_open_conns"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
}

// SessionConfig - настройки сессий
type SessionConfig struct {
	SecretKey string `yaml:"secret_key"`
	MaxAge    int    `yaml:"max_age"`
}

// BackgroundConfig - настройки фоновых задач
type BackgroundConfig struct {
	InterestAccrualInterval time.Duration `yaml:"interest_accrual_interval"`
}

// DepositType - тип вклада (НОВОЕ)
type DepositType struct {
	ID           int     `yaml:"id"`
	Name         string  `yaml:"name"`
	Description  string  `yaml:"description"`
	MinAmount    float64 `yaml:"min_amount"`
	InterestRate float64 `yaml:"interest_rate"`
	CanDeposit   bool    `yaml:"can_deposit"`
	CanWithdraw  bool    `yaml:"can_withdraw"`
	Icon         string  `yaml:"icon"`
	Badge        string  `yaml:"badge"`
	Class        string  `yaml:"class"`
}

// LoggingConfig - настройки логирования
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// LoadConfig загружает конфигурацию из YAML файла
func LoadConfig(path string) (*Config, error) {
	// Читаем файл
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла конфигурации: %w", err)
	}

	// Парсим YAML
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга YAML: %w", err)
	}

	// Заменяем переменные окружения в строках
	replaceEnvVars(&config)

	// Устанавливаем значения по умолчанию, если они не заданы
	setDefaults(&config)

	return &config, nil
}

// replaceEnvVars заменяет ${VAR_NAME} на значения из переменных окружения
func replaceEnvVars(config *Config) {
	// База данных
	config.Database.User = os.ExpandEnv(config.Database.User)
	config.Database.Password = os.ExpandEnv(config.Database.Password)
	config.Database.Name = os.ExpandEnv(config.Database.Name)
	config.Database.Host = os.ExpandEnv(config.Database.Host)

	// Сессии
	config.Session.SecretKey = os.ExpandEnv(config.Session.SecretKey)
}

// setDefaults устанавливает значения по умолчанию
func setDefaults(config *Config) {
	// Сервер
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	if config.Server.ReadTimeout == 0 {
		config.Server.ReadTimeout = 15 * time.Second
	}
	if config.Server.WriteTimeout == 0 {
		config.Server.WriteTimeout = 15 * time.Second
	}

	// База данных
	if config.Database.Host == "" {
		config.Database.Host = "localhost"
	}
	if config.Database.Port == 0 {
		config.Database.Port = 5432
	}
	if config.Database.SSLMode == "" {
		config.Database.SSLMode = "disable"
	}
	if config.Database.MaxOpenConns == 0 {
		config.Database.MaxOpenConns = 25
	}
	if config.Database.MaxIdleConns == 0 {
		config.Database.MaxIdleConns = 5
	}

	// Сессии
	if config.Session.MaxAge == 0 {
		config.Session.MaxAge = 86400 // 24 часа
	}

	// Фоновые задачи
	if config.Background.InterestAccrualInterval == 0 {
		config.Background.InterestAccrualInterval = 10 * time.Minute
	}

	// Логирование
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
	if config.Logging.Format == "" {
		config.Logging.Format = "text"
	}
}

// GetDSN возвращает строку подключения к базе данных
func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
	)
}
