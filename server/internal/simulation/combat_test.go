package simulation

import (
	"math"
	"testing"
	"time"
)

type fixedRollSource float64

// Level-one midpoint damage remains equal to each weapon's strength bonus.
const (
	swordDamage     = swordMeleeStrength
	daggerDamage    = daggerMeleeStrength
	bowDamage       = bowRangedStrength - 1 // starter bronze ammunition
	maceDamage      = maceMeleeStrength
	battleAxeDamage = battleAxeMeleeStrength
	staffDamage     = staffMagicStrength
)

func (source fixedRollSource) Float64() float64 { return float64(source) }

type sequenceRollSource struct {
	values []float64
	index  int
}

func mustTarget(t *testing.T, snapshot Snapshot, id string) TargetSnapshot {
	t.Helper()
	for _, target := range snapshot.Targets {
		if target.ID == id {
			return target
		}
	}
	t.Fatalf("target %q not found in snapshot", id)
	return TargetSnapshot{}
}

func (source *sequenceRollSource) Float64() float64 {
	value := source.values[source.index]
	source.index++
	return value
}

func newCombatTestWorld() *World {
	return NewWorldWithRollSource(fixedRollSource(0.5))
}

func mustGiveBronzeArrows(t *testing.T, world *World, id PlayerID, quantity int) {
	t.Helper()
	player, exists := world.players[id]
	if !exists {
		t.Fatalf("player %q does not exist", id)
	}
	if !player.ammunitionPouch.add(ItemBronzeArrow, quantity) {
		t.Fatalf("could not add %d bronze arrows to player %q", quantity, id)
	}
}

func TestTargetSnapshotsAreOrderedAndDetached(t *testing.T) {
	world := newCombatTestWorld()
	snapshot := world.Snapshot()
	if got, want := len(snapshot.Targets), 9; got != want {
		t.Fatalf("len(Targets) = %d, want %d", got, want)
	}
	wantIDs := []string{cowTargetID, dummyTargetID, hillWardenTargetID, mageTyrantTargetID, overlookDummyTargetID, rangedTyrantTargetID, riftTyrantTargetID, upperDummyTargetID, wildFowlTargetID}
	for index, want := range wantIDs {
		if got := snapshot.Targets[index].ID; got != want {
			t.Fatalf("Targets[%d].ID = %q, want %q", index, got, want)
		}
	}

	snapshot.Targets[0].Health = 0
	if got, want := mustTarget(t, world.Snapshot(), dummyTargetID).Health, dummyMaximumHealth; got != want {
		t.Fatalf("authoritative target health after snapshot mutation = %d, want %d", got, want)
	}
}

func TestRoamingEnemyDetectsNearbyPlayer(t *testing.T) {
	world := newCombatTestWorld()
	home := world.combat.roamingEnemies[hillWardenTargetID].home
	mustAddPlayer(t, world, "player-one", Position{X: home.X + 4, Y: home.Y, Z: home.Z}, 4)

	world.Step(actionPulseInterval)
	enemy := mustTarget(t, world.Snapshot(), hillWardenTargetID)
	if enemy.TargetID != "player-one" {
		t.Fatalf("roaming enemy target = %q, want player-one", enemy.TargetID)
	}
	if enemy.Position != home {
		t.Fatalf("roaming enemy position = %+v, want it to attack from range at %+v", enemy.Position, home)
	}
}

func TestRoamingEnemyAttacksAfterDetectionCadence(t *testing.T) {
	world := newCombatTestWorld()
	profile := combatProfileForEnemy(hillWardenTargetID)
	home := world.combat.roamingEnemies[hillWardenTargetID].home
	mustAddPlayer(t, world, "player-one", Position{X: home.X + 1, Y: home.Y, Z: home.Z}, 4)

	world.Step(actionPulseInterval)
	if got := mustPlayer(t, world, "player-one").Health; got != playerMaximumHealth {
		t.Fatalf("player health on detection pulse = %d, want %d", got, playerMaximumHealth)
	}
	world.Step(time.Duration(profile.AttackInterval) * actionPulseInterval)
	if got, want := mustPlayer(t, world, "player-one").Health, playerMaximumHealth-profile.AttackDamage; got != want {
		t.Fatalf("player health after roaming attack = %d, want %d", got, want)
	}
}

func TestRoamingEnemyLeashesAndReturnsHome(t *testing.T) {
	world := newCombatTestWorld()
	enemyState := world.combat.roamingEnemies[hillWardenTargetID]
	home := enemyState.home
	mustAddPlayer(t, world, "player-one", home, 4)
	enemyState.target.TargetID = "player-one"
	enemyState.target.Position = Position{X: home.X + 1, Y: home.Y, Z: home.Z}
	world.players["player-one"].position = Position{
		X: home.X + combatProfileForEnemy(hillWardenTargetID).LeashRange + 1,
		Y: home.Y,
		Z: home.Z,
	}
	world.Step(actionPulseInterval)
	if got := mustTarget(t, world.Snapshot(), hillWardenTargetID).TargetID; got != "" {
		t.Fatalf("roaming enemy retained target outside leash: %q", got)
	}
	world.Step(2 * time.Second)
	if got := mustTarget(t, world.Snapshot(), hillWardenTargetID).Position; got != home {
		t.Fatalf("roaming enemy returned to %+v, want home %+v", got, home)
	}
}

func TestEngagedSwordAutoAttacksOnWeaponCadence(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{X: -4, Z: -1}, 4)
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}

	world.Step(599 * time.Millisecond)
	if got := mustTarget(t, world.Snapshot(), dummyTargetID).Health; got != dummyMaximumHealth {
		t.Fatalf("dummy health before pulse = %d, want %d", got, dummyMaximumHealth)
	}

	world.Step(time.Millisecond)
	if got, want := mustTarget(t, world.Snapshot(), dummyTargetID).Health, dummyMaximumHealth-swordDamage; got != want {
		t.Fatalf("dummy health after first pulse = %d, want %d", got, want)
	}
	events := world.Snapshot().CombatEvents
	if len(events) != 1 {
		t.Fatalf("combat event count = %d, want 1", len(events))
	}
	event := events[0]
	if event.ID != 1 || event.Pulse != 1 || event.Type != CombatEventDamage ||
		event.SourceID != "player-one" || event.TargetID != dummyTargetID || event.Amount != swordDamage {
		t.Fatalf("combat event = %+v", event)
	}
	events[0].Amount = 999
	if got := world.Snapshot().CombatEvents[0].Amount; got != swordDamage {
		t.Fatalf("authoritative event amount after snapshot mutation = %d, want %d", got, swordDamage)
	}
	player := mustPlayer(t, world, "player-one")
	if player.CombatTarget != dummyTargetID || player.LastActionResult != "hit" || player.NextAttackPulse != 5 {
		t.Fatalf("player combat state after first hit = %+v", player)
	}
	attack := mustSkill(t, player, SkillAttack)
	if attack.Experience != swordDamage || attack.Level != 1 || attack.NextLevelExperience != SkillExperienceForLevel(2) {
		t.Fatalf("Attack after first sword hit = %+v", attack)
	}

	world.Step(3 * actionPulseInterval)
	if got, want := mustTarget(t, world.Snapshot(), dummyTargetID).Health, dummyMaximumHealth-swordDamage; got != want {
		t.Fatalf("dummy health before next weapon pulse = %d, want %d", got, want)
	}
	world.Step(actionPulseInterval)
	if got, want := mustTarget(t, world.Snapshot(), dummyTargetID).Health, dummyMaximumHealth-2*swordDamage; got != want {
		t.Fatalf("dummy health after next weapon pulse = %d, want %d", got, want)
	}
	if got := world.Snapshot().CombatEvents[1].ID; got != 2 {
		t.Fatalf("second combat event ID = %d, want 2", got)
	}
}

