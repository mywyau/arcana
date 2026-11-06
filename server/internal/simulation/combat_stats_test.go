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
