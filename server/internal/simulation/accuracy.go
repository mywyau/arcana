package simulation

const (
	minimumDodgeChance    = 0.05
	maximumDodgeChance    = 0.90
	accuracyCurveScale    = 50.0
	accuracyCurveStrength = 0.30
	dummyAccuracyScore    = 11
	dummyEvasionScore     = 5
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
		level = effectiveCombatSkillLevel(player, weaponAccuracySkill(weapon))
		style = player.combatStyle
	}
	equipmentAccuracy, _ := equipmentOffenceForWeapon(player, weapon)
	return level + weapon.AccuracyBonus + equipmentAccuracy + combatStyleAccuracyBonus(style)
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
	accuracy := float64(max(1, accuracyScore))
	evasion := float64(max(1, evasionScore))
	relativeAdvantage := (accuracy - evasion) / (accuracy + evasion + accuracyCurveScale)
	hitChance := 1 - dodgeBonus + relativeAdvantage*accuracyCurveStrength
	hitChance = max(1-maximumDodgeChance, min(hitChance, 1-minimumDodgeChance))
	return 1 - hitChance
}

func targetDodgeChance(player *playerState, target *TargetSnapshot, weapon WeaponDefinition) float64 {
	dodgeChance := calculatedDodgeChance(
		playerAccuracyScore(player, weapon),
		targetEvasion(target),
		target.DodgeRate,
	)
	return max(minimumDodgeChance, dodgeChance-prayerHitChanceBonus(player, weapon))
}

func targetEvasion(target *TargetSnapshot) int {
	if target == nil {
		return 1
	}
	return target.Evasion
}

func playerDodgeChance(attackerAccuracy int, player *playerState) float64 {
	return calculatedDodgeChance(attackerAccuracy, playerEvasionScore(player), effectivePlayerDodgeRate(player))
}
