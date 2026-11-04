package simulation

import (
	"testing"
	"time"
)

func TestEngagedSwordAutoAttacksOnWeaponCadence(t *testing.T) {
	world := NewWorld()
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

func TestEngagedPlayerApproachesBeforeAutoAttacking(t *testing.T) {
	world := NewWorld()
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

func TestDisengageStopsAutomaticApproach(t *testing.T) {
	world := NewWorld()
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

func TestManualMovementKeepsTargetAndApproachResumes(t *testing.T) {
	world := NewWorld()
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
	world.Step(50 * time.Millisecond)
	player := mustPlayer(t, world, "player-one")
	if player.CombatTarget != dummyTargetID || player.AutoApproaching {
		t.Fatalf("player state during manual reposition = %+v", player)
	}

	world.Step(actionPulseInterval - 50*time.Millisecond)
	player = mustPlayer(t, world, "player-one")
	if player.CombatTarget != dummyTargetID || !player.AutoApproaching || player.LastActionResult != "approaching" {
		t.Fatalf("player state after combat pulse = %+v", player)
	}
}

func TestLatestTargetCommandWinsBeforePulse(t *testing.T) {
	world := NewWorld()
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
