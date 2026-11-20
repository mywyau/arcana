package simulation

import (
	"testing"
	"time"
)

func TestDeathProtectionRejectsMoreThanFourItems(t *testing.T) {
	world := NewWorld()
	mustAddEmptyPlayer(t, world, "player-one", Position{})
	player := world.players["player-one"]
	for _, itemID := range []string{ItemCoins, ItemAirRune, ItemWaterRune, ItemEarthRune, ItemFireRune} {
		if !player.inventory.add(itemID, 1) {
			t.Fatalf("could not add %q", itemID)
		}
	}

	err := world.QueueSetDeathProtection(SetDeathProtectionCommand{
		PlayerID: "player-one", Sequence: 1, Slots: []int{0, 1, 2, 3, 4},
	})
	if err == nil {
		t.Fatal("QueueSetDeathProtection() accepted five protected items")
	}
}

func TestDeathProtectionFollowsInventorySwap(t *testing.T) {
	world := NewWorld()
	mustAddEmptyPlayer(t, world, "player-one", Position{})
	player := world.players["player-one"]
	if !player.inventory.add(ItemCoins, 25) || !player.inventory.add(ItemAirRune, 10) {
		t.Fatal("could not prepare inventory")
	}
	if err := world.QueueSetDeathProtection(SetDeathProtectionCommand{
		PlayerID: "player-one", Sequence: 1, Slots: []int{0},
	}); err != nil {
		t.Fatalf("QueueSetDeathProtection() error = %v", err)
	}
	world.Step(50 * time.Millisecond)
	if err := world.QueueMoveInventoryItem(MoveInventoryItemCommand{
		PlayerID: "player-one", Sequence: 1, FromSlot: 0, ToSlot: 1,
	}); err != nil {
		t.Fatalf("QueueMoveInventoryItem() error = %v", err)
	}
	world.Step(50 * time.Millisecond)

	snapshot := mustPlayer(t, world, "player-one")
	if len(snapshot.ProtectedInventorySlots) != 1 || snapshot.ProtectedInventorySlots[0] != 1 {
		t.Fatalf("protected slots after swap = %v, want [1]", snapshot.ProtectedInventorySlots)
	}
}

func TestPlayerDeathKeepsSelectedInventoryAndDropsOtherCarriedItems(t *testing.T) {
	world := NewWorld()
	mustAddEmptyPlayer(t, world, "player-one", Position{X: 4, Z: 6})
	player := world.players["player-one"]
	if !player.inventory.add(ItemCoins, 25) || !player.inventory.add(ItemAirRune, 10) ||
		!player.inventory.add(ItemCookedMinnow, 2) {
		t.Fatal("could not prepare inventory")
	}
	if err := world.QueueSetDeathProtection(SetDeathProtectionCommand{
		PlayerID: "player-one", Sequence: 1, Slots: []int{0, 2},
	}); err != nil {
		t.Fatalf("QueueSetDeathProtection() error = %v", err)
	}
	world.Step(50 * time.Millisecond)
	if got := protectedInventorySlots(player); len(got) != 2 || got[0] != 0 || got[1] != 2 {
		t.Fatalf("protected slots before death = %v, inventory = %+v", got, player.inventory.items)
	}

	player.equippedWeapon = WeaponSword
	player.equippedWeaponInstanceID = "equipped-sword"
	player.equippedArmour[ArmourSlotHead] = ItemBronzeHelm
	player.equippedArmourInstances[ArmourSlotHead] = "equipped-helm"
	player.equippedSpellbook = ItemStandardSpellbook
	if !player.ammunitionPouch.add(ItemBronzeArrow, 20) || !player.runePouch.add(ItemFireRune, 12) {
		t.Fatal("could not prepare carried pouches")
	}

	world.defeatPlayer(player)
	snapshot := mustPlayer(t, world, "player-one")
	if !snapshot.Defeated {
		t.Fatal("player was not defeated")
	}
	if got := snapshot.Inventory.itemsQuantity(ItemCoins); got != 25 {
		t.Fatalf("kept coin quantity = %d, want 25", got)
	}
	if got := snapshot.Inventory.itemsQuantity(ItemCookedMinnow); got != 2 {
		t.Fatalf("kept fish quantity = %d, want 2", got)
	}
	if got := snapshot.Inventory.itemsQuantity(ItemAirRune); got != 0 {
		t.Fatalf("unprotected air rune quantity = %d, want 0", got)
	}
	if snapshot.EquippedWeapon != WeaponUnarmed || len(snapshot.EquippedArmour) != 0 || snapshot.EquippedSpellbook != "" {
		t.Fatalf("equipment remained after death: weapon=%q armour=%v spellbook=%q", snapshot.EquippedWeapon, snapshot.EquippedArmour, snapshot.EquippedSpellbook)
	}
	if len(snapshot.AmmunitionPouch.Items) != 0 || len(snapshot.RunePouch.Items) != 0 {
		t.Fatalf("pouch contents remained after death: ammunition=%v runes=%v", snapshot.AmmunitionPouch.Items, snapshot.RunePouch.Items)
	}

	wantDrops := map[string]int{
		ItemAirRune: 10, string(WeaponSword): 1, ItemBronzeHelm: 1,
		ItemStandardSpellbook: 1, ItemBronzeArrow: 20, ItemFireRune: 12,
	}
	for _, drop := range world.Snapshot().Drops {
		wantDrops[drop.ItemID] -= drop.Quantity
	}
	for itemID, remaining := range wantDrops {
		if remaining != 0 {
			t.Errorf("drop remainder for %q = %d, want 0", itemID, remaining)
		}
	}
}

