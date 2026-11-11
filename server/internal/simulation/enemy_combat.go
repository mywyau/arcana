package simulation

const enemyLongRangeAttackRange = bowRange + longRangeStyleRangeBonus + 1.0

type enemyCombatProfile struct {
	CombatLevel    int
	AttackDamage   int
	AttackInterval uint64
	AttackRange    float64
	DetectionRange float64
	LeashRange     float64
	DamageType     DamageType
	AttackTypes    []DamageType
	Defence        DefenceBonuses
	Elements       ElementalResistances
}

var defaultEnemyCombatProfile = enemyCombatProfile{
	CombatLevel:    5,
	AttackDamage:   dummyAttackDamage,
	AttackInterval: dummyAttackIntervalPulses,
	AttackRange:    dummyAttackRange,
	DetectionRange: 0,
	LeashRange:     dummyLeashRange,
	DamageType:     DamageTypeCrush,
}

var enemyCombatProfiles = map[string]enemyCombatProfile{
	hillWardenTargetID: {
		CombatLevel:    18,
		AttackDamage:   18,
		AttackInterval: 5,
		AttackRange:    enemyLongRangeAttackRange,
		DetectionRange: 6,
		LeashRange:     22,
		DamageType:     DamageTypeSlash,
		Defence:        DefenceBonuses{Stab: 10, Slash: 18, Crush: 4, Ranged: 12, Magic: 6},
		Elements:       ElementalResistances{Air: -10, Earth: 10},
	},
	wildFowlTargetID: {
		CombatLevel:    2,
		AttackDamage:   wildFowlAttackDamage,
		AttackInterval: 3,
		AttackRange:    1.2,
		DetectionRange: 3,
		LeashRange:     5,
		DamageType:     DamageTypeStab,
		Defence:        DefenceBonuses{Stab: 1, Slash: 0, Crush: 3, Ranged: 1, Magic: 0},
		Elements:       ElementalResistances{Air: 5, Fire: -15},
	},
	cowTargetID: {
		CombatLevel:  4,
		AttackDamage: cowAttackDamage, AttackInterval: 4, AttackRange: 1.4,
		DetectionRange: 3.5, LeashRange: 6, DamageType: DamageTypeCrush,
		Defence:  DefenceBonuses{Stab: 2, Slash: 2, Crush: 1, Ranged: 2, Magic: 1},
		Elements: ElementalResistances{Earth: 5, Fire: -10},
	},
	riftTyrantTargetID: {
		CombatLevel:    35,
		AttackDamage:   22,
		AttackInterval: 4,
		AttackRange:    enemyLongRangeAttackRange,
		DetectionRange: 18,
		LeashRange:     24,
		DamageType:     DamageTypeStab,
		AttackTypes: []DamageType{
			DamageTypeStab, DamageTypeSlash, DamageTypeCrush, DamageTypeRanged, DamageTypeMagic,
		},
		Defence:  DefenceBonuses{Stab: 24, Slash: 28, Crush: 20, Ranged: 26, Magic: 22},
		Elements: ElementalResistances{Air: 10, Water: 10, Earth: 10, Fire: 10, Holy: -10, Dark: 20},
	},
	rangedTyrantTargetID: {
		CombatLevel:    35,
		AttackDamage:   20,
		AttackInterval: 3,
		AttackRange:    enemyLongRangeAttackRange,
		DetectionRange: 18,
		LeashRange:     24,
		DamageType:     DamageTypeRanged,
		Defence:        DefenceBonuses{Stab: 18, Slash: 20, Crush: 16, Ranged: 30, Magic: 18},
		Elements:       ElementalResistances{Air: 5, Water: 10, Earth: -10, Fire: 5, Holy: 0, Dark: 10},
	},
	mageTyrantTargetID: {
		CombatLevel:    35,
		AttackDamage:   24,
		AttackInterval: 5,
		AttackRange:    enemyLongRangeAttackRange,
		DetectionRange: 18,
		LeashRange:     24,
		DamageType:     DamageTypeMagic,
		Defence:        DefenceBonuses{Stab: 16, Slash: 18, Crush: 14, Ranged: 18, Magic: 32},
		Elements:       ElementalResistances{Air: 15, Water: 15, Earth: 15, Fire: 15, Holy: -20, Dark: 30},
	},
}

func enemyAttackType(profile enemyCombatProfile, attackSequence uint64) DamageType {
	if len(profile.AttackTypes) == 0 {
		return profile.DamageType
	}
	return profile.AttackTypes[attackSequence%uint64(len(profile.AttackTypes))]
}

func combatProfileForEnemy(targetID string) enemyCombatProfile {
	if profile, exists := enemyCombatProfiles[targetID]; exists {
		return profile
	}
	return defaultEnemyCombatProfile
}
