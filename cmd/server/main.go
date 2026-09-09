package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"projectOzonBank/internal/api"
	"projectOzonBank/internal/app"
	"projectOzonBank/internal/domain"
	"projectOzonBank/internal/storage/memory"
	"projectOzonBank/internal/storage/postgres"
)

// StorageType — допустимые бэкенды хранения. Типизация вместо голых строк
// защищает от опечаток на этапе компиляции.
type StorageType string

const (
	StorageMemory   StorageType = "memory"
	StoragePostgres StorageType = "postgres"
)

// Config собирает все настраиваемые параметры сервиса в одном месте.
type Config struct {
	Storage         StorageType
	Addr            string
	DSN             string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

func loadConfig() Config {
	// .env — best effort, отсутствие файла не является ошибкой
	// (например, в Docker переменные приходят напрямую из окружения)
	_ = godotenv.Load()

	storageType := flag.String("storage", getEnv("STORAGE", string(StorageMemory)), "storage backend: memory or postgres")
	addr := flag.String("addr", getEnv("ADDR", ":8080"), "http server address")
	dsn := flag.String("dsn", os.Getenv("DATABASE_URL"), "PostgreSQL connection string")
	flag.Parse()

	return Config{
		Storage:         StorageType(*storageType),
		Addr:            *addr,
		DSN:             *dsn,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: 5 * time.Second,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("fatal error: %v", err)
	}
}

func run() error {
	cfg := loadConfig()
	ctx := context.Background()

	storage, cleanup, err := newStorage(ctx, cfg)
	if err != nil {
		return err
	}
	defer cleanup()

	service := app.New(storage)
	handler := api.NewHandler(service)
	router := api.NewRouter(handler)

	server := &http.Server{
		Addr:         cfg.Addr,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("server started on %s", cfg.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	shutdownCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("server failed: %w", err)
		}
	case <-shutdownCtx.Done():
		log.Println("shutdown signal received")
	}

	shutdownTimeoutCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownTimeoutCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}

	log.Println("server stopped")
	return nil
}

// newStorage создаёт нужную реализацию domain.Storage по конфигу.
// Возвращает функцию очистки ресурсов — вызывающий код обязан
// вызвать её через defer.
func newStorage(ctx context.Context, cfg Config) (domain.Storage, func(), error) {
	switch cfg.Storage {
	case StorageMemory:
		return memory.New(), func() {}, nil

	case StoragePostgres:
		if cfg.DSN == "" {
			return nil, nil, errors.New("DATABASE_URL or -dsn is required for postgres storage")
		}
		pgStorage, err := postgres.New(ctx, cfg.DSN)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create postgres storage: %w", err)
		}
		return pgStorage, pgStorage.Close, nil

	default:
		return nil, nil, fmt.Errorf("unknown storage type: %s", cfg.Storage)
	}
}
