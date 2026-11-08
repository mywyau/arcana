package simulation

import (
	"testing"
	"time"
)

func TestEnemyCombatProfilesCreateDistinctEncounters(t *testing.T) {
	fowl := combatProfileForEnemy(wildFowlTargetID)
	warden := combatProfileForEnemy(hillWardenTargetID)
	if fowl.CombatLevel != 2 || fowl.AttackDamage != 4 || fowl.AttackInterval != 3 || fowl.DamageType != DamageTypeStab ||
		fowl.DetectionRange != 3 || fowl.LeashRange != 5 {
		t.Fatalf("Wild Fowl profile = %+v", fowl)
	}
	if warden.CombatLevel != 18 || warden.AttackDamage != 18 || warden.AttackInterval != 5 || warden.DamageType != DamageTypeSlash ||
		warden.DetectionRange != 6 || warden.LeashRange != 10 {
		t.Fatalf("Hill Warden profile = %+v", warden)
	}
}

func TestWildFowlUsesItsOwnDamageAndAttackInterval(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.99))
	fowl := world.combat.targets[wildFowlTargetID]
	mustAddPlayer(t, world, "player-one", fowl.Position, 4)
	fowl.TargetID = "player-one"
	fowl.NextAttackPulse = uint64(combatProfileForEnemy(wildFowlTargetID).AttackInterval)

	world.Step(2 * actionPulseInterval)
	if got := mustPlayer(t, world, "player-one").Health; got != playerMaximumHealth {
		t.Fatalf("health before Wild Fowl attack = %d, want %d", got, playerMaximumHealth)
	}
	world.Step(actionPulseInterval)
	if got, want := mustPlayer(t, world, "player-one").Health, playerMaximumHealth-wildFowlAttackDamage; got != want {
		t.Fatalf("health after Wild Fowl attack = %d, want %d", got, want)
	}
}

func TestHillWardenUsesSlashDamageAgainstArmour(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.99))
	warden := world.combat.targets[hillWardenTargetID]
	mustAddPlayer(t, world, "player-one", warden.Position, 4)
	player := world.players["player-one"]
	player.equippedArmour[ArmourSlotBody] = ItemIronPlate
	warden.TargetID = player.id
	warden.NextAttackPulse = uint64(combatProfileForEnemy(hillWardenTargetID).AttackInterval)

	profile := combatProfileForEnemy(hillWardenTargetID)
	wantDamage := reduceIncomingDamage(player, profile.AttackDamage, DamageTypeSlash)
	if wantDamage >= profile.AttackDamage {
		t.Fatalf("iron armour did not reduce Slash damage: got %d from %d", wantDamage, profile.AttackDamage)
	}
	world.Step(time.Duration(profile.AttackInterval) * actionPulseInterval)
	if got, want := mustPlayer(t, world, "player-one").Health, playerMaximumHealth-wantDamage; got != want {
		t.Fatalf("health after Hill Warden attack = %d, want %d", got, want)
	}
}
