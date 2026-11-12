package simulation

import (
	"reflect"
	"testing"
	"time"
)

func TestPersistentPlayerStateRoundTripsEveryDurableSubsystem(t *testing.T) {
	originalWorld := NewWorld()
	state := DefaultPersistentPlayerState()
	state.Position = Position{X: 3, Y: 2, Z: -1}
	state.Health = 91
	state.Prayer = 37
	state.ActivePrayerSet = PrayerSetWar
	state.EquippedWeapon = WeaponRuneSword
	state.EquippedWeaponInstanceID = "equipped-rune-sword"
	state.EquippedWeaponReinforcement = 5
	state.CombatStyle = CombatStyleSlash
	state.EquippedArmour = map[ArmourSlot]string{
		ArmourSlotHead: ItemIronHelm, ArmourSlotBody: ItemIronPlate, ArmourSlotLeftHand: ItemIronShield,
	}
	state.EquippedArmourInstances = map[ArmourSlot]string{
		ArmourSlotHead: "equipped-iron-helm", ArmourSlotBody: "equipped-iron-plate", ArmourSlotLeftHand: "equipped-iron-shield",
	}
	state.EquippedArmourReinforcements = map[ArmourSlot]int{
		ArmourSlotHead: 2, ArmourSlotBody: 3, ArmourSlotLeftHand: 1,
	}
	state.EquippedSpellbook = ItemHolyBook
	state.Inventory = InventorySnapshot{Capacity: 35, Items: []InventoryItem{
		{Slot: 0, ItemID: ItemCoins, Quantity: 12_345},
		{Slot: 7, ItemID: string(WeaponDagger), Quantity: 1, InstanceID: "inventory-dagger", Reinforcement: 2},
	}}
	state.AmmunitionPouch = AmmunitionPouchSnapshot{
		Capacity: AmmunitionPouchCapacity, UnlockedSlots: 4,
		Items: []InventoryItem{
			{Slot: 0, ItemID: ItemBronzeArrow, Quantity: 125},
			{Slot: 3, ItemID: ItemIronBolt, Quantity: 80},
		},
	}
	state.RingRack = RingRackSnapshot{Capacity: RingRackCapacity, Items: []InventoryItem{
		{Slot: 0, ItemID: ItemTyrantsRing, Quantity: 1, InstanceID: "tyrant-ring"},
		{Slot: 4, ItemID: ItemEmeraldRing, Quantity: 1, InstanceID: "emerald-ring"},
	}}
	state.RunePouch = RunePouchSnapshot{
		Capacity: RunePouchCapacity, UnlockedSlots: 4,
		Items: []InventoryItem{
			{Slot: 0, ItemID: ItemAirRune, Quantity: 500},
			{Slot: 3, ItemID: ItemFocusRune, Quantity: 75},
		},
	}
	state.Bank = BankSnapshot{Capacity: BankCapacity + BankExpansionSize, Items: []InventoryItem{
		{Slot: 12, ItemID: ItemWaterRune, Quantity: 2_500},
		{Slot: 80, ItemID: string(WeaponSteelSword), Quantity: 1, InstanceID: "bank-steel-sword", Reinforcement: 4},
	}}
	state.ExperienceLocked = true
	for index := range state.Skills {
		switch state.Skills[index].ID {
		case SkillAttack:
			state.Skills[index].Experience = SkillExperienceForLevel(22)
		case SkillAgility:
			state.Skills[index].Experience = SkillExperienceForLevel(14)
		case SkillSmithing:
			state.Skills[index].Experience = SkillExperienceForLevel(9)
		}
	}

	if err := originalWorld.QueueJoinWithPersistentState("original", state, DefaultPlayerMovementSpeed); err != nil {
		t.Fatalf("join original character: %v", err)
	}
	originalWorld.Step(time.Millisecond)
	original := mustPlayer(t, originalWorld, "original")
	saved := PersistentStateFromSnapshot(original)

	restoredWorld := NewWorld()
	if err := restoredWorld.QueueJoinWithPersistentState("restored", saved, DefaultPlayerMovementSpeed); err != nil {
		t.Fatalf("join restored character: %v", err)
	}
	restoredWorld.Step(time.Millisecond)
	restored := PersistentStateFromSnapshot(mustPlayer(t, restoredWorld, "restored"))

	if !reflect.DeepEqual(restored, saved) {
		t.Fatalf("restored durable state differs\nrestored: %+v\n   saved: %+v", restored, saved)
	}
}