func TestEnemyDefenceMitigatesSuccessfulPlayerHit(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{X: -4, Z: -1}, 4)
	target := world.combat.targets[dummyTargetID]
	target.Defence = DefenceBonuses{Stab: 100}
	if err := world.QueueTarget(TargetCommand{
		PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID,
	}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}

	world.Step(actionPulseInterval)
	if got, want := target.Health, dummyMaximumHealth-5; got != want {
		t.Fatalf("target health after 50%% mitigation = %d, want %d", got, want)
	}
	events := world.Snapshot().CombatEvents
	if len(events) != 1 || events[0].Type != CombatEventDamage || events[0].Amount != 5 {
		t.Fatalf("mitigated combat events = %+v", events)
	}
}

func TestOverkillAwardsOnlyActualDamageAsXP(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{X: -4, Z: -1}, 4)
	target, found := world.combat.dummyByID(dummyTargetID)
	if !found {
		t.Fatal("ground dummy is missing")
	}
	target.Health = 3
	if err := world.QueueTarget(TargetCommand{
		PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID,
	}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}

	world.Step(actionPulseInterval)
	player := mustPlayer(t, world, "player-one")
	if attack := mustSkill(t, player, SkillAttack); attack.Experience != 3 {
		t.Fatalf("Attack after overkill = %+v, want 3 XP", attack)
	}
	events := world.Snapshot().CombatEvents
	if len(events) != 1 || events[0].Amount != 3 {
		t.Fatalf("overkill combat events = %+v, want actual damage 3", events)
	}
}

func TestConsecutiveSwordHitsCanRollDifferentDamage(t *testing.T) {
	rolls := &sequenceRollSource{values: []float64{
		0.5, 0.0, // First attack lands and rolls minimum damage.
		0.5, 1.0, // Second attack lands and rolls maximum damage.
		0.5, // Dummy attack roll on the second attack pulse.
	}}
	world := NewWorldWithRollSource(rolls)
	mustAddPlayer(t, world, "player-one", Position{X: -4, Z: -1}, 4)
	if err := world.QueueTarget(TargetCommand{
		PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID,
	}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}

	world.Step(actionPulseInterval)
	world.Step(4 * actionPulseInterval)
	events := world.Snapshot().CombatEvents
	if len(events) < 2 || events[0].Amount != 8 || events[1].Amount != 12 {
		t.Fatalf("sword damage events = %+v, want consecutive amounts 8 and 12", events)
	}
	if got, want := mustTarget(t, world.Snapshot(), dummyTargetID).Health, 80; got != want {
		t.Fatalf("dummy health after varied hits = %d, want %d", got, want)
	}
}

func TestCombatEventHistoryIsBounded(t *testing.T) {
	combat := newCombatState()
	for i := 0; i < combatEventHistoryLimit+3; i++ {
		combat.recordEvent(CombatEvent{Type: CombatEventDamage, Amount: i})
	}

	if got := len(combat.events); got != combatEventHistoryLimit {
		t.Fatalf("event history length = %d, want %d", got, combatEventHistoryLimit)
	}
	if got, want := combat.events[0].ID, uint64(4); got != want {
		t.Fatalf("first retained event ID = %d, want %d", got, want)
	}
	if got, want := combat.events[len(combat.events)-1].ID, uint64(combatEventHistoryLimit+3); got != want {
		t.Fatalf("last retained event ID = %d, want %d", got, want)
	}
}

func TestDummyDodgeAvoidsDamageAndConsumesAttackCadence(t *testing.T) {
	world := NewWorldWithRollSource(&sequenceRollSource{values: []float64{0.001}})
	mustAddPlayer(t, world, "player-one", Position{X: -4, Z: -1}, 4)
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}

	world.Step(actionPulseInterval)
	snapshot := world.Snapshot()
	if mustTarget(t, snapshot, dummyTargetID).Health != dummyMaximumHealth || mustTarget(t, snapshot, dummyTargetID).TargetID != "" {
		t.Fatalf("dummy state after dodge = %+v", mustTarget(t, snapshot, dummyTargetID))
	}
	player := mustPlayer(t, world, "player-one")
	if player.LastActionResult != "dodged" || player.NextAttackPulse != 5 {
		t.Fatalf("player state after dodge = %+v", player)
	}
	if attack := mustSkill(t, player, SkillAttack); attack.Experience != 0 {
		t.Fatalf("Attack after dodged hit = %+v, want zero XP", attack)
	}
	if len(snapshot.CombatEvents) != 1 {
		t.Fatalf("combat event count = %d, want 1", len(snapshot.CombatEvents))
	}
	event := snapshot.CombatEvents[0]
	if event.Type != CombatEventDodged || event.SourceID != "player-one" || event.TargetID != dummyTargetID || event.Amount != 0 {
		t.Fatalf("dodge event = %+v", event)
	}
}

func TestMissIsDistinctFromTargetDodge(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.99))
	mustAddPlayer(t, world, "player-one", Position{X: -4, Z: -1}, 4)
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}

	world.Step(actionPulseInterval)
	snapshot := world.Snapshot()
	if got := mustTarget(t, snapshot, dummyTargetID).Health; got != dummyMaximumHealth {
		t.Fatalf("dummy health after miss = %d, want %d", got, dummyMaximumHealth)
	}
	player := mustPlayer(t, world, "player-one")
	if player.LastActionResult != "missed" || player.NextAttackPulse != 5 {
		t.Fatalf("player state after miss = %+v", player)
	}
	if len(snapshot.CombatEvents) != 1 || snapshot.CombatEvents[0].Type != CombatEventMiss {
		t.Fatalf("combat events after miss = %+v", snapshot.CombatEvents)
	}
}

func TestPlayerDodgeAvoidsDummyDamageAndConsumesCadence(t *testing.T) {
	rolls := &sequenceRollSource{values: []float64{0.5, 0.5, 0.5, 0.5, 0.001}}
	world := NewWorldWithRollSource(rolls)
	mustAddPlayer(t, world, "player-one", Position{X: -4, Z: -1}, 4)
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}

	world.Step(actionPulseInterval)
	world.Step(4 * actionPulseInterval)
	snapshot := world.Snapshot()
	if got := mustPlayer(t, world, "player-one").Health; got != playerMaximumHealth {
		t.Fatalf("player health after dodge = %d, want %d", got, playerMaximumHealth)
	}
	if got, want := mustTarget(t, snapshot, dummyTargetID).NextAttackPulse, uint64(5+dummyAttackIntervalPulses); got != want {
		t.Fatalf("dummy NextAttackPulse = %d, want %d", got, want)
	}
	event := snapshot.CombatEvents[len(snapshot.CombatEvents)-1]
	if event.Type != CombatEventDodged || event.SourceID != dummyTargetID || event.TargetID != "player-one" {
		t.Fatalf("player dodge event = %+v", event)
	}
}

func TestDaggerUsesItsDamageAndCadence(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{X: -4, Z: -1}, 4)
	if err := world.QueueEquip(EquipCommand{PlayerID: "player-one", Sequence: 1, ItemID: WeaponDagger}); err != nil {
		t.Fatalf("QueueEquip() returned an unexpected error: %v", err)
	}
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}

	world.Step(actionPulseInterval)
	player := mustPlayer(t, world, "player-one")
	if player.EquippedWeapon != WeaponDagger || player.LastEquipmentSequence != 1 || player.LastEquipmentPulse != 1 {
		t.Fatalf("equipment state = %+v", player)
	}
	if got, want := mustTarget(t, world.Snapshot(), dummyTargetID).Health, dummyMaximumHealth-daggerDamage; got != want {
		t.Fatalf("dummy health after dagger hit = %d, want %d", got, want)
	}
	if got, want := player.NextAttackPulse, uint64(1+daggerAttackIntervalPulses); got != want {
		t.Fatalf("NextAttackPulse = %d, want %d", got, want)
	}

	world.Step(2 * actionPulseInterval)
	if got, want := mustTarget(t, world.Snapshot(), dummyTargetID).Health, dummyMaximumHealth-daggerDamage; got != want {
		t.Fatalf("dummy health before dagger cadence = %d, want %d", got, want)
	}
	world.Step(actionPulseInterval)
	if got, want := mustTarget(t, world.Snapshot(), dummyTargetID).Health, dummyMaximumHealth-2*daggerDamage; got != want {
		t.Fatalf("dummy health after second dagger hit = %d, want %d", got, want)
	}
}

