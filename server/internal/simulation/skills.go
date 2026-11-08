package simulation

import (
	"fmt"
	"sort"
)

const (
	SkillAttack      SkillID = "attack"
	SkillAgility     SkillID = "agility"
	SkillDefence     SkillID = "defence"
	SkillMagic       SkillID = "magic"
	SkillRange       SkillID = "range"
	SkillStrength    SkillID = "strength"
	SkillFishing     SkillID = "fishing"
	SkillCooking     SkillID = "cooking"
	SkillMining      SkillID = "mining"
	SkillSmithing    SkillID = "smithing"
	SkillWoodcutting SkillID = "woodcutting"
	SkillHerblore    SkillID = "herblore"
	SkillFletching   SkillID = "fletching"
	SkillPrayer      SkillID = "prayer"

	MaximumSkillLevel               = 100
	skillLevelExperienceScale int64 = 100
	agilitySpeedPerLevel            = 0.0025
	maximumAgilitySpeedBonus        = 0.25
)

type SkillID string

type SkillSnapshot struct {
	ID                     SkillID
	Experience             int64
	Level                  int
	CurrentLevelExperience int64
	NextLevelExperience    int64
}

type ExperienceLockCommand struct {
	PlayerID PlayerID
	Sequence uint64
	Locked   bool
}

// PersistentSkillState stores only total experience. Levels are derived so a
// future curve adjustment does not require rewriting every character.
type PersistentSkillState struct {
	ID         SkillID
	Experience int64
}

type skillState map[SkillID]int64

func newDefaultSkills() skillState {
	return skillState{
		SkillAttack: 0, SkillAgility: 0, SkillDefence: 0, SkillMagic: 0, SkillRange: 0, SkillStrength: 0,
		SkillFishing: 0, SkillCooking: 0, SkillMining: 0, SkillSmithing: 0, SkillWoodcutting: 0, SkillHerblore: 0, SkillFletching: 0, SkillPrayer: 0,
	}
}

func skillsFromPersistent(states []PersistentSkillState) (skillState, error) {
	skills := newDefaultSkills()
	seen := make(map[SkillID]struct{}, len(states))
	for _, state := range states {
		if !skillIDValid(state.ID) {
			return nil, fmt.Errorf("unknown skill %q", state.ID)
		}
		if state.Experience < 0 {
			return nil, fmt.Errorf("skill %q experience cannot be negative", state.ID)
		}
		if _, exists := seen[state.ID]; exists {
			return nil, fmt.Errorf("skill %q is duplicated", state.ID)
		}
		seen[state.ID] = struct{}{}
		skills[state.ID] = state.Experience
	}
	return skills, nil
}