func TestQueueJoinWithPersistentStateRestoresCharacter(t *testing.T) {
	world := NewWorld()
	state := DefaultPersistentPlayerState()
	state.Position = Position{X: 3, Y: 2, Z: -1}
	state.Health = 64
	state.EquippedWeapon = WeaponBow
	state.CombatStyle = CombatStyleLongRange
	state.Inventory.Items = []InventoryItem{{Slot: 0, ItemID: string(WeaponBow), Quantity: 1}}
	state.Skills = []PersistentSkillState{{ID: SkillAttack, Experience: 125}}
	state.ExperienceLocked = true
	state.Bank.Items = []InventoryItem{{Slot: 0, ItemID: ItemAirRune, Quantity: 250}}
	state.AmmunitionPouch.Items = []InventoryItem{{Slot: 0, ItemID: ItemIronArrow, Quantity: 80}}
	state.RingRack.Items = []InventoryItem{{Slot: 0, ItemID: ItemRubyRing, Quantity: 1, InstanceID: "ring-1"}}
	state.RunePouch.Items = []InventoryItem{{Slot: 0, ItemID: ItemFireRune, Quantity: 60}}

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
	if !player.ExperienceLocked {
		t.Fatal("experience lock was not restored")
	}
	if snapshotInventoryContains(player.Inventory, string(WeaponBow)) {
		t.Fatal("restored equipped bow still occupies an inventory slot")
	}
	if attack := mustSkill(t, player, SkillAttack); attack.Experience != 125 || attack.Level != SkillLevelForExperience(125) {
		t.Fatalf("restored Attack = %+v", attack)
	}
	if agility := mustSkill(t, player, SkillAgility); agility.Experience != 0 || agility.Level != 1 {
		t.Fatalf("defaulted Agility = %+v", agility)
	}
	if len(player.Bank.Items) != 1 || player.Bank.Items[0].Quantity != 250 {
		t.Fatalf("restored bank = %+v", player.Bank)
	}
	if len(player.AmmunitionPouch.Items) != 1 || player.AmmunitionPouch.Items[0].Quantity != 80 {
		t.Fatalf("restored ammunition pouch = %+v", player.AmmunitionPouch)
	}
	if len(player.RingRack.Items) != 1 || player.RingRack.Items[0].ItemID != ItemRubyRing {
		t.Fatalf("restored ring rack = %+v", player.RingRack)
	}
	if len(player.RunePouch.Items) != 1 || player.RunePouch.Items[0].Quantity != 60 {
		t.Fatalf("restored rune pouch = %+v", player.RunePouch)
	}
}

func TestQueueJoinMigratesLegacySpearSlashStyle(t *testing.T) {
	world := NewWorld()
	state := DefaultPersistentPlayerState()
	state.EquippedWeapon = WeaponSpear
	state.CombatStyle = CombatStyleSlash
	state.Inventory.Items = []InventoryItem{{Slot: 0, ItemID: string(WeaponSpear), Quantity: 1}}

	if err := world.QueueJoinWithPersistentState("player-one", state, 4); err != nil {
		t.Fatalf("QueueJoinWithPersistentState() returned an unexpected error: %v", err)
	}
	world.Step(time.Millisecond)
	if got := mustPlayer(t, world, "player-one").CombatStyle; got != CombatStyleHeavyStab {
		t.Fatalf("restored spear style = %q, want %q", got, CombatStyleHeavyStab)
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

func TestQueueJoinUpgradesLegacyInventoryCapacity(t *testing.T) {
	world := NewWorld()
	state := DefaultPersistentPlayerState()
	state.Inventory.Capacity = 8
	state.Inventory.Items = []InventoryItem{{Slot: 7, ItemID: ItemCoins, Quantity: 1}}

	if err := world.QueueJoinWithPersistentState("player-one", state, 4); err != nil {
		t.Fatalf("QueueJoinWithPersistentState() returned an unexpected error: %v", err)
	}
	world.Step(time.Millisecond)
	player := mustPlayer(t, world, "player-one")
	if player.Inventory.Capacity != BaseInventoryCapacity {
		t.Fatalf("restored inventory capacity = %d, want %d", player.Inventory.Capacity, BaseInventoryCapacity)
	}
	if player.Inventory.itemsQuantity(ItemCoins) != 1 {
		t.Fatalf("restored inventory lost its items: %+v", player.Inventory)
	}
}

func TestQueueJoinDerivesMovementSpeedFromPersistedAgility(t *testing.T) {
	world := NewWorld()
	state := DefaultPersistentPlayerState()
	for index := range state.Skills {
		if state.Skills[index].ID == SkillAgility {
			state.Skills[index].Experience = SkillExperienceForLevel(20)
		}
	}
	if err := world.QueueJoinWithPersistentState("player-one", state, 6); err != nil {
		t.Fatalf("QueueJoinWithPersistentState() returned an unexpected error: %v", err)
	}
	world.Step(time.Millisecond)
	if got, want := mustPlayer(t, world, "player-one").Speed, movementSpeedForAgility(6, 20); got != want {
		t.Fatalf("restored movement speed = %.3f, want %.3f", got, want)
	}
}

func TestQueueJoinClampsHealthSavedUnderOlderVitalityCurve(t *testing.T) {
	world := NewWorld()
	state := DefaultPersistentPlayerState()
	state.Health = 200
	for index := range state.Skills {
		if state.Skills[index].ID == SkillVitality {
			state.Skills[index].Experience = SkillExperienceForLevel(2)
		}
	}

	if err := world.QueueJoinWithPersistentState("player-one", state, 4); err != nil {
		t.Fatalf("QueueJoinWithPersistentState() returned an unexpected error: %v", err)
	}
	world.Step(time.Millisecond)
	player := mustPlayer(t, world, "player-one")
	if got, want := player.MaximumHealth, 110; got != want {
		t.Fatalf("restored maximum health = %d, want %d", got, want)
	}
	if got, want := player.Health, 110; got != want {
		t.Fatalf("restored health = %d, want clamped %d", got, want)
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
	if snapshotInventoryContains(player.Inventory, ItemStandardSpellbook) {
		t.Fatal("restored equipped spellbook still occupies an inventory slot")
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
