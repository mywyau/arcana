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

func (source *sequenceRollSource) Float64() float64 {
	value := source.values[source.index]
	source.index++
	return value
}

func newCombatTestWorld() *World {
	return NewWorldWithRollSource(fixedRollSource(0.5))
}

func TestEngagedSwordAutoAttacksOnWeaponCadence(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{X: -4, Z: -1}, 4)
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}

	world.Step(599 * time.Millisecond)
	if got := world.Snapshot().Dummy.Health; got != dummyMaximumHealth {
		t.Fatalf("dummy health before pulse = %d, want %d", got, dummyMaximumHealth)
	}

	world.Step(time.Millisecond)
	if got, want := world.Snapshot().Dummy.Health, dummyMaximumHealth-swordDamage; got != want {
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

	world.Step(3 * actionPulseInterval)
	if got, want := world.Snapshot().Dummy.Health, dummyMaximumHealth-swordDamage; got != want {
		t.Fatalf("dummy health before next weapon pulse = %d, want %d", got, want)
	}
	world.Step(actionPulseInterval)
	if got, want := world.Snapshot().Dummy.Health, dummyMaximumHealth-2*swordDamage; got != want {
		t.Fatalf("dummy health after next weapon pulse = %d, want %d", got, want)
	}
	if got := world.Snapshot().CombatEvents[1].ID; got != 2 {
		t.Fatalf("second combat event ID = %d, want 2", got)
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
	if snapshot.Dummy.Health != dummyMaximumHealth || snapshot.Dummy.TargetID != "" {
		t.Fatalf("dummy state after dodge = %+v", snapshot.Dummy)
	}
	player := mustPlayer(t, world, "player-one")
	if player.LastActionResult != "dodged" || player.NextAttackPulse != 5 {
		t.Fatalf("player state after dodge = %+v", player)
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
	if got, want := snapshot.Dummy.NextAttackPulse, uint64(5+dummyAttackIntervalPulses); got != want {
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
	if got, want := world.Snapshot().Dummy.Health, dummyMaximumHealth-daggerDamage; got != want {
		t.Fatalf("dummy health after dagger hit = %d, want %d", got, want)
	}
	if got, want := player.NextAttackPulse, uint64(1+daggerAttackIntervalPulses); got != want {
		t.Fatalf("NextAttackPulse = %d, want %d", got, want)
	}

	world.Step(2 * actionPulseInterval)
	if got, want := world.Snapshot().Dummy.Health, dummyMaximumHealth-daggerDamage; got != want {
		t.Fatalf("dummy health before dagger cadence = %d, want %d", got, want)
	}
	world.Step(actionPulseInterval)
	if got, want := world.Snapshot().Dummy.Health, dummyMaximumHealth-2*daggerDamage; got != want {
		t.Fatalf("dummy health after second dagger hit = %d, want %d", got, want)
	}
}

func TestBowAttacksFromRangeWithoutApproaching(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{X: -4, Z: 3}, 4)
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
	if got, want := world.Snapshot().Dummy.Health, dummyMaximumHealth-bowDamage; got != want {
		t.Fatalf("dummy health after bow hit = %d, want %d", got, want)
	}
	if got, want := player.NextAttackPulse, uint64(1+bowAttackIntervalPulses); got != want {
		t.Fatalf("NextAttackPulse = %d, want %d", got, want)
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
	if got, want := world.Snapshot().Dummy.Health, dummyMaximumHealth-swordDamage-daggerDamage; got != want {
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
	if got := world.Snapshot().Dummy.Health; got != dummyMaximumHealth {
		t.Fatalf("dummy health while out of range = %d, want %d", got, dummyMaximumHealth)
	}

	for range 4 {
		world.Step(actionPulseInterval)
		if world.Snapshot().Dummy.Health < dummyMaximumHealth {
			break
		}
	}
	if got, want := world.Snapshot().Dummy.Health, dummyMaximumHealth-swordDamage; got != want {
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
	if got := world.Snapshot().Dummy.Health; got != dummyMaximumHealth {
		t.Fatalf("dummy health at 1.7-unit range = %d, want %d", got, dummyMaximumHealth)
	}
	player := mustPlayer(t, world, "player-one")
	if !player.AutoApproaching || player.Destination == nil {
		t.Fatalf("player approach state = %+v", player)
	}
	if got := distance(*player.Destination, world.Snapshot().Dummy.Position); math.Abs(got-swordApproachRange) > 0.001 {
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
	if got := world.Snapshot().Dummy.Health; got != dummyMaximumHealth {
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
	if player.CombatTarget != "" || world.Snapshot().Dummy.TargetID != "player-one" {
		t.Fatalf("combat ownership after disengage: player=%+v dummy=%+v", player, world.Snapshot().Dummy)
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
	if got := world.Snapshot().Dummy.Health; got != dummyMaximumHealth {
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
	if got, want := world.Snapshot().Dummy.TargetID, PlayerID("player-one"); got != want {
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
	world.combat.dummy.Health = swordDamage
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}

	world.Step(actionPulseInterval)
	dummy := world.Snapshot().Dummy
	if !dummy.Defeated || dummy.Health != 0 || dummy.RespawnAtPulse != 6 || dummy.TargetID != "" {
		t.Fatalf("defeated dummy state = %+v", dummy)
	}
	if got := mustPlayer(t, world, "player-one").CombatTarget; got != "" {
		t.Fatalf("player target after dummy defeat = %q, want empty", got)
	}

	world.Step((respawnDelayPulses - 1) * actionPulseInterval)
	if !world.Snapshot().Dummy.Defeated {
		t.Fatal("dummy respawned before the configured pulse")
	}
	world.Step(actionPulseInterval)
	dummy = world.Snapshot().Dummy
	if dummy.Defeated || dummy.Health != dummyMaximumHealth || dummy.RespawnAtPulse != 0 {
		t.Fatalf("respawned dummy state = %+v", dummy)
	}
}
