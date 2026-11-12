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
		warden.DetectionRange != 6 || warden.LeashRange != 22 {
		t.Fatalf("Hill Warden profile = %+v", warden)
	}
}

func TestElementalMonstersHaveDistinctSpellWeaknesses(t *testing.T) {
	tests := []struct {
		targetID        string
		kind            string
		weakElement     MagicElement
		resistedElement MagicElement
		position        Position
	}{
		{dustWraithTargetID, targetKindDust, MagicElementAir, MagicElementEarth, Position{X: -8, Z: 34}},
		{cinderImpTargetID, targetKindCinder, MagicElementWater, MagicElementFire, Position{X: 36, Z: 4}},
		{tideWispTargetID, targetKindTide, MagicElementEarth, MagicElementWater, Position{X: -36, Z: 4}},
		{thornlingTargetID, targetKindThorn, MagicElementFire, MagicElementEarth, Position{X: 8, Z: 34}},
	}
	world := NewWorld()
	for _, test := range tests {
		t.Run(test.targetID, func(t *testing.T) {
			target := world.combat.targets[test.targetID]
			if target == nil || target.Kind != test.kind || target.Position != test.position {
				t.Fatalf("elemental target = %+v", target)
			}
			if _, err := world.locate(target.Position); err != nil {
				t.Fatalf("elemental target is outside navigable world: %v", err)
			}
			if got := elementalDamageMultiplier(target, test.weakElement); got != 1.30 {
				t.Fatalf("weakness multiplier = %.2f, want 1.30", got)
			}
			if got := elementalDamageMultiplier(target, test.resistedElement); got >= 1 {
				t.Fatalf("resistance multiplier = %.2f, want below 1", got)
			}
			if _, roaming := world.combat.roamingEnemies[test.targetID]; !roaming {
				t.Fatal("elemental monster is not registered as roaming")
			}
		})
	}
}

func TestRiftTyrantHasLongRangeAndMeleeAttackStyles(t *testing.T) {
	profile := combatProfileForEnemy(riftTyrantTargetID)
	wantStyles := []DamageType{
		DamageTypeStab, DamageTypeSlash, DamageTypeCrush,
	}
	if profile.CombatLevel != 35 || profile.AttackRange != enemyLongRangeAttackRange || profile.DetectionRange != 18 ||
		profile.AttackDamage != 22 || len(profile.AttackTypes) != len(wantStyles) {
		t.Fatalf("Rift Tyrant profile = %+v", profile)
	}
	for index, want := range wantStyles {
		if got := enemyAttackType(profile, uint64(index)); got != want {
			t.Errorf("attack type %d = %q, want %q", index, got, want)
		}
	}
	if got := enemyAttackType(profile, uint64(len(wantStyles))); got != wantStyles[0] {
		t.Fatalf("cycled attack type = %q, want %q", got, wantStyles[0])
	}
}

func TestSpecialistTyrantsUseTheirOwnCombatStyles(t *testing.T) {
	tests := []struct {
		targetID       string
		wantDamageType DamageType
		wantDamage     int
		wantInterval   uint64
	}{
		{targetID: rangedTyrantTargetID, wantDamageType: DamageTypeRanged, wantDamage: 20, wantInterval: 3},
		{targetID: mageTyrantTargetID, wantDamageType: DamageTypeMagic, wantDamage: 24, wantInterval: 5},
	}
	for _, test := range tests {
		t.Run(test.targetID, func(t *testing.T) {
			profile := combatProfileForEnemy(test.targetID)
			if profile.CombatLevel != 35 || profile.DamageType != test.wantDamageType ||
				profile.AttackDamage != test.wantDamage || profile.AttackInterval != test.wantInterval ||
				profile.AttackRange != enemyLongRangeAttackRange || profile.DetectionRange != 18 || profile.LeashRange != 24 {
				t.Fatalf("profile = %+v", profile)
			}
			if len(profile.AttackTypes) != 0 {
				t.Fatalf("specialist attack types = %v, want fixed %q", profile.AttackTypes, test.wantDamageType)
			}
		})
	}
}

