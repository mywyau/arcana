package websockettransport

import (
	"testing"
	"time"

	"github.com/mywyau/arcana/server/internal/protocol"
)

func TestSnapshotDeltaTrackerSendsInitialStateAndKeepsLocalPlayerFull(t *testing.T) {
	tracker := newSnapshotDeltaTracker()
	now := time.Unix(100, 0)
	message := deltaTestSnapshot()

	stats := tracker.apply(message, "local", now)

	if !stats.Full || !message.FullReconciliation {
		t.Fatal("initial snapshot was not a full reconciliation")
	}
	if len(message.Players) != 1 || message.Players[0].ID != "local" || message.Players[0].Inventory == nil {
		t.Fatalf("local private player was not retained: %#v", message.Players)
	}
	if len(message.PlayerUpdates) != 1 || message.PlayerUpdates[0].ID != "remote" {
		t.Fatalf("remote player updates = %#v", message.PlayerUpdates)
	}
	if message.PlayerUpdates[0].Position == nil || message.PlayerUpdates[0].EquippedArmour == nil {
		t.Fatalf("initial remote update was incomplete: %#v", message.PlayerUpdates[0])
	}
	if len(message.Targets) != 0 || len(message.TargetUpdates) != 1 {
		t.Fatalf("targets=%d updates=%d, want 0 and 1", len(message.Targets), len(message.TargetUpdates))
	}
}

func TestSnapshotDeltaTrackerSendsOnlyChangedFields(t *testing.T) {
	tracker := newSnapshotDeltaTracker()
	now := time.Unix(200, 0)
	tracker.apply(deltaTestSnapshot(), "local", now)

	unchanged := deltaTestSnapshot()
	stats := tracker.apply(unchanged, "local", now.Add(time.Second))
	if stats.Full || len(unchanged.PlayerUpdates) != 0 || len(unchanged.TargetUpdates) != 0 {
		t.Fatalf("unchanged delta = players %#v targets %#v", unchanged.PlayerUpdates, unchanged.TargetUpdates)
	}

	changed := deltaTestSnapshot()
	changed.Players[1].Position.X = 9
	changed.Players[1].CombatTarget = ""
	changed.Targets[0].Health = 0
	stats = tracker.apply(changed, "local", now.Add(2*time.Second))
	if stats.PlayerUpdates != 1 || stats.TargetUpdates != 1 {
		t.Fatalf("delta stats = %#v", stats)
	}
	player := changed.PlayerUpdates[0]
	if player.Position == nil || player.Position.X != 9 || player.CombatTarget == nil || *player.CombatTarget != "" {
		t.Fatalf("player delta = %#v", player)
	}
	if player.Health != nil || player.EquippedWeapon != nil || player.EquippedArmour != nil {
		t.Fatalf("unchanged player fields were sent: %#v", player)
	}
	target := changed.TargetUpdates[0]
	if target.Health == nil || *target.Health != 0 {
		t.Fatalf("target health delta = %#v", target)
	}
	if target.Name != nil || target.Position != nil {
		t.Fatalf("unchanged target fields were sent: %#v", target)
	}
}

func TestSnapshotDeltaTrackerSendsRemovalsAndPeriodicFullReconciliation(t *testing.T) {
	tracker := newSnapshotDeltaTracker()
	now := time.Unix(300, 0)
	tracker.apply(deltaTestSnapshot(), "local", now)

	removed := deltaTestSnapshot()
	removed.Players = removed.Players[:1]
	removed.Targets = nil
	stats := tracker.apply(removed, "local", now.Add(time.Second))
	if stats.PlayerRemovals != 1 || len(removed.RemovedPlayerIDs) != 1 || removed.RemovedPlayerIDs[0] != "remote" {
		t.Fatalf("player removals = %#v, stats %#v", removed.RemovedPlayerIDs, stats)
	}
	if stats.TargetRemovals != 1 || len(removed.RemovedTargetIDs) != 1 || removed.RemovedTargetIDs[0] != "target" {
		t.Fatalf("target removals = %#v, stats %#v", removed.RemovedTargetIDs, stats)
	}

	reintroduced := deltaTestSnapshot()
	stats = tracker.apply(reintroduced, "local", now.Add(fullReconciliationInterval))
	if !stats.Full || len(reintroduced.PlayerUpdates) != 1 || len(reintroduced.TargetUpdates) != 1 {
		t.Fatalf("periodic reconciliation stats=%#v players=%d targets=%d", stats, len(reintroduced.PlayerUpdates), len(reintroduced.TargetUpdates))
	}
}

