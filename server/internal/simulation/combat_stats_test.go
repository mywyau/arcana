package simulation

import (
	"math"
	"testing"
)

func TestPlayerCombatLevelUsesStrongestOffenceAndGeneralDefence(t *testing.T) {
	player := &playerState{skills: newDefaultSkills()}
	player.skills[SkillAttack] = SkillExperienceForLevel(20)
	player.skills[SkillStrength] = SkillExperienceForLevel(10)
	player.skills[SkillRange] = SkillExperienceForLevel(30)
	player.skills[SkillDefence] = SkillExperienceForLevel(10)
	player.skills[SkillAgility] = SkillExperienceForLevel(20)

	if got, want := playerCombatLevel(player), 23; got != want {
		t.Fatalf("playerCombatLevel() = %d, want %d", got, want)
	}
}

func TestCombatMatchupUsesAuthoritativeAccuracyDamageAndRanges(t *testing.T) {
	player := &playerState{
		id:             "player-one",
		equippedWeapon: WeaponSword,
		combatStyle:    CombatStyleStab,
		equippedArmour: make(map[ArmourSlot]string),
		skills:         newDefaultSkills(),
	}
	target := &TargetSnapshot{
		ID: dummyTargetID, Accuracy: dummyAccuracyScore, Evasion: dummyEvasionScore,
		DodgeRate: dummyDodgeRate,
	}
	weapon, _ := equippedWeaponDefinitionForStyle(player)
	wantMinimum, wantMaximum := damageRange(weaponPoweredDamage(player, weapon), weapon.DamageVariance)
	wantTargetDamage := reduceIncomingDamage(player, dummyAttackDamage, DamageTypeCrush, "")

	got := combatMatchup(player, target)
	if got.TargetID != dummyTargetID || got.PlayerMinimumDamage != wantMinimum ||
		got.PlayerMaximumDamage != wantMaximum || got.PlayerAttackRange != weapon.Range {
		t.Fatalf("player matchup = %+v", got)
	}
	if math.Abs(got.PlayerHitChance-targetHitChance(player, target, weapon)) > 0.000001 {
		t.Fatalf("PlayerHitChance = %f", got.PlayerHitChance)
	}
	if got.TargetMinimumDamage != wantTargetDamage || got.TargetMaximumDamage != wantTargetDamage ||
		got.TargetAttackRange != dummyAttackRange {
		t.Fatalf("target matchup = %+v", got)
	}
	if math.Abs(got.TargetHitChance-playerHitChance(target.Accuracy, player)) > 0.000001 {
		t.Fatalf("TargetHitChance = %f", got.TargetHitChance)
	}
}

func TestPlayerSnapshotIncludesCurrentCombatMatchup(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 6)
	world.mu.Lock()
	world.players["player-one"].combatTarget = dummyTargetID
	world.mu.Unlock()

	player, exists := world.Player("player-one")
	if !exists {
		t.Fatal("Player() did not return the joined player")
	}
	if player.CombatMatchup.TargetID != dummyTargetID || player.CombatMatchup.PlayerHitChance <= 0 ||
		player.CombatMatchup.PlayerMaximumDamage <= 0 {
		t.Fatalf("Player() combat matchup = %+v", player.CombatMatchup)
	}
}

func TestPlayerCombatMatchupUsesOpponentDefences(t *testing.T) {
	attacker := &playerState{
		id: "attacker", equippedWeapon: WeaponSword, combatStyle: CombatStyleSlash,
		equippedArmour: make(map[ArmourSlot]string), skills: newDefaultSkills(),
	}
	defender := &playerState{
		id: "defender", equippedWeapon: WeaponSword, combatStyle: CombatStyleStab,
		equippedArmour: map[ArmourSlot]string{ArmourSlotBody: ItemTrainingPlate},
		skills:         newDefaultSkills(),
	}
	matchup := combatMatchupBetweenPlayers(attacker, defender)
	if matchup.TargetID != "defender" || matchup.PlayerHitChance <= 0 ||
		matchup.PlayerMaximumDamage <= 0 || matchup.TargetHitChance <= 0 ||
		matchup.TargetMaximumDamage <= 0 {
		t.Fatalf("player matchup = %+v", matchup)
	}
	weapon, _ := equippedWeaponDefinitionForStyle(attacker)
	_, unmitigatedMaximum := damageRange(weaponPoweredDamage(attacker, weapon), weapon.DamageVariance)
	if matchup.PlayerMaximumDamage >= unmitigatedMaximum {
		t.Fatalf("armour did not mitigate matchup damage: got %d, unmitigated %d", matchup.PlayerMaximumDamage, unmitigatedMaximum)
	}
}

