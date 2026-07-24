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

// shutdownTimeout bounds how long we wait for in-flight requests to drain before
// forcing the process to exit.
const shutdownTimeout = 15 * time.Second

func main() {
	log.Println("Starting Backend Server for error logging")

	container := di.BuildContainer()

	if err := container.Invoke(startServer); err != nil {
		log.Fatalf("fatal: %v", err)
	}
}

// startServer builds the HTTP server, runs it, and blocks until either the server
// fails to start or a shutdown signal arrives, at which point it drains in-flight
// requests within shutdownTimeout and releases resources before returning.
//
// mysql and redis are injected so they can be closed on shutdown. Injecting them
// also forces their construction at startup: mysql is required (a failure fails
// startup), while redis is degradable (provideRedis logs and continues).
func startServer(
	r *router.Router,
	appCfg config.AppConfig,
	mysqlC *mysqlclient.Client,
	redisC *redisclient.Client,
	kafkaC *kafkaclient.Client,
) error {
	// Release pools and flush the async Kafka writer on every exit path once the
	// server has drained.
	defer closeResources(mysqlC, redisC, kafkaC)

	srv := &http.Server{
		Addr:    appCfg.Addr(),
		Handler: r.Setup(),
		// Timeouts protect against slow/stuck clients (e.g. Slowloris) exhausting
		// connections.
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Run the listener in the background so we can concurrently wait on both a
	// startup error and OS shutdown signals.
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
		// Listener never came up (e.g. port already in use).
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

// closeResources closes each client, logging (but not failing on) individual errors.
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
