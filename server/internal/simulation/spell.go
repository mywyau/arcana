package simulation

import (
	"fmt"
	"sort"
)

type SpellID string

const (
	SpellWindBolt                SpellID = "wind_bolt"
	windBoltAttackIntervalPulses         = 5
)

type SpellCommand struct {
	PlayerID PlayerID
	Sequence uint64
	SpellID  SpellID
}

func spellAccuracyScore(player *playerState, spell SpellDefinition) int {
	level := 1
	focusAccuracy := 0
	if player != nil {
		level = SkillLevelForExperience(player.skills[SkillMagic])
		if weapon, exists := weaponDefinition(player.equippedWeapon); exists && weapon.MagicStrength > 0 {
			focusAccuracy = weapon.AccuracyBonus
		}
	}
	return level + spell.AccuracyBonus + focusAccuracy
}

func targetSpellDodgeChance(player *playerState, target *TargetSnapshot, spell SpellDefinition) float64 {
	return calculatedDodgeChance(spellAccuracyScore(player, spell), target.Evasion, target.DodgeRate)
}

func spellPoweredDamage(player *playerState, spell SpellDefinition) float64 {
	level := 1
	focusStrength := 0
	if player != nil {
		level = SkillLevelForExperience(player.skills[SkillMagic])
		if weapon, exists := weaponDefinition(player.equippedWeapon); exists {
			focusStrength = weapon.MagicStrength
		}
	}
	return float64(spell.BasePower+focusStrength) + float64(level-1)*skillDamagePointsPerLevel
}

func canPayItemCosts(inventory inventoryState, costs []ItemCost) bool {
	totals := make(map[string]int, len(costs))
	for _, cost := range costs {
		if cost.ItemID == "" || cost.Quantity <= 0 {
			return false
		}
		totals[cost.ItemID] += cost.Quantity
	}
	for itemID, quantity := range totals {
		if inventory.quantity(itemID) < quantity {
			return false
		}
	}
	return true
}

func payItemCosts(inventory *inventoryState, costs []ItemCost) bool {
	if inventory == nil || !canPayItemCosts(*inventory, costs) {
		return false
	}
	totals := make(map[string]int, len(costs))
	for _, cost := range costs {
		totals[cost.ItemID] += cost.Quantity
	}
	for itemID, quantity := range totals {
		inventory.remove(itemID, quantity)
	}
	return true
}

func activeSpellbook(player *playerState) SpellbookID {
	if player == nil || player.equippedSpellbook == "" {
		return ""
	}
	item, exists := itemDefinition(player.equippedSpellbook)
	if !exists || item.Kind != ItemKindSpellbook {
		return ""
	}
	return item.Spellbook
}

// QueueSpell records a spell selection for the next action pulse. Spell
// selection deliberately does not require a particular weapon.
func (w *World) QueueSpell(command SpellCommand) error {
	if command.Sequence == 0 {
		return fmt.Errorf("spell sequence must be positive")
	}
	definition := SpellDefinition{}
	if command.SpellID != "" {
		var exists bool
		definition, exists = spellDefinition(command.SpellID)
		if !exists {
			return fmt.Errorf("spell %q does not exist", command.SpellID)
		}
		if definition.CastMode != SpellCastAutocast {
			return fmt.Errorf("spell %q cannot be selected for autocast", command.SpellID)
		}
		if definition.EffectType != SpellEffectDirectDamage {
			return fmt.Errorf("spell %q does not have an autocast effect handler", command.SpellID)
		}
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	if _, leaving := w.pendingLeaves[command.PlayerID]; leaving {
		return fmt.Errorf("player %q is leaving", command.PlayerID)
	}
	player, exists := w.players[command.PlayerID]
	if !exists {
		player, exists = w.pendingJoins[command.PlayerID]
		if !exists {
			return fmt.Errorf("player %q does not exist", command.PlayerID)
		}
	}
	if player.defeated {
		return fmt.Errorf("player %q is defeated", command.PlayerID)
	}
	magicLevel := SkillLevelForExperience(player.skills[SkillMagic])
	if command.SpellID != "" && definition.BookID != activeSpellbook(player) {
		return fmt.Errorf("spell %q is not in the equipped spellbook", command.SpellID)
	}
	if command.SpellID != "" && magicLevel < definition.RequiredMagicLevel {
		return fmt.Errorf("spell %q requires Magic level %d", command.SpellID, definition.RequiredMagicLevel)
	}
	if command.Sequence <= player.lastSpellSequence {
		return fmt.Errorf("spell sequence %d is not newer than %d", command.Sequence, player.lastSpellSequence)
	}
	if pending, found := w.combat.pendingSpells[command.PlayerID]; found && command.Sequence <= pending.Sequence {
		return fmt.Errorf("spell sequence %d is not newer than pending sequence %d", command.Sequence, pending.Sequence)
	}
	w.combat.pendingSpells[command.PlayerID] = command
	return nil
}

func (w *World) applySpellCommands() {
	playerIDs := make([]PlayerID, 0, len(w.combat.pendingSpells))
	for id := range w.combat.pendingSpells {
		playerIDs = append(playerIDs, id)
	}
	sort.Slice(playerIDs, func(i, j int) bool { return playerIDs[i] < playerIDs[j] })
	for _, id := range playerIDs {
		command := w.combat.pendingSpells[id]
		if player, exists := w.players[id]; exists && !player.defeated {
			if command.SpellID == "" {
				player.selectedSpell = ""
				player.lastSpellSequence = command.Sequence
			} else if definition, found := spellDefinition(command.SpellID); found &&
				definition.CastMode == SpellCastAutocast &&
				definition.EffectType == SpellEffectDirectDamage &&
				definition.BookID == activeSpellbook(player) &&
				SkillLevelForExperience(player.skills[SkillMagic]) >= definition.RequiredMagicLevel {
				player.selectedSpell = command.SpellID
				player.lastSpellSequence = command.Sequence
			}
		}
		delete(w.combat.pendingSpells, id)
	}
}