func persistentSkills(skills skillState) []PersistentSkillState {
	ids := make([]SkillID, 0, len(skills))
	for id := range skills {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	states := make([]PersistentSkillState, 0, len(ids))
	for _, id := range ids {
		states = append(states, PersistentSkillState{ID: id, Experience: skills[id]})
	}
	return states
}

func snapshotSkills(skills skillState) []SkillSnapshot {
	states := persistentSkills(skills)
	snapshots := make([]SkillSnapshot, 0, len(states))
	for _, state := range states {
		level := SkillLevelForExperience(state.Experience)
		nextExperience := int64(0)
		if level < MaximumSkillLevel {
			nextExperience = SkillExperienceForLevel(level + 1)
		}
		snapshots = append(snapshots, SkillSnapshot{
			ID:                     state.ID,
			Experience:             state.Experience,
			Level:                  level,
			CurrentLevelExperience: SkillExperienceForLevel(level),
			NextLevelExperience:    nextExperience,
		})
	}
	return snapshots
}

// SkillExperienceForLevel uses a deliberately simple prototype curve:
// level N starts at 100 × (N-1)² total XP.
func SkillExperienceForLevel(level int) int64 {
	if level <= 1 {
		return 0
	}
	if level > MaximumSkillLevel {
		level = MaximumSkillLevel
	}
	offset := int64(level - 1)
	return skillLevelExperienceScale * offset * offset
}

func SkillLevelForExperience(experience int64) int {
	if experience <= 0 {
		return 1
	}
	level := 1
	for level < MaximumSkillLevel && experience >= SkillExperienceForLevel(level+1) {
		level++
	}
	return level
}

func grantSkillExperience(player *playerState, skill SkillID, amount int64) {
	if player == nil || player.experienceLocked || amount <= 0 {
		return
	}
	if player.skills == nil {
		player.skills = newDefaultSkills()
	}
	player.skills[skill] += amount
	if skill == SkillAgility {
		player.speed = movementSpeedForAgility(player.baseSpeed, SkillLevelForExperience(player.skills[SkillAgility]))
	}
}

func (w *World) QueueExperienceLock(command ExperienceLockCommand) error {
	if command.Sequence == 0 {
		return fmt.Errorf("experience lock sequence must be positive")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	player, exists := w.players[command.PlayerID]
	if !exists {
		player, exists = w.pendingJoins[command.PlayerID]
	}
	if !exists {
		return fmt.Errorf("player %q does not exist", command.PlayerID)
	}
	if command.Sequence <= player.lastExperienceLockSequence {
		return fmt.Errorf("experience lock sequence %d is not newer than %d", command.Sequence, player.lastExperienceLockSequence)
	}
	if pending, exists := w.pendingExperienceLocks[command.PlayerID]; exists && command.Sequence <= pending.Sequence {
		return fmt.Errorf("experience lock sequence %d is not newer than pending sequence %d", command.Sequence, pending.Sequence)
	}
	w.pendingExperienceLocks[command.PlayerID] = command
	return nil
}

func (w *World) applyPendingExperienceLocks() {
	for id, command := range w.pendingExperienceLocks {
		if player, exists := w.players[id]; exists {
			player.experienceLocked = command.Locked
			player.lastExperienceLockSequence = command.Sequence
		}
		delete(w.pendingExperienceLocks, id)
	}
}

func movementSpeedForAgility(baseSpeed float64, agilityLevel int) float64 {
	bonus := min(maximumAgilitySpeedBonus, float64(max(1, agilityLevel)-1)*agilitySpeedPerLevel)
	return baseSpeed * (1 + bonus)
}

func skillIDValid(skill SkillID) bool {
	return skill == SkillAttack || skill == SkillAgility || skill == SkillStrength || skill == SkillDefence ||
		skill == SkillRange || skill == SkillMagic || skill == SkillFishing || skill == SkillCooking || skill == SkillMining ||
		skill == SkillSmithing || skill == SkillWoodcutting || skill == SkillHerblore || skill == SkillFletching || skill == SkillPrayer
}

func skillForCombatStyle(weapon WeaponID, style CombatStyleID) (SkillID, bool) {
	switch weapon {
	case WeaponBow, WeaponCrossbow:
		return SkillRange, true
	case WeaponStaff:
		return SkillMagic, true
	case WeaponSword, WeaponBronzeSword, WeaponSteelSword, WeaponMithrilSword, WeaponRuneSword, WeaponDagger, WeaponSpear, WeaponHalberd:
		switch style {
		case CombatStyleStab:
			return SkillAttack, true
		case CombatStyleSlash:
			return SkillStrength, true
		case CombatStyleDefensive:
			return SkillDefence, true
		}
	case WeaponMace:
		switch style {
		case CombatStyleCrush:
			return SkillAttack, true
		case CombatStyleHeavyCrush:
			return SkillStrength, true
		case CombatStyleDefensive:
			return SkillDefence, true
		}
	case WeaponUnarmed:
		switch style {
		case CombatStyleCrush:
			return SkillAttack, true
		case CombatStyleHeavyCrush:
			return SkillStrength, true
		case CombatStyleDefensive:
			return SkillDefence, true
		}
	case WeaponBattleAxe:
		switch style {
		case CombatStyleSlash:
			return SkillAttack, true
		case CombatStyleHeavyCrush:
			return SkillStrength, true
		case CombatStyleDefensive:
			return SkillDefence, true
		}
	default:
	}
	return "", false
}
