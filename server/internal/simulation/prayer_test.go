package simulation

import (
	"math"
	"testing"
)

func TestEvasionPrayerActivatesAndDrainsOnCadence(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.5))
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	if err := world.QueuePrayer(PrayerCommand{
		PlayerID: "player-one",
		Sequence: 1,
		PrayerID: PrayerEvasion,
		Active:   true,
	}); err != nil {
		t.Fatalf("QueuePrayer() returned an unexpected error: %v", err)
	}

	world.Step(actionPulseInterval)
	player := mustPlayer(t, world, "player-one")
	if len(player.ActivePrayers) != 1 || player.ActivePrayers[0] != PrayerEvasion {
		t.Fatalf("active prayers = %v, want [%s]", player.ActivePrayers, PrayerEvasion)
	}
	if player.LastPrayerSequence != 1 || player.Prayer != playerMaximumPrayer {
		t.Fatalf("prayer state after activation = sequence %d, points %d", player.LastPrayerSequence, player.Prayer)
	}
	wantDodge := calculatedDodgeChance(dummyAccuracyScore, 1, playerDodgeRate+evasionDodgeBonus)
	if math.Abs(player.DodgeRate-wantDodge) > 0.000_001 {
		t.Fatalf("effective dodge rate = %.3f, want %.3f", player.DodgeRate, wantDodge)
	}

	for range evasionDrainIntervalPulses - 1 {
		world.Step(actionPulseInterval)
	}
	player = mustPlayer(t, world, "player-one")
	if player.Prayer != playerMaximumPrayer-1 {
		t.Fatalf("prayer points after one drain interval = %d, want %d", player.Prayer, playerMaximumPrayer-1)
	}
}

func TestEvasionPrayerDeactivatesAtZeroPoints(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.5))
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	world.players["player-one"].prayer = 1
	if err := world.QueuePrayer(PrayerCommand{
		PlayerID: "player-one",
		Sequence: 1,
		PrayerID: PrayerEvasion,
		Active:   true,
	}); err != nil {
		t.Fatalf("QueuePrayer() returned an unexpected error: %v", err)
	}

	for range evasionDrainIntervalPulses {
		world.Step(actionPulseInterval)
	}
	player := mustPlayer(t, world, "player-one")
	if player.Prayer != 0 || len(player.ActivePrayers) != 0 {
		t.Fatalf("prayer state at depletion = points %d, active %v", player.Prayer, player.ActivePrayers)
	}
	wantDodge := calculatedDodgeChance(dummyAccuracyScore, 1, playerDodgeRate)
	if player.DodgeRate != wantDodge {
		t.Fatalf("dodge rate after depletion = %.3f, want %.3f", player.DodgeRate, wantDodge)
	}
}

func TestPrayerCanBeToggledOffBeforeDrain(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.5))
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	if err := world.QueuePrayer(PrayerCommand{
		PlayerID: "player-one", Sequence: 1, PrayerID: PrayerEvasion, Active: true,
	}); err != nil {
		t.Fatalf("QueuePrayer(activate) returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	if err := world.QueuePrayer(PrayerCommand{
		PlayerID: "player-one", Sequence: 2, PrayerID: PrayerEvasion, Active: false,
	}); err != nil {
		t.Fatalf("QueuePrayer(deactivate) returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)

	player := mustPlayer(t, world, "player-one")
	if len(player.ActivePrayers) != 0 || player.Prayer != playerMaximumPrayer || player.LastPrayerSequence != 2 {
		t.Fatalf("prayer state after deactivation = %+v", player)
	}
}

func TestEvasionPrayerAppliesBeforeIncomingAttack(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.12))
	dummy := world.combat.targets[dummyTargetID]
	mustAddPlayer(t, world, "player-one", dummy.Position, 4)
	dummy.TargetID = "player-one"
	if err := world.QueuePrayer(PrayerCommand{
		PlayerID: "player-one", Sequence: 1, PrayerID: PrayerEvasion, Active: true,
	}); err != nil {
		t.Fatalf("QueuePrayer() returned an unexpected error: %v", err)
	}

	world.Step(actionPulseInterval)
	player := mustPlayer(t, world, "player-one")
	if player.Health != playerMaximumHealth {
		t.Fatalf("health after roll inside Evasion dodge chance = %d, want %d", player.Health, playerMaximumHealth)
	}
	events := world.Snapshot().CombatEvents
	if len(events) != 1 || events[0].Type != CombatEventDodged || events[0].TargetID != "player-one" {
		t.Fatalf("combat events after Evasion dodge = %+v", events)
	}
}
