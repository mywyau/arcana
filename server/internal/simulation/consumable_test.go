package simulation

import (
	"testing"
	"time"
)

func TestCookedMinnowHealsConsumesOneAndDelaysAttack(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	player := world.players["player-one"]
	player.health = 90
	player.nextAttackPulse = 0
	if !player.inventory.add(ItemCookedMinnow, 3) {
		t.Fatal("could not add cooked minnows")
	}
	if err := world.QueueConsumeItem(ConsumeItemCommand{
		PlayerID: "player-one", Sequence: 1, ItemID: ItemCookedMinnow,
	}); err != nil {
		t.Fatalf("QueueConsumeItem() returned an unexpected error: %v", err)
	}

	world.Step(50 * time.Millisecond)
	snapshot := mustPlayer(t, world, "player-one")
	if snapshot.Health != 98 || snapshot.Inventory.itemsQuantity(ItemCookedMinnow) != 2 {
		t.Fatalf("player after eating = %+v", snapshot)
	}
	if snapshot.LastConsumeSequence != 1 || snapshot.NextFoodConsumePulse != 2 ||
		snapshot.NextPotionConsumePulse != 0 || snapshot.NextAttackPulse != 2 {
		t.Fatalf("consume cadence after eating = %+v", snapshot)
	}
}

func TestFoodAndPotionCooldownsAreIndependent(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	player := world.players["player-one"]
	player.health = 50
	player.prayer = 50
	if !player.inventory.add(ItemCookedMinnow, 1) || !player.inventory.add(ItemPrayerPotion, 1) {
		t.Fatal("could not add test consumables")
	}
	if err := world.QueueConsumeItem(ConsumeItemCommand{
		PlayerID: "player-one", Sequence: 1, ItemID: ItemCookedMinnow, InventorySlot: 0,
	}); err != nil {
		t.Fatalf("QueueConsumeItem(food) returned an unexpected error: %v", err)
	}
	world.Step(50 * time.Millisecond)

	if err := world.QueueConsumeItem(ConsumeItemCommand{
		PlayerID: "player-one", Sequence: 2, ItemID: ItemPrayerPotion, InventorySlot: 1,
	}); err != nil {
		t.Fatalf("QueueConsumeItem(potion) during food cooldown returned an unexpected error: %v", err)
	}
	world.Step(50 * time.Millisecond)

	snapshot := mustPlayer(t, world, "player-one")
	if snapshot.Health != 58 || snapshot.Prayer != 75 {
		t.Fatalf("player after food and potion = %+v", snapshot)
	}
	if snapshot.NextFoodConsumePulse != 2 || snapshot.NextPotionConsumePulse != 2 {
		t.Fatalf("independent consume cadence = %+v", snapshot)
	}
}

func TestCookedMinnowCanBeEatenAtFullHealthButNotDuringCooldown(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	player := world.players["player-one"]
	if !player.inventory.add(ItemCookedMinnow, 2) {
		t.Fatal("could not add cooked minnows")
	}
	if err := world.QueueConsumeItem(ConsumeItemCommand{
		PlayerID: "player-one", Sequence: 1, ItemID: ItemCookedMinnow,
	}); err != nil {
		t.Fatalf("QueueConsumeItem() rejected food at full health: %v", err)
	}
	world.Step(50 * time.Millisecond)
	snapshot := mustPlayer(t, world, "player-one")
	if snapshot.Health != snapshot.MaximumHealth || snapshot.Inventory.itemsQuantity(ItemCookedMinnow) != 1 {
		t.Fatalf("player after eating at full health = %+v", snapshot)
	}
	if err := world.QueueConsumeItem(ConsumeItemCommand{
		PlayerID: "player-one", Sequence: 2, ItemID: ItemCookedMinnow,
	}); err == nil {
		t.Fatal("QueueConsumeItem() error = nil during cooldown")
	}
}

func TestPrayerPotionRestoresPrayerAndConsumesOneDose(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	player := world.players["player-one"]
	player.prayer = 60
	player.nextAttackPulse = 0
	if !player.inventory.add(ItemPrayerPotion, 5) {
		t.Fatal("could not add prayer potion doses")
	}
	if err := world.QueueConsumeItem(ConsumeItemCommand{
		PlayerID: "player-one", Sequence: 1, ItemID: ItemPrayerPotion,
	}); err != nil {
		t.Fatalf("QueueConsumeItem() returned an unexpected error: %v", err)
	}

	world.Step(50 * time.Millisecond)
	snapshot := mustPlayer(t, world, "player-one")
	if snapshot.Prayer != 85 || snapshot.Inventory.itemsQuantity(ItemPrayerPotion) != 4 {
		t.Fatalf("player after drinking a prayer potion = %+v", snapshot)
	}
	if snapshot.NextAttackPulse != 0 {
		t.Fatalf("prayer potion delayed attack until pulse %d", snapshot.NextAttackPulse)
	}
}

