package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/trigold786/92-Account-Center/pkg/logging"
)

type ServerConfig struct {
	Port            string
	ShutdownTimeout time.Duration
	ServiceName     string
}

type StandardServer struct {
	engine          *gin.Engine
	httpServer      *http.Server
	config          ServerConfig
	logger          *slog.Logger
	requestCount    uint64
	durationSumNano uint64
	durationCount   uint64
}

func NewStandardServer(cfg ServerConfig) *StandardServer {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	logger := logging.NewLogger(cfg.ServiceName)

	s := &StandardServer{
		engine: engine,
		config: cfg,
		logger: logger,
	}

	engine.Use(gin.RecoveryWithWriter(os.Stderr, func(c *gin.Context, err any) {
		logger.Error("panic recovered", "error", fmt.Sprintf("%v", err))
	}))
	engine.Use(logging.Middleware(logger))

	engine.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		atomic.AddUint64(&s.requestCount, 1)
		atomic.AddUint64(&s.durationSumNano, uint64(time.Since(start).Nanoseconds()))
		atomic.AddUint64(&s.durationCount, 1)
	})

	return s
}

func (s *StandardServer) Engine() *gin.Engine {
	return s.engine
}

func (s *StandardServer) Logger() *slog.Logger {
	return s.logger
}

func (s *StandardServer) SetupMetrics() {
	s.engine.GET("/metrics", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/plain", []byte(fmt.Sprintf(
			"http_requests_total{service=\"%s\"} %d\n",
			s.config.ServiceName, atomic.LoadUint64(&s.requestCount),
		)))
	})
}

func (s *StandardServer) SetupHealth(checker func(ctx context.Context) error) {
	s.engine.GET("/health", func(c *gin.Context) {
		if checker != nil {
			if err := checker(c.Request.Context()); err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"status": "down", "error": err.Error()})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"status": "up"})
	})
}

func (s *StandardServer) Start() error {
	s.httpServer = &http.Server{
		Addr:    ":" + s.config.Port,
		Handler: s.engine,
	}

	go func() {
		defer logging.RecoverGoroutine(s.logger, "signal-handler")
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		s.logger.Info("shutting down gracefully")
		s.Shutdown()
	}()

	s.logger.Info("starting server", "port", s.config.Port)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

func (s *StandardServer) Shutdown() {
	if s.httpServer == nil {
		return
	}
	timeout := s.config.ShutdownTimeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	s.httpServer.Shutdown(ctx)
}