func TestOwnedArmourEquipsIntoItsSlot(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	world.players["player-one"].inventory.add(ItemTrainingHelm, 1)

	if err := world.QueueEquip(EquipCommand{
		PlayerID: "player-one", Sequence: 1, ItemID: WeaponID(ItemTrainingHelm),
	}); err != nil {
		t.Fatalf("QueueEquip() returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)

	player := mustPlayer(t, world, "player-one")
	if got := player.EquippedArmour[ArmourSlotHead]; got != ItemTrainingHelm {
		t.Fatalf("equipped head item = %q, want %q", got, ItemTrainingHelm)
	}
	if snapshotInventoryContains(player.Inventory, ItemTrainingHelm) {
		t.Fatal("equipped helm still occupies an inventory slot")
	}
	if player.EquippedWeapon != WeaponSword {
		t.Fatalf("equipping armour changed weapon to %q", player.EquippedWeapon)
	}
	if err := world.QueueUnequip(UnequipCommand{
		PlayerID: "player-one", Sequence: 2, ItemID: ItemTrainingHelm,
	}); err != nil {
		t.Fatalf("QueueUnequip() returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	player = mustPlayer(t, world, "player-one")
	if _, equipped := player.EquippedArmour[ArmourSlotHead]; equipped {
		t.Fatalf("head armour remains equipped after toggle: %+v", player.EquippedArmour)
	}
	if !snapshotInventoryContains(player.Inventory, ItemTrainingHelm) {
		t.Fatal("unequipped helm was removed from inventory")
	}
}

func TestWeaponSwapMovesOnlyUnequippedWeaponBackToInventory(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	if snapshotInventoryContains(mustPlayer(t, world, "player-one").Inventory, string(WeaponSword)) {
		t.Fatal("equipped sword occupies an inventory slot")
	}
	if err := world.QueueEquip(EquipCommand{PlayerID: "player-one", Sequence: 1, ItemID: WeaponDagger}); err != nil {
		t.Fatalf("QueueEquip() returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	player := mustPlayer(t, world, "player-one")
	if player.EquippedWeapon != WeaponDagger || snapshotInventoryContains(player.Inventory, string(WeaponDagger)) ||
		!snapshotInventoryContains(player.Inventory, string(WeaponSword)) {
		t.Fatalf("inventory after weapon swap = %+v; equipped = %q", player.Inventory, player.EquippedWeapon)
	}
}

func TestUnequipRejectsFullInventory(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	player := world.players["player-one"]
	player.inventory.capacity = 8
	player.inventory.add(ItemTrainingHelm, 1)
	if err := world.QueueEquip(EquipCommand{PlayerID: "player-one", Sequence: 1, ItemID: WeaponID(ItemTrainingHelm)}); err != nil {
		t.Fatalf("QueueEquip() returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	for _, itemID := range []string{ItemAirRune, ItemIronArrow, ItemCoins, ItemBronzePickaxe, ItemBronzeAxe, ItemCopperOre, ItemTinOre, ItemIronOre} {
		if !player.inventory.add(itemID, 1) {
			t.Fatalf("failed to fill inventory with %q", itemID)
		}
	}
	if err := world.QueueUnequip(UnequipCommand{PlayerID: "player-one", Sequence: 2, ItemID: ItemTrainingHelm}); err == nil {
		t.Fatal("QueueUnequip() error = nil with a full inventory")
	}
	if player.equippedArmour[ArmourSlotHead] != ItemTrainingHelm {
		t.Fatal("failed unequip cleared the equipped helm")
	}
}

func TestSwiftArmourCanBeUnequippedWhileRemainingInInventory(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	player := world.players["player-one"]
	player.inventory.capacity = 12
	for _, itemID := range []string{ItemSwiftHelm, ItemSwiftBody, ItemSwiftBoots} {
		player.inventory.add(itemID, 1)
	}
	player.equippedArmour = map[ArmourSlot]string{
		ArmourSlotHead: ItemSwiftHelm,
		ArmourSlotBody: ItemSwiftBody,
		ArmourSlotFeet: ItemSwiftBoots,
	}

	if err := world.QueueUnequip(UnequipCommand{
		PlayerID: "player-one", Sequence: 1, ItemID: ItemSwiftBody,
	}); err != nil {
		t.Fatalf("QueueUnequip() returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)

	snapshot := mustPlayer(t, world, "player-one")
	if _, equipped := snapshot.EquippedArmour[ArmourSlotBody]; equipped {
		t.Fatalf("swift body remains equipped: %+v", snapshot.EquippedArmour)
	}
	if !snapshotInventoryContains(snapshot.Inventory, ItemSwiftBody) {
		t.Fatal("unequipped swift body was removed from inventory")
	}
	if snapshot.EquippedArmour[ArmourSlotHead] != ItemSwiftHelm ||
		snapshot.EquippedArmour[ArmourSlotFeet] != ItemSwiftBoots {
		t.Fatalf("unequipping swift body changed other slots: %+v", snapshot.EquippedArmour)
	}
}

func TestUnequippingWeaponUsesUnarmedAndAllowsDrop(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", world.combat.targets[dummyTargetID].Position, 4)
	if err := world.QueueUnequip(UnequipCommand{
		PlayerID: "player-one", Sequence: 1, ItemID: string(WeaponSword),
	}); err != nil {
		t.Fatalf("QueueUnequip() returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	player := mustPlayer(t, world, "player-one")
	if player.EquippedWeapon != WeaponUnarmed || player.CombatStyle != CombatStyleCrush {
		t.Fatalf("state after weapon unequip = %+v", player)
	}
	if !snapshotInventoryContains(player.Inventory, string(WeaponSword)) {
		t.Fatal("unequipped sword is missing from inventory")
	}
	if err := world.QueueTarget(TargetCommand{
		PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID,
	}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	if got, want := mustTarget(t, world.Snapshot(), dummyTargetID).Health, dummyMaximumHealth-1; got != want {
		t.Fatalf("dummy health after Unarmed hit = %d, want %d", got, want)
	}
	if err := world.QueueDropItem(DropItemCommand{
		PlayerID: "player-one", Sequence: 1, ItemID: string(WeaponSword),
	}); err != nil {
		t.Fatalf("QueueDropItem() returned an unexpected error: %v", err)
	}
	world.Step(50 * time.Millisecond)
	player = mustPlayer(t, world, "player-one")
	if snapshotInventoryContains(player.Inventory, string(WeaponSword)) {
		t.Fatal("dropped sword remains in inventory")
	}
	drops := world.Snapshot().Drops
	if len(drops) != 1 || drops[0].ItemID != string(WeaponSword) {
		t.Fatalf("drops after unequipping sword = %+v", drops)
	}
}

func TestUnequipRejectsItemThatIsNotEquipped(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	if err := world.QueueUnequip(UnequipCommand{
		PlayerID: "player-one", Sequence: 1, ItemID: string(WeaponDagger),
	}); err == nil {
		t.Fatal("QueueUnequip() error = nil for unequipped dagger")
	}
}

func TestBowAttacksFromRangeWithoutApproaching(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{X: -4, Z: 9.5}, 4)
	mustGiveBronzeArrows(t, world, "player-one", 10)
	if err := world.QueueEquip(EquipCommand{PlayerID: "player-one", Sequence: 1, ItemID: WeaponBow}); err != nil {
		t.Fatalf("QueueEquip() returned an unexpected error: %v", err)
	}
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}

	world.Step(actionPulseInterval)
	player := mustPlayer(t, world, "player-one")
	if player.EquippedWeapon != WeaponBow {
		t.Fatalf("EquippedWeapon = %q, want %q", player.EquippedWeapon, WeaponBow)
	}
	if player.AutoApproaching || player.Destination != nil {
		t.Fatalf("bow user approached while already in ranged distance: %+v", player)
	}
	if got, want := mustTarget(t, world.Snapshot(), dummyTargetID).Health, dummyMaximumHealth-bowDamage; got != want {
		t.Fatalf("dummy health after bow hit = %d, want %d", got, want)
	}
	if got, want := player.NextAttackPulse, uint64(1+bowAttackIntervalPulses); got != want {
		t.Fatalf("NextAttackPulse = %d, want %d", got, want)
	}
}

func TestBowRequiresAndConsumesBronzeArrows(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{X: -4, Z: -1}, 4)
	if err := world.QueueEquip(EquipCommand{
		PlayerID: "player-one", Sequence: 1, ItemID: WeaponBow,
	}); err != nil {
		t.Fatalf("QueueEquip() returned an unexpected error: %v", err)
	}
	if err := world.QueueTarget(TargetCommand{
		PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID,
	}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}

	world.Step(actionPulseInterval)
	player := mustPlayer(t, world, "player-one")
	if player.LastActionResult != "missing_ammunition" || player.NextAttackPulse != 0 {
		t.Fatalf("bow state without arrows = %+v", player)
	}
	if got := mustTarget(t, world.Snapshot(), dummyTargetID).Health; got != dummyMaximumHealth {
		t.Fatalf("dummy health without arrows = %d, want %d", got, dummyMaximumHealth)
	}

	mustGiveBronzeArrows(t, world, "player-one", 1)
	world.Step(actionPulseInterval)
	if got := (InventorySnapshot{Items: mustPlayer(t, world, "player-one").AmmunitionPouch.Items}).itemsQuantity(ItemBronzeArrow); got != 0 {
		t.Fatalf("bronze arrows after firing = %d, want 0", got)
	}
	if got, want := mustTarget(t, world.Snapshot(), dummyTargetID).Health, dummyMaximumHealth-bowDamage; got != want {
		t.Fatalf("dummy health after firing = %d, want %d", got, want)
	}
}

func TestDodgedBowShotStillConsumesBronzeArrow(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.0))
	mustAddPlayer(t, world, "player-one", Position{X: -4, Z: -1}, 4)
	mustGiveBronzeArrows(t, world, "player-one", 1)
	if err := world.QueueEquip(EquipCommand{
		PlayerID: "player-one", Sequence: 1, ItemID: WeaponBow,
	}); err != nil {
		t.Fatalf("QueueEquip() returned an unexpected error: %v", err)
	}
	if err := world.QueueTarget(TargetCommand{
		PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID,
	}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}

	world.Step(actionPulseInterval)
	player := mustPlayer(t, world, "player-one")
	if player.LastActionResult != "dodged" || (InventorySnapshot{Items: player.AmmunitionPouch.Items}).itemsQuantity(ItemBronzeArrow) != 0 {
		t.Fatalf("bow state after dodged shot = %+v", player)
	}
}

func TestBowCanAttackUpperDummyFromLowerLevel(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	mustGiveBronzeArrows(t, world, "player-one", 10)
	if err := world.QueueEquip(EquipCommand{PlayerID: "player-one", Sequence: 1, ItemID: WeaponBow}); err != nil {
		t.Fatalf("QueueEquip() returned an unexpected error: %v", err)
	}
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: upperDummyTargetID}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}

	world.Step(actionPulseInterval)
	snapshot := world.Snapshot()
	if got, want := mustTarget(t, snapshot, upperDummyTargetID).Health, dummyMaximumHealth-bowDamage; got != want {
		t.Fatalf("upper dummy health after bow hit = %d, want %d", got, want)
	}
	if got := mustTarget(t, snapshot, dummyTargetID).Health; got != dummyMaximumHealth {
		t.Fatalf("ground dummy health = %d after upper dummy hit, want %d", got, dummyMaximumHealth)
	}
	player := mustPlayer(t, world, "player-one")
	if player.AutoApproaching || player.Destination != nil {
		t.Fatalf("player approached upper dummy despite being in bow range: %+v", player)
	}
	event := snapshot.CombatEvents[len(snapshot.CombatEvents)-1]
	if event.TargetID != upperDummyTargetID || event.Amount != bowDamage {
		t.Fatalf("upper dummy combat event = %+v", event)
	}
}

func TestBowCanAttackOverlookDummyFromUpperLevel(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{X: -2, Y: 3, Z: 0}, 4)
	mustGiveBronzeArrows(t, world, "player-one", 10)
	if err := world.QueueEquip(EquipCommand{PlayerID: "player-one", Sequence: 1, ItemID: WeaponBow}); err != nil {
		t.Fatalf("QueueEquip() returned an unexpected error: %v", err)
	}
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: overlookDummyTargetID}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}

	world.Step(actionPulseInterval)
	snapshot := world.Snapshot()
	if got, want := mustTarget(t, snapshot, overlookDummyTargetID).Health, dummyMaximumHealth-bowDamage; got != want {
		t.Fatalf("overlook dummy health after bow hit = %d, want %d", got, want)
	}
	if mustTarget(t, snapshot, dummyTargetID).Health != dummyMaximumHealth || mustTarget(t, snapshot, upperDummyTargetID).Health != dummyMaximumHealth {
		t.Fatalf("another dummy was damaged: ground=%d upper=%d", mustTarget(t, snapshot, dummyTargetID).Health, mustTarget(t, snapshot, upperDummyTargetID).Health)
	}
	player := mustPlayer(t, world, "player-one")
	if player.AutoApproaching || player.Destination != nil {
		t.Fatalf("player approached overlook dummy despite being in bow range: %+v", player)
	}
}

func TestBowCooldownSurvivesDisengageAndReengage(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	mustGiveBronzeArrows(t, world, "player-one", 10)
	if err := world.QueueEquip(EquipCommand{PlayerID: "player-one", Sequence: 1, ItemID: WeaponBow}); err != nil {
		t.Fatalf("QueueEquip() returned an unexpected error: %v", err)
	}
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: upperDummyTargetID}); err != nil {
		t.Fatalf("initial QueueTarget() returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	if got, want := mustTarget(t, world.Snapshot(), upperDummyTargetID).Health, dummyMaximumHealth-bowDamage; got != want {
		t.Fatalf("upper dummy health after first attack = %d, want %d", got, want)
	}
	if got, want := mustPlayer(t, world, "player-one").NextAttackPulse, uint64(1+bowAttackIntervalPulses); got != want {
		t.Fatalf("NextAttackPulse after first attack = %d, want %d", got, want)
	}

	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 2}); err != nil {
		t.Fatalf("disengage QueueTarget() returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 3, TargetID: upperDummyTargetID}); err != nil {
		t.Fatalf("re-engage QueueTarget() returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)

	if got, want := mustTarget(t, world.Snapshot(), upperDummyTargetID).Health, dummyMaximumHealth-bowDamage; got != want {
		t.Fatalf("upper dummy health after early re-engage = %d, want %d", got, want)
	}
	if got, want := mustPlayer(t, world, "player-one").NextAttackPulse, uint64(1+bowAttackIntervalPulses); got != want {
		t.Fatalf("NextAttackPulse was reset by target changes: got %d, want %d", got, want)
	}

	world.Step(2 * actionPulseInterval)
	if got, want := mustTarget(t, world.Snapshot(), upperDummyTargetID).Health, dummyMaximumHealth-bowDamage; got != want {
		t.Fatalf("upper dummy health before original cooldown = %d, want %d", got, want)
	}
	world.Step(actionPulseInterval)
	if got, want := mustTarget(t, world.Snapshot(), upperDummyTargetID).Health, dummyMaximumHealth-2*bowDamage; got != want {
		t.Fatalf("upper dummy health when original cooldown elapsed = %d, want %d", got, want)
	}
}

func TestEngagingTargetReplacesExistingMovementWithImmediateApproach(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{X: -4, Z: 3}, 4)
	mustQueueMove(t, world, "player-one", 1, Position{X: -4, Z: 8})
	world.Step(50 * time.Millisecond)

	if err := world.QueueEquip(EquipCommand{PlayerID: "player-one", Sequence: 1, ItemID: WeaponBow}); err != nil {
		t.Fatalf("QueueEquip() returned an unexpected error: %v", err)
	}
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}
	world.Step(50 * time.Millisecond)

	player := mustPlayer(t, world, "player-one")
	if !player.AutoApproaching || player.Destination == nil || player.Destination.Z >= 3 {
		t.Fatalf("engagement did not immediately replace the existing movement route: %+v", player)
	}
}

