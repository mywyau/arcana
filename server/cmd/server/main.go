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
	initialDrops := []struct {
		itemID   string
		quantity int
	}{
		{itemID: string(simulation.WeaponSword), quantity: 1},
		{itemID: string(simulation.WeaponDagger), quantity: 1},
		{itemID: string(simulation.WeaponBow), quantity: 1},
		{itemID: simulation.ItemIronArrow, quantity: 100},
		{itemID: string(simulation.WeaponMace), quantity: 1},
		{itemID: string(simulation.WeaponBattleAxe), quantity: 1},
		{itemID: string(simulation.WeaponStaff), quantity: 1},
		{itemID: string(simulation.WeaponSteelSword), quantity: 1},
		{itemID: simulation.ItemAirRune, quantity: 50},
		{itemID: simulation.ItemSwiftHelm, quantity: 1},
		{itemID: simulation.ItemSwiftBody, quantity: 1},
		{itemID: simulation.ItemSwiftBoots, quantity: 1},
		{itemID: simulation.ItemStandardSpellbook, quantity: 1},
		{itemID: simulation.ItemCookedMinnow, quantity: 10},
		{itemID: simulation.ItemCoins, quantity: 1000},
		{itemID: simulation.ItemBronzePickaxe, quantity: 1},
		{itemID: simulation.ItemBronzeAxe, quantity: 1},
		{itemID: simulation.ItemKnife, quantity: 1},
		{itemID: simulation.ItemChisel, quantity: 1},
		{itemID: simulation.ItemSewingNeedle, quantity: 1},
		{itemID: simulation.ItemThread, quantity: 50},
		{itemID: simulation.ItemTinderbox, quantity: 1},
		{itemID: simulation.ItemHammer, quantity: 1},
		{itemID: simulation.ItemPrayerPotion, quantity: 5},
		{itemID: simulation.ItemFishingRod, quantity: 1},
		{itemID: simulation.ItemLobsterCage, quantity: 1},
		{itemID: simulation.ItemVialOfWater, quantity: 25},
		{itemID: simulation.ItemSunleaf, quantity: 10},
		{itemID: simulation.ItemMoonleaf, quantity: 10},
		{itemID: simulation.ItemSacredDust, quantity: 25},
		{itemID: simulation.ItemFishOil, quantity: 25},
		{itemID: simulation.ItemFeathers, quantity: 200},
		{itemID: simulation.ItemIronArrowTips, quantity: 200},
		{itemID: simulation.ItemBronzePlate, quantity: 1},
		{itemID: simulation.ItemIronHelm, quantity: 1},
		{itemID: simulation.ItemIronPlate, quantity: 1},
		{itemID: string(simulation.WeaponMithrilSword), quantity: 1},
		{itemID: simulation.ItemMithrilHelm, quantity: 1},
		{itemID: simulation.ItemMithrilPlate, quantity: 1},
		{itemID: simulation.ItemAmuletFerocity, quantity: 1},
		{itemID: simulation.ItemAmuletHawk, quantity: 1},
		{itemID: simulation.ItemAmuletEmbers, quantity: 1},
		{itemID: simulation.ItemAmuletFury, quantity: 1},
	}
	for _, drop := range initialDrops {
		if err := world.AddWorldDrop(simulation.TestSupplyChestPosition, drop.itemID, drop.quantity); err != nil {
			log.Fatal(err)
		}
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
