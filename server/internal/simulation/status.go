package simulation

import "sort"

const maximumActiveStatuses = 32

type StatusEffectID string
type StatusCategory string
type statusAction string

const (
	StatusStun StatusEffectID = "stun"

	StatusCategoryControl StatusCategory = "control"

	statusActionMove       statusAction = "move"
	statusActionAttack     statusAction = "attack"
	statusActionGather     statusAction = "gather"
	statusActionProduction statusAction = "production"
)

type statusDefinition struct {
	ID       StatusEffectID
	Name     string
	Category StatusCategory
	Public   bool
	Blocks   map[statusAction]bool
}

type activeStatus struct {
	EffectID       StatusEffectID
	SourceID       string
	AppliedAtPulse uint64
	ExpiresAtPulse uint64
}

type ActiveStatusSnapshot struct {
	EffectID       StatusEffectID
	SourceID       string
	AppliedAtPulse uint64
	ExpiresAtPulse uint64
}

var statusDefinitions = map[StatusEffectID]statusDefinition{
	StatusStun: {
		ID: StatusStun, Name: "Stunned", Category: StatusCategoryControl, Public: true,
		Blocks: map[statusAction]bool{
			statusActionMove: true, statusActionAttack: true,
			statusActionGather: true, statusActionProduction: true,
		},
	},
}

func statusActive(player *playerState, effectID StatusEffectID, pulse uint64) bool {
	if player == nil {
		return false
	}
	status, exists := player.activeStatuses[effectID]
	return exists && pulse <= status.ExpiresAtPulse
}

func statusBlockingAction(player *playerState, action statusAction, pulse uint64) (statusDefinition, bool) {
	if player == nil {
		return statusDefinition{}, false
	}
	ids := make([]string, 0, len(player.activeStatuses))
	for id := range player.activeStatuses {
		ids = append(ids, string(id))
	}
	sort.Strings(ids)
	for _, rawID := range ids {
		id := StatusEffectID(rawID)
		definition, known := statusDefinitions[id]
		if known && definition.Blocks[action] && statusActive(player, id, pulse) {
			return definition, true
		}
	}
	return statusDefinition{}, false
}

func statusActionError(player *playerState, action statusAction, pulse uint64) error {
	definition, blocked := statusBlockingAction(player, action, pulse)
	if !blocked {
		return nil
	}
	return &statusBlockedError{statusName: definition.Name}
}

type statusBlockedError struct {
	statusName string
}

func (err *statusBlockedError) Error() string {
	if err == nil || err.statusName == "" {
		return "action is blocked by a status effect"
	}
	return "you are " + lowerFirst(err.statusName)
}

func lowerFirst(value string) string {
	if value == "" {
		return value
	}
	bytes := []byte(value)
	if bytes[0] >= 'A' && bytes[0] <= 'Z' {
		bytes[0] += 'a' - 'A'
	}
	return string(bytes)
}

func (w *World) applyStatus(player *playerState, effectID StatusEffectID, sourceID string, durationPulses uint64) bool {
	if player == nil || durationPulses == 0 {
		return false
	}
	definition, known := statusDefinitions[effectID]
	if !known {
		return false
	}
	if player.activeStatuses == nil {
		player.activeStatuses = make(map[StatusEffectID]activeStatus)
	}
	if _, exists := player.activeStatuses[effectID]; !exists && len(player.activeStatuses) >= maximumActiveStatuses {
		return false
	}
	expiresAt := w.combat.pulse + durationPulses
	if current, exists := player.activeStatuses[effectID]; exists && current.ExpiresAtPulse > expiresAt {
		expiresAt = current.ExpiresAtPulse
	}
	player.activeStatuses[effectID] = activeStatus{
		EffectID: effectID, SourceID: sourceID,
		AppliedAtPulse: w.combat.pulse, ExpiresAtPulse: expiresAt,
	}
	player.statusesRevision++
	if definition.Blocks[statusActionMove] || definition.Blocks[statusActionGather] || definition.Blocks[statusActionProduction] {
		w.cancelActionsBlockedByStatus(player)
	}
	return true
}

func (w *World) cancelActionsBlockedByStatus(player *playerState) {
	stopPlayer(player)
	player.autoApproaching = false
	player.autoPickingUp = false
	delete(w.pendingMoves, player.id)
	delete(w.pendingPickups, player.id)
	delete(w.pendingGathering, player.id)
	delete(w.combat.pendingTargets, player.id)
	clearGathering(player, "stunned")

	if player.brewRemaining > 0 {
		delete(w.pendingBrews, player.id)
		player.brewRemaining = 0
		player.nextBrewPulse = 0
		player.brewingResult = "brewing_stopped"
		player.brewingRevision++
	}
	if player.fletchRemaining > 0 {
		delete(w.pendingFletches, player.id)
		player.fletchRemaining = 0
		player.nextFletchPulse = 0
		player.fletchingResult = "fletching_stopped"
		player.fletchingRevision++
	}
}

func (w *World) expireStatuses() {
	for _, player := range w.players {
		changed := false
		for id, status := range player.activeStatuses {
			if w.combat.pulse > status.ExpiresAtPulse {
				delete(player.activeStatuses, id)
				changed = true
			}
		}
		if changed {
			player.statusesRevision++
		}
	}
}

func clearStatuses(player *playerState) {
	if player == nil || len(player.activeStatuses) == 0 {
		return
	}
	clear(player.activeStatuses)
	player.statusesRevision++
}

func activeStatusSnapshots(player *playerState) []ActiveStatusSnapshot {
	if player == nil {
		return nil
	}
	ids := make([]string, 0, len(player.activeStatuses))
	for id := range player.activeStatuses {
		if definition, known := statusDefinitions[id]; known && definition.Public {
			ids = append(ids, string(id))
		}
	}
	sort.Strings(ids)
	snapshots := make([]ActiveStatusSnapshot, 0, len(ids))
	for _, rawID := range ids {
		status := player.activeStatuses[StatusEffectID(rawID)]
		snapshots = append(snapshots, ActiveStatusSnapshot{
			EffectID: status.EffectID, SourceID: status.SourceID,
			AppliedAtPulse: status.AppliedAtPulse, ExpiresAtPulse: status.ExpiresAtPulse,
		})
	}
	return snapshots
}
