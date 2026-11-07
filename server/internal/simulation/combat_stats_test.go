package simulation

import "testing"

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
	if stats.DamageSkill != SkillStrength || stats.DamageLevel != 1 || stats.StrengthBonus != 10 ||
		stats.MinimumDamage != 8 || stats.MaximumDamage != 12 {
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
	if stats.TotalAccuracy != 30 || stats.SkillDamageBonus != 10 || stats.MinimumDamage != 11 || stats.MaximumDamage != 18 {
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
