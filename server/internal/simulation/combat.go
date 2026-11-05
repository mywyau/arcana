package simulation

import (
	"fmt"
	"math"
	"sort"
	"time"
)

const (
	actionPulseInterval           = 600 * time.Millisecond
	swordAttackIntervalPulses     = 4
	swordRange                    = 1.55
	swordApproachRange            = 1.25
	swordDamage                   = 10
	daggerRange                   = 1.3
	daggerApproachRange           = 1.05
	daggerDamage                  = 7
	daggerAttackIntervalPulses    = 3
	bowRange                      = 12.0
	bowApproachRange              = 11.25
	bowDamage                     = 4
	bowAttackIntervalPulses       = 5
	maceRange                     = 1.45
	maceApproachRange             = 1.15
	maceDamage                    = 12
	maceAttackIntervalPulses      = 5
	battleAxeRange                = 1.55
	battleAxeApproachRange        = 1.25
	battleAxeDamage               = 15
	battleAxeAttackIntervalPulses = 6
	staffRange                    = 10.0
	staffApproachRange            = 9.25
	staffDamage                   = 6
	staffAttackIntervalPulses     = 5
	playerMaximumHealth           = 100
	playerMaximumMana             = 100
	playerMaximumPrayer           = 100
	playerDodgeRate               = 0.10
	dummyMaximumHealth            = 100
	dummyDodgeRate                = 0.10
	dummyAttackDamage             = 15
	dummyAttackIntervalPulses     = 4
	dummyAttackRange              = 1.55
	dummyLeashRange               = 6.0
	roamingEnemySpeed             = 3.0
	roamingDetectionRange         = 5.0
	roamingLeashRange             = 8.0
	roamingRepathPulses           = 1
	respawnDelayPulses            = 5
	dummyTargetID                 = "dummy"
	upperDummyTargetID            = "upper_dummy"
	overlookDummyTargetID         = "overlook_dummy"
	hillWardenTargetID            = "hill_warden"
	combatEventHistoryLimit       = 32
)

const (
	targetKindDummy    = "dummy"
	targetKindCreature = "creature"
)

type WeaponID string

const (
	WeaponSword     WeaponID = "sword"
	WeaponDagger    WeaponID = "dagger"
	WeaponBow       WeaponID = "bow"
	WeaponMace      WeaponID = "mace"
	WeaponBattleAxe WeaponID = "battle_axe"
	WeaponStaff     WeaponID = "staff"
)

type CombatStyleID string

const (
	CombatStyleSlash      CombatStyleID = "slash"
	CombatStyleStab       CombatStyleID = "stab"
	CombatStyleRapid      CombatStyleID = "rapid"
	CombatStyleAccurate   CombatStyleID = "accurate"
	CombatStyleLongRange  CombatStyleID = "long_range"
	CombatStyleCrush      CombatStyleID = "crush"
	CombatStyleHeavyCrush CombatStyleID = "heavy_crush"
	CombatStyleDefensive  CombatStyleID = "defensive"
	CombatStyleUnfocused  CombatStyleID = "unfocused"
)

func defaultCombatStyle(weapon WeaponID) CombatStyleID {
	if weapon == WeaponBow || weapon == WeaponStaff {
		return CombatStyleAccurate
	}
	if weapon == WeaponMace {
		return CombatStyleCrush
	}
	if weapon == WeaponDagger {
		return CombatStyleStab
	}
	if weapon == WeaponSword {
		return CombatStyleStab
	}
	return CombatStyleSlash
}

func combatStyleValidForWeapon(style CombatStyleID, weapon WeaponID) bool {
	if style == CombatStyleUnfocused {
		_, exists := weaponDefinition(weapon)
		return exists
	}
	switch weapon {
	case WeaponBow, WeaponStaff:
		return style == CombatStyleRapid || style == CombatStyleAccurate || style == CombatStyleLongRange
	case WeaponMace:
		return style == CombatStyleCrush || style == CombatStyleHeavyCrush || style == CombatStyleDefensive
	case WeaponBattleAxe:
		return style == CombatStyleSlash || style == CombatStyleHeavyCrush || style == CombatStyleDefensive
	case WeaponSword, WeaponDagger:
		return style == CombatStyleStab || style == CombatStyleSlash || style == CombatStyleDefensive
	default:
		return false
	}
}