func TestWardenAndTyrantOutrangeLongRangeBow(t *testing.T) {
	longRangeBow, exists := weaponDefinitionForStyle(WeaponBow, CombatStyleLongRange)
	if !exists {
		t.Fatal("long-range bow definition is missing")
	}
	for _, targetID := range []string{hillWardenTargetID, riftTyrantTargetID, rangedTyrantTargetID, mageTyrantTargetID} {
		profile := combatProfileForEnemy(targetID)
		if profile.AttackRange <= longRangeBow.Range {
			t.Errorf("%s attack range = %.2f, want greater than long-range bow %.2f", targetID, profile.AttackRange, longRangeBow.Range)
		}
		if profile.LeashRange <= longRangeBow.Range {
			t.Errorf("%s leash range = %.2f, want greater than long-range bow %.2f", targetID, profile.LeashRange, longRangeBow.Range)
		}
	}
}

func TestSpecialistTyrantsHaveDistinctWorldSpawnsAndVisualKinds(t *testing.T) {
	world := NewWorld()
	ranged := world.combat.targets[rangedTyrantTargetID]
	mage := world.combat.targets[mageTyrantTargetID]
	if ranged == nil || ranged.Position != (Position{X: -40, Z: -35}) || ranged.Kind != targetKindRangedBoss {
		t.Fatalf("Ranged Tyrant = %+v", ranged)
	}
	if mage == nil || mage.Position != (Position{X: 38, Z: -34}) || mage.Kind != targetKindMageBoss {
		t.Fatalf("Mage Tyrant = %+v", mage)
	}
}

func TestRiftTyrantMovesSlightlyFasterThanNewPlayer(t *testing.T) {
	world := NewWorld()
	boss := world.combat.roamingEnemies[riftTyrantTargetID]
	if boss.speed != DefaultPlayerMovementSpeed+0.5 {
		t.Fatalf("Rift Tyrant speed = %.2f, want %.2f", boss.speed, DefaultPlayerMovementSpeed+0.5)
	}
}

func TestRiftTyrantCyclesAttackStyleOnEachAttack(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.5))
	boss := world.combat.targets[riftTyrantTargetID]
	mustAddPlayer(t, world, "player-one", boss.Position, 4)
	player := world.players["player-one"]
	player.maximumHealth = 1_000
	player.health = 1_000
	boss.TargetID = player.id
	profile := combatProfileForEnemy(riftTyrantTargetID)

	wantStyles := []DamageType{DamageTypeStab, DamageTypeSlash, DamageTypeCrush}
	for index, want := range wantStyles {
		world.combat.pulse = uint64(index+1) * profile.AttackInterval
		boss.NextAttackPulse = world.combat.pulse
		world.resolveOneDummyAttack(riftTyrantTargetID, boss)
		if boss.AttackType != want {
			t.Fatalf("attack %d type = %q, want %q", index+1, boss.AttackType, want)
		}
	}
}

func TestTyrantsHaveBossTierDefence(t *testing.T) {
	tests := []struct {
		targetID string
		type_    DamageType
		minimum  int
	}{
		{targetID: riftTyrantTargetID, type_: DamageTypeSlash, minimum: 60},
		{targetID: rangedTyrantTargetID, type_: DamageTypeRanged, minimum: 65},
		{targetID: mageTyrantTargetID, type_: DamageTypeMagic, minimum: 70},
	}
	for _, test := range tests {
		profile := combatProfileForEnemy(test.targetID)
		if got := defenceBonusForDamageType(profile.Defence, test.type_); got < test.minimum {
			t.Errorf("%s %s defence = %d, want at least %d", test.targetID, test.type_, got, test.minimum)
		}
	}
}

func TestWildFowlUsesItsOwnDamageAndAttackInterval(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.5))
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
	world := NewWorldWithRollSource(fixedRollSource(0.5))
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
