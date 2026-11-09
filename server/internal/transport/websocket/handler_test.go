package websockettransport

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/mywyau/arcana/server/internal/character"
	"github.com/mywyau/arcana/server/internal/identity"
	"github.com/mywyau/arcana/server/internal/protocol"
	"github.com/mywyau/arcana/server/internal/simulation"
)

func TestHandlerMovesConnectionPlayerAndSendsSnapshots(t *testing.T) {
	world := simulation.NewWorld()
	const interval = 5 * time.Millisecond
	runner, err := simulation.NewRunner(world, interval)
	if err != nil {
		t.Fatalf("NewRunner() returned an unexpected error: %v", err)
	}
	runnerCtx, stopRunner := context.WithCancel(context.Background())
	defer stopRunner()
	go runner.Run(runnerCtx)

	handler, err := NewHandler(world, interval)
	if err != nil {
		t.Fatalf("NewHandler() returned an unexpected error: %v", err)
	}
	server := startTestServer(t, handler)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	connection, _, err := websocket.Dial(ctx, websocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("websocket.Dial() returned an unexpected error: %v", err)
	}
	defer connection.CloseNow()

	authenticateGuest(t, ctx, connection, "")
	welcome := readServerMessage(t, ctx, connection)
	if welcome.Type != protocol.MessageWelcome || welcome.PlayerID == "" || welcome.GuestToken == "" {
		t.Fatalf("welcome message = %+v", welcome)
	}
	if len(welcome.ProductionRecipes) != len(simulation.ProductionRecipeCatalog()) {
		t.Fatalf("welcome production recipe count = %d, want %d", len(welcome.ProductionRecipes), len(simulation.ProductionRecipeCatalog()))
	}
	if recipe := welcome.ProductionRecipes[0]; recipe.Station == "" || recipe.ID == "" || len(recipe.Costs) == 0 {
		t.Fatalf("welcome production recipe is incomplete: %+v", recipe)
	}
	foundFletchingRecipe := false
	for _, recipe := range welcome.ProductionRecipes {
		if recipe.Station == "fletching" {
			foundFletchingRecipe = true
			if recipe.RequiredLevel < 1 || recipe.RewardXP < 1 || recipe.Interval < 1 {
				t.Fatalf("welcome fletching recipe is incomplete: %+v", recipe)
			}
		}
	}
	if !foundFletchingRecipe {
		t.Fatal("welcome production catalogue has no fletching recipe")
	}
	if len(welcome.Spellbooks) != len(simulation.SpellbookCatalog()) || len(welcome.Spells) != len(simulation.SpellCatalog()) {
		t.Fatalf("welcome spell catalogue has %d books and %d spells", len(welcome.Spellbooks), len(welcome.Spells))
	}
	if spell := welcome.Spells[0]; spell.ID == "" || spell.BookID == "" || spell.IconID == "" || len(spell.Costs) == 0 {
		t.Fatalf("welcome spell is incomplete: %+v", spell)
	}
	if len(welcome.Prayers) != len(simulation.PrayerCatalog()) || len(welcome.PrayerSets) != len(simulation.PrayerSetCatalog()) {
		t.Fatalf("welcome prayer catalogue has %d prayers and %d sets", len(welcome.Prayers), len(welcome.PrayerSets))
	}
	if prayer := welcome.Prayers[0]; prayer.ID == "" || prayer.Name == "" || prayer.DrainAmount <= 0 || prayer.DrainIntervalSeconds <= 0 {
		t.Fatalf("welcome prayer is incomplete: %+v", prayer)
	}
	if set := welcome.PrayerSets[0]; set.ID == "" || set.Name == "" || len(set.PrayerIDs) == 0 {
		t.Fatalf("welcome prayer set is incomplete: %+v", set)
	}

	command := protocol.ClientMessage{
		Type:        protocol.MessageMoveTo,
		Sequence:    1,
		Destination: &protocol.Position{Y: 10},
	}
	if err := wsjson.Write(ctx, connection, command); err != nil {
		t.Fatalf("wsjson.Write() returned an unexpected error: %v", err)
	}

	for {
		message := readServerMessage(t, ctx, connection)
		if message.Type != protocol.MessageSnapshot || message.Snapshot == nil {
			continue
		}
		player, exists := findPlayer(message.Snapshot, welcome.PlayerID)
		if exists && player.LastCommandSequence == 1 {
			if player.Speed != playerSpeed {
				t.Fatalf("player speed = %v, want %v", player.Speed, playerSpeed)
			}
			if player.Position.Y <= 0 {
				t.Fatalf("player position = %+v, want positive Y movement", player.Position)
			}
			break
		}
	}
}

