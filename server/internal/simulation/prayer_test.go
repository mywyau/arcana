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
	if player.Health != 85 {
		t.Fatalf("health after regeneration = %d, want 85", player.Health)
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
	withPrayer := 1 - targetDodgeChance(player, target, sword)
	player.activePrayers = nil
	withoutPrayer := 1 - targetDodgeChance(player, target, sword)
	if math.Abs(withPrayer-withoutPrayer-accuracyPrayerMeleeHitChanceBonus) > 0.000_001 {
		t.Fatalf("Accuracy prayer hit chance change = %.3f, want %.3f", withPrayer-withoutPrayer, accuracyPrayerMeleeHitChanceBonus)
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
	reducedDamage := reduceIncomingDamage(world.players["player-one"], dummyAttackDamage, DamageTypeCrush)
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
