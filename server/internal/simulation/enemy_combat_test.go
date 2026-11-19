package simulation

import (
	"fmt"
	"math"
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
		{dustWraithTargetID, targetKindDust, MagicElementAir, MagicElementEarth, Position{X: -5, Z: 30}},
		{cinderImpTargetID, targetKindCinder, MagicElementWater, MagicElementFire, Position{X: 5, Z: 30}},
		{tideWispTargetID, targetKindTide, MagicElementEarth, MagicElementWater, Position{X: 15, Z: 30}},
		{thornlingTargetID, targetKindThorn, MagicElementFire, MagicElementEarth, Position{X: 25, Z: 30}},
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
	if profile.CombatLevel != 35 || profile.AttackRange != enemyLongRangeAttackRange || profile.DetectionRange != 26 ||
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
				profile.AttackRange != enemyLongRangeAttackRange || profile.DetectionRange != 26 || profile.LeashRange != 27 {
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
	rift := world.combat.targets[riftTyrantTargetID]
	ranged := world.combat.targets[rangedTyrantTargetID]
	mage := world.combat.targets[mageTyrantTargetID]
	if rift == nil || rift.Position != (Position{X: -52, Z: 55}) || rift.Kind != targetKindBoss {
		t.Fatalf("Rift Tyrant = %+v", rift)
	}
	if ranged == nil || ranged.Position != (Position{X: 0, Z: 55}) || ranged.Kind != targetKindRangedBoss {
		t.Fatalf("Ranged Tyrant = %+v", ranged)
	}
	if mage == nil || mage.Position != (Position{X: 52, Z: 55}) || mage.Kind != targetKindMageBoss {
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

func TestRangedTyrantRepositionsInsideArenaBeforeAttacking(t *testing.T) {
	world := NewWorld()
	boss := world.combat.roamingEnemies[rangedTyrantTargetID]
	arena := tyrantCombatArenas[rangedTyrantTargetID]
	playerPosition := arena.Center
	playerPosition.X -= 2
	mustAddPlayer(t, world, "player-one", playerPosition, 4)

	world.updateRoamingEnemies()
	if boss.target.TargetID != "player-one" {
		t.Fatalf("Ranged Tyrant target = %q, want player-one", boss.target.TargetID)
	}
	if len(boss.path) == 0 {
		t.Fatal("Ranged Tyrant did not choose a reposition path during its attack cooldown")
	}
	destination := boss.path[len(boss.path)-1]
	if !arena.contains(destination) || destination == boss.target.Position {
		t.Fatalf("Ranged Tyrant reposition destination = %+v, arena = %+v", destination, arena)
	}
	if destination.X <= boss.target.Position.X {
		t.Fatalf("Ranged Tyrant moved toward player at %+v instead of kiting to %+v", playerPosition, destination)
	}

	start := boss.target.Position
	world.moveRoamingEnemies(0.5)
	if boss.target.Position == start {
		t.Fatal("Ranged Tyrant did not move along its reposition path")
	}
	if !arena.contains(boss.target.Position) {
		t.Fatalf("Ranged Tyrant moved outside its arena to %+v", boss.target.Position)
	}

	world.combat.pulse = boss.target.NextAttackPulse
	world.updateRoamingEnemies()
	if len(boss.path) != 0 {
		t.Fatalf("Ranged Tyrant retained reposition path when ready to attack: %+v", boss.path)
	}
}

func TestRangedTyrantRapidlyHealsAfterPlayerLeavesArena(t *testing.T) {
	world := NewWorld()
	boss := world.combat.roamingEnemies[rangedTyrantTargetID]
	arena := tyrantCombatArenas[rangedTyrantTargetID]
	mustAddPlayer(t, world, "player-one", arena.Center, 4)
	boss.target.TargetID = "player-one"
	boss.target.Health = boss.target.MaximumHealth / 2

	world.players["player-one"].position = Position{X: arena.Center.X, Z: arena.MinimumZ - 1}
	world.updateRoamingEnemies()

	if boss.target.TargetID != "" {
		t.Fatalf("Ranged Tyrant retained player outside arena: %q", boss.target.TargetID)
	}
	wantHealth := boss.target.MaximumHealth/2 + boss.target.MaximumHealth/10
	if boss.target.Health != wantHealth {
		t.Fatalf("Ranged Tyrant health after disengaging = %d, want %d", boss.target.Health, wantHealth)
	}
}

func TestRangedTyrantSideStepsOutOfArenaCorner(t *testing.T) {
	world := NewWorld()
	boss := world.combat.roamingEnemies[rangedTyrantTargetID]
	arena := tyrantCombatArenas[rangedTyrantTargetID]
	boss.target.Position = Position{X: arena.MaximumX - 1, Z: arena.MaximumZ - 1}
	playerPosition := Position{X: arena.Center.X, Z: arena.Center.Z}

	world.setRangedTyrantRepositionPath(boss, playerPosition)
	if len(boss.path) == 0 {
		t.Fatal("Ranged Tyrant did not find a side-step out of the arena corner")
	}
	destination := boss.path[len(boss.path)-1]
	if !arena.contains(destination) || distance(destination, boss.target.Position) < 1 {
		t.Fatalf("corner escape destination = %+v from %+v", destination, boss.target.Position)
	}
	if destination.X >= boss.target.Position.X && destination.Z >= boss.target.Position.Z {
		t.Fatalf("corner escape did not move inward or sideways: %+v", destination)
	}
}

func TestRangedTyrantFollowsClockwiseOrbitAfterOpeningKite(t *testing.T) {
	world := NewWorld()
	boss := world.combat.roamingEnemies[rangedTyrantTargetID]
	arena := tyrantCombatArenas[rangedTyrantTargetID]
	playerPosition := Position{X: arena.Center.X - 2, Z: arena.Center.Z}

	world.setRangedTyrantRepositionPath(boss, playerPosition)
	first := boss.path[len(boss.path)-1]
	boss.path = nil
	world.setRangedTyrantRepositionPath(boss, playerPosition)
	second := boss.path[len(boss.path)-1]

	for _, point := range []Position{first, second} {
		if got := distance(point, arena.Center); math.Abs(got-rangedTyrantRepositionOffset) > 0.001 {
			t.Fatalf("orbit point %+v is %.2f from centre, want %.2f", point, got, rangedTyrantRepositionOffset)
		}
	}
	firstX, firstZ := first.X-arena.Center.X, first.Z-arena.Center.Z
	secondX, secondZ := second.X-arena.Center.X, second.Z-arena.Center.Z
	dot := firstX*secondX + firstZ*secondZ
	cross := firstX*secondZ - firstZ*secondX
	if first == second || math.Abs(dot-0.5*rangedTyrantRepositionOffset*rangedTyrantRepositionOffset) > 0.001 || cross >= 0 {
		t.Fatalf("clockwise orbit opening = first %+v second %+v around %+v", first, second, arena.Center)
	}
}

func TestRangedTyrantCentersBeforePiercingPlayerAtArenaWall(t *testing.T) {
	world := NewWorld()
	boss := world.combat.roamingEnemies[rangedTyrantTargetID]
	arena := tyrantCombatArenas[rangedTyrantTargetID]
	playerPosition := Position{X: arena.MaximumX, Z: arena.Center.Z}
	mustAddPlayer(t, world, "player-one", playerPosition, 4)
	boss.target.TargetID = "player-one"
	boss.target.attackSequence = 7
	boss.target.NextAttackPulse = 10
	boss.target.Position = Position{X: arena.MinimumX + 2, Z: arena.Center.Z}
	world.combat.pulse = 10

	world.updateRoamingEnemies()
	if boss.target.NextAttackPulse != 11 || len(boss.path) == 0 || boss.path[len(boss.path)-1] != arena.Center {
		t.Fatalf("piercing centering state = next %d path %+v", boss.target.NextAttackPulse, boss.path)
	}

	boss.target.Position = arena.Center
	boss.path = nil
	world.combat.pulse = boss.target.NextAttackPulse
	world.updateRoamingEnemies()
	world.resolveOneDummyAttack(rangedTyrantTargetID, boss.target)
	if boss.target.TelegraphType != "piercing_shot" {
		t.Fatalf("centred Tyrant did not target arena-wall player with piercing volley: %+v", boss.target)
	}
}

func TestRangedTyrantRetainsCircularHeavyStrike(t *testing.T) {
	world := NewWorld()
	boss := world.combat.targets[rangedTyrantTargetID]
	mustAddPlayer(t, world, "player-one", Position{X: boss.Position.X, Z: boss.Position.Z + 5}, 4)
	boss.TargetID = "player-one"
	boss.attackSequence = 3
	boss.NextAttackPulse = 10
	world.combat.pulse = 10

	world.resolveOneDummyAttack(rangedTyrantTargetID, boss)
	if boss.TelegraphType != "heavy_strike" || boss.TelegraphRadius != 5 {
		t.Fatalf("Ranged Tyrant fourth attack = %+v, want circular heavy strike", boss)
	}
}

func TestRangedTyrantKeepsMovingWhenHeavyStrikeMissesItself(t *testing.T) {
	world := NewWorld()
	bossState := world.combat.roamingEnemies[rangedTyrantTargetID]
	boss := bossState.target
	mustAddPlayer(t, world, "player-one", Position{X: boss.Position.X, Z: boss.Position.Z + 10}, 4)
	boss.TargetID = "player-one"
	boss.attackSequence = 3
	boss.NextAttackPulse = 10
	world.combat.pulse = 10

	world.resolveOneDummyAttack(rangedTyrantTargetID, boss)
	world.updateRoamingEnemies()
	if len(bossState.path) == 0 {
		t.Fatal("Ranged Tyrant stopped moving during its heavy-strike warning")
	}
	world.combat.pulse = boss.TelegraphImpactPulse
	world.resolveOneDummyAttack(rangedTyrantTargetID, boss)
	if boss.VulnerableUntilPulse != 0 || boss.stunnedUntilPulse != 0 {
		t.Fatalf("missed self-hit applied vulnerability or stun: %+v", boss)
	}
	if len(bossState.path) == 0 {
		t.Fatal("Ranged Tyrant lost its movement path after the circle missed it")
	}
}

func TestRangedTyrantSelfHitCreatesLongStunWindow(t *testing.T) {
	world := NewWorld()
	bossState := world.combat.roamingEnemies[rangedTyrantTargetID]
	boss := bossState.target
	mustAddPlayer(t, world, "player-one", boss.Position, 4)
	boss.TargetID = "player-one"
	boss.attackSequence = 3
	boss.NextAttackPulse = 10
	world.combat.pulse = 10

	world.resolveOneDummyAttack(rangedTyrantTargetID, boss)
	world.combat.pulse = boss.TelegraphImpactPulse
	world.resolveOneDummyAttack(rangedTyrantTargetID, boss)
	wantUntil := world.combat.pulse + 7
	if boss.VulnerableUntilPulse != wantUntil || boss.stunnedUntilPulse != wantUntil {
		t.Fatalf("self-hit window = vulnerable %d stunned %d, want %d",
			boss.VulnerableUntilPulse, boss.stunnedUntilPulse, wantUntil)
	}

	start := boss.Position
	bossState.path = []Position{{X: start.X + 5, Z: start.Z}}
	world.moveRoamingEnemies(1)
	if boss.Position != start || len(bossState.path) != 0 {
		t.Fatalf("self-stunned Tyrant moved from %+v to %+v along %+v", start, boss.Position, bossState.path)
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
		targetID      string
		type_         DamageType
		minimum       int
		maximumHealth int
	}{
		{targetID: riftTyrantTargetID, type_: DamageTypeSlash, minimum: 110, maximumHealth: 1100},
		{targetID: rangedTyrantTargetID, type_: DamageTypeRanged, minimum: 115, maximumHealth: 5000},
		{targetID: mageTyrantTargetID, type_: DamageTypeMagic, minimum: 120, maximumHealth: 950},
	}
	for _, test := range tests {
		profile := combatProfileForEnemy(test.targetID)
		if got := defenceBonusForDamageType(profile.Defence, test.type_); got < test.minimum {
			t.Errorf("%s %s defence = %d, want at least %d", test.targetID, test.type_, got, test.minimum)
		}
		if got := NewWorld().combat.targets[test.targetID].MaximumHealth; got != test.maximumHealth {
			t.Errorf("%s maximum health = %d, want %d", test.targetID, got, test.maximumHealth)
		}
		if test.targetID == rangedTyrantTargetID {
			volley := profile.PiercingVolley
			if volley.CycleAttacks != 12 || volley.TriggerAttack != 8 || volley.Shots != 3 || volley.WindupPulses != 3 ||
				volley.BetweenPulses != 2 || volley.Length != 32 || volley.Width != 2 {
				t.Errorf("%s piercing volley = %+v", test.targetID, volley)
			}
			arrows := profile.ArrowVolley
			if arrows.CycleAttacks != 12 || arrows.TriggerAttack != 12 || arrows.Waves != 3 ||
				arrows.WindupPulses != 3 || arrows.BetweenPulses != 2 || arrows.Radius != 1.8 {
				t.Errorf("%s arrow volley = %+v", test.targetID, arrows)
			}
			if profile.Defence.Ranged <= profile.Defence.Slash || profile.Defence.Magic <= profile.Defence.Slash {
				t.Errorf("%s should resist ranged and magic more than melee: %+v", test.targetID, profile.Defence)
			}
			if profile.Defence.Ranged != 115 || profile.Defence.Magic != 150 {
				t.Errorf("%s specialist defences = %+v", test.targetID, profile.Defence)
			}
		}
		wantVulnerabilityPulses := uint64(3)
		if test.targetID == rangedTyrantTargetID {
			wantVulnerabilityPulses = 7
		}
		if profile.HeavyStrike.EveryAttacks != 4 || profile.HeavyStrike.WindupPulses != 4 ||
			profile.HeavyStrike.Radius != 5 || profile.HeavyStrike.DamageMultiplier != 2 ||
			profile.HeavyStrike.VulnerabilityPulses != wantVulnerabilityPulses ||
			profile.HeavyStrike.VulnerabilityDamageMultiplier != 1.25 {
			t.Errorf("%s heavy strike = %+v", test.targetID, profile.HeavyStrike)
		}
	}
}

func TestRangedTyrantArrowVolleyUsesThreeShiftedWaves(t *testing.T) {
	world := NewWorld()
	boss := world.combat.targets[rangedTyrantTargetID]
	arena := tyrantCombatArenas[rangedTyrantTargetID]
	mustAddPlayer(t, world, "player-one", Position{X: arena.Center.X, Z: arena.Center.Z + 5}, 4)
	mustAddPlayer(t, world, "player-safe", Position{X: arena.MaximumX - 0.5, Z: arena.MaximumZ - 0.5}, 4)
	player := world.players["player-one"]
	player.maximumHealth = 1000
	player.health = 1000
	boss.TargetID = player.id
	boss.attackSequence = 11
	boss.NextAttackPulse = 10
	world.combat.pulse = 10

	world.resolveOneDummyAttack(rangedTyrantTargetID, boss)
	if boss.TelegraphType != "arrow_volley_1" || boss.TelegraphImpactPulse != 13 {
		t.Fatalf("arrow volley opening = %+v", boss)
	}
	var previousPattern []Position
	for wave := 0; wave < 3; wave++ {
		pattern := arrowVolleyPositions(boss.TelegraphPosition, wave)
		if len(pattern) == 0 {
			t.Fatalf("arrow volley wave %d has no circles", wave+1)
		}
		if wave > 0 && pattern[1] == previousPattern[1] {
			t.Fatalf("arrow volley wave %d did not shift", wave+1)
		}
		previousPattern = pattern
		player.position = pattern[0]
		before := player.health
		world.combat.pulse = boss.TelegraphImpactPulse
		world.resolveOneDummyAttack(rangedTyrantTargetID, boss)
		if player.health >= before {
			t.Fatalf("player in wave %d circle was not hit", wave+1)
		}
		if world.players["player-safe"].health != playerMaximumHealth {
			t.Fatalf("safe player was hit during wave %d", wave+1)
		}
		if wave < 2 && (boss.TelegraphType != fmt.Sprintf("arrow_volley_%d", wave+2) ||
			boss.TelegraphImpactPulse != world.combat.pulse+2) {
			t.Fatalf("arrow volley wave %d transition = %+v", wave+1, boss)
		}
	}
	if boss.TelegraphType != "" || boss.arrowVolleyWave != 0 {
		t.Fatalf("completed arrow volley was not cleared: %+v", boss)
	}
}

func TestRangedTyrantPiercingVolleyFansOutAndHitsEveryPlayerInLane(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.5))
	boss := world.combat.targets[rangedTyrantTargetID]
	mustAddPlayer(t, world, "player-one", Position{X: boss.Position.X, Z: boss.Position.Z + 5}, 4)
	mustAddPlayer(t, world, "player-two", Position{X: boss.Position.X, Z: boss.Position.Z + 10}, 4)
	mustAddPlayer(t, world, "player-off", Position{X: boss.Position.X + 5, Z: boss.Position.Z}, 4)
	boss.TargetID = "player-one"
	boss.attackSequence = 7
	boss.NextAttackPulse = 10
	world.combat.pulse = 10

	world.resolveOneDummyAttack(rangedTyrantTargetID, boss)
	if boss.TelegraphType != "piercing_shot" || len(boss.piercingDirections) != 3 ||
		boss.TelegraphImpactPulse != 13 {
		t.Fatalf("piercing volley windup = %+v", boss)
	}
	if boss.piercingDirections[0].X <= 0 || math.Abs(boss.piercingDirections[1].X) > 0.001 ||
		boss.piercingDirections[2].X >= 0 {
		t.Fatalf("piercing directions do not form a left-to-right fan: %+v", boss.piercingDirections)
	}

	direction := boss.piercingDirections[0]
	world.players["player-one"].position = Position{
		X: boss.Position.X + direction.X*5, Z: boss.Position.Z + direction.Z*5,
	}
	world.players["player-two"].position = Position{
		X: boss.Position.X + direction.X*10, Z: boss.Position.Z + direction.Z*10,
	}
	world.combat.pulse = boss.TelegraphImpactPulse
	world.resolveOneDummyAttack(rangedTyrantTargetID, boss)

	if world.players["player-one"].health >= playerMaximumHealth ||
		world.players["player-two"].health >= playerMaximumHealth {
		t.Fatalf("aligned players were not both pierced: one=%d two=%d",
			world.players["player-one"].health, world.players["player-two"].health)
	}
	if world.players["player-off"].health != playerMaximumHealth {
		t.Fatalf("player outside fired lane was hit: health=%d", world.players["player-off"].health)
	}
	if boss.TelegraphType != "piercing_shot" || boss.piercingShotIndex != 1 ||
		boss.TelegraphImpactPulse != world.combat.pulse+2 {
		t.Fatalf("second piercing shot was not delayed: %+v", boss)
	}
}

func TestTyrantHeavyStrikeTelegraphsAndCanBeEscaped(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.5))
	boss := world.combat.targets[riftTyrantTargetID]
	mustAddPlayer(t, world, "player-one", boss.Position, 4)
	player := world.players["player-one"]
	player.maximumHealth = 200
	player.health = 200
	boss.TargetID = player.id
	boss.attackSequence = 3
	boss.NextAttackPulse = 10
	world.combat.pulse = 10

	world.resolveOneDummyAttack(riftTyrantTargetID, boss)
	if player.health != 200 || boss.TelegraphType != "heavy_strike" ||
		boss.TelegraphPosition != player.position || boss.TelegraphRadius != 5 || boss.TelegraphImpactPulse != 14 {
		t.Fatalf("heavy strike windup: player health=%d boss=%+v", player.health, boss)
	}

	player.position.X += boss.TelegraphRadius + 1
	world.combat.pulse = boss.TelegraphImpactPulse
	world.resolveOneDummyAttack(riftTyrantTargetID, boss)
	if player.health != 200 {
		t.Fatalf("escaped heavy strike health = %d, want 200", player.health)
	}
	if boss.TelegraphType != "" || boss.TelegraphRadius != 0 || boss.TelegraphImpactPulse != 0 {
		t.Fatalf("resolved telegraph was not cleared: %+v", boss)
	}
	if boss.VulnerableUntilPulse != world.combat.pulse+3 {
		t.Fatalf("vulnerability ends at pulse %d, want %d", boss.VulnerableUntilPulse, world.combat.pulse+3)
	}
	if statusActive(player, StatusStun, world.combat.pulse) {
		t.Fatal("escaped player was stunned by the heavy strike")
	}
}

