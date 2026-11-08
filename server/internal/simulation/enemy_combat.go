package simulation

type enemyCombatProfile struct {
	CombatLevel    int
	AttackDamage   int
	AttackInterval uint64
	AttackRange    float64
	DetectionRange float64
	LeashRange     float64
	DamageType     DamageType
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
		AttackRange:    1.7,
		DetectionRange: 6,
		LeashRange:     10,
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
}

func combatProfileForEnemy(targetID string) enemyCombatProfile {
	if profile, exists := enemyCombatProfiles[targetID]; exists {
		return profile
	}
	return defaultEnemyCombatProfile
}
