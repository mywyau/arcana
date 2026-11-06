package simulation

const (
	minimumDodgeChance    = 0.05
	maximumDodgeChance    = 0.90
	dodgeChancePerScore   = 0.002
	dummyAccuracyScore    = 11
	dummyEvasionScore     = 1
	defaultBaseDodgeBonus = 0.12
)

func weaponAccuracySkill(weapon WeaponDefinition) SkillID {
	if weapon.RangedStrength > 0 {
		return SkillRange
	}
	if weapon.MagicStrength > 0 {
		return SkillMagic
	}
	return SkillAttack
}

func playerAccuracyScore(player *playerState, weapon WeaponDefinition) int {
	level := 1
	style := CombatStyleID("")
	if player != nil {
		level = SkillLevelForExperience(player.skills[weaponAccuracySkill(weapon)])
		style = player.combatStyle
	}
	return level + weapon.AccuracyBonus + combatStyleAccuracyBonus(style) + prayerAccuracyBonus(player, weapon)
}

func playerEvasionScore(player *playerState) int {
	if player == nil {
		return 1
	}
	return SkillLevelForExperience(player.skills[SkillAgility])
}

func combatStyleAccuracyBonus(style CombatStyleID) int {
	switch style {
	case CombatStyleAccurate:
		return 4
	case CombatStyleStab:
		return 2
	case CombatStyleCrush:
		return 1
	case CombatStyleRapid, CombatStyleHeavyCrush:
		return -3
	case CombatStyleDefensive:
		return -2
	default:
		return 0
	}
}

func calculatedDodgeChance(accuracyScore, evasionScore int, dodgeBonus float64) float64 {
	chance := dodgeBonus + float64(evasionScore-accuracyScore)*dodgeChancePerScore
	return max(minimumDodgeChance, min(chance, maximumDodgeChance))
}

func targetDodgeChance(player *playerState, target *TargetSnapshot, weapon WeaponDefinition) float64 {
	return calculatedDodgeChance(playerAccuracyScore(player, weapon), target.Evasion, target.DodgeRate)
}

func playerDodgeChance(attackerAccuracy int, player *playerState) float64 {
	return calculatedDodgeChance(attackerAccuracy, playerEvasionScore(player), effectivePlayerDodgeRate(player))
}
