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

type testSupply struct {
	itemID   string
	quantity int
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	world := simulation.NewWorldWithPathfinder(navigation.NewPrototypeMap())
	combatSupplies := []testSupply{
		{itemID: string(simulation.WeaponSword), quantity: 1},
		{itemID: string(simulation.WeaponDagger), quantity: 1},
		{itemID: string(simulation.WeaponSpear), quantity: 1},
		{itemID: string(simulation.WeaponHalberd), quantity: 1},
		{itemID: simulation.ItemBones, quantity: 25},
		{itemID: string(simulation.WeaponMace), quantity: 1},
		{itemID: string(simulation.WeaponBattleAxe), quantity: 1},
		{itemID: string(simulation.WeaponSteelSword), quantity: 1},
		{itemID: simulation.ItemSwiftHelm, quantity: 1},
		{itemID: simulation.ItemSwiftBody, quantity: 1},
		{itemID: simulation.ItemSwiftBoots, quantity: 1},
		{itemID: simulation.ItemCoins, quantity: 1000},
		{itemID: simulation.ItemPrayerPotion, quantity: 5},
		{itemID: simulation.ItemBronzePlate, quantity: 1},
		{itemID: simulation.ItemIronHelm, quantity: 1},
		{itemID: simulation.ItemIronPlate, quantity: 1},
		{itemID: string(simulation.WeaponMithrilSword), quantity: 1},
		{itemID: simulation.ItemMithrilHelm, quantity: 1},
		{itemID: simulation.ItemMithrilPlate, quantity: 1},
		{itemID: simulation.ItemBronzeShield, quantity: 1}, {itemID: simulation.ItemBronzeLegs, quantity: 1},
		{itemID: simulation.ItemBronzeGloves, quantity: 1}, {itemID: simulation.ItemBronzeBoots, quantity: 1},
		{itemID: simulation.ItemIronShield, quantity: 1}, {itemID: simulation.ItemIronLegs, quantity: 1},
		{itemID: simulation.ItemIronGloves, quantity: 1}, {itemID: simulation.ItemIronBoots, quantity: 1},
		{itemID: simulation.ItemMithrilShield, quantity: 1}, {itemID: simulation.ItemMithrilLegs, quantity: 1},
		{itemID: simulation.ItemMithrilGloves, quantity: 1}, {itemID: simulation.ItemMithrilBoots, quantity: 1},
		{itemID: simulation.ItemAmuletFerocity, quantity: 1},
		{itemID: simulation.ItemAmuletFury, quantity: 1},
	}
	magicRangedSupplies := []testSupply{
		{string(simulation.WeaponBow), 1}, {simulation.ItemIronArrow, 100},
		{string(simulation.WeaponCrossbow), 1}, {simulation.ItemIronBolt, 100},
		{string(simulation.WeaponStaff), 1}, {simulation.ItemStandardSpellbook, 1},
		{simulation.ItemAirRune, 50}, {simulation.ItemWaterRune, 50}, {simulation.ItemEarthRune, 50},
		{simulation.ItemFireRune, 50}, {simulation.ItemHolyRune, 50}, {simulation.ItemDarkRune, 50},
		{simulation.ItemNoviceRobeHood, 1}, {simulation.ItemNoviceRobeTop, 1},
		{simulation.ItemNoviceRobeGloves, 1}, {simulation.ItemNoviceRobeBottoms, 1}, {simulation.ItemNoviceRobeBoots, 1},
		{simulation.ItemNoviceRangerHood, 1}, {simulation.ItemNoviceRangerBody, 1},
		{simulation.ItemNoviceRangerGloves, 1}, {simulation.ItemNoviceRangerLegs, 1}, {simulation.ItemNoviceRangerBoots, 1},
		{simulation.ItemAmuletHawk, 1}, {simulation.ItemAmuletEmbers, 1},
	}
	craftingSupplies := []testSupply{
		{simulation.ItemCookedMinnow, 10}, {simulation.ItemBronzePickaxe, 1}, {simulation.ItemBronzeAxe, 1},
		{simulation.ItemKnife, 1}, {simulation.ItemChisel, 1}, {simulation.ItemSewingNeedle, 1},
		{simulation.ItemThread, 50}, {simulation.ItemTinderbox, 1}, {simulation.ItemHammer, 1},
		{simulation.ItemFishingRod, 1}, {simulation.ItemLobsterCage, 1},
		{simulation.ItemVialOfWater, 25}, {simulation.ItemSunleaf, 10}, {simulation.ItemMoonleaf, 10},
		{simulation.ItemSacredDust, 25}, {simulation.ItemFishOil, 25},
		{simulation.ItemFeathers, 200}, {simulation.ItemIronArrowTips, 200},
		{simulation.ItemGoldOre, 10}, {simulation.ItemUncutSapphire, 10},
		{simulation.ItemUncutEmerald, 10}, {simulation.ItemUncutRuby, 10}, {simulation.ItemUncutDiamond, 10},
	}
	for _, group := range []struct {
		position simulation.Position
		items    []testSupply
	}{
		{simulation.TestCombatChestPosition, combatSupplies},
		{simulation.TestMagicRangedChestPosition, magicRangedSupplies},
		{simulation.TestCraftingChestPosition, craftingSupplies},
	} {
		for _, supply := range group.items {
			if err := world.AddWorldDrop(group.position, supply.itemID, supply.quantity); err != nil {
				log.Fatal(err)
			}
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
