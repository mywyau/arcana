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

	"github.com/mywyau/arcana/server/internal/account"
	"github.com/mywyau/arcana/server/internal/character"
	"github.com/mywyau/arcana/server/internal/identity"
	"github.com/mywyau/arcana/server/internal/navigation"
	postgrespersistence "github.com/mywyau/arcana/server/internal/persistence/postgres"
	"github.com/mywyau/arcana/server/internal/serverconfig"
	"github.com/mywyau/arcana/server/internal/serverhealth"
	"github.com/mywyau/arcana/server/internal/simulation"
	httpaccount "github.com/mywyau/arcana/server/internal/transport/httpaccount"
	websockettransport "github.com/mywyau/arcana/server/internal/transport/websocket"
)

const (
	tickInterval = 50 * time.Millisecond
)

type testSupply struct {
	itemID   string
	quantity int
}

func main() {
	config, err := serverconfig.Load(os.Getenv)
	if err != nil {
		log.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	world := simulation.NewWorldWithPathfinder(navigation.NewPrototypeMap())
	world.SetDebugContentEnabled(config.DebugContentEnabled)
	if err := world.AddStarterToolSpawns(); err != nil {
		log.Fatal(err)
	}
	meleeSupplies := []testSupply{
		{itemID: string(simulation.WeaponSword), quantity: 1},
		{itemID: string(simulation.WeaponDagger), quantity: 1},
		{itemID: string(simulation.WeaponSpear), quantity: 1},
		{itemID: string(simulation.WeaponHalberd), quantity: 1},
		{itemID: string(simulation.WeaponMace), quantity: 1},
		{itemID: string(simulation.WeaponBattleAxe), quantity: 1},
		{itemID: string(simulation.WeaponSteelSword), quantity: 1},
		{itemID: simulation.ItemSwiftHelm, quantity: 1},
		{itemID: simulation.ItemSwiftBody, quantity: 1},
		{itemID: simulation.ItemSwiftBoots, quantity: 1},
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
		{itemID: simulation.ItemVanguardCape, quantity: 1},
	}
	rangedSupplies := []testSupply{
		{string(simulation.WeaponBow), 1}, {string(simulation.WeaponOakShortbow), 1},
		{string(simulation.WeaponWillowShortbow), 1}, {string(simulation.WeaponMapleShortbow), 1},
		{string(simulation.WeaponYewShortbow), 1},
		{simulation.ItemBronzeArrow, 100}, {simulation.ItemIronArrow, 100},
		{simulation.ItemSteelArrow, 100}, {simulation.ItemMithrilArrow, 100}, {simulation.ItemAdamantiteArrow, 100},
		{string(simulation.WeaponCrossbow), 1},
		{simulation.ItemBronzeBolt, 100}, {simulation.ItemIronBolt, 100},
		{simulation.ItemSteelBolt, 100}, {simulation.ItemMithrilBolt, 100}, {simulation.ItemAdamantiteBolt, 100},
		{simulation.ItemNoviceRangerHood, 1}, {simulation.ItemNoviceRangerBody, 1},
		{simulation.ItemNoviceRangerGloves, 1}, {simulation.ItemNoviceRangerLegs, 1}, {simulation.ItemNoviceRangerBoots, 1},
		{simulation.ItemAmuletHawk, 1},
		{simulation.ItemWayfarerCape, 1},
	}
	magicSupplies := []testSupply{
		{string(simulation.WeaponStaff), 1}, {simulation.ItemStandardSpellbook, 1},
		{simulation.ItemSkinBible, 1}, {simulation.ItemHolyBook, 1},
		{simulation.ItemClericsTome, 1},
		{simulation.ItemAirRune, 50}, {simulation.ItemWaterRune, 50}, {simulation.ItemEarthRune, 50},
		{simulation.ItemFireRune, 50}, {simulation.ItemHolyRune, 50}, {simulation.ItemDarkRune, 50},
		{simulation.ItemFocusRune, 250},
		{simulation.ItemBlankRune, 100},
		{simulation.ItemNoviceRobeHood, 1}, {simulation.ItemNoviceRobeTop, 1},
		{simulation.ItemNoviceRobeGloves, 1}, {simulation.ItemNoviceRobeBottoms, 1}, {simulation.ItemNoviceRobeBoots, 1},
		{simulation.ItemAmuletEmbers, 1},
		{simulation.ItemArcanistCape, 1},
	}
	oresSupplies := []testSupply{
		{simulation.ItemCopperOre, 20}, {simulation.ItemTinOre, 20}, {simulation.ItemIronOre, 20},
		{simulation.ItemCoal, 20}, {simulation.ItemMithrilOre, 20}, {simulation.ItemGoldOre, 20},
		{simulation.ItemSilverOre, 20}, {simulation.ItemClay, 20}, {simulation.ItemAdamantiteOre, 20},
		{simulation.ItemUncutSapphire, 10}, {simulation.ItemUncutEmerald, 10},
		{simulation.ItemUncutRuby, 10}, {simulation.ItemUncutDiamond, 10},
	}
	barsSupplies := []testSupply{
		{simulation.ItemBronzeBar, 20}, {simulation.ItemIronBar, 20},
		{simulation.ItemMithrilBar, 20}, {simulation.ItemGoldBar, 20},
		{simulation.ItemSteelBar, 20}, {simulation.ItemSilverBar, 20}, {simulation.ItemAdamantiteBar, 20},
	}
	toolsSupplies := []testSupply{
		{simulation.ItemBronzePickaxe, 1}, {simulation.ItemBronzeAxe, 1},
		{simulation.ItemKnife, 1}, {simulation.ItemChisel, 1}, {simulation.ItemSewingNeedle, 1},
		{simulation.ItemTinderbox, 1}, {simulation.ItemHammer, 1},
		{simulation.ItemRingMould, 1},
		{simulation.ItemAmuletMould, 1},
		{simulation.ItemFishingRod, 1}, {simulation.ItemLobsterCage, 1},
	}
	herbsSupplies := []testSupply{
		{simulation.ItemVialOfWater, 25}, {simulation.ItemSunleaf, 10}, {simulation.ItemMoonleaf, 10},
		{simulation.ItemSacredDust, 25}, {simulation.ItemFishOil, 25},
		{simulation.ItemPrayerPotion, 5}, {simulation.ItemMinorHealingPotion, 5}, {simulation.ItemMurkyBrew, 5},
		{simulation.ItemMeleePotion, 5}, {simulation.ItemRangingPotion, 5}, {simulation.ItemMagicPotion, 5},
	}
	craftingSupplies := []testSupply{
		{simulation.ItemRubyRing, 1}, {simulation.ItemEmeraldRing, 1},
		{simulation.ItemSapphireRing, 1}, {simulation.ItemDiamondRing, 1},
		{simulation.ItemThread, 50}, {simulation.ItemBowstring, 50}, {simulation.ItemFeathers, 200},
		{simulation.ItemBronzeArrowTips, 200}, {simulation.ItemIronArrowTips, 200},
		{simulation.ItemSteelArrowTips, 200}, {simulation.ItemMithrilArrowTips, 200}, {simulation.ItemAdamantiteArrowTips, 200},
		{simulation.ItemNormalLogs, 10}, {simulation.ItemOakLogs, 10}, {simulation.ItemWillowLogs, 10},
		{simulation.ItemMapleLogs, 10}, {simulation.ItemYewLogs, 10},
		{simulation.ItemBones, 25}, {simulation.ItemLeather, 25}, {simulation.ItemCoins, 1000},
		{simulation.ItemSapphire, 10}, {simulation.ItemEmerald, 10}, {simulation.ItemRuby, 10}, {simulation.ItemDiamond, 10},
	}
	foodSupplies := []testSupply{
		{simulation.ItemRawMinnow, 10}, {simulation.ItemCookedMinnow, 10},
		{simulation.ItemRawTrout, 10}, {simulation.ItemCookedTrout, 10},
		{simulation.ItemRawSalmon, 10}, {simulation.ItemCookedSalmon, 10},
		{simulation.ItemRawTuna, 10}, {simulation.ItemCookedTuna, 10},
		{simulation.ItemRawLobster, 10}, {simulation.ItemCookedLobster, 10},
	}
	debugSupplyGroups := []struct {
		position simulation.Position
		items    []testSupply
	}{
		{simulation.TestMeleeChestPosition, meleeSupplies},
		{simulation.TestRangedChestPosition, rangedSupplies},
		{simulation.TestMagicChestPosition, magicSupplies},
		{simulation.TestOresChestPosition, oresSupplies},
		{simulation.TestBarsChestPosition, barsSupplies},
		{simulation.TestToolsChestPosition, toolsSupplies},
		{simulation.TestHerbsChestPosition, herbsSupplies},
		{simulation.TestCraftingChestPosition, craftingSupplies},
		{simulation.TestFoodChestPosition, foodSupplies},
	}
	if config.DebugContentEnabled {
		for _, group := range debugSupplyGroups {
			for _, supply := range group.items {
				if err := world.AddTestSupply(group.position, supply.itemID, supply.quantity); err != nil {
					log.Fatal(err)
				}
			}
		}
	}
	runner, err := simulation.NewRunner(world, tickInterval)
	if err != nil {
		log.Fatal(err)
	}
	var readinessChecker serverhealth.Checker
	var guestIdentities *identity.GuestManager
	var characters character.Repository
	var accountManager *account.Manager
	databaseURL := config.DatabaseURL
	if databaseURL == "" {
		guestIdentities, err = identity.NewGuestManager(identity.NewMemoryGuestRepository())
		if err != nil {
			log.Fatal(err)
		}
		characters = character.NewMemoryRepository()
		accountRepository := account.NewMemoryRepository()
		accountManager, err = account.NewManager(accountRepository, accountRepository)
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

		guestIdentities, err = identity.NewGuestManager(store)
		if err != nil {
			log.Fatal(err)
		}
		accountManager, err = account.NewManager(store, store)
		if err != nil {
			log.Fatal(err)
		}
		characters = store
		log.Print("PostgreSQL persistence is enabled")
	}
	handler, err := websockettransport.NewHandlerWithRepositoriesAndOptions(
		world,
		tickInterval,
		guestIdentities,
		characters,
		websockettransport.HandlerOptions{
			AllowDebugCommands: config.DebugContentEnabled,
			AllowedOrigins:     config.AllowedOrigins,
			AccountManager:     accountManager,
			RequireAccounts:    !config.GuestAuthenticationEnabled,
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	accountHTTP, err := httpaccount.NewHandlerWithOptions(
		accountManager,
		guestIdentities,
		characters,
		httpaccount.HandlerOptions{TrustProxyHeaders: config.TrustProxyHeaders},
	)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/ws", handler)
	accountHTTP.RegisterRoutes(mux)
	mux.Handle("/health", serverhealth.Liveness())
	mux.Handle("/ready", serverhealth.Readiness(readinessChecker))
	server := &http.Server{
		Addr:              config.ListenAddress,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf(
		"simulation started in %s mode with a %s tick interval (debug content: %t)",
		config.Environment, tickInterval, config.DebugContentEnabled,
	)
	go runner.Run(ctx)

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()
	log.Printf("WebSocket endpoint listening on %s/ws", config.ListenAddress)
	log.Printf("Health endpoints listening on %s/health and /ready", config.ListenAddress)

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
