package simulation

const (
	minimumAccuracyChance = 0.20
	maximumAccuracyChance = 0.95
	baseAccuracyChance    = 0.50
	accuracyCurveStrength = 0.45
	accuracyCurveScale    = 35.0
	minimumDodgeChance    = 0.0
	maximumDodgeChance    = 0.60
	evasionCurveScale     = 50.0
	evasionCurveStrength  = 0.35
	dummyAccuracyScore    = 11
	dummyEvasionScore     = 5
	defaultBaseDodgeBonus = 0.0
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
	return max(0, SkillLevelForExperience(player.skills[SkillAgility])+equippedEvasionBonusForPlayer(player))
}

func combatStyleAccuracyBonus(style CombatStyleID) int {
	switch style {
	case CombatStyleAccurate:
		return 4
	case CombatStyleStab:
		return 2
	case CombatStyleCrush:
		return 1
	case CombatStyleRapid, CombatStyleHeavyCrush, CombatStyleHeavyStab:
		return -3
	case CombatStyleDefensive:
		return -2
	default:
		return 0
	}
}

func calculatedAccuracyChance(accuracyScore int) float64 {
	accuracy := float64(max(0, accuracyScore))
	chance := baseAccuracyChance + accuracyCurveStrength*accuracy/(accuracy+accuracyCurveScale)
	return max(minimumAccuracyChance, min(chance, maximumAccuracyChance))
}

func calculatedDodgeChance(accuracyScore, evasionScore int, dodgeBonus float64) float64 {
	accuracy := float64(max(0, accuracyScore))
	evasion := float64(max(0, evasionScore))
	// Dodge is deliberately separate from the attacker's chance to miss. This
	// keeps early accuracy and Agility/equipment evasion independently tunable.
	chance := dodgeBonus
	if evasion > 0 {
		chance += evasionCurveStrength * evasion / (accuracy + evasion + evasionCurveScale)
	}
	return max(minimumDodgeChance, min(chance, maximumDodgeChance))
}

func targetAccuracyChance(player *playerState, weapon WeaponDefinition) float64 {
	chance := calculatedAccuracyChance(playerAccuracyScore(player, weapon)) + prayerHitChanceBonus(player, weapon)
	return min(maximumAccuracyChance, chance)
}

func targetDodgeChance(player *playerState, target *TargetSnapshot, weapon WeaponDefinition) float64 {
	return calculatedDodgeChance(
		playerAccuracyScore(player, weapon),
		targetEvasion(target),
		target.DodgeRate,
	)
}

func targetHitChance(player *playerState, target *TargetSnapshot, weapon WeaponDefinition) float64 {
	return targetAccuracyChance(player, weapon) * (1 - targetDodgeChance(player, target, weapon))
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

func playerHitChance(attackerAccuracy int, player *playerState) float64 {
	return calculatedAccuracyChance(attackerAccuracy) * (1 - playerDodgeChance(attackerAccuracy, player))
}