type WeaponDefinition struct {
	ID             WeaponID
	Damage         int
	AttackInterval uint64
	Range          float64
	ApproachRange  float64
}

func weaponDefinition(id WeaponID) (WeaponDefinition, bool) {
	switch id {
	case WeaponSword:
		return WeaponDefinition{ID: id, Damage: swordDamage, AttackInterval: swordAttackIntervalPulses, Range: swordRange, ApproachRange: swordApproachRange}, true
	case WeaponDagger:
		return WeaponDefinition{ID: id, Damage: daggerDamage, AttackInterval: daggerAttackIntervalPulses, Range: daggerRange, ApproachRange: daggerApproachRange}, true
	case WeaponBow:
		return WeaponDefinition{ID: id, Damage: bowDamage, AttackInterval: bowAttackIntervalPulses, Range: bowRange, ApproachRange: bowApproachRange}, true
	case WeaponMace:
		return WeaponDefinition{ID: id, Damage: maceDamage, AttackInterval: maceAttackIntervalPulses, Range: maceRange, ApproachRange: maceApproachRange}, true
	case WeaponBattleAxe:
		return WeaponDefinition{ID: id, Damage: battleAxeDamage, AttackInterval: battleAxeAttackIntervalPulses, Range: battleAxeRange, ApproachRange: battleAxeApproachRange}, true
	case WeaponStaff:
		return WeaponDefinition{ID: id, Damage: staffDamage, AttackInterval: staffAttackIntervalPulses, Range: staffRange, ApproachRange: staffApproachRange}, true
	default:
		return WeaponDefinition{}, false
	}
}

func weaponDefinitionForStyle(id WeaponID, style CombatStyleID) (WeaponDefinition, bool) {
	weapon, exists := weaponDefinition(id)
	if !exists {
		return WeaponDefinition{}, false
	}
	if id == WeaponBow {
		switch style {
		case CombatStyleRapid:
			weapon.AttackInterval = 4
		case CombatStyleLongRange:
			weapon.Range = 14
			weapon.ApproachRange = 13.25
		}
	}
	return weapon, true
}

type CombatEventType string

const (
	CombatEventDamage CombatEventType = "damage"
	CombatEventDodged CombatEventType = "dodged"
)

// CombatEvent describes an authoritative outcome for client presentation.
type CombatEvent struct {
	ID       uint64
	Pulse    uint64
	Type     CombatEventType
	SourceID string
	TargetID string
	Amount   int
}

type TargetCommand struct {
	PlayerID PlayerID
	Sequence uint64
	TargetID string
}

type EquipCommand struct {
	PlayerID PlayerID
	Sequence uint64
	ItemID   WeaponID
}

type CombatStyleCommand struct {
	PlayerID PlayerID
	Sequence uint64
	Style    CombatStyleID
}

type TargetSnapshot struct {
	ID              string
	Name            string
	Kind            string
	Position        Position
	Health          int
	MaximumHealth   int
	TargetID        PlayerID
	NextAttackPulse uint64
	Defeated        bool
	RespawnAtPulse  uint64
	DodgeRate       float64
}

type combatState struct {
	elapsed          time.Duration
	pulse            uint64
	targets          map[string]*TargetSnapshot
	pendingTargets   map[PlayerID]TargetCommand
	pendingEquipment map[PlayerID]EquipCommand
	pendingStyles    map[PlayerID]CombatStyleCommand
	nextEventID      uint64
	events           []CombatEvent
	roamingEnemies   map[string]*roamingEnemyState
}

type roamingEnemyState struct {
	target          *TargetSnapshot
	home            Position
	path            []Position
	speed           float64
	nextRepathPulse uint64
}

type dummyTarget struct {
	id    string
	dummy *TargetSnapshot
}