func TestDerivedCombatStatsUseAuthoritativeCombatRules(t *testing.T) {
	player := &playerState{
		equippedWeapon: WeaponSword,
		combatStyle:    CombatStyleStab,
		equippedArmour: map[ArmourSlot]string{
			ArmourSlotHead: ItemTrainingHelm,
			ArmourSlotBody: ItemTrainingPlate,
		},
		skills: newDefaultSkills(),
	}

	stats := derivedCombatStats(player)
	if stats.AccuracySkill != SkillAttack || stats.AccuracyLevel != 1 ||
		stats.WeaponAccuracyBonus != 10 || stats.StyleAccuracyBonus != 2 || stats.TotalAccuracy != 13 {
		t.Fatalf("accuracy stats = %+v", stats)
	}
	if stats.DamageSkill != SkillStrength || stats.DamageLevel != 1 || stats.StrengthBonus != 21 ||
		stats.MinimumDamage != 17 || stats.MaximumDamage != 25 {
		t.Fatalf("damage stats = %+v", stats)
	}
	if stats.DefenceLevel != 1 {
		t.Fatalf("DefenceLevel = %d, want 1", stats.DefenceLevel)
	}
	wantDefence := (DefenceBonuses{Stab: 10, Slash: 13, Crush: 9, Ranged: 7, Magic: -2})
	if stats.DefenceBonuses != wantDefence {
		t.Fatalf("DefenceBonuses = %+v, want %+v", stats.DefenceBonuses, wantDefence)
	}
}

func TestDerivedCombatStatsFollowRangedSkillAndStyle(t *testing.T) {
	player := &playerState{
		equippedWeapon: WeaponBow,
		combatStyle:    CombatStyleRapid,
		equippedArmour: make(map[ArmourSlot]string),
		skills:         newDefaultSkills(),
	}
	player.skills[SkillRange] = SkillExperienceForLevel(21)

	stats := derivedCombatStats(player)
	if stats.AccuracySkill != SkillRange || stats.DamageSkill != SkillRange || stats.AccuracyLevel != 21 || stats.DamageLevel != 21 {
		t.Fatalf("ranged skill stats = %+v", stats)
	}
	if stats.TotalAccuracy != 30 || stats.SkillDamageBonus != 10 || stats.MinimumDamage != 8 || stats.MaximumDamage != 20 {
		t.Fatalf("ranged derived stats = %+v", stats)
	}
}

func TestDerivedCombatStatsDescribeSelectedSpell(t *testing.T) {
	player := &playerState{
		equippedWeapon: WeaponSword,
		combatStyle:    CombatStyleStab,
		selectedSpell:  SpellWindBolt,
		equippedArmour: make(map[ArmourSlot]string),
		skills:         newDefaultSkills(),
	}
	stats := derivedCombatStats(player)
	if stats.Spell != SpellWindBolt || stats.AccuracySkill != SkillMagic || stats.DamageSkill != SkillMagic ||
		stats.SpellAccuracyBonus != 10 || stats.TotalAccuracy != 11 || stats.SpellPower != 4 ||
		stats.StrengthBonus != 0 || stats.MinimumDamage != 3 || stats.MaximumDamage != 5 {
		t.Fatalf("Wind Bolt combat stats = %+v", stats)
	}
}