func TestSnapshotDeltaTrackerOmitsUnchangedPrivateContainers(t *testing.T) {
	tracker := newSnapshotDeltaTracker()
	now := time.Unix(400, 0)

	initial := deltaTestSnapshot()
	stats := tracker.apply(initial, "local", now)
	tracker.applyPrivateState(initial, "local", stats.Full)
	local := initial.Players[0]
	if local.Inventory == nil || local.Bank == nil || local.EquippedWeaponInstanceID == "" || local.EquippedArmourInstances == nil {
		t.Fatalf("initial private state was incomplete: %#v", local)
	}

	unchanged := deltaTestSnapshot()
	stats = tracker.apply(unchanged, "local", now.Add(time.Second))
	tracker.applyPrivateState(unchanged, "local", stats.Full)
	local = unchanged.Players[0]
	if local.Inventory != nil || local.Bank != nil || local.EquippedWeaponInstanceID != "" || local.EquippedArmourInstances != nil {
		t.Fatalf("unchanged private state was sent: %#v", local)
	}
	if local.EquippedWeapon != "sword" || local.EquippedArmour["head"] != "iron_helm" {
		t.Fatalf("public equipment was stripped with private equipment: %#v", local)
	}

	inventoryChanged := deltaTestSnapshot()
	inventoryChanged.Players[0].InventoryRevision++
	stats = tracker.apply(inventoryChanged, "local", now.Add(2*time.Second))
	tracker.applyPrivateState(inventoryChanged, "local", stats.Full)
	local = inventoryChanged.Players[0]
	if local.Inventory == nil || local.Bank != nil || local.EquippedWeaponInstanceID != "" || local.EquippedArmourInstances != nil {
		t.Fatalf("independent inventory revision was not respected: %#v", local)
	}

	bankChanged := deltaTestSnapshot()
	bankChanged.Players[0].InventoryRevision++
	bankChanged.Players[0].BankRevision++
	stats = tracker.apply(bankChanged, "local", now.Add(3*time.Second))
	tracker.applyPrivateState(bankChanged, "local", stats.Full)
	local = bankChanged.Players[0]
	if local.Bank == nil || local.Inventory != nil || local.EquippedWeaponInstanceID != "" {
		t.Fatalf("independent bank revision was not respected: %#v", local)
	}

	full := deltaTestSnapshot()
	stats = tracker.apply(full, "local", now.Add(fullReconciliationInterval))
	tracker.applyPrivateState(full, "local", stats.Full)
	local = full.Players[0]
	if local.Inventory == nil || local.Bank == nil || local.EquippedWeaponInstanceID == "" || local.EquippedArmourInstances == nil {
		t.Fatalf("full reconciliation omitted private state: %#v", local)
	}
}

func deltaTestSnapshot() *protocol.WorldSnapshot {
	return &protocol.WorldSnapshot{
		Players: []protocol.PlayerSnapshot{
			{
				ID: "local", Position: protocol.Position{X: 1},
				EquippedWeapon: "sword", EquippedWeaponInstanceID: "sword-instance",
				EquippedArmour:          map[string]string{"head": "iron_helm"},
				EquippedArmourInstances: map[string]string{"head": "helm-instance"},
				Inventory:               &protocol.InventorySnapshot{Capacity: 8},
				InventoryRevision:       1, EquipmentRevision: 1,
				Bank: &protocol.InventorySnapshot{Capacity: 100}, BankRevision: 1,
			},
			{
				ID: "remote", Position: protocol.Position{X: 2}, Speed: 6, CombatTarget: "target",
				Health: 20, MaximumHealth: 20, EquippedWeapon: "sword",
				EquippedArmour: map[string]string{"head": "iron_helm"}, CombatLevel: 5,
			},
		},
		Targets: []protocol.TargetSnapshot{{
			ID: "target", Name: "Target", Kind: "creature", Position: protocol.Position{X: 3},
			Health: 10, MaximumHealth: 10, Accuracy: 4, Evasion: 3, CombatLevel: 2,
		}},
	}
}
