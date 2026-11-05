package simulation

import "fmt"

// PersistentPlayerState is the durable subset of an active player's state.
// Transient movement, targeting, cooldown, and sequence state starts fresh on join.
type PersistentPlayerState struct {
	Position       Position
	Health         int
	EquippedWeapon WeaponID
	CombatStyle    CombatStyleID
	Inventory      InventorySnapshot
	Skills         []PersistentSkillState
}

// DefaultPersistentPlayerState is used for a newly created character.
func DefaultPersistentPlayerState() PersistentPlayerState {
	return PersistentPlayerState{
		Health:         playerMaximumHealth,
		EquippedWeapon: WeaponSword,
		CombatStyle:    defaultCombatStyle(WeaponSword),
		Inventory:      snapshotInventory(newStarterInventory()),
		Skills:         persistentSkills(newDefaultSkills()),
	}
}

// PersistentStateFromSnapshot extracts the state that survives a disconnect.
func PersistentStateFromSnapshot(snapshot PlayerSnapshot) PersistentPlayerState {
	state := PersistentPlayerState{
		Position:       snapshot.Position,
		Health:         snapshot.Health,
		EquippedWeapon: snapshot.EquippedWeapon,
		CombatStyle:    snapshot.CombatStyle,
		Inventory: InventorySnapshot{
			Capacity: snapshot.Inventory.Capacity,
			Items:    append([]InventoryItem(nil), snapshot.Inventory.Items...),
		},
		Skills: make([]PersistentSkillState, 0, len(snapshot.Skills)),
	}
	for _, skill := range snapshot.Skills {
		state.Skills = append(state.Skills, PersistentSkillState{ID: skill.ID, Experience: skill.Experience})
	}
	// A defeated character reconnects at the normal spawn rather than restoring
	// an incomplete respawn timer from the previous server process.
	if snapshot.Defeated || state.Health <= 0 {
		state.Position = Position{}
		state.Health = playerMaximumHealth
	}
	return state
}

// QueueJoinWithPersistentState validates durable state and adds the character
// on the next simulation step.
func (w *World) QueueJoinWithPersistentState(id PlayerID, state PersistentPlayerState, speed float64) error {
	if err := validatePlayer(id, state.Position, speed); err != nil {
		return err
	}
	if state.Health <= 0 || state.Health > playerMaximumHealth {
		return fmt.Errorf("player health must be between 1 and %d", playerMaximumHealth)
	}
	if _, exists := weaponDefinition(state.EquippedWeapon); !exists {
		return fmt.Errorf("unknown equipped weapon %q", state.EquippedWeapon)
	}
	if !combatStyleValidForWeapon(state.CombatStyle, state.EquippedWeapon) {
		return fmt.Errorf("combat style %q is invalid for %q", state.CombatStyle, state.EquippedWeapon)
	}
	inventory, err := inventoryFromSnapshot(state.Inventory)
	if err != nil {
		return fmt.Errorf("restore inventory: %w", err)
	}
	if !inventory.contains(string(state.EquippedWeapon)) {
		return fmt.Errorf("equipped weapon %q is not in inventory", state.EquippedWeapon)
	}
	skills, err := skillsFromPersistent(state.Skills)
	if err != nil {
		return fmt.Errorf("restore skills: %w", err)
	}
	location, err := w.locate(state.Position)
	if err != nil {
		return fmt.Errorf("locate player: %w", err)
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	if _, exists := w.players[id]; exists {
		return fmt.Errorf("player %q already exists", id)
	}
	if _, exists := w.pendingJoins[id]; exists {
		return fmt.Errorf("player %q is already waiting to join", id)
	}
	w.pendingJoins[id] = &playerState{
		id: id, position: state.Position, spawnPosition: Position{}, speed: speed,
		navigationLocation: location, health: state.Health, maximumHealth: playerMaximumHealth,
		equippedWeapon: state.EquippedWeapon, dodgeRate: playerDodgeRate, inventory: inventory,
		combatStyle: state.CombatStyle,
		skills:      skills,
	}
	return nil
}
