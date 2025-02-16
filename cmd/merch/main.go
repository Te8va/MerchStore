package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/golang-migrate/migrate/v4"
	"go.uber.org/zap"

	"github.com/Te8va/MerchStore/internal/config"
	"github.com/Te8va/MerchStore/internal/handler"
	"github.com/Te8va/MerchStore/internal/middleware"
	"github.com/Te8va/MerchStore/internal/repository"
	"github.com/Te8va/MerchStore/internal/service"
	"github.com/Te8va/MerchStore/pkg/logger"
)

func main() {
	cfg := config.Config{}
	if err := env.Parse(&cfg); err != nil {
		logger.Logger().Fatalln("Failed to parse env: %v", err)
	}

	m, err := migrate.New("file://migrations", cfg.PostgresConn)
	if err != nil {
		logger.Logger().Fatalln(zap.Error(err))
	}

	err = repository.ApplyMigrations(m)
	if err != nil {
		logger.Logger().Fatalln(zap.Error(err))
	}

	logger.Logger().Infoln("Migrations applied successfully")

	pool, err := repository.GetPgxPool(cfg.PostgresConn)
	if err != nil {
		logger.Logger().Fatalln(zap.Error(err))
	}

	logger.Logger().Infoln("Postgres connection pool created")

	var wg sync.WaitGroup

	merchRep := repository.NewMerchService(pool)
	merchService := service.NewMerch(merchRep)
	merchHandler := handler.NewMerchHandler(merchService, cfg.JWTKey)

	authRepository := repository.NewAuthorizationService(pool)
	authService := service.NewAuthorization(authRepository, cfg.JWTKey)
	authHandler := handler.NewAuthorizationHandler(authService)

	_, cancelDeleteCtx := context.WithCancel(context.Background())

	mux := http.NewServeMux()

	mux.Handle("GET /api/info", middleware.Log(http.HandlerFunc(merchHandler.GetUserInfoHandler)))
	mux.Handle("POST /api/sendCoin", middleware.Log(http.HandlerFunc(merchHandler.SendCoinHandler)))
	mux.Handle("GET /api/buy/{item}", middleware.Log(http.HandlerFunc(merchHandler.BuyMerchHandler)))
	mux.Handle("POST /api/auth", middleware.Log(http.HandlerFunc(authHandler.AuthHandler)))

	server := &http.Server{
		Addr:     fmt.Sprintf("%s:%d", cfg.ServiceHost, cfg.ServicePort),
		ErrorLog: log.New(logger.Logger(), "", 0),
		Handler:  mux,
	}

	go func() {
		logger.Logger().Infoln("Server started, listening on port", cfg.ServicePort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger().Fatalln("ListenAndServe failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit

	logger.Logger().Infoln("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Logger().Fatalln("Server was forced to shutdown:", zap.Error(err))
	}

	waitGroupChan := make(chan struct{})
	go func() {
		wg.Wait()
		waitGroupChan <- struct{}{}
	}()

	select {
	case <-waitGroupChan:
		logger.Logger().Infoln("All delete goroutines successfully finished")
	case <-time.After(time.Second * 3):
		cancelDeleteCtx()
		logger.Logger().Infoln("Some of delete goroutines have not completed their job due to shutdown timeout")
	}

	logger.Logger().Infoln("Server was shut down")
}
