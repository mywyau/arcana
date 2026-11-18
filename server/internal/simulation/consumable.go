package simulation

import (
	"fmt"
	"sort"
)

const (
	combatPotionBoostLevels    = 3
	combatPotionDurationPulses = 500 // Five minutes at one 600 ms action pulse.
)

type activeCombatBoost struct {
	Levels         int
	ExpiresAtPulse uint64
}

type ActiveCombatBoostSnapshot struct {
	Skill          SkillID
	Levels         int
	ExpiresAtPulse uint64
}

type ConsumeItemCommand struct {
	PlayerID      PlayerID
	Sequence      uint64
	ItemID        string
	InventorySlot int
}

func (w *World) QueueConsumeItem(command ConsumeItemCommand) error {
	if command.Sequence == 0 {
		return fmt.Errorf("consume sequence must be positive")
	}
	item, exists := itemDefinition(command.ItemID)
	if !exists || item.Kind != ItemKindConsumable {
		return fmt.Errorf("item %q is not consumable", command.ItemID)
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	player, exists := w.players[command.PlayerID]
	if !exists {
		return fmt.Errorf("player %q does not exist", command.PlayerID)
	}
	if player.defeated {
		return fmt.Errorf("player %q is defeated", command.PlayerID)
	}
	if command.Sequence <= player.lastConsumeSequence {
		return fmt.Errorf("consume sequence %d is not newer than %d", command.Sequence, player.lastConsumeSequence)
	}
	if pending, exists := w.pendingConsumables[command.PlayerID]; exists && command.Sequence <= pending.Sequence {
		return fmt.Errorf("consume sequence %d is not newer than pending sequence %d", command.Sequence, pending.Sequence)
	}
	itemAtSlot, exists := inventoryItemAt(player.inventory, command.InventorySlot)
	if !exists || itemAtSlot.ItemID != command.ItemID || itemAtSlot.Quantity <= 0 {
		return fmt.Errorf("inventory slot %d does not contain item %q", command.InventorySlot, command.ItemID)
	}
	if !playerCanBenefitFromConsumable(player, item.Consumable) {
		return fmt.Errorf("player %q is already full for item %q", command.PlayerID, command.ItemID)
	}
	nextConsumePulse := consumableCooldownPulse(player, item.Consumable.CooldownGroup)
	if w.combat.pulse < nextConsumePulse {
		return fmt.Errorf("%s cooldown lasts until pulse %d", item.Consumable.CooldownGroup, nextConsumePulse)
	}
	w.pendingConsumables[command.PlayerID] = command
	return nil
}

func (w *World) applyPendingConsumables() {
	playerIDs := make([]PlayerID, 0, len(w.pendingConsumables))
	for id := range w.pendingConsumables {
		playerIDs = append(playerIDs, id)
	}
	sort.Slice(playerIDs, func(i, j int) bool { return playerIDs[i] < playerIDs[j] })
	for _, id := range playerIDs {
		command := w.pendingConsumables[id]
		delete(w.pendingConsumables, id)
		player, exists := w.players[id]
		if !exists || player.defeated {
			continue
		}
		item, exists := itemDefinition(command.ItemID)
		if !exists || item.Kind != ItemKindConsumable ||
			w.combat.pulse < consumableCooldownPulse(player, item.Consumable.CooldownGroup) ||
			!playerCanBenefitFromConsumable(player, item.Consumable) ||
			!player.inventory.removeFromSlot(command.InventorySlot, command.ItemID, 1) {
			continue
		}
		player.health = min(player.maximumHealth, player.health+item.Consumable.Healing)
		player.prayer = min(player.maximumPrayer, player.prayer+item.Consumable.PrayerRestoration)
		if player.combatBoosts == nil {
			player.combatBoosts = make(map[SkillID]activeCombatBoost)
		}
		for _, boost := range item.Consumable.SkillBoosts {
			player.combatBoosts[boost.Skill] = activeCombatBoost{
				Levels: boost.Levels, ExpiresAtPulse: w.combat.pulse + item.Consumable.BoostDuration,
			}
		}
		if len(item.Consumable.SkillBoosts) > 0 {
			player.combatBoostsRevision++
		}
		player.lastConsumeSequence = command.Sequence
		nextConsumePulse := w.combat.pulse + item.Consumable.CooldownPulses
		setConsumableCooldownPulse(player, item.Consumable.CooldownGroup, nextConsumePulse)
		if item.Consumable.DelaysAttacks {
			player.nextAttackPulse = max(player.nextAttackPulse, nextConsumePulse)
		}
	}
}

func consumableCooldownPulse(player *playerState, group ConsumableCooldownGroup) uint64 {
	if group == ConsumableCooldownPotion {
		return player.nextPotionConsumePulse
	}
	return player.nextFoodConsumePulse
}

func setConsumableCooldownPulse(player *playerState, group ConsumableCooldownGroup, pulse uint64) {
	if group == ConsumableCooldownPotion {
		player.nextPotionConsumePulse = pulse
		return
	}
	player.nextFoodConsumePulse = pulse
}

func playerCanBenefitFromConsumable(player *playerState, consumable ConsumableDefinition) bool {
	if player == nil {
		return false
	}
	return consumable.Healing > 0 && player.health < player.maximumHealth ||
		consumable.PrayerRestoration > 0 && player.prayer < player.maximumPrayer ||
		len(consumable.SkillBoosts) > 0
}

func effectiveCombatSkillLevel(player *playerState, skill SkillID) int {
	if player == nil {
		return 1
	}
	level := SkillLevelForExperience(player.skills[skill])
	if boost, exists := player.combatBoosts[skill]; exists {
		level += boost.Levels
	}
	return level
}

func (w *World) expireCombatBoosts() {
	for _, player := range w.players {
		changed := false
		for skill, boost := range player.combatBoosts {
			if w.combat.pulse >= boost.ExpiresAtPulse {
				delete(player.combatBoosts, skill)
				changed = true
			}
		}
		if changed {
			player.combatBoostsRevision++
		}
	}
}

func activeCombatBoostSnapshots(player *playerState) []ActiveCombatBoostSnapshot {
	if player == nil {
		return nil
	}
	boosts := make([]ActiveCombatBoostSnapshot, 0, len(player.combatBoosts))
	for skill, boost := range player.combatBoosts {
		boosts = append(boosts, ActiveCombatBoostSnapshot{
			Skill: skill, Levels: boost.Levels, ExpiresAtPulse: boost.ExpiresAtPulse,
		})
	}
	sort.Slice(boosts, func(i, j int) bool { return boosts[i].Skill < boosts[j].Skill })
	return boosts
}
