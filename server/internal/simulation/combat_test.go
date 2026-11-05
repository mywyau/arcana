package simulation

import (
	"math"
	"testing"
	"time"
)

type fixedRollSource float64

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

func TestTargetSnapshotsAreOrderedAndDetached(t *testing.T) {
	world := newCombatTestWorld()
	snapshot := world.Snapshot()
	if got, want := len(snapshot.Targets), 4; got != want {
		t.Fatalf("len(Targets) = %d, want %d", got, want)
	}
	wantIDs := []string{dummyTargetID, hillWardenTargetID, overlookDummyTargetID, upperDummyTargetID}
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

func TestRoamingEnemyDetectsAndChasesNearbyPlayer(t *testing.T) {
	world := newCombatTestWorld()
	home := world.combat.roamingEnemies[hillWardenTargetID].home
	mustAddPlayer(t, world, "player-one", Position{X: home.X + 4, Y: home.Y, Z: home.Z}, 4)

	world.Step(actionPulseInterval)
	enemy := mustTarget(t, world.Snapshot(), hillWardenTargetID)
	if enemy.TargetID != "player-one" {
		t.Fatalf("roaming enemy target = %q, want player-one", enemy.TargetID)
	}
	world.Step(100 * time.Millisecond)
	moved := mustTarget(t, world.Snapshot(), hillWardenTargetID)
	if moved.Position.X <= home.X {
		t.Fatalf("roaming enemy position = %+v, want movement toward player", moved.Position)
	}
}

func TestRoamingEnemyAttacksAfterDetectionCadence(t *testing.T) {
	world := newCombatTestWorld()
	home := world.combat.roamingEnemies[hillWardenTargetID].home
	mustAddPlayer(t, world, "player-one", Position{X: home.X + 1, Y: home.Y, Z: home.Z}, 4)

	world.Step(actionPulseInterval)
	if got := mustPlayer(t, world, "player-one").Health; got != playerMaximumHealth {
		t.Fatalf("player health on detection pulse = %d, want %d", got, playerMaximumHealth)
	}
	world.Step(dummyAttackIntervalPulses * actionPulseInterval)
	if got, want := mustPlayer(t, world, "player-one").Health, playerMaximumHealth-dummyAttackDamage; got != want {
		t.Fatalf("player health after roaming attack = %d, want %d", got, want)
	}
}

func TestRoamingEnemyLeashesAndReturnsHome(t *testing.T) {
	world := newCombatTestWorld()
	enemyState := world.combat.roamingEnemies[hillWardenTargetID]
	home := enemyState.home
	mustAddPlayer(t, world, "player-one", Position{X: home.X + 4, Y: home.Y, Z: home.Z}, 4)

	world.Step(actionPulseInterval)
	world.Step(500 * time.Millisecond)
	if got := mustTarget(t, world.Snapshot(), hillWardenTargetID).Position; got == home {
		t.Fatal("roaming enemy did not leave home before leash test")
	}
	world.players["player-one"].position = Position{X: home.X + roamingLeashRange + 1, Y: home.Y, Z: home.Z}
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
	if attack.Experience != swordDamage || attack.Level != 1 || attack.NextLevelExperience != 100 {
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
	world := NewWorldWithRollSource(&sequenceRollSource{values: []float64{0.05}})
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

func TestPlayerDodgeAvoidsDummyDamageAndConsumesCadence(t *testing.T) {
	rolls := &sequenceRollSource{values: []float64{0.5, 0.5, 0.05}}
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

func TestBowAttacksFromRangeWithoutApproaching(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{X: -4, Z: 9.5}, 4)
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

func TestBowCanAttackUpperDummyFromLowerLevel(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
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

func TestEngagingTargetStopsExistingMovementBeforeCombatPulse(t *testing.T) {
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
	if player.Destination != nil || len(player.Path) != 0 {
		t.Fatalf("movement continued while engagement waited for its action pulse: %+v", player)
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
	if err := world.QueueEquip(EquipCommand{PlayerID: "player-one", Sequence: 2, ItemID: WeaponSword}); err != nil {
		t.Fatalf("QueueEquip(2) returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	if got := mustPlayer(t, world, "player-one").EquippedWeapon; got != WeaponSword {
		t.Fatalf("EquippedWeapon = %q, want %q", got, WeaponSword)
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
	if longRange.Range != 14 || longRange.ApproachRange != 13.25 || accurate.Range != bowRange {
		t.Fatalf("bow ranges: long=%+v accurate=%+v", longRange, accurate)
	}
}

func TestBowLongRangeHitsBeyondAccurateRange(t *testing.T) {
	world := newCombatTestWorld()
	target := world.combat.targets[dummyTargetID]
	mustAddPlayer(t, world, "player-one", Position{X: target.Position.X + 13, Z: target.Position.Z}, 4)
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

func TestSwordStylesRouteOrSuppressCombatXP(t *testing.T) {
	tests := []struct {
		name       string
		style      CombatStyleID
		trained    SkillID
		experience int64
	}{
		{name: "stab", style: CombatStyleStab, trained: SkillAttack, experience: swordDamage},
		{name: "slash", style: CombatStyleSlash, trained: SkillStrength, experience: swordDamage},
		{name: "defensive", style: CombatStyleDefensive, trained: SkillDefence, experience: swordDamage},
		{name: "unfocused", style: CombatStyleUnfocused},
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
		damage   int
		interval uint64
		range_   float64
	}{
		{id: WeaponMace, damage: maceDamage, interval: maceAttackIntervalPulses, range_: maceRange},
		{id: WeaponBattleAxe, damage: battleAxeDamage, interval: battleAxeAttackIntervalPulses, range_: battleAxeRange},
		{id: WeaponStaff, damage: staffDamage, interval: staffAttackIntervalPulses, range_: staffRange},
	}
	for _, test := range tests {
		weapon, exists := weaponDefinition(test.id)
		if !exists || weapon.Damage != test.damage || weapon.AttackInterval != test.interval || weapon.Range != test.range_ {
			t.Fatalf("weaponDefinition(%q) = %+v, exists %t", test.id, weapon, exists)
		}
	}
	if !combatStyleValidForWeapon(CombatStyleHeavyCrush, WeaponMace) ||
		!combatStyleValidForWeapon(CombatStyleHeavyCrush, WeaponBattleAxe) ||
		!combatStyleValidForWeapon(CombatStyleLongRange, WeaponStaff) ||
		!combatStyleValidForWeapon(CombatStyleDefensive, WeaponSword) ||
		!combatStyleValidForWeapon(CombatStyleUnfocused, WeaponBow) {
		t.Fatal("additional weapon combat styles are not valid")
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

func TestSwordApproachesAtPreviouslyValidLongRange(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{X: -4, Z: -0.3}, 4)
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}

	world.Step(actionPulseInterval)
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
	if player.CombatTarget != dummyTargetID || player.AutoApproaching {
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
