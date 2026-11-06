package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mywyau/arcana/server/internal/identity"
	"github.com/mywyau/arcana/server/internal/navigation"
	postgrespersistence "github.com/mywyau/arcana/server/internal/persistence/postgres"
	"github.com/mywyau/arcana/server/internal/serverhealth"
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
	if err := world.AddWorldDrop(
		simulation.Position{X: -2, Y: 0, Z: -5},
		string(simulation.WeaponSteelSword),
		1,
	); err != nil {
		log.Fatal(err)
	}
	runner, err := simulation.NewRunner(world, tickInterval)
	if err != nil {
		log.Fatal(err)
	}
	var handler *websockettransport.Handler
	var readinessChecker serverhealth.Checker
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		handler, err = websockettransport.NewHandler(world, tickInterval)
		if err != nil {
			log.Fatal(err)
		}
		log.Print("DATABASE_URL is not set; guest identities and characters are in memory")
	} else {
		startupCtx, cancelStartup := context.WithTimeout(ctx, 10*time.Second)
		store, openErr := postgrespersistence.Open(startupCtx, databaseURL)
		if openErr != nil {
			cancelStartup()
			log.Fatal(openErr)
		}
		if migrateErr := store.Migrate(startupCtx); migrateErr != nil {
			cancelStartup()
			store.Close()
			log.Fatal(migrateErr)
		}
		cancelStartup()
		defer store.Close()
		readinessChecker = store

		guestIdentities, identityErr := identity.NewGuestManager(store)
		if identityErr != nil {
			log.Fatal(identityErr)
		}
		handler, err = websockettransport.NewHandlerWithRepositories(
			world,
			tickInterval,
			guestIdentities,
			store,
		)
		if err != nil {
			log.Fatal(err)
		}
		log.Print("PostgreSQL persistence is enabled")
	}

	mux := http.NewServeMux()
	mux.Handle("/ws", handler)
	mux.Handle("/health", serverhealth.Liveness())
	mux.Handle("/ready", serverhealth.Readiness(readinessChecker))
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
	log.Printf("Health endpoints listening on http://localhost%s/health and /ready", listenAddress)

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
	if err := handler.SaveAll(shutdownCtx); err != nil {
		log.Printf("save characters during shutdown: %v", err)
	}
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown: %v", err)
	}
	log.Printf("simulation stopped after %d ticks", world.Tick())
}
