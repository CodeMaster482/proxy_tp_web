package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"proxy/cmd/app/init/server"
	"proxy/pkg/config"
	"proxy/pkg/logger"

	postgresql "proxy/cmd/app/init/db"

	handlerRequest "proxy/internal/api/handler/http"
	repositoryRequest "proxy/internal/api/repository/requests"
	usecaseRequest "proxy/internal/api/usecase/requests"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/sync/errgroup"
)

func main() {
	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error .env load: %v\n", err)
	}

	// Load config
	cfg, err := config.GetConfig(".")
	if err != nil {
		log.Fatalf("Error loading configuration: %v\n", err)
	}

	// Setup Context
	ctx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelCtx()

	logger := logger.NewLogger(ctx, cfg.Logger)

	router := gin.Default()
	// ------------------- DB ---------------------
	db, err := postgresql.InitPostgresDB(ctx, cfg)
	if err != nil {
		logger.Errorf("Error Initializing PostgreSQL database: %v", err)
		return
	}
	defer func() {
		db.Close()
		logger.Info("Db closed without errors")
	}()

	// ----------------- PROXY ------------------------
	// ------------------------------------------------

	r := repositoryRequest.NewRepository(db, logger)
	u := usecaseRequest.NewUsecase(r, logger)
	h := handlerRequest.NewHandler(logger, u)

	server := server.NewServer(signalCtx, &cfg, router, logger, h)
	// server.ListenAndServe()
	// go server.ApiRun()

	// addrPxy := fmt.Sprintf("%s:%s", cfg.Proxy.Addr, cfg.Proxy.Port)
	// if err := http.ListenAndServe(addrPxy, proxy); err != nil {
	// 	logger.Fatalf("Error: %s", err)
	// }

	g, gCtx := errgroup.WithContext(signalCtx)
	g.Go(func() error {
		return server.ListenAndServe()
	})
	g.Go(func() error {
		<-gCtx.Done()
		return server.Shutdown(context.Background())
	})

	if err := g.Wait(); err != nil {
		logger.Infof("exit reason: %v\n", err)
	}
}
