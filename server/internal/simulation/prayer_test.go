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

func TestRegenerationPrayerHealsFivePercentOnDrainCadence(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.5))
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	world.players["player-one"].health = 80
	if err := world.QueuePrayer(PrayerCommand{
		PlayerID: "player-one", Sequence: 1, PrayerID: PrayerRegeneration, Active: true,
	}); err != nil {
		t.Fatalf("QueuePrayer() returned an unexpected error: %v", err)
	}

	for range regenerationDrainIntervalPulses {
		world.Step(actionPulseInterval)
	}
	player := mustPlayer(t, world, "player-one")
	if player.Health != 86 {
		t.Fatalf("health after prayer and passive regeneration = %d, want 86", player.Health)
	}
	if player.Prayer != playerMaximumPrayer-1 {
		t.Fatalf("prayer after regeneration = %d, want %d", player.Prayer, playerMaximumPrayer-1)
	}
	world.players["player-one"].health = 98
	for range regenerationDrainIntervalPulses {
		world.Step(actionPulseInterval)
	}
	if player := mustPlayer(t, world, "player-one"); player.Health != playerMaximumHealth {
		t.Fatalf("capped regeneration health = %d, want %d", player.Health, playerMaximumHealth)
	}
}

func TestEvasionPrayerAppliesBeforeIncomingAttack(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.02))
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

func TestStrengthPrayerBoostsMeleeDamageBeforeVariance(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.5))
	dummy := world.combat.targets[dummyTargetID]
	mustAddPlayer(t, world, "player-one", dummy.Position, 4)
	if err := world.QueuePrayer(PrayerCommand{
		PlayerID: "player-one", Sequence: 1, PrayerID: PrayerStrength, Active: true,
	}); err != nil {
		t.Fatalf("QueuePrayer() returned an unexpected error: %v", err)
	}
	if err := world.QueueTarget(TargetCommand{
		PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID,
	}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}

	world.Step(actionPulseInterval)
	player := mustPlayer(t, world, "player-one")
	if got, want := mustTarget(t, world.Snapshot(), dummyTargetID).Health, dummyMaximumHealth-11; got != want {
		t.Fatalf("dummy health after Strength hit = %d, want %d", got, want)
	}
	if player.CombatStats.PrayerDamageBonus != strengthPrayerMeleeDamageBonus ||
		player.CombatStats.MinimumDamage != 8 || player.CombatStats.MaximumDamage != 13 {
		t.Fatalf("Strength combat stats = %+v", player.CombatStats)
	}
	bow, _ := weaponDefinition(WeaponBow)
	if got := weaponPoweredDamage(world.players["player-one"], bow); got != bowRangedStrength {
		t.Fatalf("ranged damage with Strength prayer = %.1f, want %d", got, bowRangedStrength)
	}
}

func TestRangedAndMagicStrengthPrayersBoostOnlyMatchingDamage(t *testing.T) {
	player := &playerState{
		activePrayers: map[PrayerID]bool{PrayerRangedStrength: true},
		skills:        newDefaultSkills(), equippedWeapon: WeaponBow, combatStyle: CombatStyleAccurate,
		equippedArmour: make(map[ArmourSlot]string), equippedArmourInstances: make(map[ArmourSlot]string),
	}
	bow, _ := weaponDefinition(WeaponBow)
	sword, _ := weaponDefinition(WeaponSword)
	if got, want := weaponPoweredDamage(player, bow), float64(bow.RangedStrength)*(1+rangedPrayerDamageBonus); math.Abs(got-want) > 0.000_001 {
		t.Fatalf("ranged prayer damage = %.2f, want %.2f", got, want)
	}
	if got := weaponPoweredDamage(player, sword); got != float64(sword.MeleeStrength) {
		t.Fatalf("ranged prayer changed melee damage to %.2f", got)
	}
	if got := derivedCombatStats(player).PrayerDamageBonus; got != rangedPrayerDamageBonus {
		t.Fatalf("ranged prayer stats bonus = %.2f", got)
	}

	player.activePrayers = map[PrayerID]bool{PrayerMagicStrength: true}
	player.equippedWeapon = WeaponStaff
	player.combatStyle = CombatStyleAccurate
	spell, _ := spellDefinition(SpellWindBolt)
	basePlayer := *player
	basePlayer.activePrayers = nil
	if got, want := spellPoweredDamage(player, spell), spellPoweredDamage(&basePlayer, spell)*(1+magicPrayerDamageBonus); math.Abs(got-want) > 0.000_001 {
		t.Fatalf("magic prayer damage = %.2f, want %.2f", got, want)
	}
	player.selectedSpell = SpellWindBolt
	if got := derivedCombatStats(player).PrayerDamageBonus; got != magicPrayerDamageBonus {
		t.Fatalf("magic prayer stats bonus = %.2f", got)
	}
}