func TestHandlerReportsInvalidMessages(t *testing.T) {
	world := simulation.NewWorld()
	handler, err := NewHandler(world, 5*time.Millisecond)
	if err != nil {
		t.Fatalf("NewHandler() returned an unexpected error: %v", err)
	}
	server := startTestServer(t, handler)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	connection, _, err := websocket.Dial(ctx, websocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("websocket.Dial() returned an unexpected error: %v", err)
	}
	defer connection.CloseNow()
	authenticateGuest(t, ctx, connection, "")
	_ = readServerMessage(t, ctx, connection)

	if err := wsjson.Write(ctx, connection, map[string]any{"type": "dance"}); err != nil {
		t.Fatalf("wsjson.Write() returned an unexpected error: %v", err)
	}

	for {
		message := readServerMessage(t, ctx, connection)
		if message.Type == protocol.MessageError {
			if message.Error == nil || message.Error.Code != "invalid_message" {
				t.Fatalf("error message = %+v", message)
			}
			break
		}
	}
}

func TestHandlerRegistersAndRecoversAccountCharacter(t *testing.T) {
	world := simulation.NewWorld()
	const interval = 5 * time.Millisecond
	runner, err := simulation.NewRunner(world, interval)
	if err != nil {
		t.Fatalf("NewRunner() error = %v", err)
	}
	runnerCtx, stopRunner := context.WithCancel(context.Background())
	defer stopRunner()
	go runner.Run(runnerCtx)

	handler, err := NewHandler(world, interval)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	server := startTestServer(t, handler)
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	guestConnection, _, err := websocket.Dial(ctx, websocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial guest: %v", err)
	}
	authenticateGuest(t, ctx, guestConnection, "")
	guestWelcome := readServerMessage(t, ctx, guestConnection)
	if err := wsjson.Write(ctx, guestConnection, protocol.ClientMessage{
		Type: protocol.MessageRegisterAccount, Sequence: 1,
		Email: "player@example.com", Password: "strong-password",
	}); err != nil {
		t.Fatalf("write registration: %v", err)
	}
	var registration protocol.ServerMessage
	for registration.Type != protocol.MessageAccountSession {
		registration = readServerMessage(t, ctx, guestConnection)
	}
	if registration.SessionToken == "" || registration.AccountEmail != "player@example.com" {
		t.Fatalf("registration response = %+v", registration)
	}
	if err := guestConnection.Close(websocket.StatusNormalClosure, "registered"); err != nil {
		t.Fatalf("close guest connection: %v", err)
	}
	waitForPlayerToLeave(t, world, simulation.PlayerID(guestWelcome.PlayerID))

	loginConnection, _, err := websocket.Dial(ctx, websocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial login: %v", err)
	}
	if err := wsjson.Write(ctx, loginConnection, protocol.ClientMessage{
		Type:  protocol.MessageAuthenticate,
		Email: "PLAYER@example.com", Password: "strong-password",
	}); err != nil {
		t.Fatalf("write login: %v", err)
	}
	loginWelcome := readServerMessage(t, ctx, loginConnection)
	if loginWelcome.PlayerID != guestWelcome.PlayerID || loginWelcome.SessionToken == "" {
		t.Fatalf("login welcome = %+v, original player = %s", loginWelcome, guestWelcome.PlayerID)
	}
	if err := loginConnection.Close(websocket.StatusNormalClosure, "login complete"); err != nil {
		t.Fatalf("close login connection: %v", err)
	}
	waitForPlayerToLeave(t, world, simulation.PlayerID(guestWelcome.PlayerID))

	sessionConnection, _, err := websocket.Dial(ctx, websocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial session: %v", err)
	}
	defer sessionConnection.CloseNow()
	if err := wsjson.Write(ctx, sessionConnection, protocol.ClientMessage{
		Type: protocol.MessageAuthenticate, SessionToken: loginWelcome.SessionToken,
	}); err != nil {
		t.Fatalf("write session authentication: %v", err)
	}
	sessionWelcome := readServerMessage(t, ctx, sessionConnection)
	if sessionWelcome.PlayerID != guestWelcome.PlayerID {
		t.Fatalf("session player = %s, want %s", sessionWelcome.PlayerID, guestWelcome.PlayerID)
	}
	if err := wsjson.Write(ctx, sessionConnection, protocol.ClientMessage{
		Type: protocol.MessageLogoutAccount, Sequence: 1,
	}); err != nil {
		t.Fatalf("write logout: %v", err)
	}
	var logout protocol.ServerMessage
	for logout.Type != protocol.MessageAccountLoggedOut {
		logout = readServerMessage(t, ctx, sessionConnection)
	}
	if err := sessionConnection.Close(websocket.StatusNormalClosure, "logged out"); err != nil {
		t.Fatalf("close logged-out connection: %v", err)
	}
	waitForPlayerToLeave(t, world, simulation.PlayerID(guestWelcome.PlayerID))

	revokedConnection, _, err := websocket.Dial(ctx, websocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial revoked session: %v", err)
	}
	defer revokedConnection.CloseNow()
	if err := wsjson.Write(ctx, revokedConnection, protocol.ClientMessage{
		Type: protocol.MessageAuthenticate, SessionToken: loginWelcome.SessionToken,
	}); err != nil {
		t.Fatalf("write revoked session authentication: %v", err)
	}
	revokedResponse := readServerMessage(t, ctx, revokedConnection)
	if revokedResponse.Type != protocol.MessageError || revokedResponse.Error == nil ||
		revokedResponse.Error.Code != "invalid_account_session" {
		t.Fatalf("revoked session response = %+v", revokedResponse)
	}
}

