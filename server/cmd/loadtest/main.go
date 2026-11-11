package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	config := loadTestConfig{}
	flag.StringVar(&config.URL, "url", "ws://localhost:8080/ws", "game WebSocket endpoint")
	flag.IntVar(&config.Clients, "clients", 50, "number of simulated clients")
	flag.DurationVar(&config.RampDuration, "ramp", 10*time.Second, "time over which clients connect")
	flag.DurationVar(&config.HoldDuration, "duration", time.Minute, "time to hold after all clients connect")
	flag.DurationVar(&config.MoveInterval, "move-every", 2*time.Second, "movement interval; zero disables movement")
	flag.StringVar(&config.Scenario, "scenario", scenarioClustered, "movement layout: clustered or spread")
	flag.BoolVar(&config.AllowRemote, "allow-remote", false, "allow a non-local WebSocket endpoint")
	flag.Parse()

	if err := config.validate(); err != nil {
		fmt.Fprintln(os.Stderr, "load test configuration:", err)
		flag.Usage()
		os.Exit(2)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	logger := log.New(os.Stdout, "", log.LstdFlags)
	report := runLoadTest(ctx, config, logger)
	fmt.Fprintln(os.Stdout, report.String())
	if report.Connected == 0 || report.ConnectFailures > 0 || report.UnexpectedDisconnects > 0 {
		os.Exit(1)
	}
}