func TestWeaponSwapDoesNotResetCurrentAttackCooldown(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{X: -4, Z: -1}, 4)
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	if err := world.QueueEquip(EquipCommand{PlayerID: "player-one", Sequence: 1, ItemID: WeaponDagger}); err != nil {
		t.Fatalf("QueueEquip() returned an unexpected error: %v", err)
	}

	world.Step(actionPulseInterval)
	player := mustPlayer(t, world, "player-one")
	if player.EquippedWeapon != WeaponDagger || player.NextAttackPulse != 5 {
		t.Fatalf("state after weapon swap = %+v", player)
	}
	world.Step(3 * actionPulseInterval)
	if got, want := mustTarget(t, world.Snapshot(), dummyTargetID).Health, dummyMaximumHealth-swordDamage-daggerDamage; got != want {
		t.Fatalf("dummy health after swapped attack = %d, want %d", got, want)
	}
}

func TestLatestEquipmentCommandWinsBeforePulse(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	if err := world.QueueEquip(EquipCommand{PlayerID: "player-one", Sequence: 1, ItemID: WeaponDagger}); err != nil {
		t.Fatalf("QueueEquip(1) returned an unexpected error: %v", err)
	}
	if err := world.QueueEquip(EquipCommand{PlayerID: "player-one", Sequence: 2, ItemID: WeaponBow}); err != nil {
		t.Fatalf("QueueEquip(2) returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	if got := mustPlayer(t, world, "player-one").EquippedWeapon; got != WeaponBow {
		t.Fatalf("EquippedWeapon = %q, want %q", got, WeaponBow)
	}
	if err := world.QueueEquip(EquipCommand{PlayerID: "player-one", Sequence: 2, ItemID: WeaponDagger}); err == nil {
		t.Fatal("QueueEquip() error = nil for stale sequence")
	}
}

func TestCombatStyleSelectionIsAuthoritativeAndWeaponSpecific(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	if got := mustPlayer(t, world, "player-one").CombatStyle; got != CombatStyleStab {
		t.Fatalf("default sword style = %q, want stab", got)
	}
	if err := world.QueueCombatStyle(CombatStyleCommand{
		PlayerID: "player-one", Sequence: 1, Style: CombatStyleStab,
	}); err != nil {
		t.Fatalf("QueueCombatStyle(stab) returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	player := mustPlayer(t, world, "player-one")
	if player.CombatStyle != CombatStyleStab || player.LastCombatStyleSequence != 1 {
		t.Fatalf("sword style state = %+v", player)
	}
	if err := world.QueueCombatStyle(CombatStyleCommand{
		PlayerID: "player-one", Sequence: 2, Style: CombatStyleRapid,
	}); err == nil {
		t.Fatal("QueueCombatStyle(rapid) error = nil while sword equipped")
	}

	if err := world.QueueEquip(EquipCommand{PlayerID: "player-one", Sequence: 1, ItemID: WeaponBow}); err != nil {
		t.Fatalf("QueueEquip(bow) returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	if got := mustPlayer(t, world, "player-one").CombatStyle; got != CombatStyleAccurate {
		t.Fatalf("default bow style = %q, want accurate", got)
	}
	if err := world.QueueCombatStyle(CombatStyleCommand{
		PlayerID: "player-one", Sequence: 2, Style: CombatStyleLongRange,
	}); err != nil {
		t.Fatalf("QueueCombatStyle(long range) returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	if got := mustPlayer(t, world, "player-one").CombatStyle; got != CombatStyleLongRange {
		t.Fatalf("selected bow style = %q, want long range", got)
	}
}

func TestBowStylesModifyCadenceAndRange(t *testing.T) {
	accurate, _ := weaponDefinitionForStyle(WeaponBow, CombatStyleAccurate)
	rapid, _ := weaponDefinitionForStyle(WeaponBow, CombatStyleRapid)
	longRange, _ := weaponDefinitionForStyle(WeaponBow, CombatStyleLongRange)
	if rapid.AttackInterval != 4 || accurate.AttackInterval != bowAttackIntervalPulses {
		t.Fatalf("bow intervals: rapid=%d accurate=%d", rapid.AttackInterval, accurate.AttackInterval)
	}
	if rapid.Range != bowRange ||
		rapid.ApproachRange != bowApproachRange ||
		accurate.Range != bowRange || accurate.ApproachRange != bowApproachRange ||
		longRange.Range != bowRange+longRangeStyleRangeBonus ||
		longRange.ApproachRange != bowApproachRange+longRangeStyleRangeBonus {
		t.Fatalf("bow ranges: rapid=%+v accurate=%+v long=%+v", rapid, accurate, longRange)
	}
	if rapid.DamageVariance != bowDamageVariance+wideStyleDamageVarianceBonus ||
		accurate.DamageVariance != bowDamageVariance-accurateStyleDamageVarianceReduction ||
		longRange.DamageVariance != bowDamageVariance+wideStyleDamageVarianceBonus {
		t.Fatalf("bow variances: rapid=%.2f accurate=%.2f long=%.2f", rapid.DamageVariance, accurate.DamageVariance, longRange.DamageVariance)
	}
}

func TestCrossbowStylesUseRangedRangeTradeoffs(t *testing.T) {
	accurate, _ := weaponDefinitionForStyle(WeaponCrossbow, CombatStyleAccurate)
	rapid, _ := weaponDefinitionForStyle(WeaponCrossbow, CombatStyleRapid)
	longRange, _ := weaponDefinitionForStyle(WeaponCrossbow, CombatStyleLongRange)

	if rapid.Range != crossbowRange ||
		accurate.Range != crossbowRange ||
		longRange.Range != crossbowRange+longRangeStyleRangeBonus {
		t.Fatalf("crossbow ranges: rapid=%.2f accurate=%.2f long=%.2f", rapid.Range, accurate.Range, longRange.Range)
	}
}

func TestStaffStylesUseProjectileStyleTradeoffs(t *testing.T) {
	accurate, _ := weaponDefinitionForStyle(WeaponStaff, CombatStyleAccurate)
	rapid, _ := weaponDefinitionForStyle(WeaponStaff, CombatStyleRapid)
	longRange, _ := weaponDefinitionForStyle(WeaponStaff, CombatStyleLongRange)

	if rapid.AttackInterval != staffAttackIntervalPulses-1 || accurate.AttackInterval != staffAttackIntervalPulses {
		t.Fatalf("staff intervals: rapid=%d accurate=%d", rapid.AttackInterval, accurate.AttackInterval)
	}
	if rapid.Range != staffRange || accurate.Range != staffRange || longRange.Range != staffRange+longRangeStyleRangeBonus {
		t.Fatalf("staff ranges: rapid=%.2f accurate=%.2f long=%.2f", rapid.Range, accurate.Range, longRange.Range)
	}
	if accurate.AccuracyBonus != 11+accurateStyleAccuracyBonus ||
		math.Abs(rapid.DamageVariance-(staffDamageVariance+wideStyleDamageVarianceBonus)) > 0.000_001 ||
		math.Abs(accurate.DamageVariance-(staffDamageVariance-accurateStyleDamageVarianceReduction)) > 0.000_001 ||
		math.Abs(longRange.DamageVariance-(staffDamageVariance+wideStyleDamageVarianceBonus)) > 0.000_001 {
		t.Fatalf("staff style definitions: rapid=%+v accurate=%+v", rapid, accurate)
	}
}

func TestTwoHandedBowUnequipsShieldWhileCrossbowKeepsIt(t *testing.T) {
	t.Run("bow", func(t *testing.T) {
		world := newCombatTestWorld()
		mustAddPlayer(t, world, "player-one", Position{}, 4)
		player := world.players["player-one"]
		if !player.inventory.add(ItemBronzeShield, 1) {
			t.Fatal("could not add shield")
		}
		shieldInstance := resolveOwnedInstanceID(player.inventory, ItemBronzeShield, "")
		if err := world.QueueEquip(EquipCommand{PlayerID: "player-one", Sequence: 1, ItemID: WeaponID(ItemBronzeShield), InstanceID: shieldInstance}); err != nil {
			t.Fatalf("QueueEquip(shield): %v", err)
		}
		world.Step(actionPulseInterval)
		bowInstance := resolveOwnedInstanceID(player.inventory, string(WeaponBow), "")
		if err := world.QueueEquip(EquipCommand{PlayerID: "player-one", Sequence: 2, ItemID: WeaponBow, InstanceID: bowInstance}); err != nil {
			t.Fatalf("QueueEquip(bow): %v", err)
		}
		world.Step(actionPulseInterval)
		if player.equippedWeapon != WeaponBow || player.equippedArmour[ArmourSlotLeftHand] != "" || !player.inventory.contains(ItemBronzeShield) {
			t.Fatalf("equipment after bow = weapon %q, armour %+v, inventory %+v", player.equippedWeapon, player.equippedArmour, player.inventory)
		}
	})

	t.Run("crossbow", func(t *testing.T) {
		world := newCombatTestWorld()
		mustAddPlayer(t, world, "player-one", Position{}, 4)
		player := world.players["player-one"]
		if !player.inventory.add(ItemBronzeShield, 1) || !player.inventory.add(string(WeaponCrossbow), 1) {
			t.Fatal("could not add crossbow loadout")
		}
		shieldInstance := resolveOwnedInstanceID(player.inventory, ItemBronzeShield, "")
		if err := world.QueueEquip(EquipCommand{PlayerID: "player-one", Sequence: 1, ItemID: WeaponID(ItemBronzeShield), InstanceID: shieldInstance}); err != nil {
			t.Fatalf("QueueEquip(shield): %v", err)
		}
		world.Step(actionPulseInterval)
		crossbowInstance := resolveOwnedInstanceID(player.inventory, string(WeaponCrossbow), "")
		if err := world.QueueEquip(EquipCommand{PlayerID: "player-one", Sequence: 2, ItemID: WeaponCrossbow, InstanceID: crossbowInstance}); err != nil {
			t.Fatalf("QueueEquip(crossbow): %v", err)
		}
		world.Step(actionPulseInterval)
		if player.equippedWeapon != WeaponCrossbow || player.equippedArmour[ArmourSlotLeftHand] != ItemBronzeShield {
			t.Fatalf("crossbow loadout = weapon %q, armour %+v", player.equippedWeapon, player.equippedArmour)
		}
	})
}

func TestLongMeleeAndCrossbowDefinitions(t *testing.T) {
	spear, _ := weaponDefinition(WeaponSpear)
	halberd, _ := weaponDefinition(WeaponHalberd)
	crossbow, _ := weaponDefinition(WeaponCrossbow)
	if spear.Range != polearmRange || halberd.Range != polearmRange || spear.Range <= swordRange {
		t.Fatalf("polearm ranges: spear %.2f, halberd %.2f, sword %.2f", spear.Range, halberd.Range, swordRange)
	}
	if !halberd.TwoHanded || spear.TwoHanded {
		t.Fatalf("polearm handedness: spear %t, halberd %t", spear.TwoHanded, halberd.TwoHanded)
	}
	if crossbow.TwoHanded || crossbow.Ammunition != (ItemCost{ItemID: ItemIronBolt, Quantity: 1}) {
		t.Fatalf("crossbow definition = %+v", crossbow)
	}
}

func TestBowLongRangeHitsBeyondAccurateRange(t *testing.T) {
	world := newCombatTestWorld()
	target := world.combat.targets[dummyTargetID]
	mustAddPlayer(t, world, "player-one", Position{X: target.Position.X + 13, Z: target.Position.Z}, 4)
	mustGiveBronzeArrows(t, world, "player-one", 10)
	if err := world.QueueEquip(EquipCommand{PlayerID: "player-one", Sequence: 1, ItemID: WeaponBow}); err != nil {
		t.Fatalf("QueueEquip(bow) returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	if err := world.QueueCombatStyle(CombatStyleCommand{
		PlayerID: "player-one", Sequence: 1, Style: CombatStyleLongRange,
	}); err != nil {
		t.Fatalf("QueueCombatStyle(long range) returned an unexpected error: %v", err)
	}
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	if got, want := target.Health, dummyMaximumHealth-bowDamage; got != want {
		t.Fatalf("dummy health after 13-unit long-range attack = %d, want %d", got, want)
	}
}

func TestBowRapidSchedulesFasterNextAttack(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{X: -4, Z: -1}, 4)
	mustGiveBronzeArrows(t, world, "player-one", 10)
	if err := world.QueueEquip(EquipCommand{PlayerID: "player-one", Sequence: 1, ItemID: WeaponBow}); err != nil {
		t.Fatalf("QueueEquip(bow) returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	if err := world.QueueCombatStyle(CombatStyleCommand{
		PlayerID: "player-one", Sequence: 1, Style: CombatStyleRapid,
	}); err != nil {
		t.Fatalf("QueueCombatStyle(rapid) returned an unexpected error: %v", err)
	}
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	if got, want := mustPlayer(t, world, "player-one").NextAttackPulse, uint64(6); got != want {
		t.Fatalf("rapid bow next attack pulse = %d, want %d", got, want)
	}
	player := mustPlayer(t, world, "player-one")
	if attack := mustSkill(t, player, SkillAttack); attack.Experience != 0 {
		t.Fatalf("Attack after bow hit = %+v, want zero XP", attack)
	}
	if ranged := mustSkill(t, player, SkillRange); ranged.Experience != bowDamage {
		t.Fatalf("Range after bow hit = %+v, want %d XP", ranged, bowDamage)
	}
}

func TestStaffHitTrainsMagic(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{X: -4, Z: -1}, 4)
	world.players["player-one"].inventory.add(string(WeaponStaff), 1)
	if err := world.QueueEquip(EquipCommand{PlayerID: "player-one", Sequence: 1, ItemID: WeaponStaff}); err != nil {
		t.Fatalf("QueueEquip(staff) returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	player := mustPlayer(t, world, "player-one")
	if magic := mustSkill(t, player, SkillMagic); magic.Experience != staffDamage {
		t.Fatalf("Magic after staff hit = %+v, want %d XP", magic, staffDamage)
	}
	if attack := mustSkill(t, player, SkillAttack); attack.Experience != 0 {
		t.Fatalf("Attack after staff hit = %+v, want zero XP", attack)
	}
}

func TestSwordStylesRouteCombatXP(t *testing.T) {
	tests := []struct {
		name       string
		style      CombatStyleID
		trained    SkillID
		experience int64
	}{
		{name: "stab", style: CombatStyleStab, trained: SkillAttack, experience: swordDamage},
		{name: "slash", style: CombatStyleSlash, trained: SkillStrength, experience: swordDamage},
		{name: "defensive", style: CombatStyleDefensive, trained: SkillDefence, experience: swordDamage},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			world := newCombatTestWorld()
			mustAddPlayer(t, world, "player-one", Position{X: -4, Z: -1}, 4)
			if err := world.QueueCombatStyle(CombatStyleCommand{
				PlayerID: "player-one", Sequence: 1, Style: test.style,
			}); err != nil {
				t.Fatalf("QueueCombatStyle(%s): %v", test.style, err)
			}
			world.Step(actionPulseInterval)
			if err := world.QueueTarget(TargetCommand{
				PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID,
			}); err != nil {
				t.Fatalf("QueueTarget(): %v", err)
			}
			world.Step(actionPulseInterval)

			player := mustPlayer(t, world, "player-one")
			for _, skillID := range []SkillID{SkillAttack, SkillStrength, SkillDefence} {
				want := int64(0)
				if skillID == test.trained {
					want = test.experience
				}
				if got := mustSkill(t, player, skillID).Experience; got != want {
					t.Errorf("%s XP = %d, want %d", skillID, got, want)
				}
			}
		})
	}
}

func TestAdditionalWeaponDefinitions(t *testing.T) {
	tests := []struct {
		id       WeaponID
		strength int
		interval uint64
		range_   float64
	}{
		{id: WeaponMace, strength: maceMeleeStrength, interval: maceAttackIntervalPulses, range_: maceRange},
		{id: WeaponSteelSword, strength: steelSwordMeleeStrength, interval: swordAttackIntervalPulses, range_: swordRange},
		{id: WeaponMithrilSword, strength: 18, interval: swordAttackIntervalPulses, range_: swordRange},
		{id: WeaponRuneSword, strength: runeSwordMeleeStrength, interval: swordAttackIntervalPulses, range_: swordRange},
		{id: WeaponBattleAxe, strength: battleAxeMeleeStrength, interval: battleAxeAttackIntervalPulses, range_: battleAxeRange},
		{id: WeaponStaff, strength: staffMagicStrength, interval: staffAttackIntervalPulses, range_: staffRange},
	}
	for _, test := range tests {
		weapon, exists := weaponDefinition(test.id)
		strength := weapon.MeleeStrength + weapon.RangedStrength + weapon.MagicStrength
		if !exists || strength != test.strength || weapon.AttackInterval != test.interval || weapon.Range != test.range_ {
			t.Fatalf("weaponDefinition(%q) = %+v, exists %t", test.id, weapon, exists)
		}
	}
	if !combatStyleValidForWeapon(CombatStyleHeavyCrush, WeaponMace) ||
		!combatStyleValidForWeapon(CombatStyleHeavyCrush, WeaponBattleAxe) ||
		!combatStyleValidForWeapon(CombatStyleLongRange, WeaponStaff) ||
		!combatStyleValidForWeapon(CombatStyleDefensive, WeaponSword) ||
		!combatStyleValidForWeapon(CombatStyleDefensive, WeaponSteelSword) ||
		!combatStyleValidForWeapon(CombatStyleDefensive, WeaponMithrilSword) ||
		!combatStyleValidForWeapon(CombatStyleDefensive, WeaponRuneSword) {
		t.Fatal("additional weapon combat styles are not valid")
	}
}

func TestWeaponsUseConfiguredDamageVariance(t *testing.T) {
	tests := []struct {
		id       WeaponID
		minimum  int
		average  int
		maximum  int
		variance float64
	}{
		{id: WeaponSword, minimum: 8, average: 10, maximum: 12, variance: swordDamageVariance},
		{id: WeaponSteelSword, minimum: 11, average: 14, maximum: 17, variance: swordDamageVariance},
		{id: WeaponRuneSword, minimum: 32, average: 40, maximum: 48, variance: swordDamageVariance},
		{id: WeaponDagger, minimum: 5, average: 7, maximum: 9, variance: daggerDamageVariance},
		{id: WeaponBow, minimum: 3, average: 4, maximum: 5, variance: bowDamageVariance},
		{id: WeaponMace, minimum: 9, average: 12, maximum: 15, variance: maceDamageVariance},
		{id: WeaponBattleAxe, minimum: 11, average: 15, maximum: 19, variance: battleAxeDamageVariance},
		{id: WeaponStaff, minimum: 4, average: 6, maximum: 8, variance: staffDamageVariance},
	}
	for _, test := range tests {
		t.Run(string(test.id), func(t *testing.T) {
			weapon, exists := weaponDefinition(test.id)
			if !exists {
				t.Fatalf("weaponDefinition(%q) does not exist", test.id)
			}
			if math.Abs(weapon.DamageVariance-test.variance) > 0.000_001 {
				t.Fatalf("damage variance = %.3f, want %.3f", weapon.DamageVariance, test.variance)
			}
			poweredDamage := weaponPoweredDamage(&playerState{skills: newDefaultSkills()}, weapon)
			if got := rollDamageWithVariance(poweredDamage, weapon.DamageVariance, 0); got != test.minimum {
				t.Fatalf("minimum damage = %d, want %d", got, test.minimum)
			}
			if got := rollDamageWithVariance(poweredDamage, weapon.DamageVariance, 0.5); got != test.average {
				t.Fatalf("midpoint damage = %d, want %d", got, test.average)
			}
			if got := rollDamageWithVariance(poweredDamage, weapon.DamageVariance, 1); got != test.maximum {
				t.Fatalf("maximum damage = %d, want %d", got, test.maximum)
			}
		})
	}
}

func TestWeaponStrengthUsesRelevantPowerSkill(t *testing.T) {
	tests := []struct {
		name   string
		weapon WeaponID
		skill  SkillID
		want   float64
	}{
		{name: "melee uses Strength", weapon: WeaponSword, skill: SkillStrength, want: 34.5},
		{name: "ranged uses Range", weapon: WeaponBow, skill: SkillRange, want: 28.5},
		{name: "magic uses Magic", weapon: WeaponStaff, skill: SkillMagic, want: 30.5},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			weapon, exists := weaponDefinition(test.weapon)
			if !exists {
				t.Fatalf("weaponDefinition(%q) does not exist", test.weapon)
			}
			player := &playerState{skills: newDefaultSkills()}
			player.skills[test.skill] = SkillExperienceForLevel(50)
			if got := weaponPoweredDamage(player, weapon); math.Abs(got-test.want) > 0.000_001 {
				t.Fatalf("powered damage = %.3f, want %.3f", got, test.want)
			}
		})
	}
}

func TestAttackLevelDoesNotIncreaseMeleeDamage(t *testing.T) {
	weapon, _ := weaponDefinition(WeaponSword)
	player := &playerState{skills: newDefaultSkills()}
	player.skills[SkillAttack] = SkillExperienceForLevel(50)
	if got := weaponPoweredDamage(player, weapon); got != swordMeleeStrength {
		t.Fatalf("powered damage with Attack level 50 = %.3f, want %d", got, swordMeleeStrength)
	}
}

func TestEngagedPlayerApproachesBeforeAutoAttacking(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}

	world.Step(actionPulseInterval)
	player := mustPlayer(t, world, "player-one")
	if player.LastActionResult != "approaching" || !player.AutoApproaching || player.Destination == nil {
		t.Fatalf("player approach state = %+v", player)
	}
	if got := mustTarget(t, world.Snapshot(), dummyTargetID).Health; got != dummyMaximumHealth {
		t.Fatalf("dummy health while out of range = %d, want %d", got, dummyMaximumHealth)
	}

	for range 4 {
		world.Step(actionPulseInterval)
		if mustTarget(t, world.Snapshot(), dummyTargetID).Health < dummyMaximumHealth {
			break
		}
	}
	if got, want := mustTarget(t, world.Snapshot(), dummyTargetID).Health, dummyMaximumHealth-swordDamage; got != want {
		t.Fatalf("dummy health after approach = %d, want %d", got, want)
	}
	player = mustPlayer(t, world, "player-one")
	if player.AutoApproaching || player.Destination != nil || player.LastActionResult != "hit" {
		t.Fatalf("player state after reaching target = %+v", player)
	}
}

