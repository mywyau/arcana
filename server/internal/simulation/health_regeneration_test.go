package simulation

import "testing"

func TestPassiveHealthRegenerationUsesOutOfCombatCadence(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	world.players["player-one"].health = 80

	for range passiveHealthRegenerationPulses - 1 {
		world.Step(actionPulseInterval)
	}
	if got := mustPlayer(t, world, "player-one").Health; got != 80 {
		t.Fatalf("health before regeneration cadence = %d, want 80", got)
	}
	world.Step(actionPulseInterval)
	if got := mustPlayer(t, world, "player-one").Health; got != 81 {
		t.Fatalf("health after regeneration cadence = %d, want 81", got)
	}

	world.players["player-one"].health = playerMaximumHealth
	for range passiveHealthRegenerationPulses {
		world.Step(actionPulseInterval)
	}
	if got := mustPlayer(t, world, "player-one").Health; got != playerMaximumHealth {
		t.Fatalf("regeneration exceeded maximum health: got %d", got)
	}
}

func TestPassiveHealthRegenerationScalesWithMaximumHealth(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	player := world.players["player-one"]
	player.skills[SkillVitality] = SkillExperienceForLevel(20)
	refreshPlayerMaximumHealth(player)
	player.health = 200

	for range passiveHealthRegenerationPulses {
		world.regeneratePlayersOutOfCombat()
	}
	if got, want := player.health, 203; got != want {
		t.Fatalf("level 20 passive healing = %d, want %d", got, want)
	}
}

func TestPassiveHealthRegenerationStopsDuringEitherCombatDirection(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	player := world.players["player-one"]
	player.health = 80

	player.combatTarget = dummyTargetID
	world.regeneratePlayersOutOfCombat()
	if player.health != 80 {
		t.Fatalf("health regenerated while attacking: %d", player.health)
	}

	player.combatTarget = ""
	world.combat.targets[dummyTargetID].TargetID = player.id
	world.regeneratePlayersOutOfCombat()
	if player.health != 80 {
		t.Fatalf("health regenerated while targeted: %d", player.health)
	}

	world.combat.targets[dummyTargetID].TargetID = ""
	for range passiveHealthRegenerationPulses {
		world.regeneratePlayersOutOfCombat()
	}
	if player.health != 81 {
		t.Fatalf("health after leaving combat = %d, want 81", player.health)
	}
}