func TestAmuletsApplyOnlyTheirMatchingOffensiveBonuses(t *testing.T) {
	tests := []struct {
		name         string
		amulet       string
		weapon       WeaponID
		wantAccuracy int
		wantStrength int
	}{
		{name: "melee", amulet: ItemAmuletFerocity, weapon: WeaponSword, wantAccuracy: 5, wantStrength: 5},
		{name: "ranged", amulet: ItemAmuletHawk, weapon: WeaponBow, wantAccuracy: 5, wantStrength: 5},
		{name: "magic", amulet: ItemAmuletEmbers, weapon: WeaponStaff, wantAccuracy: 5, wantStrength: 5},
		{name: "hybrid", amulet: ItemAmuletFury, weapon: WeaponSword, wantAccuracy: 2, wantStrength: 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			player := &playerState{
				equippedWeapon: test.weapon,
				combatStyle:    defaultCombatStyle(test.weapon),
				equippedArmour: map[ArmourSlot]string{ArmourSlotNeck: test.amulet},
				skills:         newDefaultSkills(),
			}
			weapon, _ := weaponDefinition(test.weapon)
			stats := derivedCombatStats(player)
			if stats.EquipmentAccuracyBonus != test.wantAccuracy || stats.EquipmentStrengthBonus != test.wantStrength {
				t.Fatalf("amulet combat stats = %+v", stats)
			}
			baseStrength := weapon.MeleeStrength + weapon.RangedStrength + weapon.MagicStrength
			if got, want := weaponPoweredDamage(player, weapon), float64(baseStrength+test.wantStrength); got != want {
				t.Fatalf("powered damage = %.1f, want %.1f", got, want)
			}
		})
	}
}

func TestMagicAmuletBoostsStafflessSpellcasting(t *testing.T) {
	player := &playerState{
		equippedWeapon: WeaponSword,
		selectedSpell:  SpellWindBolt,
		equippedArmour: map[ArmourSlot]string{ArmourSlotNeck: ItemAmuletEmbers},
		skills:         newDefaultSkills(),
	}
	spell, _ := spellDefinition(SpellWindBolt)
	if got, want := spellAccuracyScore(player, spell), 16; got != want {
		t.Fatalf("spell accuracy = %d, want %d", got, want)
	}
	if got, want := spellPoweredDamage(player, spell), float64(spell.BasePower+5); got != want {
		t.Fatalf("spell damage = %.1f, want %.1f", got, want)
	}
}

func TestElementalResistanceStacksWithProtectionAndMatchesGear(t *testing.T) {
	w := NewWorld()
	mustAddEmptyPlayer(t, w, "defender", Position{})
	player := w.players["defender"]
	if got := derivedCombatStats(player).ElementalResistances; got != (ElementalResistances{}) {
		t.Fatalf("base resistances = %+v", got)
	}
	player.ringRack.items = []InventoryItem{{ItemID: ItemRubyRing}, {ItemID: ItemRubyRing}}
	player.activePrayers = map[PrayerID]bool{PrayerProtectMagic: true}
	if got := derivedCombatStats(player).ElementalResistances; got != (ElementalResistances{Fire: 20}) {
		t.Fatalf("equipped resistances = %+v", got)
	}
	for _, tc := range []struct {
		damageType DamageType
		element    MagicElement
		want       int
	}{
		{DamageTypeMagic, MagicElementFire, 56},
		{DamageTypeMagic, MagicElementWater, 70},
		{DamageTypeMagic, "", 70},
		{DamageTypeSlash, MagicElementFire, 100},
	} {
		if got := reduceIncomingDamage(player, 100, tc.damageType, tc.element); got != tc.want {
			t.Errorf("%s/%s damage = %d, want %d", tc.damageType, tc.element, got, tc.want)
		}
	}
	player.activePrayers = nil
	imp := TargetSnapshot{ID: cinderImpTargetID}
	if damage := combatMatchup(player, &imp).TargetMaximumDamage; damage != 6 {
		t.Errorf("cinder imp hit preview = %d, want 6 with two rings", damage)
	}
}

