package simulation

import (
	"fmt"
	"sort"
)

type PrayerID string

const (
	PrayerEvasion PrayerID = "evasion"

	evasionDodgeBonus          = 0.05
	evasionDrainIntervalPulses = 10 // Five 600 ms action pulses is three seconds.

	PrayerStrength PrayerID = "strength"

	strengthBonus          = 0.10
	strengthDrainIntervalPulses = 10 // Five 600 ms action pulses is three seconds.
)

type PrayerCommand struct {
	PlayerID PlayerID
	Sequence uint64
	PrayerID PrayerID
	Active   bool
}

type prayerDefinition struct {
	StrenghtBonus       float64
	DodgeBonus          float64
	DrainIntervalPulses uint64
}

func prayerDefinitionFor(id PrayerID) (prayerDefinition, bool) {
	switch id {
	case PrayerEvasion:
		return prayerDefinition{
			StrenghtBonus:       0,
			DodgeBonus:          evasionDodgeBonus,
			DrainIntervalPulses: evasionDrainIntervalPulses,
		}, true
	case PrayerStrength:
		return prayerDefinition{
				StrenghtBonus:       strengthBonus,
				DodgeBonus:          0,
				DrainIntervalPulses: strengthDrainIntervalPulses,
		}, true	
	default:
		return prayerDefinition{}, false
	}
}

// QueuePrayer records an authoritative prayer toggle for the next action pulse.
func (w *World) QueuePrayer(command PrayerCommand) error {
	if command.Sequence == 0 {
		return fmt.Errorf("prayer sequence must be positive")
	}
	if _, exists := prayerDefinitionFor(command.PrayerID); !exists {
		return fmt.Errorf("prayer %q does not exist", command.PrayerID)
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
	if command.Active && player.prayer <= 0 {
		return fmt.Errorf("player %q has no prayer points", command.PlayerID)
	}
	if command.Sequence <= player.lastPrayerSequence {
		return fmt.Errorf("prayer sequence %d is not newer than %d", command.Sequence, player.lastPrayerSequence)
	}
	if pending, found := w.combat.pendingPrayers[command.PlayerID]; found && command.Sequence <= pending.Sequence {
		return fmt.Errorf("prayer sequence %d is not newer than pending sequence %d", command.Sequence, pending.Sequence)
	}
	w.combat.pendingPrayers[command.PlayerID] = command
	return nil
}

func (w *World) applyPrayerCommands() {
	playerIDs := make([]PlayerID, 0, len(w.combat.pendingPrayers))
	for id := range w.combat.pendingPrayers {
		playerIDs = append(playerIDs, id)
	}
	sort.Slice(playerIDs, func(i, j int) bool { return playerIDs[i] < playerIDs[j] })
	for _, id := range playerIDs {
		command := w.combat.pendingPrayers[id]
		if player, exists := w.players[id]; exists && !player.defeated {
			if command.Active && player.prayer > 0 {
				player.activePrayers[command.PrayerID] = true
			} else {
				delete(player.activePrayers, command.PrayerID)
				delete(player.prayerDrainProgress, command.PrayerID)
			}
			player.lastPrayerSequence = command.Sequence
		}
		delete(w.combat.pendingPrayers, id)
	}
}

func (w *World) drainActivePrayers() {
	for _, player := range w.players {
		if player.defeated || len(player.activePrayers) == 0 {
			continue
		}
		prayerIDs := activePrayerIDs(player)
		for _, id := range prayerIDs {
			definition, exists := prayerDefinitionFor(id)
			if !exists {
				delete(player.activePrayers, id)
				delete(player.prayerDrainProgress, id)
				continue
			}
			player.prayerDrainProgress[id]++
			if player.prayerDrainProgress[id] < definition.DrainIntervalPulses {
				continue
			}
			player.prayerDrainProgress[id] = 0
			player.prayer = max(0, player.prayer-1)
			if player.prayer == 0 {
				clearActivePrayers(player)
				break
			}
		}
	}
}

func effectivePlayerDodgeRate(player *playerState) float64 {
	rate := player.dodgeRate
	for id := range player.activePrayers {
		if definition, exists := prayerDefinitionFor(id); exists {
			rate += definition.DodgeBonus
		}
	}
	return min(0.95, rate)
}

func activePrayerIDs(player *playerState) []PrayerID {
	ids := make([]PrayerID, 0, len(player.activePrayers))
	for id, active := range player.activePrayers {
		if active {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func clearActivePrayers(player *playerState) {
	clear(player.activePrayers)
	clear(player.prayerDrainProgress)
}
