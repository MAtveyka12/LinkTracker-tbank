package httpserver

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

type HTTPServer struct {
	logger *slog.Logger
	server *http.Server
	config *Config
}

type Config struct {
	Host              string
	Port              string
	StartMsg          string
	Handler           http.Handler
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	ReadTimeout       time.Duration
	ShutdownTimeout   time.Duration
}

func NewHTTPServer(logger *slog.Logger, config *Config) *HTTPServer {
	server := &http.Server{
		Handler:           config.Handler,
		ReadTimeout:       config.ReadTimeout,
		WriteTimeout:      config.WriteTimeout,
		ReadHeaderTimeout: config.ReadHeaderTimeout,
		Addr:              config.Host + ":" + config.Port,
	}

	s := &HTTPServer{
		logger: logger,
		server: server,
		config: config,
	}

	return s
}

func (a *HTTPServer) Start(ctx context.Context) error {
	a.logger.Info(a.config.StartMsg)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		select {
		case <-ctx.Done():
			a.logger.Info("Context cancelled, shutting down gracefully...")
		case <-sigCh:
			a.logger.Info("Shutdown signal received, shutting down gracefully...")
		}

		ctx, cancel := context.WithTimeout(context.Background(), a.config.ShutdownTimeout)
		defer cancel()

		err := a.server.Shutdown(ctx)
		if err != nil {
			a.logger.Error("Failed to shutdown the server", slog.String("error", err.Error()))
			return err
		}

		a.logger.Info("Server is shutdown!")

		return nil
	})

	g.Go(func() error {
		err := a.server.ListenAndServe()
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}

			return err
		}

		return nil
	})

	return g.Wait()
}