func waitForPlayerToLeave(t *testing.T, world *simulation.World, playerID simulation.PlayerID) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if _, found := world.Player(playerID); !found {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("player %s did not leave the world", playerID)
}

func TestHandlerRejectsDebugCommandsWhenDisabled(t *testing.T) {
	world := simulation.NewWorld()
	handler, err := NewHandlerWithOptions(
		world, 5*time.Millisecond, HandlerOptions{AllowDebugCommands: false},
	)
	if err != nil {
		t.Fatalf("NewHandlerWithOptions() error = %v", err)
	}
	server := startTestServer(t, handler)
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	connection, _, err := websocket.Dial(ctx, websocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("websocket.Dial() error = %v", err)
	}
	defer connection.CloseNow()
	authenticateGuest(t, ctx, connection, "")
	_ = readServerMessage(t, ctx, connection)
	if err := wsjson.Write(ctx, connection, protocol.ClientMessage{
		Type: protocol.MessageDebugSetSkill, Sequence: 1, SkillID: "attack", Level: 50,
	}); err != nil {
		t.Fatalf("write debug command: %v", err)
	}
	for {
		message := readServerMessage(t, ctx, connection)
		if message.Type != protocol.MessageError {
			continue
		}
		if message.Error == nil || message.Error.Code != "debug_skill_rejected" {
			t.Fatalf("error message = %+v", message)
		}
		break
	}
}

func TestHandlerRejectsGuestAuthenticationWhenAccountsAreRequired(t *testing.T) {
	handler, err := NewHandlerWithOptions(
		simulation.NewWorld(),
		5*time.Millisecond,
		HandlerOptions{RequireAccounts: true},
	)
	if err != nil {
		t.Fatalf("NewHandlerWithOptions() error = %v", err)
	}
	server := startTestServer(t, handler)
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	connection, _, err := websocket.Dial(ctx, websocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("websocket.Dial() error = %v", err)
	}
	defer connection.CloseNow()
	authenticateGuest(t, ctx, connection, "")
	message := readServerMessage(t, ctx, connection)
	if message.Type != protocol.MessageError || message.Error == nil ||
		message.Error.Code != "account_required" {
		t.Fatalf("guest rejection = %+v", message)
	}
}