func TestEngagedPlayerApproachesBeforeNextCombatPulse(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	world.players["player-one"].nextAttackPulse = 3
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}

	world.Step(50 * time.Millisecond)
	player := mustPlayer(t, world, "player-one")
	if player.LastActionResult != "approaching" || !player.AutoApproaching || player.Destination == nil {
		t.Fatalf("player did not begin approaching before the combat pulse: %+v", player)
	}
	if player.Position == (Position{}) {
		t.Fatalf("player position did not advance on the immediate approach: %+v", player)
	}
	if got := mustTarget(t, world.Snapshot(), dummyTargetID).Health; got != dummyMaximumHealth {
		t.Fatalf("dummy health before attack cooldown = %d, want %d", got, dummyMaximumHealth)
	}
	if player.NextAttackPulse != 3 {
		t.Fatalf("next attack pulse = %d, want 3", player.NextAttackPulse)
	}
}

func TestSwordApproachesAtPreviouslyValidLongRange(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{X: -4, Z: -0.3}, 4)
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}

	world.Step(50 * time.Millisecond)
	if got := mustTarget(t, world.Snapshot(), dummyTargetID).Health; got != dummyMaximumHealth {
		t.Fatalf("dummy health at 1.7-unit range = %d, want %d", got, dummyMaximumHealth)
	}
	player := mustPlayer(t, world, "player-one")
	if !player.AutoApproaching || player.Destination == nil {
		t.Fatalf("player approach state = %+v", player)
	}
	if got := distance(*player.Destination, mustTarget(t, world.Snapshot(), dummyTargetID).Position); math.Abs(got-swordApproachRange) > 0.001 {
		t.Fatalf("approach distance = %.3f, want %.3f", got, swordApproachRange)
	}
}

