package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/mywyau/arcana/server/internal/simulation"
)

const tickInterval = 50 * time.Millisecond

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	world := simulation.NewWorld()
	runner, err := simulation.NewRunner(world, tickInterval)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("simulation started with a %s tick interval", tickInterval)
	runner.Run(ctx)
	log.Printf("simulation stopped after %d ticks", world.Tick())
}