func TestAccuracyPrayerBoostsOnlyMeleeAccuracy(t *testing.T) {
	player := &playerState{
		equippedWeapon: WeaponSword,
		combatStyle:    CombatStyleStab,
		activePrayers:  map[PrayerID]bool{PrayerAccuracy: true},
		skills:         newDefaultSkills(),
	}
	sword, _ := weaponDefinition(WeaponSword)
	if got, want := playerAccuracyScore(player, sword), 13; got != want {
		t.Fatalf("melee accuracy with Accuracy prayer = %d, want %d", got, want)
	}
	stats := derivedCombatStats(player)
	if stats.PrayerHitChanceBonus != accuracyPrayerMeleeHitChanceBonus || stats.TotalAccuracy != 13 {
		t.Fatalf("Accuracy combat stats = %+v", stats)
	}
	target := &TargetSnapshot{Evasion: hillWardenEvasionScore, DodgeRate: defaultBaseDodgeBonus}
	withPrayer := targetHitChance(player, target, sword)
	player.activePrayers = nil
	withoutPrayer := targetHitChance(player, target, sword)
	wantIncrease := accuracyPrayerMeleeHitChanceBonus * (1 - targetDodgeChance(player, target, sword))
	if math.Abs(withPrayer-withoutPrayer-wantIncrease) > 0.000_001 {
		t.Fatalf("Accuracy prayer hit chance change = %.3f, want %.3f", withPrayer-withoutPrayer, wantIncrease)
	}
	player.activePrayers = map[PrayerID]bool{PrayerAccuracy: true}
	player.equippedWeapon = WeaponBow
	player.combatStyle = CombatStyleAccurate
	bow, _ := weaponDefinitionForStyle(WeaponBow, CombatStyleAccurate)
	if got, want := playerAccuracyScore(player, bow), 19; got != want {
		t.Fatalf("ranged accuracy with melee Accuracy prayer = %d, want %d", got, want)
	}
}

func TestDefencePrayerReducesIncomingDamageWithoutChangingDodge(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.5))
	dummy := world.combat.targets[dummyTargetID]
	mustAddPlayer(t, world, "player-one", dummy.Position, 4)
	dummy.TargetID = "player-one"
	if err := world.QueuePrayer(PrayerCommand{
		PlayerID: "player-one", Sequence: 1, PrayerID: PrayerDefence, Active: true,
	}); err != nil {
		t.Fatalf("QueuePrayer() returned an unexpected error: %v", err)
	}

	world.Step(actionPulseInterval)
	player := mustPlayer(t, world, "player-one")
	reducedDamage := reduceIncomingDamage(world.players["player-one"], dummyAttackDamage, DamageTypeCrush, "")
	if got, want := player.Health, playerMaximumHealth-reducedDamage; got != want {
		t.Fatalf("health after Defence prayer hit = %d, want %d", got, want)
	}
	if player.CombatStats.PrayerDamageReduction != defencePrayerDamageReduction {
		t.Fatalf("Defence damage reduction = %.2f, want %.2f", player.CombatStats.PrayerDamageReduction, defencePrayerDamageReduction)
	}
	wantDodge := calculatedDodgeChance(dummyAccuracyScore, 1, playerDodgeRate)
	if player.DodgeRate != wantDodge {
		t.Fatalf("dodge with Defence prayer = %.3f, want unchanged %.3f", player.DodgeRate, wantDodge)
	}
	events := world.Snapshot().CombatEvents
	if len(events) != 1 || events[0].Type != CombatEventDamage || events[0].Amount != reducedDamage {
		t.Fatalf("combat events after Defence prayer hit = %+v", events)
	}
}