func TestDisengageStopsAutomaticApproach(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID}); err != nil {
		t.Fatalf("QueueTarget(engage) returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 2}); err != nil {
		t.Fatalf("QueueTarget(disengage) returned an unexpected error: %v", err)
	}

	world.Step(actionPulseInterval)
	player := mustPlayer(t, world, "player-one")
	if player.CombatTarget != "" || player.AutoApproaching || player.Destination != nil || player.LastActionResult != "disengaged" {
		t.Fatalf("player state after disengage = %+v", player)
	}
	if got := mustTarget(t, world.Snapshot(), dummyTargetID).Health; got != dummyMaximumHealth {
		t.Fatalf("dummy health after disengage = %d, want %d", got, dummyMaximumHealth)
	}
}

func TestManualMovementWithDisengageDoesNotResumeApproach(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)

	if err := world.QueueMove(MoveCommand{
		PlayerID: "player-one", Sequence: 1, Destination: Position{X: 5},
	}); err != nil {
		t.Fatalf("QueueMove() returned an unexpected error: %v", err)
	}
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 2}); err != nil {
		t.Fatalf("QueueTarget(disengage) returned an unexpected error: %v", err)
	}
	world.Step(50 * time.Millisecond)
	player := mustPlayer(t, world, "player-one")
	if player.CombatTarget != "" || player.AutoApproaching {
		t.Fatalf("player state during manual reposition = %+v", player)
	}

	world.Step(actionPulseInterval - 50*time.Millisecond)
	player = mustPlayer(t, world, "player-one")
	if player.CombatTarget != "" || player.AutoApproaching || player.LastActionResult != "disengaged" {
		t.Fatalf("player state after combat pulse = %+v", player)
	}
	if player.Destination == nil || player.Destination.X != 5 {
		t.Fatalf("manual destination after disengage = %+v, want X=5", player.Destination)
	}
}

