package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/gef3dx/api_workinghub/internal/config"
)

// Storage хранит активное GORM-подключение к PostgreSQL.
type Storage struct {
	DB *gorm.DB
}

var pgIdentifierPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// New создаёт подключение к PostgreSQL, при необходимости создаёт целевую БД и проверяет доступность соединения.
func New(cfg config.PostgresConfig) (*Storage, error) {
	if err := ensureDatabaseExists(cfg); err != nil {
		return nil, err
	}

	// Формируем DSN уже для целевой рабочей базы данных.
	dsn := buildDSN(cfg, cfg.Database)

	gormCfg := &gorm.Config{}

	// При включённом логировании отдаём GORM более подробный уровень логов.
	if cfg.Logging.LogQueries || cfg.Logging.LogSlowQueries {
		gormCfg.Logger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(postgres.Open(dsn), gormCfg)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Настраиваем пул соединений после открытия основного подключения.
	sqlDB.SetMaxOpenConns(cfg.Pool.MaxConnections)
	sqlDB.SetMaxIdleConns(cfg.Pool.MinConnections)
	sqlDB.SetConnMaxIdleTime(cfg.Pool.IdleTimeout())
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Проверяем, что база действительно отвечает до старта приложения.
	if err := pingWithRetry(sqlDB, cfg); err != nil {
		return nil, err
	}

	log.Println("postgres connected")

	return &Storage{DB: db}, nil
}

// Close закрывает underlying sql.DB, который использует GORM.
func (s *Storage) Close() error {
	sqlDB, err := s.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// pingWithRetry несколько раз пытается выполнить Ping к БД с таймаутом и паузой между попытками.
func pingWithRetry(sqlDB *sql.DB, cfg config.PostgresConfig) error {
	attempts := cfg.Retry.MaxRetries
	if attempts < 1 {
		attempts = 1
	}

	timeout := cfg.Pool.ConnectionTimeout()
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	var lastErr error

	for attempt := 1; attempt <= attempts; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		lastErr = sqlDB.PingContext(ctx)
		cancel()
		if lastErr == nil {
			return nil
		}

		if attempt == attempts {
			break
		}

		log.Printf("postgres ping failed, retrying (%d/%d): %v", attempt, attempts, lastErr)
		time.Sleep(cfg.Retry.Delay())
	}

	return fmt.Errorf("postgres is unavailable after %d attempts: %w", attempts, lastErr)
}

// ensureDatabaseExists подключается к служебной БД и создаёт рабочую БД, если та ещё не существует.
func ensureDatabaseExists(cfg config.PostgresConfig) error {
	if cfg.Database == "" {
		return fmt.Errorf("postgres database name is empty")
	}

	maintenanceDB := "postgres"
	if cfg.Database == maintenanceDB {
		return nil
	}

	adminDB, err := sql.Open("pgx", buildDSN(cfg, maintenanceDB))
	if err != nil {
		return fmt.Errorf("open maintenance connection: %w", err)
	}
	defer func() { _ = adminDB.Close() }()

	// Сначала убеждаемся, что сервер PostgreSQL в целом доступен.
	if err := pingWithRetry(adminDB, cfg); err != nil {
		return fmt.Errorf("ping maintenance database: %w", err)
	}

	exists, err := databaseExists(adminDB, cfg)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	databaseName, err := quoteIdentifier(cfg.Database)
	if err != nil {
		return err
	}

	userName, err := quoteIdentifier(cfg.User)
	if err != nil {
		return err
	}

	createQuery := fmt.Sprintf(
		"CREATE DATABASE %s OWNER %s",
		databaseName,
		userName,
	)

	if _, err := adminDB.Exec(createQuery); err != nil {
		return fmt.Errorf("create database %q: %w", cfg.Database, err)
	}

	log.Printf("postgres database %q created for user %q", cfg.Database, cfg.User)

	return nil
}

// databaseExists проверяет наличие рабочей БД по имени в системном каталоге PostgreSQL.
func databaseExists(sqlDB *sql.DB, cfg config.PostgresConfig) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Pool.ConnectionTimeout())
	defer cancel()

	var exists bool
	err := sqlDB.QueryRowContext(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)",
		cfg.Database,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check database %q: %w", cfg.Database, err)
	}

	return exists, nil
}

// buildDSN собирает строку подключения для заданной базы данных.
func buildDSN(cfg config.PostgresConfig, database string) string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Host,
		cfg.User,
		cfg.Password,
		database,
		cfg.Port,
		cfg.SSLMode,
	)
}

// quoteIdentifier безопасно подготавливает имя БД или пользователя для SQL-команды.
func quoteIdentifier(name string) (string, error) {
	if !pgIdentifierPattern.MatchString(name) {
		return "", fmt.Errorf("invalid postgres identifier: %q", name)
	}

	return `"` + name + `"`, nil
}
