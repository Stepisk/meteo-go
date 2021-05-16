package app

import (
	"context"
	"errors"
	"gitlab.com/peleng-meteo/meteo-go/internal/config"
	"gitlab.com/peleng-meteo/meteo-go/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Run initializes whole application
func Run(configPath string) {
	cfg, err := config.Init(configPath)
	if err != nil {
		logger.Error(err)
		return
	}

	// Services, Repos & API Handlers
	repos := repository.NewRepositories(db)
	services := service.NewServices(service.Deps{
		Repos:       repos,
		Environment: cfg.Environment,
	})
	handlers := delivery.NewHandler(services, tokenManager)

	// HTTP Server
	webServer := server.NewServer(cfg, handlers.Init(cfg))
	go func() {
		if err := server.Run(); !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("error occurred while running htto server: %s\n", err.Error())
		}
	}()

	logger.Info("Server started")

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	const timeout = 5 * time.Second

	ctx, shutdown := context.WithTimeout(context.Background(), timeout)
	defer shutdown()

	if err := webServer.Stop(ctx); err != nil {
		logger.Errorf("failed to stop server: %v", err)
	}
}