func TestDummyAggroPersistsAfterPlayerDisengages(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{X: -4, Z: -1}, 4)
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID}); err != nil {
		t.Fatalf("QueueTarget(engage) returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 2}); err != nil {
		t.Fatalf("QueueTarget(disengage) returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)

	player := mustPlayer(t, world, "player-one")
	if player.CombatTarget != "" || mustTarget(t, world.Snapshot(), dummyTargetID).TargetID != "player-one" {
		t.Fatalf("combat ownership after disengage: player=%+v dummy=%+v", player, mustTarget(t, world.Snapshot(), dummyTargetID))
	}
	world.Step(3 * actionPulseInterval)
	if got, want := mustPlayer(t, world, "player-one").Health, playerMaximumHealth-dummyAttackDamage; got != want {
		t.Fatalf("player health after retained dummy aggro = %d, want %d", got, want)
	}
}

func TestLatestTargetCommandWinsBeforePulse(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{X: -4, Z: -1}, 4)
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID}); err != nil {
		t.Fatalf("QueueTarget(1) returned an unexpected error: %v", err)
	}
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 2}); err != nil {
		t.Fatalf("QueueTarget(2) returned an unexpected error: %v", err)
	}

	world.Step(actionPulseInterval)
	if got := mustTarget(t, world.Snapshot(), dummyTargetID).Health; got != dummyMaximumHealth {
		t.Fatalf("dummy health = %d, want %d", got, dummyMaximumHealth)
	}
	if got := mustPlayer(t, world, "player-one").LastActionSequence; got != 2 {
		t.Fatalf("LastActionSequence = %d, want 2", got)
	}
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 2}); err == nil {
		t.Fatal("QueueTarget() error = nil for stale sequence")
	}
}