func TestPvPFireBoltResistanceMatchesDamagePreview(t *testing.T) {
	w := NewWorldWithRollSource(&sequenceRollSource{values: []float64{0.25, 0.999}})
	mustAddEmptyPlayer(t, w, "attacker", Position{Z: -100})
	mustAddEmptyPlayer(t, w, "defender", Position{Z: -100, X: 1})
	attacker, defender := w.players["attacker"], w.players["defender"]
	attacker.selectedSpell = SpellFireBolt
	attacker.skills[SkillMagic] = SkillExperienceForLevel(25)
	spell, _ := spellDefinition(SpellFireBolt)
	for _, cost := range spell.Costs {
		attacker.inventory.add(cost.ItemID, cost.Quantity)
	}
	_, _, baseMax, _ := playerAttackMatchup(attacker, defender)
	defender.ringRack.items = []InventoryItem{{ItemID: ItemRubyRing}}
	_, _, maxDamage, _ := playerAttackMatchup(attacker, defender)
	if maxDamage >= baseMax {
		t.Fatalf("maximum fire hit %d with ring, %d without", maxDamage, baseMax)
	}
	before := defender.health
	w.resolvePlayerAutoAttack(attacker, defender)
	if got := before - defender.health; got != maxDamage {
		t.Fatalf("maximum fire hit dealt %d, preview %d (%s)", got, maxDamage, attacker.lastActionResult)
	}
}

func TestMeleeSwordLeadsSustainedDamageWhileInReach(t *testing.T) {
	// Compare neutral targets, equal offensive levels, and continuous attacks.
	// Attack range, pursuit, and enemy-specific weaknesses matter in encounters.
	for _, tier := range []struct {
		level            int
		sword, bow, wand WeaponID
		arrow            string
		spell            SpellID
	}{
		{1, WeaponBronzeSword, WeaponBow, WeaponWand, ItemBronzeArrow, SpellWindBolt},
		{5, WeaponSword, WeaponOakShortbow, WeaponWand, ItemIronArrow, SpellHolyBolt},
		{10, WeaponSteelSword, WeaponWillowShortbow, WeaponMoonspunWand, ItemSteelArrow, SpellHolyBolt},
		{15, WeaponMithrilSword, WeaponMapleShortbow, WeaponMoonspunWand, ItemMithrilArrow, SpellHolyBolt},
		{20, WeaponMithrilSword, WeaponYewShortbow, WeaponMoonspunWand, ItemAdamantiteArrow, SpellHolyBolt},
		{25, WeaponRuneSword, WeaponYewShortbow, WeaponMoonspunWand, ItemAdamantiteArrow, SpellHolyBolt},
	} {
		player := &playerState{skills: newDefaultSkills(), ammunitionPouch: newAmmunitionPouch()}
		for _, skill := range []SkillID{SkillAttack, SkillStrength, SkillRange, SkillMagic} {
			player.skills[skill] = SkillExperienceForLevel(tier.level)
		}
		player.ammunitionPouch.add(tier.arrow, 1)
		dps := func(weapon WeaponID, style CombatStyleID, spell SpellID) float64 {
			player.equippedWeapon, player.combatStyle, player.selectedSpell = weapon, style, spell
			matchup := combatMatchup(player, &TargetSnapshot{})
			average := float64(matchup.PlayerMinimumDamage+matchup.PlayerMaximumDamage) / 2
			interval := float64(derivedCombatStats(player).AttackInterval) * actionPulseInterval.Seconds()
			return matchup.PlayerHitChance * average / interval
		}
		melee := dps(tier.sword, CombatStyleSlash, "")
		bow := dps(tier.bow, CombatStyleRapid, "")
		magic := dps(tier.wand, CombatStyleRapid, tier.spell)
		t.Logf("level %d: melee %.2f, bow %.2f, wand %.2f DPS", tier.level, melee, bow, magic)
		if melee <= max(bow, magic) {
			t.Errorf("level %d sword should lead while continuously in reach", tier.level)
		}
	}
}
