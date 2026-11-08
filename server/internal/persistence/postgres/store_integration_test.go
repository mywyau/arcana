package postgres

import (
	"context"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/mywyau/arcana/server/internal/account"
	"github.com/mywyau/arcana/server/internal/identity"
	"github.com/mywyau/arcana/server/internal/simulation"
)

func TestStoreGuestAndCharacterRoundTrip(t *testing.T) {
	databaseURL := os.Getenv("ARCANA_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("ARCANA_TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	store, err := Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("Open() returned an unexpected error: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := store.Close(); closeErr != nil {
			t.Errorf("close integration store: %v", closeErr)
		}
	})
	if err := store.Migrate(ctx); err != nil {
		t.Fatalf("Migrate() returned an unexpected error: %v", err)
	}
	if err := store.Migrate(ctx); err != nil {
		t.Fatalf("second Migrate() returned an unexpected error: %v", err)
	}

	manager, err := identity.NewGuestManager(store)
	if err != nil {
		t.Fatalf("NewGuestManager() returned an unexpected error: %v", err)
	}
	guest, err := manager.Authenticate(ctx, "")
	if err != nil {
		t.Fatalf("create guest identity: %v", err)
	}
	accountID := ""
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if accountID != "" {
			if _, cleanupErr := store.database.ExecContext(
				cleanupCtx,
				"DELETE FROM accounts WHERE account_id = $1",
				accountID,
			); cleanupErr != nil {
				t.Errorf("clean up integration account: %v", cleanupErr)
			}
		}
		if _, cleanupErr := store.database.ExecContext(
			cleanupCtx,
			"DELETE FROM guest_identities WHERE player_id = $1",
			guest.PlayerID,
		); cleanupErr != nil {
			t.Errorf("clean up integration guest: %v", cleanupErr)
		}
	})

	want := simulation.DefaultPersistentPlayerState()
	want.Position = simulation.Position{X: 7.5, Y: 3, Z: -2.25}
	want.Health = 73
	want.EquippedWeapon = simulation.WeaponBow
	want.EquippedWeaponInstanceID = "integration-bow-instance"
	want.EquippedWeaponReinforcement = 1
	want.CombatStyle = simulation.CombatStyleLongRange
	want.Skills = []simulation.PersistentSkillState{{ID: simulation.SkillAttack, Experience: 125}}
	want.Inventory.Items = []simulation.InventoryItem{
		{Slot: 0, ItemID: "coins", Quantity: 3},
	}
	want.Bank.Items = []simulation.InventoryItem{
		{Slot: 0, ItemID: simulation.ItemAirRune, Quantity: 250},
	}
	want.RingRack.Items = []simulation.InventoryItem{
		{Slot: 0, ItemID: simulation.ItemRubyRing, Quantity: 1, InstanceID: "integration-ring-instance"},
	}
	want.RunePouch.Items = []simulation.InventoryItem{
		{Slot: 0, ItemID: simulation.ItemFireRune, Quantity: 75},
	}
	want.RunePouch.UnlockedSlots = 4
	want.EquippedSpellbook = simulation.ItemStandardSpellbook
	if err := store.Save(ctx, guest.PlayerID, want); err != nil {
		t.Fatalf("Save() returned an unexpected error: %v", err)
	}

	secondStore, err := Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open second store: %v", err)
	}
	defer secondStore.Close()
	loaded, found, err := secondStore.Load(ctx, guest.PlayerID)
	if err != nil || !found {
		t.Fatalf("Load() = found %t, error %v", found, err)
	}
	if !reflect.DeepEqual(loaded, want) {
		t.Fatalf("loaded state = %+v, want %+v", loaded, want)
	}

	accountID = "account-" + strings.TrimPrefix(guest.PlayerID, "guest-")
	passwordHash, err := account.NewPasswordHasher().Hash("integration-test-password")
	if err != nil {
		t.Fatalf("hash integration password: %v", err)
	}
	wantAccount := account.Account{
		ID:           accountID,
		Email:        accountID + "@example.invalid",
		PasswordHash: passwordHash,
		PlayerID:     guest.PlayerID,
	}
	if err := store.CreateAccount(ctx, wantAccount); err != nil {
		t.Fatalf("CreateAccount() returned an unexpected error: %v", err)
	}
	loadedAccount, found, err := secondStore.FindAccountByEmail(ctx, strings.ToUpper(wantAccount.Email))
	if err != nil || !found {
		t.Fatalf("FindAccountByEmail() = found %t, error %v", found, err)
	}
	if !reflect.DeepEqual(loadedAccount, wantAccount) {
		t.Fatalf("loaded account = %+v, want %+v", loadedAccount, wantAccount)
	}

	secondManager, err := identity.NewGuestManager(secondStore)
	if err != nil {
		t.Fatalf("create second guest manager: %v", err)
	}
	resumed, err := secondManager.Authenticate(ctx, guest.Token)
	if err != nil {
		t.Fatalf("resume guest identity: %v", err)
	}
	if resumed.PlayerID != guest.PlayerID || resumed.Token != "" {
		t.Fatalf("resumed identity = %+v, want player ID %q", resumed, guest.PlayerID)
	}
}
