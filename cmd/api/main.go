// Package main содержит точку входа в API-приложение.
//
//go:generate go run github.com/swaggo/swag/cmd/swag@v1.16.4 init -g cmd/api/main.go -d ../.. -o ../../docs --parseInternal
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/gef3dx/it_courses/docs"
	"github.com/gef3dx/it_courses/internal/bootstrap"
	"github.com/gef3dx/it_courses/internal/config"
)

// @title API it_courses
// @version 1.0
// @description Документация API сервиса it_courses.
// @host localhost:3000
// @BasePath /

// main загружает конфигурацию, собирает приложение и запускает HTTP-сервер.
func main() {
	cfg := config.MustLoad()

	app, err := bootstrap.NewApp(cfg)
	if err != nil {
		log.Fatalf("startup failed: %v", err)
	}
	defer func() {
		if err := app.Close(); err != nil {
			log.Printf("close storage: %v", err)
		}
	}()

	// Отдельная горутина слушает системные сигналы и завершает Fiber корректно.
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		<-stop

		if err := app.Fiber.Shutdown(); err != nil {
			log.Printf("shutdown failed: %v", err)
		}
	}()

	address := cfg.HTTP.Address()
	log.Printf("server is starting on %s", address)

	// Запускаем HTTP-сервер только после успешной инициализации всех зависимостей.
	if err := app.Fiber.Listen(address); err != nil {
		fmt.Println(err)
	}
}
