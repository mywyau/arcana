package simulation

import "sort"

const maximumActiveStatuses = 32

// slowMovementSpeedMultiplier is catalogue balance rather than runtime state,
// so future content can tune the shared Slow effect in one place.
const slowMovementSpeedMultiplier = 0.7

type StatusEffectID string
type StatusCategory string
type statusAction string

const (
	StatusStun    StatusEffectID = "stun"
	StatusRoot    StatusEffectID = "root"
	StatusSlow    StatusEffectID = "slow"
	StatusSilence StatusEffectID = "silence"

	StatusCategoryControl          StatusCategory = "control"
	StatusCategoryMovementModifier StatusCategory = "movement_modifier"

	statusActionMove       statusAction = "move"
	statusActionAttack     statusAction = "attack"
	statusActionCast       statusAction = "cast"
	statusActionGather     statusAction = "gather"
	statusActionProduction statusAction = "production"
)

type statusDefinition struct {
	ID                      StatusEffectID
	Name                    string
	Category                StatusCategory
	Public                  bool
	Blocks                  map[statusAction]bool
	MovementSpeedMultiplier float64
}

type activeStatus struct {
	SlowPenaltyReductionPercent int
	EffectID                    StatusEffectID
	SourceID                    string
	AppliedAtPulse              uint64
	ExpiresAtPulse              uint64
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
	StatusRoot: {
		ID: StatusRoot, Name: "Rooted", Category: StatusCategoryControl, Public: true,
		Blocks: map[statusAction]bool{statusActionMove: true},
	},
	StatusSlow: {
		ID: StatusSlow, Name: "Slowed", Category: StatusCategoryMovementModifier, Public: true,
		MovementSpeedMultiplier: slowMovementSpeedMultiplier,
	},
	StatusSilence: {
		ID: StatusSilence, Name: "Silenced", Category: StatusCategoryControl, Public: true,
		Blocks: map[statusAction]bool{statusActionCast: true},
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

// statusCurrentlyBlocks is for derived snapshots built immediately after the
// pulse-expiration pass. Authoritative commands should use statusActionError
// with the current pulse instead.
func statusCurrentlyBlocks(player *playerState, action statusAction) bool {
	if player == nil {
		return false
	}
	for id := range player.activeStatuses {
		if definition, known := statusDefinitions[id]; known && definition.Blocks[action] {
			return true
		}
	}
	return false
}

func effectivePlayerMovementSpeed(player *playerState) float64 {
	if player == nil {
		return 0
	}
	multiplier := 1.0
	for id, status := range player.activeStatuses {
		definition, known := statusDefinitions[id]
		if !known || definition.MovementSpeedMultiplier <= 0 {
			continue
		}
		penalty := 1 - definition.MovementSpeedMultiplier
		multiplier *= 1 - penalty*(1-float64(status.SlowPenaltyReductionPercent)/100)
	}
	return player.speed * max(0.1, multiplier)
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
	if len(definition.Blocks) > 0 {
		w.cancelActionsBlockedByStatus(player, definition)
	}
	return true
}

func (w *World) cancelActionsBlockedByStatus(player *playerState, definition statusDefinition) {
	if definition.Blocks[statusActionMove] {
		stopPlayer(player)
		player.autoApproaching = false
		player.autoPickingUp = false
		delete(w.pendingMoves, player.id)
		delete(w.pendingPickups, player.id)
		delete(w.pendingGathering, player.id)
	}
	if definition.Blocks[statusActionAttack] {
		delete(w.combat.pendingTargets, player.id)
	}
	if definition.Blocks[statusActionGather] {
		delete(w.pendingGathering, player.id)
		clearGathering(player, lowerFirst(definition.Name))
	}

	if definition.Blocks[statusActionProduction] && player.brewRemaining > 0 {
		delete(w.pendingBrews, player.id)
		player.brewRemaining = 0
		player.nextBrewPulse = 0
		player.brewingResult = "brewing_stopped"
		player.brewingRevision++
	}
	if definition.Blocks[statusActionProduction] && player.fletchRemaining > 0 {
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
