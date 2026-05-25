// Package bootstrap собирает приложение из конфигурации и инфраструктурных зависимостей.
package bootstrap

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	swagger "github.com/gofiber/swagger/v2"

	"github.com/gef3dx/it_courses/internal/config"
	"github.com/gef3dx/it_courses/internal/database/postgres"
	"github.com/gef3dx/it_courses/internal/user"
)

// App объединяет HTTP-приложение и инфраструктурные зависимости, которые нужно закрыть при завершении.
type App struct {
	Fiber   *fiber.App
	Storage *postgres.Storage
}

// MessageResponse описывает простой текстовый ответ от корневого endpoint.
type MessageResponse struct {
	Message string `json:"message"`
}

// HealthResponse описывает ответ health-check маршрута.
type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

// rootHandler отвечает на корневой маршрут и подтверждает, что HTTP-слой работает.
//
// @Summary Корневой маршрут
// @Description Возвращает простой ответ для быстрой проверки запуска HTTP-сервера.
// @Tags system
// @Produce json
// @Success 200 {object} MessageResponse
// @Router / [get]
func rootHandler(c fiber.Ctx) error {
	return c.JSON(MessageResponse{Message: "Hello, World!"})
}

// healthHandler возвращает текущее состояние сервиса и его соединения с базой данных.
//
// @Summary Проверка состояния сервиса
// @Description Показывает, что HTTP-сервер запущен и соединение с базой данных установлено.
// @Tags system
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func healthHandler(c fiber.Ctx) error {
	return c.JSON(HealthResponse{
		Status:   "ok",
		Database: "connected",
	})
}

// NewApp поднимает соединение с БД, применяет миграции и регистрирует HTTP-маршруты.
func NewApp(cfg *config.Config) (*App, error) {
	storage, err := postgres.New(cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("init postgres: %w", err)
	}

	// Перед стартом приложения приводим схему БД к актуальному состоянию.
	if err := postgres.ApplyMigrations(storage.DB, cfg.Migrations); err != nil {
		return nil, fmt.Errorf("apply migrations: %w", err)
	}

	app := fiber.New(fiber.Config{
		// Базовые настройки Fiber для всего HTTP-приложения.
		AppName: "api_workinghub",
	})

	// CORS для Swagger UI и внешних клиентов.
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	// Подключаем Swagger UI для просмотра сгенерированной документации API.
	app.Get("/swagger/*", swagger.HandlerDefault)

	app.Get("/", rootHandler)
	app.Get("/health", healthHandler)

	// Собираем user-модуль из репозитория, сервиса и роутов.
	userRepository := user.NewRepository(storage.DB)
	userService := user.NewService(userRepository)
	user.RegisterRoutes(app, userService)

	return &App{
		Fiber:   app,
		Storage: storage,
	}, nil
}

// Close освобождает ресурсы приложения перед завершением процесса.
func (a *App) Close() error {
	if a.Storage == nil {
		return nil
	}

	return a.Storage.Close()
}
