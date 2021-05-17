package app

import (
	"context"
	"errors"
	"gitlab.com/peleng-meteo/meteo-go/internal/config"
	"gitlab.com/peleng-meteo/meteo-go/internal/repository"
	"gitlab.com/peleng-meteo/meteo-go/internal/server"
	"gitlab.com/peleng-meteo/meteo-go/pkg/database/mongodb"
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

	// Dependencies
	mongoClient, err := mongodb.NewClient(cfg.Mongo.URI, cfg.Mongo.User, cfg.Mongo.Password)
	if err != nil {
		logger.Error(err)
		return
	}

	db := mongoClient.Database(cfg.Mongo.Name)

	memCache := cache.NewMemoryCache()
	hasher := hash.NewSHA1Hasher(cfg.Auth.PasswordSalt)
	emailProvider := sendpulse.NewClient(cfg.Email.SendPulse.ClientID, cfg.Email.SendPulse.ClientSecret, memCache)
	emailSender, err := smtp.NewSMTPSender(cfg.SMTP.From, cfg.SMTP.Pass, cfg.SMTP.Host, cfg.SMTP.Port)
	if err != nil {
		logger.Error(err)
		return
	}

	tokenManager, err := auth.NewManager(cfg.Auth.JWT.SigningKey)
	if err != nil {
		logger.Error(err)
		return
	}

	otpGenerator := otp.NewGOTPGenerator()

	// Services, Repos & API Handlers
	repos := repository.NewRepositories(db)
	services := service.NewServices(service.Deps{
		Repos:                  repos,
		Cache:                  memCache,
		Hasher:                 hasher,
		TokenManager:           tokenManager,
		EmailProvider:          emailProvider,
		EmailSender:            emailSender,
		EmailConfig:            cfg.Email,
		AccessTokenTTL:         cfg.Auth.JWT.AccessTokenTTL,
		RefreshTokenTTL:        cfg.Auth.JWT.RefreshTokenTTL,
		CacheTTL:               int64(cfg.CacheTTL.Seconds()),
		OtpGenerator:           otpGenerator,
		VerificationCodeLength: cfg.Auth.VerificationCodeLength,
		FrontendURL:            cfg.FrontendURL,
		Environment:            cfg.Environment,
	})
	handlers := delivery.NewHandler(services, tokenManager)

	// HTTP Server
	srv := server.NewServer(cfg, handlers.Init(cfg))
	go func() {
		if err := srv.Run(); !errors.Is(err, http.ErrServerClosed) {
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

	if err := srv.Stop(ctx); err != nil {
		logger.Errorf("failed to stop server: %v", err)
	}

	if err := mongoClient.Disconnect(context.Background()); err != nil {
		logger.Error(err.Error())
	}
}