func (combat *combatState) dummyTargets() []dummyTarget {
	ids := make([]string, 0, len(combat.targets))
	for id := range combat.targets {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	targets := make([]dummyTarget, 0, len(ids))
	for _, id := range ids {
		targets = append(targets, dummyTarget{id: id, dummy: combat.targets[id]})
	}
	return targets
}

func (combat *combatState) dummyByID(id string) (*TargetSnapshot, bool) {
	target, exists := combat.targets[id]
	return target, exists
}

func newCombatState() *combatState {
	hillWardenHome := Position{X: -12, Y: 2, Z: 10}
	hillWarden := &TargetSnapshot{
		ID:            hillWardenTargetID,
		Name:          "Hill Warden",
		Kind:          targetKindCreature,
		Position:      hillWardenHome,
		Health:        dummyMaximumHealth,
		MaximumHealth: dummyMaximumHealth,
		DodgeRate:     dummyDodgeRate,
	}
	return &combatState{
		targets: map[string]*TargetSnapshot{
			dummyTargetID: {
				ID:            dummyTargetID,
				Name:          "Ground Dummy",
				Kind:          targetKindDummy,
				Position:      Position{X: -4, Z: -2},
				Health:        dummyMaximumHealth,
				MaximumHealth: dummyMaximumHealth,
				DodgeRate:     dummyDodgeRate,
			},
			upperDummyTargetID: {
				ID:            upperDummyTargetID,
				Name:          "Upper Dummy",
				Kind:          targetKindDummy,
				Position:      Position{X: -4, Y: 3, Z: 2},
				Health:        dummyMaximumHealth,
				MaximumHealth: dummyMaximumHealth,
				DodgeRate:     dummyDodgeRate,
			},
			overlookDummyTargetID: {
				ID:            overlookDummyTargetID,
				Name:          "Overlook Dummy",
				Kind:          targetKindDummy,
				Position:      Position{X: 2.75, Y: 0, Z: 3.5},
				Health:        dummyMaximumHealth,
				MaximumHealth: dummyMaximumHealth,
				DodgeRate:     dummyDodgeRate,
			},
			hillWardenTargetID: hillWarden,
		},
		pendingTargets:   make(map[PlayerID]TargetCommand),
		pendingEquipment: make(map[PlayerID]EquipCommand),
		pendingStyles:    make(map[PlayerID]CombatStyleCommand),
		roamingEnemies: map[string]*roamingEnemyState{
			hillWardenTargetID: {
				target: hillWarden,
				home:   hillWardenHome,
				speed:  roamingEnemySpeed,
			},
		},
	}
}

// QueueCombatStyle records a technique selection for the equipped weapon.
// Techniques route combat XP and may also modify weapon behaviour.
func (w *World) QueueCombatStyle(command CombatStyleCommand) error {
	if command.Sequence == 0 {
		return fmt.Errorf("combat style sequence must be positive")
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
	if !combatStyleValidForWeapon(command.Style, player.equippedWeapon) {
		return fmt.Errorf("style %q is invalid for weapon %q", command.Style, player.equippedWeapon)
	}
	if command.Sequence <= player.lastCombatStyleSequence {
		return fmt.Errorf("combat style sequence %d is not newer than %d", command.Sequence, player.lastCombatStyleSequence)
	}
	if pending, found := w.combat.pendingStyles[command.PlayerID]; found && command.Sequence <= pending.Sequence {
		return fmt.Errorf("combat style sequence %d is not newer than pending sequence %d", command.Sequence, pending.Sequence)
	}
	w.combat.pendingStyles[command.PlayerID] = command
	return nil
}

// QueueEquip records an owned weapon or armour item to equip on the next action pulse.
func (w *World) QueueEquip(command EquipCommand) error {
	if command.Sequence == 0 {
		return fmt.Errorf("equipment sequence must be positive")
	}
	_, weaponExists := weaponDefinition(command.ItemID)
	_, armourExists := armourSlot(string(command.ItemID))
	if !weaponExists && !armourExists {
		return fmt.Errorf("item %q does not exist", command.ItemID)
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
	if !player.inventory.contains(string(command.ItemID)) {
		return fmt.Errorf("player %q does not own item %q", command.PlayerID, command.ItemID)
	}
	if command.Sequence <= player.lastEquipmentSequence {
		return fmt.Errorf("equipment sequence %d is not newer than %d", command.Sequence, player.lastEquipmentSequence)
	}
	if pending, exists := w.combat.pendingEquipment[command.PlayerID]; exists && command.Sequence <= pending.Sequence {
		return fmt.Errorf("equipment sequence %d is not newer than pending sequence %d", command.Sequence, pending.Sequence)
	}

	w.combat.pendingEquipment[command.PlayerID] = command
	return nil
}

// QueueTarget records an intention to engage TargetID, or to disengage when it
// is empty. Target changes are applied on the next action pulse.
func (w *World) QueueTarget(command TargetCommand) error {
	if command.Sequence == 0 {
		return fmt.Errorf("action sequence must be positive")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if command.TargetID != "" {
		if _, exists := w.combat.dummyByID(command.TargetID); !exists {
			return fmt.Errorf("target %q does not exist", command.TargetID)
		}
	}

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
	if command.Sequence <= player.lastActionSequence {
		return fmt.Errorf("action sequence %d is not newer than %d", command.Sequence, player.lastActionSequence)
	}
	if pending, exists := w.combat.pendingTargets[command.PlayerID]; exists && command.Sequence <= pending.Sequence {
		return fmt.Errorf("action sequence %d is not newer than pending sequence %d", command.Sequence, pending.Sequence)
	}

	w.combat.pendingTargets[command.PlayerID] = command
	return nil
}

func (w *World) advanceCombat(elapsed time.Duration) {
	w.combat.elapsed += elapsed
	for w.combat.elapsed >= actionPulseInterval {
		w.combat.elapsed -= actionPulseInterval
		w.combat.pulse++
		w.respawnDefeatedEntities()
		w.applyEquipmentCommands()
		w.applyCombatStyleCommands()
		w.applyTargetCommands()
		w.resolveAutoAttacks()
		w.updateRoamingEnemies()
		w.resolveDummyAttack()
	}
}

func (w *World) applyEquipmentCommands() {
	playerIDs := make([]PlayerID, 0, len(w.combat.pendingEquipment))
	for id := range w.combat.pendingEquipment {
		playerIDs = append(playerIDs, id)
	}
	sort.Slice(playerIDs, func(i, j int) bool { return playerIDs[i] < playerIDs[j] })
	for _, id := range playerIDs {
		command := w.combat.pendingEquipment[id]
		if player, exists := w.players[id]; exists && !player.defeated {
			if slot, armourExists := armourSlot(string(command.ItemID)); armourExists {
				if player.equippedArmour[slot] == string(command.ItemID) {
					delete(player.equippedArmour, slot)
				} else {
					player.equippedArmour[slot] = string(command.ItemID)
				}
			} else {
				player.equippedWeapon = command.ItemID
				player.combatStyle = defaultCombatStyle(command.ItemID)
			}
			player.lastEquipmentSequence = command.Sequence
			player.lastEquipmentPulse = w.combat.pulse
		}
		delete(w.combat.pendingEquipment, id)
	}
}

func (w *World) applyCombatStyleCommands() {
	playerIDs := make([]PlayerID, 0, len(w.combat.pendingStyles))
	for id := range w.combat.pendingStyles {
		playerIDs = append(playerIDs, id)
	}
	sort.Slice(playerIDs, func(i, j int) bool { return playerIDs[i] < playerIDs[j] })
	for _, id := range playerIDs {
		command := w.combat.pendingStyles[id]
		if player, exists := w.players[id]; exists && !player.defeated {
			if combatStyleValidForWeapon(command.Style, player.equippedWeapon) {
				player.combatStyle = command.Style
			}
			player.lastCombatStyleSequence = command.Sequence
		}
		delete(w.combat.pendingStyles, id)
	}
}

func (w *World) respawnDefeatedEntities() {
	for _, target := range w.combat.dummyTargets() {
		if target.dummy.Defeated && w.combat.pulse >= target.dummy.RespawnAtPulse {
			target.dummy.Health = target.dummy.MaximumHealth
			target.dummy.Defeated = false
			target.dummy.RespawnAtPulse = 0
			if enemy, roaming := w.combat.roamingEnemies[target.id]; roaming {
				target.dummy.Position = enemy.home
				enemy.path = nil
				enemy.nextRepathPulse = 0
			}
		}
	}
	for _, player := range w.players {
		if !player.defeated || w.combat.pulse < player.respawnAtPulse {
			continue
		}
		player.position = player.spawnPosition
		player.health = player.maximumHealth
		player.defeated = false
		player.respawnAtPulse = 0
		player.lastActionPulse = w.combat.pulse
		player.lastActionResult = "respawned"
		if location, err := w.locate(player.position); err == nil {
			player.navigationLocation = location
		}
	}
}

func (w *World) applyTargetCommands() {
	playerIDs := make([]PlayerID, 0, len(w.combat.pendingTargets))
	for id := range w.combat.pendingTargets {
		playerIDs = append(playerIDs, id)
	}
	sort.Slice(playerIDs, func(i, j int) bool { return playerIDs[i] < playerIDs[j] })

	for _, id := range playerIDs {
		command := w.combat.pendingTargets[id]
		player, exists := w.players[id]
		if !exists || player.defeated {
			delete(w.combat.pendingTargets, id)
			continue
		}

		player.lastActionSequence = command.Sequence
		player.lastActionPulse = w.combat.pulse
		player.combatTarget = command.TargetID
		if command.TargetID == "" {
			player.lastActionResult = "disengaged"
			if player.autoApproaching {
				stopPlayer(player)
			}
		} else {
			player.lastActionResult = "engaged"
		}
		player.autoApproaching = false
		delete(w.combat.pendingTargets, id)
	}
}

func (w *World) resolveAutoAttacks() {
	playerIDs := make([]PlayerID, 0, len(w.players))
	for id := range w.players {
		playerIDs = append(playerIDs, id)
	}
	sort.Slice(playerIDs, func(i, j int) bool { return playerIDs[i] < playerIDs[j] })

	for _, id := range playerIDs {
		player := w.players[id]
		if player.defeated || player.combatTarget == "" {
			continue
		}
		target, exists := w.combat.dummyByID(player.combatTarget)
		if !exists {
			w.clearPlayerTarget(player, "target_missing")
			continue
		}
		if target.Defeated {
			w.clearPlayerTarget(player, "target_defeated")
			continue
		}
		weapon, _ := weaponDefinitionForStyle(player.equippedWeapon, player.combatStyle)
		if distance(player.position, target.Position) > weapon.Range {
			if err := w.approachDummy(player, target.Position, weapon.ApproachRange); err != nil {
				player.lastActionResult = "target_unreachable"
				player.autoApproaching = false
			} else {
				player.lastActionResult = "approaching"
			}
			continue
		}
		if player.autoApproaching {
			stopPlayer(player)
			player.autoApproaching = false
		}
		if w.combat.pulse < player.nextAttackPulse {
			continue
		}

		player.lastActionPulse = w.combat.pulse
		player.nextAttackPulse = w.combat.pulse + weapon.AttackInterval
		if w.rollSource.Float64() < target.DodgeRate {
			player.lastActionResult = "dodged"
			w.combat.recordEvent(CombatEvent{
				Pulse:    w.combat.pulse,
				Type:     CombatEventDodged,
				SourceID: string(player.id),
				TargetID: player.combatTarget,
			})
			continue
		}

		damageDealt := min(weapon.Damage, target.Health)
		target.Health -= damageDealt
		if skill, trains := skillForCombatStyle(player.equippedWeapon, player.combatStyle); trains {
			grantSkillExperience(player, skill, int64(damageDealt))
		}
		player.lastActionResult = "hit"
		w.combat.recordEvent(CombatEvent{
			Pulse:    w.combat.pulse,
			Type:     CombatEventDamage,
			SourceID: string(player.id),
			TargetID: player.combatTarget,
			Amount:   damageDealt,
		})
		if target.TargetID == "" {
			target.TargetID = player.id
			target.NextAttackPulse = w.combat.pulse + dummyAttackIntervalPulses
		}
		if target.Health == 0 {
			w.defeatDummy(player.combatTarget, target)
		}
	}
}

func (w *World) resolveDummyAttack() {
	for _, target := range w.combat.dummyTargets() {
		w.resolveOneDummyAttack(target.id, target.dummy)
	}
}

func (w *World) updateRoamingEnemies() {
	ids := make([]string, 0, len(w.combat.roamingEnemies))
	for id := range w.combat.roamingEnemies {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		enemy := w.combat.roamingEnemies[id]
		if enemy.target.Defeated {
			continue
		}

		hadTarget := enemy.target.TargetID != ""
		player, hasTarget := w.players[enemy.target.TargetID]
		if hadTarget && (!hasTarget || player.defeated || distance(player.position, enemy.home) > roamingLeashRange) {
			clearDummyTarget(enemy.target)
			enemy.path = nil
			hasTarget = false
		}
		if !hasTarget {
			player = w.nearestDetectablePlayer(enemy)
			if player != nil {
				enemy.target.TargetID = player.id
				enemy.target.NextAttackPulse = w.combat.pulse + dummyAttackIntervalPulses
				hasTarget = true
			}
		}

		if hasTarget {
			if distance(enemy.target.Position, player.position) <= dummyAttackRange {
				enemy.path = nil
				continue
			}
			if w.combat.pulse >= enemy.nextRepathPulse {
				if err := w.setRoamingPath(enemy, player.position); err != nil {
					clearDummyTarget(enemy.target)
					enemy.path = nil
					w.setRoamingReturnPath(enemy)
					continue
				}
				enemy.nextRepathPulse = w.combat.pulse + roamingRepathPulses
			}
			continue
		}

		w.setRoamingReturnPath(enemy)
	}
}

func (w *World) nearestDetectablePlayer(enemy *roamingEnemyState) *playerState {
	var nearest *playerState
	nearestDistance := math.Inf(1)
	for _, player := range w.players {
		if player.defeated || distance(player.position, enemy.home) > roamingLeashRange {
			continue
		}
		candidateDistance := distance(enemy.target.Position, player.position)
		if candidateDistance > roamingDetectionRange {
			continue
		}
		if candidateDistance < nearestDistance ||
			(candidateDistance == nearestDistance && (nearest == nil || player.id < nearest.id)) {
			nearest = player
			nearestDistance = candidateDistance
		}
	}
	return nearest
}

func (w *World) setRoamingPath(enemy *roamingEnemyState, destination Position) error {
	path := []Position{destination}
	if w.pathfinder != nil {
		var err error
		path, err = w.pathfinder.FindPath(enemy.target.Position, destination)
		if err != nil {
			return err
		}
	}
	if len(path) == 0 {
		return fmt.Errorf("empty roaming path")
	}
	enemy.path = clonePositions(path)
	return nil
}

func (w *World) setRoamingReturnPath(enemy *roamingEnemyState) {
	if len(enemy.path) > 0 || distance(enemy.target.Position, enemy.home) <= 0.01 {
		return
	}
	_ = w.setRoamingPath(enemy, enemy.home)
}

func (w *World) resolveOneDummyAttack(dummyID string, dummy *TargetSnapshot) {
	if dummy.Defeated || dummy.TargetID == "" {
		return
	}
	player, exists := w.players[dummy.TargetID]
	if !exists || player.defeated {
		clearDummyTarget(dummy)
		return
	}
	distanceToPlayer := distance(dummy.Position, player.position)
	_, roaming := w.combat.roamingEnemies[dummyID]
	if (!roaming && distanceToPlayer > dummyLeashRange) ||
		(roaming && distance(player.position, w.combat.roamingEnemies[dummyID].home) > roamingLeashRange) {
		clearDummyTarget(dummy)
		return
	}
	if distanceToPlayer > dummyAttackRange || w.combat.pulse < dummy.NextAttackPulse {
		return
	}

	dummy.NextAttackPulse = w.combat.pulse + dummyAttackIntervalPulses
	if w.rollSource.Float64() < player.dodgeRate {
		w.combat.recordEvent(CombatEvent{
			Pulse:    w.combat.pulse,
			Type:     CombatEventDodged,
			SourceID: dummyID,
			TargetID: string(player.id),
		})
		return
	}

	player.health = max(0, player.health-dummyAttackDamage)
	w.combat.recordEvent(CombatEvent{
		Pulse:    w.combat.pulse,
		Type:     CombatEventDamage,
		SourceID: dummyID,
		TargetID: string(player.id),
		Amount:   dummyAttackDamage,
	})
	if player.health == 0 {
		w.defeatPlayer(player)
	}
}

func (w *World) defeatPlayer(player *playerState) {
	player.defeated = true
	player.respawnAtPulse = w.combat.pulse + respawnDelayPulses
	player.lastActionPulse = w.combat.pulse
	player.lastActionResult = "defeated"
	player.combatTarget = ""
	player.nextAttackPulse = 0
	player.autoApproaching = false
	stopPlayer(player)
	delete(w.pendingMoves, player.id)
	delete(w.combat.pendingTargets, player.id)
	delete(w.combat.pendingEquipment, player.id)
	delete(w.combat.pendingStyles, player.id)
	for _, target := range w.combat.dummyTargets() {
		if target.dummy.TargetID == player.id {
			clearDummyTarget(target.dummy)
		}
	}
}

func (w *World) defeatDummy(dummyID string, dummy *TargetSnapshot) {
	dummy.Defeated = true
	dummy.RespawnAtPulse = w.combat.pulse + respawnDelayPulses
	clearDummyTarget(dummy)
	if enemy, roaming := w.combat.roamingEnemies[dummyID]; roaming {
		enemy.path = nil
		enemy.nextRepathPulse = 0
	}
	itemID := trainingShardItem
	if dummyID == upperDummyTargetID {
		itemID = ItemTrainingHelm
	} else if dummyID == hillWardenTargetID {
		itemID = ItemTrainingPlate
	}
	w.spawnDrop(dummy.Position, itemID, 1)
	for _, player := range w.players {
		if player.combatTarget == dummyID {
			w.clearPlayerTarget(player, "target_defeated")
		}
	}
}

func (w *World) clearPlayerTarget(player *playerState, result string) {
	player.combatTarget = ""
	player.lastActionPulse = w.combat.pulse
	player.lastActionResult = result
	if player.autoApproaching {
		stopPlayer(player)
		player.autoApproaching = false
	}
}

func clearDummyTarget(dummy *TargetSnapshot) {
	dummy.TargetID = ""
	dummy.NextAttackPulse = 0
}

func (combat *combatState) recordEvent(event CombatEvent) {
	combat.nextEventID++
	event.ID = combat.nextEventID
	combat.events = append(combat.events, event)
	if len(combat.events) > combatEventHistoryLimit {
		combat.events = append([]CombatEvent(nil), combat.events[len(combat.events)-combatEventHistoryLimit:]...)
	}
}

func (w *World) approachDummy(player *playerState, target Position, approachRange float64) error {
	dx := player.position.X - target.X
	dz := player.position.Z - target.Z
	length := math.Hypot(dx, dz)
	if length == 0 {
		dx = 1
		length = 1
	}
	destination := Position{
		X: target.X + dx/length*approachRange,
		Y: target.Y,
		Z: target.Z + dz/length*approachRange,
	}
	path := []Position{destination}
	if w.pathfinder != nil {
		var err error
		path, err = w.pathfinder.FindPath(player.position, destination)
		if err != nil {
			return fmt.Errorf("find combat path: %w", err)
		}
	}
	if len(path) == 0 {
		return fmt.Errorf("find combat path: empty path")
	}

	player.destination = positionPointer(destination)
	player.path = clonePositions(path)
	player.autoApproaching = true
	return nil
}

func stopPlayer(player *playerState) {
	player.destination = nil
	player.path = nil
}

func distance(a, b Position) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	dz := a.Z - b.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}
