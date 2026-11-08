package simulation

type enemyCombatProfile struct {
	AttackDamage   int
	AttackInterval uint64
	AttackRange    float64
	DetectionRange float64
	LeashRange     float64
	DamageType     DamageType
}

var defaultEnemyCombatProfile = enemyCombatProfile{
	AttackDamage:   dummyAttackDamage,
	AttackInterval: dummyAttackIntervalPulses,
	AttackRange:    dummyAttackRange,
	DetectionRange: 0,
	LeashRange:     dummyLeashRange,
	DamageType:     DamageTypeCrush,
}

var enemyCombatProfiles = map[string]enemyCombatProfile{
	hillWardenTargetID: {
		AttackDamage:   18,
		AttackInterval: 5,
		AttackRange:    1.7,
		DetectionRange: 6,
		LeashRange:     10,
		DamageType:     DamageTypeSlash,
	},
	wildFowlTargetID: {
		AttackDamage:   wildFowlAttackDamage,
		AttackInterval: 3,
		AttackRange:    1.2,
		DetectionRange: 3,
		LeashRange:     5,
		DamageType:     DamageTypeStab,
	},
}

func combatProfileForEnemy(targetID string) enemyCombatProfile {
	if profile, exists := enemyCombatProfiles[targetID]; exists {
		return profile
	}
	return defaultEnemyCombatProfile
}
