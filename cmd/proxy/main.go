package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"proxy/internal/proxy"
	"proxy/pkg/config"
	"proxy/pkg/logger"

	postgresql "proxy/cmd/app/init/db"

	repositoryRequest "proxy/internal/api/repository/requests"
	usecaseRequest "proxy/internal/api/usecase/requests"

	"github.com/joho/godotenv"
	"golang.org/x/sync/errgroup"
)

// type codeRecorder struct {
// 	ResponseWriter http.ResponseWriter
// 	statusCode     int
// }

// func (cr *codeRecorder) WriteHeader(code int) {
// 	cr.statusCode = code
// 	cr.ResponseWriter.WriteHeader(code)
// }

// func (cr *codeRecorder) Write(b []byte) (int, error) {
// 	return cr.ResponseWriter.Write(b)
// }

// func (cr *codeRecorder) Header() http.Header {
// 	return cr.ResponseWriter.Header()
// }

var (
	caCertFile = flag.String("ca_cert_file", "", "certificate .pem file for trusted CA")
	caKeyFile  = flag.String("ca_key_file", "", "key .pem file for trusted CA")
)

func main() {
	//log.SetFlags(log.LstdFlags | log.Lshortfile)

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

	db, err := postgresql.InitPostgresDB(ctx, cfg)
	if err != nil {
		logger.Errorf("Error Initializing PostgreSQL database: %v", err)
		return
	}
	defer func() {
		db.Close()
		logger.Info("Db closed without errors")
	}()

	r := repositoryRequest.NewRepository(db, logger)
	u := usecaseRequest.NewUsecase(r, logger)

	proxy, err := proxy.NewProxy(*caCertFile, *caKeyFile, u, logger)
	if err != nil {
		fmt.Println(*caCertFile, *caKeyFile)
		log.Fatal(err)
	}

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.Proxy.Addr, cfg.Proxy.Port),
		Handler: proxy,
	}

	g, gCtx := errgroup.WithContext(signalCtx)
	g.Go(func() error {
		logger.Infof(fmt.Sprintf("%s:%s", cfg.Proxy.Addr, cfg.Proxy.Port))
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

func init() {
	flag.Parse()
}