func TestProtectionPrayersOnlyReduceTheirMatchingDamageStyle(t *testing.T) {
	tests := []struct {
		name        string
		prayer      PrayerID
		matching    []DamageType
		nonMatching []DamageType
	}{
		{
			name: "melee", prayer: PrayerProtectMelee,
			matching:    []DamageType{DamageTypeStab, DamageTypeSlash, DamageTypeCrush},
			nonMatching: []DamageType{DamageTypeRanged, DamageTypeMagic},
		},
		{
			name: "ranged", prayer: PrayerProtectRanged,
			matching: []DamageType{DamageTypeRanged}, nonMatching: []DamageType{DamageTypeSlash, DamageTypeMagic},
		},
		{
			name: "magic", prayer: PrayerProtectMagic,
			matching: []DamageType{DamageTypeMagic}, nonMatching: []DamageType{DamageTypeSlash, DamageTypeRanged},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			player := &playerState{activePrayers: map[PrayerID]bool{test.prayer: true}}
			for _, damageType := range test.matching {
				if got := reduceIncomingDamage(player, 100, damageType, ""); got != 70 {
					t.Errorf("matching %s damage = %d, want 70", damageType, got)
				}
			}
			for _, damageType := range test.nonMatching {
				if got := reduceIncomingDamage(player, 100, damageType, ""); got != 100 {
					t.Errorf("non-matching %s damage = %d, want 100", damageType, got)
				}
			}
		})
	}
}

func TestProtectionAndDefencePrayersCombineMultiplicatively(t *testing.T) {
	player := &playerState{activePrayers: map[PrayerID]bool{
		PrayerProtectMelee: true,
		PrayerDefence:      true,
	}}
	if got := reduceIncomingDamage(player, 100, DamageTypeSlash, ""); got != 67 {
		t.Fatalf("combined protection damage = %d, want 67", got)
	}
}

func TestPrayerExclusiveGroupsAllowOneBoostAndOneProtectionWithUtilities(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.5))
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	sequence := uint64(1)
	activate := func(prayerID PrayerID) {
		t.Helper()
		if err := world.QueuePrayer(PrayerCommand{
			PlayerID: "player-one", Sequence: sequence, PrayerID: prayerID, Active: true,
		}); err != nil {
			t.Fatalf("QueuePrayer(%s) error = %v", prayerID, err)
		}
		sequence++
		world.Step(actionPulseInterval)
	}

	activate(PrayerAccuracy)
	activate(PrayerEvasion)
	activate(PrayerStrength)
	activate(PrayerProtectMelee)
	activate(PrayerProtectMagic)

	active := activePrayerIDs(world.players["player-one"])
	want := []PrayerID{PrayerEvasion, PrayerProtectMagic, PrayerStrength}
	if len(active) != len(want) {
		t.Fatalf("active prayers = %v, want %v", active, want)
	}
	for index := range want {
		if active[index] != want[index] {
			t.Fatalf("active prayers = %v, want %v", active, want)
		}
	}
	if _, active := world.players["player-one"].prayerDrainProgress[PrayerAccuracy]; active {
		t.Fatal("replaced offensive prayer retained drain progress")
	}
	if _, active := world.players["player-one"].prayerDrainProgress[PrayerProtectMelee]; active {
		t.Fatal("replaced protection prayer retained drain progress")
	}
}

func TestProtectionPrayerUsesFasterDrainCadence(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.5))
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	if err := world.QueuePrayer(PrayerCommand{
		PlayerID: "player-one", Sequence: 1, PrayerID: PrayerProtectMelee, Active: true,
	}); err != nil {
		t.Fatalf("QueuePrayer() error = %v", err)
	}
	for range protectionDrainIntervalPulses {
		world.Step(actionPulseInterval)
	}
	if got := mustPlayer(t, world, "player-one").Prayer; got != playerMaximumPrayer-1 {
		t.Fatalf("prayer points after protection drain = %d, want %d", got, playerMaximumPrayer-1)
	}
}

