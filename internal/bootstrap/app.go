// Package bootstrap собирает приложение из конфигурации и инфраструктурных зависимостей.
package bootstrap

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	swagger "github.com/gofiber/swagger/v2"

	"github.com/gef3dx/it_courses/internal/article"
	"github.com/gef3dx/it_courses/internal/auth"
	"github.com/gef3dx/it_courses/internal/cache"
	"github.com/gef3dx/it_courses/internal/config"
	"github.com/gef3dx/it_courses/internal/course"
	"github.com/gef3dx/it_courses/internal/database/postgres"
	"github.com/gef3dx/it_courses/internal/lesson"
	"github.com/gef3dx/it_courses/internal/mailer"
	"github.com/gef3dx/it_courses/internal/page"
	"github.com/gef3dx/it_courses/internal/payment"
	"github.com/gef3dx/it_courses/internal/queue"
	"github.com/gef3dx/it_courses/internal/storage"
	testdomain "github.com/gef3dx/it_courses/internal/test"
	"github.com/gef3dx/it_courses/internal/upload"
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
	pgStorage, err := postgres.New(cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("init postgres: %w", err)
	}

	// Перед стартом приложения приводим схему БД к актуальному состоянию.
	if err := postgres.ApplyMigrations(pgStorage.DB, cfg.Migrations); err != nil {
		return nil, fmt.Errorf("apply migrations: %w", err)
	}

	app := fiber.New(fiber.Config{
		// Базовые настройки Fiber для всего HTTP-приложения.
		AppName: "it_courses",
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

	userRepository := user.NewRepository(pgStorage.DB)
	authRepository := auth.NewRepository(pgStorage.DB, userRepository)
	authService := auth.NewService(authRepository, mailer.NoopSender{}, cfg.Auth)
	userService := user.NewService(userRepository)
	courseRepository := course.NewRepository(pgStorage.DB)
	courseService := course.NewService(courseRepository)
	paymentRepository := payment.NewRepository(pgStorage.DB, courseRepository)
	paymentService := payment.NewService(paymentRepository, courseRepository)
	testRepository := testdomain.NewRepository(pgStorage.DB)
	testService := testdomain.NewService(testRepository)
	lessonRepository := lesson.NewRepository(pgStorage.DB)
	lessonService := lesson.NewService(lessonRepository)
	pageRepository := page.NewRepository(pgStorage.DB)
	pageService := page.NewService(pageRepository)
	articleRepository := article.NewRepository(pgStorage.DB)
	articleService := article.NewService(articleRepository)
	storageService := storage.New(cfg.Storage.BaseURL)
	publisher := queue.NoopPublisher{}
	_ = cache.NewCache()

	auth.RegisterRoutes(app, authService)
	user.RegisterRoutes(
		app,
		userService,
		authService.Required(),
		authService.Required(user.RoleAdmin),
		authService.OwnerOrAdmin(),
	)
	course.RegisterRoutes(app, courseService, authService)
	payment.RegisterRoutes(app, paymentService, authService)
	testdomain.RegisterRoutes(app, testService, authService)
	lesson.RegisterRoutes(app, lessonService, authService)
	page.RegisterRoutes(app, pageService, authService)
	article.RegisterRoutes(app, articleService, authService)
	upload.RegisterRoutes(app, authService, storageService, publisher)

	return &App{
		Fiber:   app,
		Storage: pgStorage,
	}, nil
}

// Close освобождает ресурсы приложения перед завершением процесса.
func (a *App) Close() error {
	if a.Storage == nil {
		return nil
	}

	return a.Storage.Close()
}
