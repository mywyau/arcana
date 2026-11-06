package simulation

// CombatStatsSnapshot contains the derived values used by combat. Sending the
// derived values keeps the client display aligned with the authoritative rules.
type CombatStatsSnapshot struct {
	Weapon              WeaponID
	Style               CombatStyleID
	AccuracySkill       SkillID
	AccuracyLevel       int
	WeaponAccuracyBonus int
	StyleAccuracyBonus  int
	TotalAccuracy       int
	DamageSkill         SkillID
	DamageLevel         int
	StrengthBonus       int
	SkillDamageBonus    float64
	MinimumDamage       int
	MaximumDamage       int
	DefenceLevel        int
	DefenceBonuses      DefenceBonuses
}

func derivedCombatStats(player *playerState) CombatStatsSnapshot {
	if player == nil {
		return CombatStatsSnapshot{}
	}
	weapon, exists := weaponDefinitionForStyle(player.equippedWeapon, player.combatStyle)
	if !exists {
		return CombatStatsSnapshot{}
	}

	accuracySkill := weaponAccuracySkill(weapon)
	damageSkill := SkillStrength
	strengthBonus := weapon.MeleeStrength
	if weapon.RangedStrength > 0 {
		damageSkill = SkillRange
		strengthBonus = weapon.RangedStrength
	} else if weapon.MagicStrength > 0 {
		damageSkill = SkillMagic
		strengthBonus = weapon.MagicStrength
	}
	minimumDamage, maximumDamage := damageRange(
		weaponPoweredDamage(player, weapon),
		weapon.DamageVariance,
	)

	return CombatStatsSnapshot{
		Weapon:              weapon.ID,
		Style:               player.combatStyle,
		AccuracySkill:       accuracySkill,
		AccuracyLevel:       SkillLevelForExperience(player.skills[accuracySkill]),
		WeaponAccuracyBonus: weapon.AccuracyBonus,
		StyleAccuracyBonus:  combatStyleAccuracyBonus(player.combatStyle),
		TotalAccuracy:       playerAccuracyScore(player, weapon),
		DamageSkill:         damageSkill,
		DamageLevel:         SkillLevelForExperience(player.skills[damageSkill]),
		StrengthBonus:       strengthBonus,
		SkillDamageBonus:    float64(SkillLevelForExperience(player.skills[damageSkill])-1) * skillDamagePointsPerLevel,
		MinimumDamage:       minimumDamage,
		MaximumDamage:       maximumDamage,
		DefenceLevel:        SkillLevelForExperience(player.skills[SkillDefence]),
		DefenceBonuses:      equippedDefenceBonuses(player.equippedArmour),
	}
}