func TestMasteryPrayersProvideModestCombinedStyleBonuses(t *testing.T) {
	tests := []struct {
		prayer PrayerID
		weapon WeaponID
		want   func(prayerBonuses) bool
	}{
		{
			prayer: PrayerMeleeMastery, weapon: WeaponSword,
			want: func(b prayerBonuses) bool {
				return b.MeleeHitChance == masteryBonus && b.MeleeDamage == masteryBonus &&
					b.DamageReduction == masteryBonus && b.RangedDamage == 0 && b.MagicDamage == 0
			},
		},
		{
			prayer: PrayerRangedMastery, weapon: WeaponBow,
			want: func(b prayerBonuses) bool {
				return b.RangedHitChance == masteryBonus && b.RangedDamage == masteryBonus &&
					b.DamageReduction == masteryBonus && b.MeleeDamage == 0 && b.MagicDamage == 0
			},
		},
		{
			prayer: PrayerMagicMastery, weapon: WeaponStaff,
			want: func(b prayerBonuses) bool {
				return b.MagicHitChance == masteryBonus && b.MagicDamage == masteryBonus &&
					b.DamageReduction == masteryBonus && b.MeleeDamage == 0 && b.RangedDamage == 0
			},
		},
	}
	for _, test := range tests {
		t.Run(string(test.prayer), func(t *testing.T) {
			player := &playerState{activePrayers: map[PrayerID]bool{test.prayer: true}}
			bonuses := activePrayerBonuses(player)
			if !test.want(bonuses) {
				t.Fatalf("mastery bonuses = %+v", bonuses)
			}
			weapon, exists := weaponDefinition(test.weapon)
			if !exists {
				t.Fatalf("weapon %q not found", test.weapon)
			}
			if got := prayerHitChanceBonus(player, weapon); got != masteryBonus {
				t.Fatalf("mastery hit chance = %.2f, want %.2f", got, masteryBonus)
			}
		})
	}
}

func TestMasteryPrayersRequireLevelFortyPrayer(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.5))
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	command := PrayerCommand{
		PlayerID: "player-one", Sequence: 1, PrayerID: PrayerMeleeMastery, Active: true,
	}
	if err := world.QueuePrayer(command); err == nil {
		t.Fatal("level-one player activated Melee Mastery")
	}
	world.players["player-one"].skills[SkillPrayer] = SkillExperienceForLevel(masteryRequiredLevel)
	if err := world.QueuePrayer(command); err != nil {
		t.Fatalf("level-forty QueuePrayer() error = %v", err)
	}
}

func TestMercyHealsAtLowHealthWithoutDrainingAllPrayer(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.5))
	dummy := world.combat.targets[dummyTargetID]
	mustAddPlayer(t, world, "player-one", dummy.Position, 4)
	player := world.players["player-one"]
	player.skills[SkillPrayer] = SkillExperienceForLevel(mercyRequiredLevel)
	player.health = 25
	if err := world.QueuePrayer(PrayerCommand{
		PlayerID: "player-one", Sequence: 1, PrayerID: PrayerMercy, Active: true,
	}); err != nil {
		t.Fatalf("QueuePrayer() error = %v", err)
	}
	dummy.TargetID = "player-one"

	world.Step(actionPulseInterval)
	got := mustPlayer(t, world, "player-one")
	wantHealth := 25 - reduceIncomingDamage(player, dummyAttackDamage, DamageTypeCrush, "") + 20
	if got.Health != wantHealth {
		t.Fatalf("health after Mercy = %d, want %d", got.Health, wantHealth)
	}
	if got.Prayer != playerMaximumPrayer-mercyPrayerCost {
		t.Fatalf("prayer after Mercy = %d, want %d", got.Prayer, playerMaximumPrayer-mercyPrayerCost)
	}
	for _, prayerID := range got.ActivePrayers {
		if prayerID == PrayerMercy {
			t.Fatalf("Mercy remained active after triggering: %v", got.ActivePrayers)
		}
	}
}

