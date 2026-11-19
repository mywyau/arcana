package simulation

// The prototype PvP field occupies the southern world extension. Keeping the
// bounds here makes entry and combat authorization server authoritative.
var prototypePvPField = worldRectangle{
	minimumX: -74,
	maximumX: 74,
	minimumZ: -144,
	maximumZ: -80,
}

type worldRectangle struct {
	minimumX float64
	maximumX float64
	minimumZ float64
	maximumZ float64
}

func (area worldRectangle) contains(position Position) bool {
	return position.X >= area.minimumX && position.X <= area.maximumX &&
		position.Z >= area.minimumZ && position.Z <= area.maximumZ
}

func playerInPvPField(player *playerState) bool {
	return player != nil && prototypePvPField.contains(player.position)
}

func playersMayFight(attacker, defender *playerState) bool {
	return attacker != nil && defender != nil && attacker != defender &&
		!attacker.defeated && !defender.defeated &&
		playerInPvPField(attacker) && playerInPvPField(defender)
}

func (w *World) resolvePlayerAutoAttack(attacker, defender *playerState) {
	if !playersMayFight(attacker, defender) {
		w.clearPlayerTarget(attacker, "left_pvp_area")
		return
	}

	weapon, _ := equippedWeaponDefinitionForStyle(attacker)
	spell, castingSpell := spellDefinition(attacker.selectedSpell)
	if castingSpell {
		_, castingBlocked := statusBlockingAction(attacker, statusActionCast, w.combat.pulse)
		castingSpell = !castingBlocked
	}
	if castingSpell {
		spell = spellDefinitionForStyle(attacker, spell)
	}
	attackRange := weapon.Range
	approachRange := weapon.ApproachRange
	attackInterval := effectiveWeaponAttackInterval(attacker, weapon)
	if castingSpell {
		attackRange = spell.Range
		approachRange = max(0.1, spell.Range-0.75)
		attackInterval = spell.AttackInterval
	}
	if distance(attacker.position, defender.position) > attackRange {
		if err := w.approachDummy(attacker, defender.position, approachRange); err != nil {
			attacker.lastActionResult = "target_unreachable"
			attacker.autoApproaching = false
		} else {
			attacker.lastActionResult = "approaching"
		}
		return
	}
	if attacker.autoApproaching {
		stopPlayer(attacker)
		attacker.autoApproaching = false
	}
	if w.combat.pulse < attacker.nextAttackPulse {
		return
	}

	if castingSpell && !paySpellCosts(attacker, spell.Costs) {
		attacker.lastActionResult = "missing_reagents"
		return
	}
	ammunitionStrength := 0
	if !castingSpell && weapon.Ammunition.Quantity > 0 {
		ammunitionCost, strength, available := ammunitionForWeapon(attacker, weapon)
		if !available || !payItemCosts(&attacker.ammunitionPouch, []ItemCost{ammunitionCost}) {
			attacker.lastActionResult = "missing_ammunition"
			return
		}
		ammunitionStrength = strength
	}

	attacker.lastActionPulse = w.combat.pulse
	attacker.nextAttackPulse = w.combat.pulse + attackInterval
	accuracyChance := targetAccuracyChance(attacker, weapon)
	attackerAccuracy := playerAccuracyScore(attacker, weapon)
	if castingSpell {
		accuracyChance = targetSpellAccuracyChance(attacker, spell)
		attackerAccuracy = spellAccuracyScore(attacker, spell)
	}
	dodgeChance := playerDodgeChance(attackerAccuracy, defender)
	castSpellID := SpellID("")
	if castingSpell {
		castSpellID = attacker.selectedSpell
	}
	outcomeRoll := w.rollSource.Float64()
	if outcomeRoll >= accuracyChance {
		attacker.lastActionResult = "missed"
		w.combat.recordEvent(CombatEvent{Pulse: w.combat.pulse, Type: CombatEventMiss,
			SourceID: string(attacker.id), TargetID: string(defender.id), SpellID: castSpellID})
		return
	}
	if outcomeRoll < accuracyChance*dodgeChance {
		attacker.lastActionResult = "dodged"
		w.combat.recordEvent(CombatEvent{Pulse: w.combat.pulse, Type: CombatEventDodged,
			SourceID: string(attacker.id), TargetID: string(defender.id), SpellID: castSpellID})
		return
	}

	poweredDamage := weaponPoweredDamageWithAmmunition(attacker, weapon, ammunitionStrength)
	damageVariance := weapon.DamageVariance
	var element MagicElement
	damageType := DamageTypeCrush
	if styleDamageType, exists := weaponDamageTypeForStyle(weapon, attacker.combatStyle); exists {
		damageType = styleDamageType
	}
	if castingSpell {
		poweredDamage = spellPoweredDamage(attacker, spell)
		damageVariance = spell.DamageVariance
		damageType = DamageTypeMagic
		element = spell.Element
	}
	rolledDamage := rollDamageWithVariance(poweredDamage, damageVariance, w.rollSource.Float64())
	damageDealt := min(reduceIncomingDamage(defender, rolledDamage, damageType, element), defender.health)
	defender.health = max(0, defender.health-damageDealt)
	triggerEmergencyPrayer(defender)
	attacker.lastActionResult = "hit"
	w.combat.recordEvent(CombatEvent{Pulse: w.combat.pulse, Type: CombatEventDamage,
		SourceID: string(attacker.id), TargetID: string(defender.id), Amount: damageDealt, SpellID: castSpellID})
	// PvP intentionally grants no combat experience in this first slice. That
	// avoids easy account boosting while the rules and rewards are still basic.
	if defender.health == 0 {
		w.defeatPlayer(defender)
	}
}
