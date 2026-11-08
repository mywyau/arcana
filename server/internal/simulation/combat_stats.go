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
	EquipmentAccuracyBonus int
	StyleAccuracyBonus     int
	SpellAccuracyBonus     int
	PrayerAccuracyBonus    int
	TotalAccuracy          int
	DamageSkill            SkillID
	DamageLevel            int
	StrengthBonus          int
	EquipmentStrengthBonus int
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
	ArmourDamageReduction  float64
	MaximumHealthBonus     int
}

type CombatMatchupSnapshot struct {
	TargetID            string
	PlayerHitChance     float64
	PlayerMinimumDamage int
	PlayerMaximumDamage int
	PlayerAttackRange   float64
	TargetHitChance     float64
	TargetMinimumDamage int
	TargetMaximumDamage int
	TargetAttackRange   float64
}

// playerCombatLevel combines general survivability with the player's strongest
// offensive discipline. This avoids penalising specialised melee, ranged, or
// magic characters for not levelling every offensive skill equally.
func playerCombatLevel(player *playerState) int {
	if player == nil {
		return 1
	}
	attack := effectiveCombatSkillLevel(player, SkillAttack)
	strength := effectiveCombatSkillLevel(player, SkillStrength)
	defence := effectiveCombatSkillLevel(player, SkillDefence)
	agility := SkillLevelForExperience(player.skills[SkillAgility])
	ranged := effectiveCombatSkillLevel(player, SkillRange)
	magic := effectiveCombatSkillLevel(player, SkillMagic)
	melee := float64(attack+strength) / 2
	offence := max(melee, float64(ranged), float64(magic))
	defenceRating := float64(defence+agility) / 2
	return max(1, int((offence+defenceRating)/2+0.5))
}

func combatMatchup(player *playerState, target *TargetSnapshot) CombatMatchupSnapshot {
	if player == nil || target == nil {
		return CombatMatchupSnapshot{}
	}
	weapon, exists := equippedWeaponDefinitionForStyle(player)
	if !exists {
		return CombatMatchupSnapshot{}
	}
	playerDodge := targetDodgeChance(player, target, weapon)
	poweredDamage := weaponPoweredDamage(player, weapon)
	variance := weapon.DamageVariance
	playerAttackRange := weapon.Range
	if spell, casting := spellDefinition(player.selectedSpell); casting {
		spell = spellDefinitionForStyle(player, spell)
		playerDodge = targetSpellDodgeChance(player, target, spell)
		poweredDamage = spellPoweredDamage(player, spell) * elementalDamageMultiplier(target, spell.Element)
		variance = spell.DamageVariance
		playerAttackRange = spell.Range
	}
	playerMinimum, playerMaximum := damageRange(poweredDamage, variance)
	profile := combatProfileForEnemy(target.ID)
	targetDamage := reduceIncomingDamage(player, profile.AttackDamage, profile.DamageType)
	return CombatMatchupSnapshot{
		TargetID:            target.ID,
		PlayerHitChance:     1 - playerDodge,
		PlayerMinimumDamage: playerMinimum, PlayerMaximumDamage: playerMaximum,
		PlayerAttackRange:   playerAttackRange,
		TargetHitChance:     1 - playerDodgeChance(target.Accuracy, player),
		TargetMinimumDamage: targetDamage, TargetMaximumDamage: targetDamage,
		TargetAttackRange: profile.AttackRange,
	}
}

func (w *World) combatMatchupForPlayer(player *playerState) CombatMatchupSnapshot {
	if player == nil {
		return CombatMatchupSnapshot{}
	}
	if target, exists := w.combat.dummyByID(player.combatTarget); exists {
		return combatMatchup(player, target)
	}
	for _, candidate := range w.combat.dummyTargets() {
		if candidate.dummy.TargetID == player.id {
			return combatMatchup(player, candidate.dummy)
		}
	}
	return CombatMatchupSnapshot{}
}

