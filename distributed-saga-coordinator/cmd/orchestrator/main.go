package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"distributed-saga-coordinator/internal/saga"
	"distributed-saga-coordinator/internal/transport"
)

const (
	defaultConfigPath = "config/sagas.json"
	defaultBindAddr   = ":8080"
)

func main() {
	cfgPath := configPathFromArgs(os.Args)

	coord, err := saga.NewCoordinatorFromFile(cfgPath)
	if err != nil {
		log.Fatalf("failed to build coordinator from %s: %v", cfgPath, err)
	}

	httpSrv := transport.NewHTTPServer(defaultBindAddr, coord)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- httpSrv.Run()
	}()

	select {
	case err := <-serverErrCh:
		if errors.Is(err, http.ErrServerClosed) {
			log.Println("http server shut down cleanly")
		} else {
			log.Fatalf("http server stopped unexpectedly: %v", err)
		}
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}
}

func configPathFromArgs(args []string) string {
	if len(args) < 2 || args[1] == "" {
		return defaultConfigPath
	}
	return args[1]
}