func TestTyrantVulnerabilityIncreasesDamageForItsShortWindow(t *testing.T) {
	world := NewWorld()
	boss := world.combat.targets[riftTyrantTargetID]
	boss.VulnerableUntilPulse = 15

	if got := targetVulnerabilityMultiplier(boss, 15); got != 1.25 {
		t.Fatalf("vulnerability multiplier = %.2f, want 1.25", got)
	}
	if got := targetVulnerabilityMultiplier(boss, 16); got != 1 {
		t.Fatalf("expired vulnerability multiplier = %.2f, want 1", got)
	}
	world.combat.pulse = 16
	world.expireTargetVulnerabilities()
	if boss.VulnerableUntilPulse != 0 {
		t.Fatalf("expired vulnerability pulse = %d, want 0", boss.VulnerableUntilPulse)
	}
}

func TestTyrantHeavyStrikeDamagesPlayersRemainingInArea(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.5))
	boss := world.combat.targets[riftTyrantTargetID]
	mustAddPlayer(t, world, "player-one", boss.Position, 4)
	player := world.players["player-one"]
	player.maximumHealth = 200
	player.health = 200
	boss.TargetID = player.id
	boss.attackSequence = 3
	boss.NextAttackPulse = 10
	world.combat.pulse = 10
	world.resolveOneDummyAttack(riftTyrantTargetID, boss)

	world.combat.pulse = boss.TelegraphImpactPulse
	world.resolveOneDummyAttack(riftTyrantTargetID, boss)
	if got, want := player.health, 156; got != want {
		t.Fatalf("health after heavy strike = %d, want %d", got, want)
	}
	if !statusActive(player, StatusStun, world.combat.pulse) {
		t.Fatal("player hit by heavy strike was not stunned")
	}
	if got, want := player.activeStatuses[StatusStun].ExpiresAtPulse, world.combat.pulse+3; got != want {
		t.Fatalf("stun expires at pulse %d, want %d", got, want)
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