func derivedCombatStats(player *playerState) CombatStatsSnapshot {
	if player == nil {
		return CombatStatsSnapshot{}
	}
	weapon, exists := equippedWeaponDefinitionForStyle(player)
	if !exists {
		return CombatStatsSnapshot{}
	}
	if spell, castingSpell := spellDefinition(player.selectedSpell); castingSpell {
		spell = spellDefinitionForStyle(player, spell)
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
		magicLevel := effectiveCombatSkillLevel(player, SkillMagic)
		equipmentBonuses := equippedOffenceBonuses(player.equippedArmour)
		return CombatStatsSnapshot{
			Weapon:                 weapon.ID,
			Style:                  player.combatStyle,
			Spell:                  spell.ID,
			AccuracySkill:          SkillMagic,
			AccuracyLevel:          magicLevel,
			WeaponAccuracyBonus:    focusAccuracy,
			EquipmentAccuracyBonus: equipmentBonuses.MagicAccuracy,
			SpellAccuracyBonus:     spell.AccuracyBonus,
			PrayerDamageReduction:  prayerBonuses.DamageReduction,
			BaseAttackInterval:     spell.AttackInterval,
			AttackInterval:         spell.AttackInterval,
			TotalAccuracy:          spellAccuracyScore(player, spell),
			DamageSkill:            SkillMagic,
			DamageLevel:            magicLevel,
			StrengthBonus:          focusStrength,
			EquipmentStrengthBonus: equipmentBonuses.MagicStrength,
			SpellPower:             spell.BasePower,
			SkillDamageBonus:       float64(magicLevel-1) * skillDamagePointsPerLevel,
			PrayerDamageBonus:      prayerBonuses.MagicDamage,
			MinimumDamage:          minimumDamage,
			MaximumDamage:          maximumDamage,
			DefenceLevel:           effectiveCombatSkillLevel(player, SkillDefence),
			DefenceBonuses:         equippedDefenceBonusesForPlayer(player),
			ArmourDamageReduction:  armourDamageReduction(player, DamageTypeCrush),
			MaximumHealthBonus:     equippedHealthBonus(player.equippedArmour) + ringRackHealthBonus(player.ringRack),
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
	equipmentAccuracy, equipmentStrength := equipmentOffenceForWeapon(player, weapon)
	attackInterval := effectiveWeaponAttackInterval(player, weapon)
	return CombatStatsSnapshot{
		Weapon:                 weapon.ID,
		Style:                  player.combatStyle,
		AccuracySkill:          accuracySkill,
		AccuracyLevel:          effectiveCombatSkillLevel(player, accuracySkill),
		WeaponAccuracyBonus:    weapon.AccuracyBonus,
		EquipmentAccuracyBonus: equipmentAccuracy,
		StyleAccuracyBonus:     combatStyleAccuracyBonus(player.combatStyle),
		PrayerAccuracyBonus:    prayerAccuracyBonus(player, weapon),
		TotalAccuracy:          playerAccuracyScore(player, weapon),
		DamageSkill:            damageSkill,
		DamageLevel:            effectiveCombatSkillLevel(player, damageSkill),
		StrengthBonus:          strengthBonus,
		EquipmentStrengthBonus: equipmentStrength,
		SkillDamageBonus:       float64(effectiveCombatSkillLevel(player, damageSkill)-1) * skillDamagePointsPerLevel,
		PrayerDamageBonus:      prayerDamageBonusForWeapon(prayerBonuses, weapon),
		PrayerDamageReduction:  prayerBonuses.DamageReduction,
		BaseAttackInterval:     weapon.AttackInterval,
		AttackInterval:         attackInterval,
		AttackIntervalModifier: int(attackInterval) - int(weapon.AttackInterval),
		MinimumDamage:          minimumDamage,
		MaximumDamage:          maximumDamage,
		DefenceLevel:           effectiveCombatSkillLevel(player, SkillDefence),
		DefenceBonuses:         equippedDefenceBonusesForPlayer(player),
		ArmourDamageReduction:  armourDamageReduction(player, DamageTypeCrush),
		MaximumHealthBonus:     equippedHealthBonus(player.equippedArmour) + ringRackHealthBonus(player.ringRack),
	}
}

func prayerDamageBonusForWeapon(bonuses prayerBonuses, weapon WeaponDefinition) float64 {
	switch {
	case weapon.RangedStrength > 0:
		return bonuses.RangedDamage
	case weapon.MagicStrength > 0:
		return bonuses.MagicDamage
	default:
		return bonuses.MeleeDamage
	}
}