func TestRemovedProtectedItemDoesNotTransferProtectionToReplacement(t *testing.T) {
	world := NewWorld()
	mustAddEmptyPlayer(t, world, "player-one", Position{})
	player := world.players["player-one"]
	if !player.inventory.add(ItemCoins, 5) {
		t.Fatal("could not add coins")
	}
	if err := world.QueueSetDeathProtection(SetDeathProtectionCommand{
		PlayerID: "player-one", Sequence: 1, Slots: []int{0},
	}); err != nil {
		t.Fatalf("QueueSetDeathProtection() error = %v", err)
	}
	world.Step(50 * time.Millisecond)
	if !player.inventory.removeFromSlot(0, ItemCoins, 5) || !player.inventory.add(ItemAirRune, 3) {
		t.Fatal("could not replace protected inventory item")
	}

	world.defeatPlayer(player)
	if got := snapshotInventory(player.inventory).itemsQuantity(ItemAirRune); got != 0 {
		t.Fatalf("replacement rune quantity kept = %d, want 0", got)
	}
}

func mustAddEmptyPlayer(t *testing.T, world *World, id PlayerID, position Position) {
	t.Helper()
	if err := world.AddPlayer(id, position, DefaultPlayerMovementSpeed); err != nil {
		t.Fatalf("AddPlayer() error = %v", err)
	}
}

func TestMixedProtectionSurvivesRelogAndDeath(t *testing.T) {
	w := NewWorld()
	mustAddEmptyPlayer(t, w, "p", Position{})
	p := w.players["p"]
	p.inventory.add(ItemCoins, 12)
	p.equippedWeapon = WeaponBronzeDagger
	p.equippedWeaponInstanceID = "knife"
	p.equippedWeaponReinforcement = 3
	p.combatStyle = CombatStyleStab
	p.equippedArmour[ArmourSlotLeftHand] = ItemIronDirk
	p.equippedArmourInstances[ArmourSlotLeftHand] = "dirk"
	p.equippedSpellbook = ItemStandardSpellbook
	p.ringRack.add(ItemRubyRing, 1)
	p.ammunitionPouch.add(ItemBronzeArrow, 20)
	p.runePouch.add(ItemAirRune, 10)
	keys := []string{"gear:knife", "book:" + ItemStandardSpellbook, "ammo:" + ItemBronzeArrow}
	cmd := SetDeathProtectionCommand{PlayerID: "p", Sequence: 1, Slots: []int{0}, Equipped: keys}
	for _, invalid := range [][]string{{"gear:missing"}, {"gear:knife", "gear:knife"}, append(append([]string{}, keys...), "gear:dirk")} {
		bad := cmd
		bad.Equipped = invalid
		if w.QueueSetDeathProtection(bad) == nil {
			t.Fatalf("accepted invalid selection %v", invalid)
		}
	}
	if err := w.QueueSetDeathProtection(cmd); err != nil {
		t.Fatal(err)
	}
	w.Step(time.Millisecond)
	w2 := NewWorld()
	state := PersistentStateFromSnapshot(mustPlayer(t, w, "p"))
	if err := w2.QueueJoinWithPersistentState("p", state, 4); err != nil {
		t.Fatal(err)
	}
	w2.Step(time.Millisecond)
	p = w2.players["p"]
	w2.defeatPlayer(p)
	if p.inventory.quantity(ItemCoins) != 12 || p.equippedWeaponInstanceID != "knife" || p.equippedWeaponReinforcement != 3 || p.equippedSpellbook != ItemStandardSpellbook || p.ammunitionPouch.quantity(ItemBronzeArrow) != 20 {
		t.Fatal("lost protected item on death")
	}
	if len(p.equippedArmour)+len(p.ringRack.items)+len(p.runePouch.items) != 0 || len(w2.Snapshot().Drops) != 3 {
		t.Fatal("incorrect unprotected drops")
	}
}

func TestProtectionFollowsWornInstance(t *testing.T) {
	w := NewWorld()
	mustAddEmptyPlayer(t, w, "p", Position{})
	p := w.players["p"]
	p.inventory.add(ItemIronDirk, 1)
	dir := p.inventory.items[0]
	def, _ := itemDefinition(ItemIronDirk)
	if err := w.QueueSetDeathProtection(SetDeathProtectionCommand{PlayerID: "p", Sequence: 1, Slots: []int{dir.Slot}}); err != nil {
		t.Fatal(err)
	}
	w.applyPendingDeathProtections()
	if !applyEquipmentChange(p, def, EquipCommand{ItemID: WeaponID(dir.ItemID), InstanceID: dir.InstanceID}) {
		t.Fatal("equip failed")
	}
	reconcileDeathProtection(p)
	if !p.protectedEquipment["gear:"+dir.InstanceID] || len(p.protectedInventory) != 0 {
		t.Fatal("protection did not follow equipped dirk")
	}
	if !applyEquipmentChange(p, def, EquipCommand{ItemID: WeaponID(dir.ItemID), InstanceID: dir.InstanceID, Unequip: true}) {
		t.Fatal("unequip failed")
	}
	reconcileDeathProtection(p)
	if len(p.protectedInventory) != 1 || len(p.protectedEquipment) != 0 {
		t.Fatal("protection did not follow returned dirk")
	}
}
