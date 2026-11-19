package simulation

import (
	"math"
	"testing"
)

func TestTyrantArenaOnlyDetectsPlayersInsideItsBounds(t *testing.T) {
	world := NewWorld()
	arena := tyrantCombatArenas[riftTyrantTargetID]
	mustAddPlayer(t, world, "player-one", Position{X: arena.Center.X, Z: arena.MinimumZ - 1}, 4)
	boss := world.combat.targets[riftTyrantTargetID]

	world.updateRoamingEnemies()
	if boss.TargetID != "" {
		t.Fatalf("Tyrant acquired player outside arena: target %q", boss.TargetID)
	}

	world.players["player-one"].position = arena.Entrance
	world.updateRoamingEnemies()
	if boss.TargetID != "player-one" {
		t.Fatalf("Tyrant target inside arena = %q, want player-one", boss.TargetID)
	}

	world.players["player-one"].position.Z = arena.MinimumZ - 1
	world.updateRoamingEnemies()
	if boss.TargetID != "" {
		t.Fatalf("Tyrant retained player after arena exit: target %q", boss.TargetID)
	}
}

func TestTargetingTyrantFromOutsideApproachesArenaEntranceFirst(t *testing.T) {
	world := NewWorld()
	arena := tyrantCombatArenas[mageTyrantTargetID]
	mustAddPlayer(t, world, "player-one", Position{X: arena.Center.X, Z: arena.MinimumZ - 4}, 4)
	player := world.players["player-one"]

	if err := world.QueueTarget(TargetCommand{
		PlayerID: "player-one", Sequence: 1, TargetID: mageTyrantTargetID,
	}); err != nil {
		t.Fatalf("QueueTarget() returned %v", err)
	}
	world.applyTargetCommands()

	if player.combatTarget != mageTyrantTargetID || !player.autoApproaching {
		t.Fatalf("arena approach state = target %q approaching %v", player.combatTarget, player.autoApproaching)
	}
	if player.destination == nil || *player.destination != arena.Entrance {
		t.Fatalf("arena destination = %v, want %+v", player.destination, arena.Entrance)
	}
}

func TestTyrantArenasAreDistinctAndContainTheirHomes(t *testing.T) {
	centers := make(map[Position]bool)
	for targetID, arena := range tyrantCombatArenas {
		if !arena.contains(arena.Center) || !arena.contains(arena.Entrance) {
			t.Fatalf("arena %q does not contain center or entrance: %+v", targetID, arena)
		}
		if arena.MaximumX-arena.MinimumX != 35 || arena.MaximumZ-arena.MinimumZ != 35 {
			t.Fatalf("arena %q dimensions = %.1fx%.1f, want 35x35", targetID, arena.MaximumX-arena.MinimumX, arena.MaximumZ-arena.MinimumZ)
		}
		if centers[arena.Center] {
			t.Fatalf("arena %q shares center %+v", targetID, arena.Center)
		}
		centers[arena.Center] = true
		profile := combatProfileForEnemy(targetID)
		cornerDistance := math.Hypot(arena.MaximumX-arena.Center.X, arena.MaximumZ-arena.Center.Z)
		if profile.DetectionRange < cornerDistance || profile.LeashRange < cornerDistance {
			t.Fatalf(
				"arena %q corner distance %.2f exceeds detection %.2f or leash %.2f",
				targetID, cornerDistance, profile.DetectionRange, profile.LeashRange,
			)
		}
	}
	if len(centers) != 3 {
		t.Fatalf("distinct arena centers = %d, want 3", len(centers))
	}
}
