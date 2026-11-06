package simulation

// CombatStatsSnapshot contains the derived values used by combat. Sending the
// derived values keeps the client display aligned with the authoritative rules.
type CombatStatsSnapshot struct {
	Weapon                 WeaponID
	Style                  CombatStyleID
	Spell                  SpellID
	AccuracySkill          SkillID
	AccuracyLevel          int
	WeaponAccuracyBonus    int
	StyleAccuracyBonus     int
	SpellAccuracyBonus     int
	PrayerAccuracyBonus    int
	TotalAccuracy          int
	DamageSkill            SkillID
	DamageLevel            int
	StrengthBonus          int
	SpellPower             int
	SkillDamageBonus       float64
	PrayerDamageBonus      float64
	PrayerDamageReduction  float64
	BaseAttackInterval     uint64
	AttackInterval         uint64
	AttackIntervalModifier int
	MinimumDamage          int
	MaximumDamage          int
	DefenceLevel           int
	DefenceBonuses         DefenceBonuses
}

func derivedCombatStats(player *playerState) CombatStatsSnapshot {
	if player == nil {
		return CombatStatsSnapshot{}
	}
	weapon, exists := weaponDefinitionForStyle(player.equippedWeapon, player.combatStyle)
	if !exists {
		return CombatStatsSnapshot{}
	}
	if spell, castingSpell := spellDefinition(player.selectedSpell); castingSpell {
		prayerBonuses := activePrayerBonuses(player)
		focusAccuracy := 0
		focusStrength := 0
		if weapon.MagicStrength > 0 {
			focusAccuracy = weapon.AccuracyBonus
			focusStrength = weapon.MagicStrength
		}
		minimumDamage, maximumDamage := damageRange(
			spellPoweredDamage(player, spell), spell.DamageVariance,
		)
		magicLevel := SkillLevelForExperience(player.skills[SkillMagic])
		return CombatStatsSnapshot{
			Weapon:                weapon.ID,
			Style:                 player.combatStyle,
			Spell:                 spell.ID,
			AccuracySkill:         SkillMagic,
			AccuracyLevel:         magicLevel,
			WeaponAccuracyBonus:   focusAccuracy,
			SpellAccuracyBonus:    spell.AccuracyBonus,
			PrayerDamageReduction: prayerBonuses.DamageReduction,
			BaseAttackInterval:    spell.AttackInterval,
			AttackInterval:        spell.AttackInterval,
			TotalAccuracy:         spellAccuracyScore(player, spell),
			DamageSkill:           SkillMagic,
			DamageLevel:           magicLevel,
			StrengthBonus:         focusStrength,
			SpellPower:            spell.BasePower,
			SkillDamageBonus:      float64(magicLevel-1) * skillDamagePointsPerLevel,
			MinimumDamage:         minimumDamage,
			MaximumDamage:         maximumDamage,
			DefenceLevel:          SkillLevelForExperience(player.skills[SkillDefence]),
			DefenceBonuses:        equippedDefenceBonuses(player.equippedArmour),
		}
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

	prayerBonuses := activePrayerBonuses(player)
	attackInterval := effectiveWeaponAttackInterval(player, weapon)
	return CombatStatsSnapshot{
		Weapon:                 weapon.ID,
		Style:                  player.combatStyle,
		AccuracySkill:          accuracySkill,
		AccuracyLevel:          SkillLevelForExperience(player.skills[accuracySkill]),
		WeaponAccuracyBonus:    weapon.AccuracyBonus,
		StyleAccuracyBonus:     combatStyleAccuracyBonus(player.combatStyle),
		PrayerAccuracyBonus:    prayerAccuracyBonus(player, weapon),
		TotalAccuracy:          playerAccuracyScore(player, weapon),
		DamageSkill:            damageSkill,
		DamageLevel:            SkillLevelForExperience(player.skills[damageSkill]),
		StrengthBonus:          strengthBonus,
		SkillDamageBonus:       float64(SkillLevelForExperience(player.skills[damageSkill])-1) * skillDamagePointsPerLevel,
		PrayerDamageBonus:      prayerBonuses.MeleeDamage,
		PrayerDamageReduction:  prayerBonuses.DamageReduction,
		BaseAttackInterval:     weapon.AttackInterval,
		AttackInterval:         attackInterval,
		AttackIntervalModifier: int(attackInterval) - int(weapon.AttackInterval),
		MinimumDamage:          minimumDamage,
		MaximumDamage:          maximumDamage,
		DefenceLevel:           SkillLevelForExperience(player.skills[SkillDefence]),
		DefenceBonuses:         equippedDefenceBonuses(player.equippedArmour),
	}
}
