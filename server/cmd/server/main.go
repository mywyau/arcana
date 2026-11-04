package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/mywyau/arcana/server/internal/navigation"
	"github.com/mywyau/arcana/server/internal/simulation"
	websockettransport "github.com/mywyau/arcana/server/internal/transport/websocket"
)

const (
	tickInterval  = 50 * time.Millisecond
	listenAddress = ":8080"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	world := simulation.NewWorldWithPathfinder(navigation.NewPrototypeMap())
	runner, err := simulation.NewRunner(world, tickInterval)
	if err != nil {
		log.Fatal(err)
	}
	handler, err := websockettransport.NewHandler(world, tickInterval)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/ws", handler)
	server := &http.Server{
		Addr:              listenAddress,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("simulation started with a %s tick interval", tickInterval)
	go runner.Run(ctx)

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()
	log.Printf("WebSocket endpoint listening on ws://localhost%s/ws", listenAddress)

	select {
	case <-ctx.Done():
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server stopped unexpectedly: %v", err)
		}
		stop()
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown: %v", err)
	}
	log.Printf("simulation stopped after %d ticks", world.Tick())
}
