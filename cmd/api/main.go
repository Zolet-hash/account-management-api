package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Zolet-hash/account-management-api/internal/config"
	"github.com/Zolet-hash/account-management-api/internal/handler"
	httpmw "github.com/Zolet-hash/account-management-api/internal/http/middleware"
	"github.com/Zolet-hash/account-management-api/internal/infrastructure/database"
	"github.com/Zolet-hash/account-management-api/internal/infrastructure/telemetry"
	"github.com/Zolet-hash/account-management-api/internal/repository/postgres"
	"github.com/Zolet-hash/account-management-api/internal/service"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// application entry point
func main() {
	if err := run(); err != nil {
		slog.Error("application stopped", "error", err)
		os.Exit(1)
	}
}

// run initializes application dependencies, registers routes,
// starts the HTTP server, and handles graceful shutdown.
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	shutdownTelemetry, err := telemetry.Init(ctx, cfg.OTLPEndpoint)
	if err != nil {
		return err
	}

	defer func() {
		shutdown, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()
		_ = shutdownTelemetry(shutdown)
	}()

	db, err := database.Open(
		ctx,
		cfg.DatabaseURL,
		telemetry.QueryTracer{},
	)
	if err != nil {
		return err
	}
	defer db.Close()

	if err = db.Ping(ctx); err != nil {
		return err
	}

	metrics := telemetry.NewMetrics(db)
	h := handler.New(
		service.New(
			postgres.New(db),
		),
	)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(
		gin.Recovery(),
		httpmw.RequestID(),
		httpmw.Logger(log),
		otelgin.Middleware("account-management"),
		metrics.Middleware(),
	)
	r.MaxMultipartMemory = 8 << 20

	// health endpoint
	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "ok"})
	})

	// readiness endpoint -> check can communicate woth postgres
	r.GET("/ready", func(ctx *gin.Context) {
		check, stop := context.WithTimeout(
			ctx.Request.Context(),
			2*time.Second,
		)
		defer stop()

		if db.Ping(check) != nil {
			ctx.JSON(503, gin.H{"status": "not_ready"})
			return

		}

		ctx.JSON(200, gin.H{"status": "ready"})
	})

	// register prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(telemetry.Handler()))
	// api structure
	api := r.Group("/api/v1")

	api.Use(func(ctx *gin.Context) {
		ctx.Request.Body = http.MaxBytesReader(
			ctx.Writer,
			ctx.Request.Body,
			1<<20)
		ctx.Next()
	})

	// Register application endpoint
	api.POST("/users", h.CreateUser)
	api.POST("/accounts", h.CreateAccount)
	api.POST("/accounts/:accountId/deposit", h.TransferDeposit())
	api.POST("/accounts/:accountId/withdrawal", h.TransferWithdrawal())
	api.GET("/users/:userId/balances", h.Balances)
	api.GET("/transactions/volume", h.Volumes)

	// configure the http server
	srv := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// create err channel
	errCh := make(chan error, 1)

	// start the http server in a go routine
	go func() {
		log.Info("server listening", "address", cfg.ListenAddr)
		errCh <- srv.ListenAndServe()
	}()

	// listen for OS termination signals
	stopSignal, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	select {
	case err = <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err

	// gracefully shutdown server
	case <-stopSignal.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		return srv.Shutdown(shutdown)
	}
}