func TestMercyRequiresLevelAndReservedPrayerPoints(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.5))
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	command := PrayerCommand{
		PlayerID: "player-one", Sequence: 1, PrayerID: PrayerMercy, Active: true,
	}
	if err := world.QueuePrayer(command); err == nil {
		t.Fatal("level-one player activated Mercy")
	}
	player := world.players["player-one"]
	player.skills[SkillPrayer] = SkillExperienceForLevel(mercyRequiredLevel)
	player.prayer = mercyPrayerCost - 1
	if err := world.QueuePrayer(command); err == nil {
		t.Fatal("Mercy activated without its reserved prayer cost")
	}
	player.prayer = mercyPrayerCost
	if err := world.QueuePrayer(command); err != nil {
		t.Fatalf("eligible QueuePrayer() error = %v", err)
	}
}

func TestMercyDoesNotReviveLethalDamage(t *testing.T) {
	player := &playerState{
		health: 0, maximumHealth: 100, prayer: 100,
		activePrayers: map[PrayerID]bool{PrayerMercy: true},
	}
	if healing := triggerEmergencyPrayer(player); healing != 0 || player.health != 0 || player.prayer != 100 {
		t.Fatalf("lethal Mercy result = healing %d, player %+v", healing, player)
	}
}

func TestMercyRemainsArmedUntilDamageLeavesHealthBelowThreshold(t *testing.T) {
	player := &playerState{
		health: 20, maximumHealth: 100, prayer: 100,
		activePrayers:       map[PrayerID]bool{PrayerMercy: true},
		prayerDrainProgress: make(map[PrayerID]uint64),
	}
	if healing := triggerEmergencyPrayer(player); healing != 0 || !player.activePrayers[PrayerMercy] {
		t.Fatalf("Mercy triggered at threshold: healing %d, player %+v", healing, player)
	}
	player.health = 19
	if healing := triggerEmergencyPrayer(player); healing != 20 || player.health != 39 || player.activePrayers[PrayerMercy] {
		t.Fatalf("Mercy below threshold: healing %d, player %+v", healing, player)
	}
}

func TestWrathExplosionCanDefeatNearbyPlayersAndMonsters(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.5))
	dummy := world.combat.targets[dummyTargetID]
	dummy.Position = Position{Z: -100}
	mustAddPlayer(t, world, "source", dummy.Position, 4)
	mustAddPlayer(t, world, "nearby", Position{X: dummy.Position.X + 1, Z: dummy.Position.Z}, 4)
	mustAddPlayer(t, world, "far", Position{X: dummy.Position.X + wrathRadius + 1, Z: dummy.Position.Z}, 4)
	source := world.players["source"]
	source.health = 0
	source.activePrayers[PrayerWrath] = true
	world.players["nearby"].health = 15
	dummy.Health = 15

	world.defeatPlayer(source)

	if nearby := mustPlayer(t, world, "nearby"); !nearby.Defeated || nearby.Health != 0 {
		t.Fatalf("nearby player after Wrath = %+v", nearby)
	}
	if far := mustPlayer(t, world, "far"); far.Defeated || far.Health != playerMaximumHealth {
		t.Fatalf("far player after Wrath = %+v", far)
	}
	if !dummy.Defeated || dummy.Health != 0 {
		t.Fatalf("nearby monster after Wrath = %+v", dummy)
	}
	events := world.Snapshot().CombatEvents
	if len(events) < 3 || events[0].Type != CombatEventWrath || events[0].SourceID != "source" {
		t.Fatalf("Wrath combat events = %+v", events)
	}
}

func TestMercyAndWrathAreMutuallyExclusiveEmergencyPrayers(t *testing.T) {
	player := &playerState{
		activePrayers:       map[PrayerID]bool{PrayerMercy: true},
		prayerDrainProgress: map[PrayerID]uint64{PrayerMercy: 1},
	}
	definition, _ := prayerDefinitionFor(PrayerWrath)
	deactivateExclusivePrayers(player, PrayerWrath, definition.ExclusiveGroup)
	player.activePrayers[PrayerWrath] = true
	if player.activePrayers[PrayerMercy] || !player.activePrayers[PrayerWrath] {
		t.Fatalf("emergency prayers after activating Wrath = %v", player.activePrayers)
	}
	if _, exists := player.prayerDrainProgress[PrayerMercy]; exists {
		t.Fatal("replaced Mercy retained drain progress")
	}
}