func TestHandlerSavesCharacterOnDisconnect(t *testing.T) {
	world := simulation.NewWorld()
	const interval = 5 * time.Millisecond
	runner, err := simulation.NewRunner(world, interval)
	if err != nil {
		t.Fatalf("NewRunner() returned an unexpected error: %v", err)
	}
	runnerCtx, stopRunner := context.WithCancel(context.Background())
	defer stopRunner()
	go runner.Run(runnerCtx)

	guestIdentities, err := identity.NewGuestManager(identity.NewMemoryGuestRepository())
	if err != nil {
		t.Fatalf("NewGuestManager() returned an unexpected error: %v", err)
	}
	characters := character.NewMemoryRepository()
	handler, err := NewHandlerWithRepositories(world, interval, guestIdentities, characters)
	if err != nil {
		t.Fatalf("NewHandlerWithRepositories() returned an unexpected error: %v", err)
	}
	server := startTestServer(t, handler)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	connection, _, err := websocket.Dial(ctx, websocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("websocket.Dial() returned an unexpected error: %v", err)
	}
	authenticateGuest(t, ctx, connection, "")
	welcome := readServerMessage(t, ctx, connection)
	if err := wsjson.Write(ctx, connection, protocol.ClientMessage{
		Type: protocol.MessageMoveTo, Sequence: 1,
		Destination: &protocol.Position{Y: 10},
	}); err != nil {
		t.Fatalf("write movement: %v", err)
	}

	for {
		message := readServerMessage(t, ctx, connection)
		if message.Type != protocol.MessageSnapshot || message.Snapshot == nil {
			continue
		}
		player, found := findPlayer(message.Snapshot, welcome.PlayerID)
		if found && player.Position.Y > 0 {
			break
		}
	}
	if err := connection.Close(websocket.StatusNormalClosure, "test complete"); err != nil {
		t.Fatalf("close connection: %v", err)
	}

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		state, found, loadErr := characters.Load(ctx, welcome.PlayerID)
		if loadErr != nil {
			t.Fatalf("load saved character: %v", loadErr)
		}
		if found && state.Position.Y > 0 {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("character movement was not saved after disconnect")
}

func TestSaveAllPersistsEveryActiveCharacter(t *testing.T) {
	world := simulation.NewWorld()
	if err := world.AddPlayer("player-one", simulation.Position{X: 2}, 4); err != nil {
		t.Fatalf("AddPlayer(player-one): %v", err)
	}
	if err := world.AddPlayer("player-two", simulation.Position{Z: -3}, 4); err != nil {
		t.Fatalf("AddPlayer(player-two): %v", err)
	}
	characters := character.NewMemoryRepository()
	guestIdentities, err := identity.NewGuestManager(identity.NewMemoryGuestRepository())
	if err != nil {
		t.Fatalf("NewGuestManager() returned an unexpected error: %v", err)
	}
	handler, err := NewHandlerWithRepositories(
		world,
		5*time.Millisecond,
		guestIdentities,
		characters,
	)
	if err != nil {
		t.Fatalf("NewHandlerWithRepositories() returned an unexpected error: %v", err)
	}

	if err := handler.SaveAll(context.Background()); err != nil {
		t.Fatalf("SaveAll() returned an unexpected error: %v", err)
	}
	for _, playerID := range []string{"player-one", "player-two"} {
		if _, found, loadErr := characters.Load(context.Background(), playerID); loadErr != nil || !found {
			t.Fatalf("Load(%s) = found %t, error %v", playerID, found, loadErr)
		}
	}
}

func authenticateGuest(t *testing.T, ctx context.Context, connection *websocket.Conn, token string) {
	t.Helper()
	if err := wsjson.Write(ctx, connection, protocol.ClientMessage{
		Type:       protocol.MessageAuthenticate,
		GuestToken: token,
	}); err != nil {
		t.Fatalf("write authentication message: %v", err)
	}
}

func TestSnapshotMessageIncludesCombatEvents(t *testing.T) {
	message := snapshotMessage(simulation.Snapshot{
		Targets: []simulation.TargetSnapshot{
			{
				ID:              "dummy",
				Kind:            "creature",
				TargetID:        "player-one",
				NextAttackPulse: 16,
				Defeated:        true,
				RespawnAtPulse:  20,
				DodgeRate:       0.2,
				Accuracy:        17,
				Evasion:         9,
			},
			{
				ID:            "upper_dummy",
				Position:      simulation.Position{X: -4, Y: 3, Z: 2},
				Health:        75,
				MaximumHealth: 100,
				DodgeRate:     0.1,
			},
		},
		Drops: []simulation.DropSnapshot{{
			ID:       "drop-1",
			Position: simulation.Position{X: 2, Y: 3, Z: 4},
			ItemID:   "coins",
			Quantity: 2,
		}},
		CombatEvents: []simulation.CombatEvent{{
			ID:       7,
			Pulse:    12,
			Type:     simulation.CombatEventDamage,
			SourceID: "player-one",
			TargetID: "dummy",
			Amount:   10,
			SpellID:  simulation.SpellWindBolt,
		}},
		Players: []simulation.PlayerSnapshot{{
			ID:                      "player-one",
			Health:                  40,
			MaximumHealth:           100,
			Prayer:                  70,
			MaximumPrayer:           100,
			Defeated:                true,
			RespawnAtPulse:          20,
			EquippedWeapon:          simulation.WeaponDagger,
			EquippedSpellbook:       simulation.ItemStandardSpellbook,
			ActiveSpellbook:         simulation.SpellbookStandard,
			LastEquipmentSequence:   3,
			LastEquipmentPulse:      11,
			CombatStyle:             simulation.CombatStyleStab,
			LastCombatStyleSequence: 4,
			ActivePrayers:           []simulation.PrayerID{simulation.PrayerEvasion},
			LastPrayerSequence:      5,
			SelectedSpell:           simulation.SpellWindBolt,
			LastSpellSequence:       6,
			DodgeRate:               0.15,
			CombatStats: simulation.CombatStatsSnapshot{
				Weapon: simulation.WeaponDagger, Style: simulation.CombatStyleStab, Spell: simulation.SpellWindBolt,
				AccuracySkill: simulation.SkillAttack, AccuracyLevel: 2,
				WeaponAccuracyBonus: 14, StyleAccuracyBonus: 2, SpellAccuracyBonus: 10, PrayerHitChanceBonus: 0.05, TotalAccuracy: 18,
				DamageSkill: simulation.SkillStrength, DamageLevel: 3, StrengthBonus: 7, SpellPower: 4, SkillDamageBonus: 1,
				PrayerDamageBonus: 0.1, PrayerDamageReduction: 0.2,
				BaseAttackInterval: 4, AttackInterval: 3, AttackIntervalModifier: -1,
				MinimumDamage: 5, MaximumDamage: 9, DefenceLevel: 4,
				DefenceBonuses: simulation.DefenceBonuses{Stab: 2, Slash: 3, Crush: 4, Ranged: 5, Magic: -1},
			},
			Inventory: simulation.InventorySnapshot{
				Capacity: 8,
				Items: []simulation.InventoryItem{{
					Slot: 0, ItemID: "sword", Quantity: 1,
				}},
			},
			Skills: []simulation.SkillSnapshot{{
				ID: simulation.SkillAttack, Experience: 125, Level: 3,
				CurrentLevelExperience: 87, NextLevelExperience: 138,
			}},
		}},
	}, "player-one")

	if len(message.CombatEvents) != 1 {
		t.Fatalf("combat event count = %d, want 1", len(message.CombatEvents))
	}
	event := message.CombatEvents[0]
	if event.ID != 7 || event.Pulse != 12 || event.Type != "damage" ||
		event.SourceID != "player-one" || event.TargetID != "dummy" || event.Amount != 10 || event.SpellID != "wind_bolt" {
		t.Fatalf("combat event = %+v", event)
	}
	dummy, found := findTarget(message, "dummy")
	if !found || dummy.Kind != "creature" || dummy.TargetID != "player-one" || dummy.NextAttackPulse != 16 ||
		!dummy.Defeated || dummy.RespawnAtPulse != 20 || dummy.DodgeRate != 0.2 ||
		dummy.Accuracy != 17 || dummy.Evasion != 9 {
		t.Fatalf("dummy combat state = %+v, found %t", dummy, found)
	}
	upperDummy, found := findTarget(message, "upper_dummy")
	if !found || upperDummy.Position.X != -4 || upperDummy.Position.Y != 3 ||
		upperDummy.Position.Z != 2 || upperDummy.Health != 75 ||
		upperDummy.MaximumHealth != 100 || upperDummy.DodgeRate != 0.1 {
		t.Fatalf("upper dummy combat state = %+v, found %t", upperDummy, found)
	}
	if len(message.Drops) != 1 || message.Drops[0].ID != "drop-1" ||
		message.Drops[0].Position.X != 2 || message.Drops[0].Position.Y != 3 ||
		message.Drops[0].Position.Z != 4 || message.Drops[0].ItemID != "coins" ||
		message.Drops[0].Quantity != 2 {
		t.Fatalf("drop snapshot = %+v", message.Drops)
	}
	player := message.Players[0]
	if player.Health != 40 || player.MaximumHealth != 100 || player.Prayer != 70 || player.MaximumPrayer != 100 ||
		!player.Defeated || player.RespawnAtPulse != 20 ||
		player.EquippedWeapon != "dagger" || player.EquippedSpellbook != "standard_spellbook" ||
		player.ActiveSpellbook != "standard" || player.LastEquipmentSequence != 3 || player.LastEquipmentPulse != 11 ||
		player.CombatStyle != "stab" || player.LastCombatStyleSequence != 4 ||
		len(player.ActivePrayers) != 1 || player.ActivePrayers[0] != "evasion" ||
		player.LastPrayerSequence != 5 || player.DodgeRate != 0.15 {
		t.Fatalf("player vital state = %+v", player)
	}
	if player.SelectedSpell != "wind_bolt" || player.LastSpellSequence != 6 {
		t.Fatalf("player spell state = %+v", player)
	}
	if player.Inventory.Capacity != 8 || len(player.Inventory.Items) != 1 ||
		player.Inventory.Items[0].Slot != 0 || player.Inventory.Items[0].ItemID != "sword" ||
		player.Inventory.Items[0].Quantity != 1 {
		t.Fatalf("player inventory = %+v", player.Inventory)
	}
	if player.CombatStats.Weapon != "dagger" || player.CombatStats.Style != "stab" || player.CombatStats.Spell != "wind_bolt" ||
		player.CombatStats.AccuracySkill != "attack" || player.CombatStats.AccuracyLevel != 2 ||
		player.CombatStats.WeaponAccuracyBonus != 14 || player.CombatStats.StyleAccuracyBonus != 2 || player.CombatStats.SpellAccuracyBonus != 10 || player.CombatStats.PrayerHitChanceBonus != 0.05 ||
		player.CombatStats.TotalAccuracy != 18 || player.CombatStats.DamageSkill != "strength" ||
		player.CombatStats.DamageLevel != 3 || player.CombatStats.StrengthBonus != 7 || player.CombatStats.SpellPower != 4 || player.CombatStats.SkillDamageBonus != 1 ||
		player.CombatStats.PrayerDamageBonus != 0.1 || player.CombatStats.PrayerDamageReduction != 0.2 ||
		player.CombatStats.BaseAttackInterval != 4 || player.CombatStats.AttackInterval != 3 || player.CombatStats.AttackIntervalModifier != -1 ||
		player.CombatStats.MinimumDamage != 5 || player.CombatStats.MaximumDamage != 9 ||
		player.CombatStats.DefenceLevel != 4 || player.CombatStats.DefenceBonuses.Magic != -1 {
		t.Fatalf("player combat stats = %+v", player.CombatStats)
	}
	if len(player.Skills) != 1 || player.Skills[0].ID != "attack" ||
		player.Skills[0].Experience != 125 || player.Skills[0].Level != 3 ||
		player.Skills[0].CurrentLevelExperience != 87 || player.Skills[0].NextLevelExperience != 138 {
		t.Fatalf("player skills = %+v", player.Skills)
	}
}

func TestSnapshotMessageOnlyIncludesLocalPlayersBank(t *testing.T) {
	bank := simulation.BankSnapshot{
		Capacity: simulation.BankCapacity,
		Items:    []simulation.InventoryItem{{Slot: 0, ItemID: simulation.ItemAirRune, Quantity: 50}},
	}
	message := snapshotMessage(simulation.Snapshot{Players: []simulation.PlayerSnapshot{
		{ID: "player-one", Bank: bank},
		{ID: "player-two", Bank: bank},
	}}, "player-one")
	if message.Players[0].Bank == nil || len(message.Players[0].Bank.Items) != 1 {
		t.Fatalf("local bank = %+v", message.Players[0].Bank)
	}
	if message.Players[1].Bank != nil {
		t.Fatalf("remote bank was exposed: %+v", message.Players[1].Bank)
	}
}

func readServerMessage(t *testing.T, ctx context.Context, connection *websocket.Conn) protocol.ServerMessage {
	t.Helper()
	var message protocol.ServerMessage
	if err := wsjson.Read(ctx, connection, &message); err != nil {
		t.Fatalf("wsjson.Read() returned an unexpected error: %v", err)
	}
	return message
}

func startTestServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Skipf("local sockets are unavailable: %v", err)
	}
	server := httptest.NewUnstartedServer(handler)
	server.Listener = listener
	server.Start()
	return server
}

func websocketURL(serverURL string) string {
	return "ws" + strings.TrimPrefix(serverURL, "http")
}

func findPlayer(snapshot *protocol.WorldSnapshot, id string) (protocol.PlayerSnapshot, bool) {
	for _, player := range snapshot.Players {
		if player.ID == id {
			return player, true
		}
	}
	return protocol.PlayerSnapshot{}, false
}

func findTarget(snapshot *protocol.WorldSnapshot, id string) (protocol.TargetSnapshot, bool) {
	for _, target := range snapshot.Targets {
		if target.ID == id {
			return target, true
		}
	}
	return protocol.TargetSnapshot{}, false
}