func TestConsumableIsRemovedFromSelectedInventoryStack(t *testing.T) {
	tests := []struct {
		name   string
		itemID string
		setup  func(*playerState)
	}{
		{
			name:   "food",
			itemID: ItemCookedMinnow,
			setup:  func(player *playerState) { player.health = 50 },
		},
		{
			name:   "potion",
			itemID: ItemPrayerPotion,
			setup:  func(player *playerState) { player.prayer = 50 },
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			world := NewWorld()
			mustAddPlayer(t, world, "player-one", Position{}, 4)
			player := world.players["player-one"]
			player.inventory.items = []InventoryItem{
				{Slot: 0, ItemID: test.itemID, Quantity: 2},
				{Slot: 4, ItemID: test.itemID, Quantity: 3},
			}
			test.setup(player)

			if err := world.QueueConsumeItem(ConsumeItemCommand{
				PlayerID: "player-one", Sequence: 1, ItemID: test.itemID, InventorySlot: 4,
			}); err != nil {
				t.Fatalf("QueueConsumeItem() returned an unexpected error: %v", err)
			}
			world.Step(actionPulseInterval)

			firstStack, firstExists := inventoryItemAt(player.inventory, 0)
			selectedStack, selectedExists := inventoryItemAt(player.inventory, 4)
			if !firstExists || firstStack.Quantity != 2 {
				t.Fatalf("first stack changed: %+v", firstStack)
			}
			if !selectedExists || selectedStack.Quantity != 2 {
				t.Fatalf("selected stack was not consumed: %+v", selectedStack)
			}
		})
	}
}

func TestCombatPotionsApplyLowLevelBoostsForFiveMinutes(t *testing.T) {
	tests := []struct {
		itemID string
		skills []SkillID
	}{
		{ItemMeleePotion, []SkillID{SkillAttack, SkillStrength, SkillDefence}},
		{ItemRangingPotion, []SkillID{SkillRange}},
		{ItemMagicPotion, []SkillID{SkillMagic}},
	}
	for _, test := range tests {
		world := NewWorld()
		mustAddPlayer(t, world, "player-one", Position{}, 4)
		player := world.players["player-one"]
		initialBoostRevision := player.combatBoostsRevision
		player.inventory.add(test.itemID, 1)
		if err := world.QueueConsumeItem(ConsumeItemCommand{PlayerID: "player-one", Sequence: 1, ItemID: test.itemID}); err != nil {
			t.Fatalf("QueueConsumeItem(%q) returned error: %v", test.itemID, err)
		}
		world.Step(actionPulseInterval)
		activeBoostRevision := player.combatBoostsRevision
		if activeBoostRevision <= initialBoostRevision {
			t.Errorf("%s boost revision = %d, want greater than %d", test.itemID, activeBoostRevision, initialBoostRevision)
		}
		for _, skill := range test.skills {
			if got := effectiveCombatSkillLevel(player, skill); got != 1+combatPotionBoostLevels {
				t.Errorf("%s after %s = %d, want %d", skill, test.itemID, got, 1+combatPotionBoostLevels)
			}
		}
		if player.inventory.quantity(test.itemID) != 0 {
			t.Errorf("%s dose was not consumed", test.itemID)
		}
		for range combatPotionDurationPulses - 1 {
			world.Step(actionPulseInterval)
		}
		if len(player.combatBoosts) == 0 {
			t.Errorf("%s expired before five minutes", test.itemID)
		}
		world.Step(actionPulseInterval)
		if len(player.combatBoosts) != 0 {
			t.Errorf("%s remained active after five minutes: %+v", test.itemID, player.combatBoosts)
		}
		if player.combatBoostsRevision <= activeBoostRevision {
			t.Errorf("%s expiry revision = %d, want greater than %d", test.itemID, player.combatBoostsRevision, activeBoostRevision)
		}
	}
}

func TestPrayerPotionCannotBeUsedAtFullPrayer(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	player := world.players["player-one"]
	player.health = 50
	if !player.inventory.add(ItemPrayerPotion, 1) {
		t.Fatal("could not add a prayer potion dose")
	}
	if err := world.QueueConsumeItem(ConsumeItemCommand{
		PlayerID: "player-one", Sequence: 1, ItemID: ItemPrayerPotion,
	}); err == nil {
		t.Fatal("QueueConsumeItem() error = nil at full prayer")
	}
}