func TestDummyTargetsFirstAttackerAndRetaliates(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{X: -4, Z: -1}, 4)
	mustAddPlayer(t, world, "player-two", Position{X: -3, Z: -2}, 4)
	for _, id := range []PlayerID{"player-one", "player-two"} {
		if err := world.QueueTarget(TargetCommand{PlayerID: id, Sequence: 1, TargetID: dummyTargetID}); err != nil {
			t.Fatalf("QueueTarget(%s) returned an unexpected error: %v", id, err)
		}
	}

	world.Step(actionPulseInterval)
	if got, want := mustTarget(t, world.Snapshot(), dummyTargetID).TargetID, PlayerID("player-one"); got != want {
		t.Fatalf("dummy target = %q, want %q", got, want)
	}
	world.Step(3 * actionPulseInterval)
	if got := mustPlayer(t, world, "player-one").Health; got != playerMaximumHealth {
		t.Fatalf("player health before retaliation pulse = %d, want %d", got, playerMaximumHealth)
	}
	world.Step(actionPulseInterval)
	if got, want := mustPlayer(t, world, "player-one").Health, playerMaximumHealth-dummyAttackDamage; got != want {
		t.Fatalf("player health after retaliation = %d, want %d", got, want)
	}
	events := world.Snapshot().CombatEvents
	event := events[len(events)-1]
	if event.SourceID != dummyTargetID || event.TargetID != "player-one" || event.Amount != dummyAttackDamage {
		t.Fatalf("dummy damage event = %+v", event)
	}
	if got := mustPlayer(t, world, "player-two").Health; got != playerMaximumHealth {
		t.Fatalf("non-targeted player health = %d, want %d", got, playerMaximumHealth)
	}
}

func TestDefeatedPlayerRespawnsAtFullHealthAndSpawn(t *testing.T) {
	world := newCombatTestWorld()
	spawn := Position{X: -4, Z: -1}
	mustAddPlayer(t, world, "player-one", spawn, 4)
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	world.players["player-one"].health = dummyAttackDamage
	world.players["player-one"].position = Position{X: -3.8, Z: -1}

	world.Step(dummyAttackIntervalPulses * actionPulseInterval)
	player := mustPlayer(t, world, "player-one")
	if !player.Defeated || player.Health != 0 || player.RespawnAtPulse != 10 {
		t.Fatalf("defeated player state = %+v", player)
	}
	if player.Destination != nil || player.CombatTarget != "" {
		t.Fatalf("defeated player retained movement or target: %+v", player)
	}
	if err := world.QueueMove(MoveCommand{PlayerID: "player-one", Sequence: 1, Destination: Position{X: 1}}); err == nil {
		t.Fatal("QueueMove() error = nil for defeated player")
	}
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 2, TargetID: dummyTargetID}); err == nil {
		t.Fatal("QueueTarget() error = nil for defeated player")
	}

	world.Step((respawnDelayPulses - 1) * actionPulseInterval)
	if !mustPlayer(t, world, "player-one").Defeated {
		t.Fatal("player respawned before the configured pulse")
	}
	world.Step(actionPulseInterval)
	player = mustPlayer(t, world, "player-one")
	if player.Defeated || player.Health != playerMaximumHealth || player.Position != spawn || player.RespawnAtPulse != 0 {
		t.Fatalf("respawned player state = %+v", player)
	}
}

func TestDefeatedDummyRespawnsAtFullHealth(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{X: -4, Z: -1}, 4)
	world.combat.targets[dummyTargetID].Health = swordDamage
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}

	world.Step(actionPulseInterval)
	dummy := mustTarget(t, world.Snapshot(), dummyTargetID)
	if !dummy.Defeated || dummy.Health != 0 || dummy.RespawnAtPulse != 6 || dummy.TargetID != "" {
		t.Fatalf("defeated dummy state = %+v", dummy)
	}
	if got := mustPlayer(t, world, "player-one").CombatTarget; got != "" {
		t.Fatalf("player target after dummy defeat = %q, want empty", got)
	}

	world.Step((respawnDelayPulses - 1) * actionPulseInterval)
	if !mustTarget(t, world.Snapshot(), dummyTargetID).Defeated {
		t.Fatal("dummy respawned before the configured pulse")
	}
	world.Step(actionPulseInterval)
	dummy = mustTarget(t, world.Snapshot(), dummyTargetID)
	if dummy.Defeated || dummy.Health != dummyMaximumHealth || dummy.RespawnAtPulse != 0 {
		t.Fatalf("respawned dummy state = %+v", dummy)
	}
}
