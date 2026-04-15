package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// Исправленные пути под твой проект
	infrastructurepostgres "github.com/medods/test-task-for-junior-backend-developer/internal/infrastructure/postgres"
	postgresrepo "github.com/medods/test-task-for-junior-backend-developer/internal/repository/postgres"
	transporthttp "github.com/medods/test-task-for-junior-backend-developer/internal/transport/http"
	swaggerdocs "github.com/medods/test-task-for-junior-backend-developer/internal/transport/http/docs"
	httphandlers "github.com/medods/test-task-for-junior-backend-developer/internal/transport/http/handlers"
	"github.com/medods/test-task-for-junior-backend-developer/internal/usecase/task"
)

func main() {
	// 1. Инициализация логгера
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg := loadConfig()

	// 2. Настройка контекста для корректного завершения (Ctrl+C)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 3. Подключение к БД
	pool, err := infrastructurepostgres.Open(ctx, cfg.DatabaseDSN)
	if err != nil {
		logger.Error("failed to open postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// 4. Сборка слоев приложения (Dependency Injection)
	// Важно: используем NewRepository, как мы писали в шаге с репозиторием
	taskRepo := postgresrepo.NewRepository(pool) 
	taskUsecase := task.NewService(taskRepo)
	taskHandler := httphandlers.NewTaskHandler(taskUsecase)
	
	// Если у тебя есть папка docs, оставляем это, если нет — можно закомментировать
	docsHandler := swaggerdocs.NewHandler() 
	
	router := transporthttp.NewRouter(taskHandler, docsHandler)

	// 5. Настройка HTTP сервера
	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// 6. Фоновая остановка сервера по сигналу
	go func() {
		<-ctx.Done()
		logger.Info("shutting down http server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown http server", "error", err)
		}
	}()

	logger.Info("http server started", "addr", cfg.HTTPAddr)

	// 7. Запуск сервера
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("listen and serve", "error", err)
		os.Exit(1)
	}
	
	logger.Info("server stopped gracefully")
}

// --- Конфигурация ---

type config struct {
	HTTPAddr    string
	DatabaseDSN string
}

func loadConfig() config {
	return config{
		HTTPAddr:    envOrDefault("HTTP_ADDR", ":8080"),
		// По умолчанию используем те данные, которые обычно в docker-compose проекта
		DatabaseDSN: envOrDefault("DATABASE_DSN", "postgres://postgres:postgres@localhost:5432/medods?sslmode=disable"),
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}