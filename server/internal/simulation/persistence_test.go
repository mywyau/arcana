package simulation

import (
	"testing"
	"time"
)

func TestQueueJoinWithPersistentStateRestoresCharacter(t *testing.T) {
	world := NewWorld()
	state := DefaultPersistentPlayerState()
	state.Position = Position{X: 3, Y: 2, Z: -1}
	state.Health = 64
	state.EquippedWeapon = WeaponBow
	state.CombatStyle = CombatStyleLongRange
	state.Inventory.Items = []InventoryItem{{Slot: 0, ItemID: string(WeaponBow), Quantity: 1}}
	state.Skills = []PersistentSkillState{{ID: SkillAttack, Experience: 125}}

	if err := world.QueueJoinWithPersistentState("player-one", state, 4); err != nil {
		t.Fatalf("QueueJoinWithPersistentState() returned an unexpected error: %v", err)
	}
	world.Step(time.Millisecond)
	player, found := world.Player("player-one")
	if !found {
		t.Fatal("restored player was not added")
	}
	if player.Position != state.Position || player.Health != 64 ||
		player.EquippedWeapon != WeaponBow || player.CombatStyle != CombatStyleLongRange {
		t.Fatalf("restored player = %+v", player)
	}
	if attack := mustSkill(t, player, SkillAttack); attack.Experience != 125 || attack.Level != 2 {
		t.Fatalf("restored Attack = %+v", attack)
	}
	if agility := mustSkill(t, player, SkillAgility); agility.Experience != 0 || agility.Level != 1 {
		t.Fatalf("defaulted Agility = %+v", agility)
	}
}

func TestQueueJoinWithPersistentStateRejectsInvalidEquipment(t *testing.T) {
	world := NewWorld()
	state := DefaultPersistentPlayerState()
	state.EquippedWeapon = WeaponBow
	state.CombatStyle = CombatStyleLongRange

	if err := world.QueueJoinWithPersistentState("player-one", state, 4); err == nil {
		t.Fatal("QueueJoinWithPersistentState() succeeded without equipped weapon in inventory")
	}
}

func TestQueueJoinWithPersistentStateAllowsInnateUnarmedProfile(t *testing.T) {
	world := NewWorld()
	state := DefaultPersistentPlayerState()
	state.EquippedWeapon = WeaponUnarmed
	state.CombatStyle = CombatStyleCrush

	if err := world.QueueJoinWithPersistentState("player-one", state, 4); err != nil {
		t.Fatalf("QueueJoinWithPersistentState() returned an unexpected error: %v", err)
	}
}

func TestQueueJoinWithPersistentStateRestoresEquippedSpellbook(t *testing.T) {
	world := NewWorld()
	state := DefaultPersistentPlayerState()
	state.Inventory.Items = []InventoryItem{{Slot: 0, ItemID: ItemStandardSpellbook, Quantity: 1}}
	state.EquippedSpellbook = ItemStandardSpellbook

	if err := world.QueueJoinWithPersistentState("player-one", state, 4); err != nil {
		t.Fatalf("QueueJoinWithPersistentState() returned an unexpected error: %v", err)
	}
	world.Step(time.Millisecond)
	player := mustPlayer(t, world, "player-one")
	if player.EquippedSpellbook != ItemStandardSpellbook || player.ActiveSpellbook != SpellbookStandard {
		t.Fatalf("restored spellbook state = %+v", player)
	}
}

func TestPersistentStateFromDefeatedSnapshotRespawnsCharacter(t *testing.T) {
	state := PersistentStateFromSnapshot(PlayerSnapshot{
		Position: Position{X: 10, Y: 3, Z: 4},
		Health:   0,
		Defeated: true,
	})
	if state.Position != (Position{}) || state.Health != playerMaximumHealth {
		t.Fatalf("persistent defeated state = %+v", state)
	}
}
