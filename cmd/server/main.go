package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"error-logging/di"
	kafkaclient "error-logging/pkg/client/kafka"
	mysqlclient "error-logging/pkg/client/mysql"
	redisclient "error-logging/pkg/client/redis"
	"error-logging/pkg/config"
	"error-logging/router"
)

const shutdownTimeout = 15 * time.Second

func main() {
	log.Println("Starting Backend Server for error logging")

	container := di.BuildContainer()

	if err := container.Invoke(startServer); err != nil {
		log.Fatalf("fatal: %v", err)
	}
}

func startServer(
	r *router.Router,
	appCfg config.AppConfig,
	mysqlC *mysqlclient.Client,
	redisC *redisclient.Client,
	kafkaC *kafkaclient.Client,
) error {

	defer closeResources(mysqlC, redisC, kafkaC)

	srv := &http.Server{
		Addr:    appCfg.Addr(),
		Handler: r.Setup(),

		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("server listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:

		return fmt.Errorf("server failed to start: %w", err)
	case sig := <-quit:
		log.Printf("received %s, shutting down gracefully...", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("graceful shutdown timed out, connections forcibly closed: %w", err)
	}

	log.Println("server stopped cleanly")
	return nil
}

func closeResources(closers ...io.Closer) {
	for _, c := range closers {
		if c == nil {
			continue
		}
		if err := c.Close(); err != nil {
			log.Printf("error closing resource: %v", err)
		}
	}
}
