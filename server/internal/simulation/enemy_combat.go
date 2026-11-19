package simulation

import "math"

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
	HeavyStrike    heavyStrikeProfile
	PiercingVolley piercingVolleyProfile
	ArrowVolley    arrowVolleyProfile
	ElementalZones elementalZoneProfile
	QuadrantBlast  quadrantBlastProfile
}

type heavyStrikeProfile struct {
	EveryAttacks                  uint64
	WindupPulses                  uint64
	Radius                        float64
	DamageMultiplier              float64
	VulnerabilityPulses           uint64
	VulnerabilityDamageMultiplier float64
	MeleeOverloadRange            float64
}

type piercingVolleyProfile struct {
	CycleAttacks     uint64
	TriggerAttack    uint64
	Shots            int
	WindupPulses     uint64
	BetweenPulses    uint64
	Length           float64
	Width            float64
	SpreadRadians    float64
	DamageMultiplier float64
}

type arrowVolleyProfile struct {
	CycleAttacks     uint64
	TriggerAttack    uint64
	Waves            int
	WindupPulses     uint64
	BetweenPulses    uint64
	Radius           float64
	DamageMultiplier float64
}

type elementalZoneProfile struct {
	CycleAttacks                  uint64
	TriggerAttack                 uint64
	Waves                         int
	WindupPulses                  uint64
	BetweenPulses                 uint64
	Radius                        float64
	DamageMultiplier              float64
	VulnerabilityPulses           uint64
	VulnerabilityDamageMultiplier float64
}

type quadrantBlastProfile struct {
	CycleAttacks     uint64
	TriggerAttack    uint64
	WindupPulses     uint64
	DamageMultiplier float64
	EnrageHealth     float64
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
	dustWraithTargetID: {
		CombatLevel:  6,
		AttackDamage: 6, AttackInterval: 4, AttackRange: 1.6,
		DetectionRange: 4, LeashRange: 7, DamageType: DamageTypeMagic,
		Defence:  DefenceBonuses{Stab: 3, Slash: 3, Crush: 2, Ranged: 4, Magic: 2},
		Elements: ElementalResistances{Air: -30, Earth: 20},
	},
	cinderImpTargetID: {
		CombatLevel:  8,
		AttackDamage: 8, AttackInterval: 4, AttackRange: 1.8,
		DetectionRange: 4.5, LeashRange: 8, DamageType: DamageTypeMagic,
		Defence:  DefenceBonuses{Stab: 5, Slash: 4, Crush: 4, Ranged: 5, Magic: 3},
		Elements: ElementalResistances{Water: -30, Fire: 35},
	},
	tideWispTargetID: {
		CombatLevel:  10,
		AttackDamage: 10, AttackInterval: 4, AttackRange: 2,
		DetectionRange: 5, LeashRange: 8, DamageType: DamageTypeMagic,
		Defence:  DefenceBonuses{Stab: 6, Slash: 6, Crush: 5, Ranged: 7, Magic: 4},
		Elements: ElementalResistances{Water: 35, Earth: -30},
	},
	thornlingTargetID: {
		CombatLevel:  12,
		AttackDamage: 12, AttackInterval: 5, AttackRange: 1.6,
		DetectionRange: 5, LeashRange: 9, DamageType: DamageTypeCrush,
		Defence:  DefenceBonuses{Stab: 8, Slash: 7, Crush: 10, Ranged: 8, Magic: 5},
		Elements: ElementalResistances{Earth: 35, Fire: -30},
	},
	riftTyrantTargetID: {
		CombatLevel:    35,
		AttackDamage:   22,
		AttackInterval: 4,
		AttackRange:    enemyLongRangeAttackRange,
		DetectionRange: 26,
		LeashRange:     27,
		DamageType:     DamageTypeStab,
		AttackTypes: []DamageType{
			DamageTypeStab, DamageTypeSlash, DamageTypeCrush,
		},
		Defence:  DefenceBonuses{Stab: 100, Slash: 115, Crush: 90, Ranged: 105, Magic: 95},
		Elements: ElementalResistances{Air: 10, Water: 10, Earth: 10, Fire: 10, Holy: -10, Dark: 20},
		HeavyStrike: heavyStrikeProfile{
			EveryAttacks: 4, WindupPulses: 4, Radius: 5, DamageMultiplier: 2,
			VulnerabilityPulses: 3, VulnerabilityDamageMultiplier: 1.25,
		},
	},
	rangedTyrantTargetID: {
		CombatLevel:    35,
		AttackDamage:   20,
		AttackInterval: 3,
		AttackRange:    enemyLongRangeAttackRange,
		DetectionRange: 26,
		LeashRange:     27,
		DamageType:     DamageTypeRanged,
		Defence:        DefenceBonuses{Stab: 85, Slash: 90, Crush: 80, Ranged: 115, Magic: 150},
		Elements:       ElementalResistances{Air: 5, Water: 10, Earth: -10, Fire: 5, Holy: 0, Dark: 10},
		HeavyStrike: heavyStrikeProfile{
			EveryAttacks: 4, WindupPulses: 4, Radius: 5, DamageMultiplier: 2,
			VulnerabilityPulses: 7, VulnerabilityDamageMultiplier: 1.25,
			MeleeOverloadRange: 6,
		},
		PiercingVolley: piercingVolleyProfile{
			CycleAttacks: 12, TriggerAttack: 8, Shots: 3, WindupPulses: 3, BetweenPulses: 2,
			Length: 32, Width: 2, SpreadRadians: 25 * math.Pi / 180, DamageMultiplier: 1.25,
		},
		ArrowVolley: arrowVolleyProfile{
			CycleAttacks: 12, TriggerAttack: 12, Waves: 3,
			WindupPulses: 3, BetweenPulses: 2, Radius: 1.8, DamageMultiplier: 1,
		},
	},
	mageTyrantTargetID: {
		CombatLevel:    35,
		AttackDamage:   24,
		AttackInterval: 5,
		AttackRange:    40,
		DetectionRange: 26,
		LeashRange:     27,
		DamageType:     DamageTypeMagic,
		Defence:        DefenceBonuses{Stab: 150, Slash: 155, Crush: 145, Ranged: 70, Magic: 105},
		Elements:       ElementalResistances{Air: 15, Water: 15, Earth: 15, Fire: 15, Holy: -20, Dark: 30},
		ElementalZones: elementalZoneProfile{
			CycleAttacks: 8, TriggerAttack: 4, Waves: 3,
			WindupPulses: 4, BetweenPulses: 3, Radius: 4, DamageMultiplier: 1.25,
			VulnerabilityPulses: 5, VulnerabilityDamageMultiplier: 1.25,
		},
		QuadrantBlast: quadrantBlastProfile{
			CycleAttacks: 8, TriggerAttack: 8, WindupPulses: 5,
			DamageMultiplier: 1.5, EnrageHealth: 1.0 / 3.0,
		},
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
