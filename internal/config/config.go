// Package config отвечает за загрузку и преобразование конфигурации приложения.
package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config объединяет все разделы конфигурационного файла приложения.
type Config struct {
	HTTP       HTTPConfig       `yaml:"http"`
	Auth       AuthConfig       `yaml:"auth"`
	Storage    StorageConfig    `yaml:"storage"`
	Postgres   PostgresConfig   `yaml:"postgres"`
	Migrations MigrationsConfig `yaml:"migrations"`
}

// HTTPConfig хранит сетевые настройки HTTP-сервера.
type HTTPConfig struct {
	Host string `yaml:"host" env-default:"0.0.0.0"`
	Port string `yaml:"port" env-default:"3000"`
}

// AuthConfig хранит настройки JWT-токенов и секрет подписи.
type AuthConfig struct {
	JWTSecret                 string `yaml:"jwt_secret" env-default:"development-secret"`
	AccessTokenTTLMinutes     int    `yaml:"access_token_ttl_minutes" env-default:"15"`
	RefreshTokenTTLHours      int    `yaml:"refresh_token_ttl_hours" env-default:"720"`
	PasswordResetTTLMinutes   int    `yaml:"password_reset_ttl_minutes" env-default:"60"`
	EmailVerificationTTLHours int    `yaml:"email_verification_ttl_hours" env-default:"24"`
}

type StorageConfig struct {
	BaseURL string `yaml:"base_url" env-default:"http://localhost:3000/uploads"`
}

// PostgresConfig хранит параметры подключения к PostgreSQL.
type PostgresConfig struct {
	Host     string `yaml:"host" env-required:"true"`
	Port     int    `yaml:"port" env-required:"true"`
	Database string `yaml:"database" env-required:"true"`
	User     string `yaml:"user" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
	SSLMode  string `yaml:"sslmode" env-default:"disable"`

	Pool PoolConfig `yaml:"pool"`

	Logging LoggingConfig `yaml:"logging"`
	Retry   RetryConfig   `yaml:"retry"`
}

// PoolConfig описывает размеры и таймауты пула соединений к БД.
type PoolConfig struct {
	MaxConnections           int `yaml:"max_connections" env-default:"20"`
	MinConnections           int `yaml:"min_connections" env-default:"5"`
	ConnectionTimeoutSeconds int `yaml:"connection_timeout_seconds" env-default:"30"`
	IdleTimeoutSeconds       int `yaml:"idle_timeout_seconds" env-default:"300"`
}

// LoggingConfig управляет подробностью логирования SQL-запросов.
type LoggingConfig struct {
	LogQueries           bool `yaml:"log_queries"`
	LogSlowQueries       bool `yaml:"log_slow_queries"`
	SlowQueryThresholdMs int  `yaml:"slow_query_threshold_ms" env-default:"500"`
}

// RetryConfig описывает повторные попытки подключения к базе данных.
type RetryConfig struct {
	MaxRetries   int `yaml:"max_retries" env-default:"3"`
	RetryDelayMs int `yaml:"retry_delay_ms" env-default:"1000"`
}

// MigrationsConfig описывает параметры автоприменения SQL-миграций.
type MigrationsConfig struct {
	AutoApply bool   `yaml:"auto_apply" env-default:"true"`
	Path      string `yaml:"path" env-default:"migrations"`
}

// MustLoad загружает конфиг по стандартному пути или из переданных параметров окружения.
func MustLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		panic("config path is empty")
	}

	return MustLoadPath(path)
}

// MustLoadPath читает конфиг из конкретного файла и паникует при ошибке.
func MustLoadPath(path string) *Config {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exist: " + path)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("cannot read config: " + err.Error())
	}

	return &cfg
}

// fetchConfigPath определяет путь к конфигу из флага, переменной окружения или значения по умолчанию.
func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	if !flag.Parsed() {
		flag.Parse()
	}

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	if res == "" {
		res = "config/config.yaml"
	}

	return res
}

// Address собирает полный адрес HTTP-сервера в формате host:port.
func (cfg HTTPConfig) Address() string {
	return cfg.Host + ":" + cfg.Port
}

// ConnectionTimeout преобразует timeout подключения к БД в time.Duration.
func (cfg PoolConfig) ConnectionTimeout() time.Duration {
	return time.Duration(cfg.ConnectionTimeoutSeconds) * time.Second
}

// IdleTimeout преобразует таймаут простоя соединения к БД в time.Duration.
func (cfg PoolConfig) IdleTimeout() time.Duration {
	return time.Duration(cfg.IdleTimeoutSeconds) * time.Second
}

// Delay преобразует задержку между ретраями подключения в time.Duration.
func (cfg RetryConfig) Delay() time.Duration {
	return time.Duration(cfg.RetryDelayMs) * time.Millisecond
}

// AccessTokenTTL возвращает TTL access token.
func (cfg AuthConfig) AccessTokenTTL() time.Duration {
	return time.Duration(cfg.AccessTokenTTLMinutes) * time.Minute
}

// RefreshTokenTTL возвращает TTL refresh token.
func (cfg AuthConfig) RefreshTokenTTL() time.Duration {
	return time.Duration(cfg.RefreshTokenTTLHours) * time.Hour
}

// PasswordResetTTL возвращает TTL токена сброса пароля.
func (cfg AuthConfig) PasswordResetTTL() time.Duration {
	return time.Duration(cfg.PasswordResetTTLMinutes) * time.Minute
}

// EmailVerificationTTL возвращает TTL токена подтверждения email.
func (cfg AuthConfig) EmailVerificationTTL() time.Duration {
	return time.Duration(cfg.EmailVerificationTTLHours) * time.Hour
}
