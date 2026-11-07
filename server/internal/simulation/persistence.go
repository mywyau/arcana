package simulation

import "fmt"

// PersistentPlayerState is the durable subset of an active player's state.
// Transient movement, targeting, cooldown, and sequence state starts fresh on join.
type PersistentPlayerState struct {
	Position                 Position
	Health                   int
	EquippedWeapon           WeaponID
	EquippedWeaponInstanceID string
	EquippedArmour           map[ArmourSlot]string
	EquippedArmourInstances  map[ArmourSlot]string
	EquippedSpellbook        string
	CombatStyle              CombatStyleID
	Inventory                InventorySnapshot
	Skills                   []PersistentSkillState
}

// DefaultPersistentPlayerState is used for a newly created character.
func DefaultPersistentPlayerState() PersistentPlayerState {
	return PersistentPlayerState{
		Health:                  playerMaximumHealth,
		EquippedWeapon:          WeaponUnarmed,
		EquippedArmour:          make(map[ArmourSlot]string),
		EquippedArmourInstances: make(map[ArmourSlot]string),
		CombatStyle:             defaultCombatStyle(WeaponUnarmed),
		Inventory:               snapshotInventory(newEmptyInventory()),
		Skills:                  persistentSkills(newDefaultSkills()),
	}
}

// PersistentStateFromSnapshot extracts the state that survives a disconnect.
func PersistentStateFromSnapshot(snapshot PlayerSnapshot) PersistentPlayerState {
	state := PersistentPlayerState{
		Position:                 snapshot.Position,
		Health:                   snapshot.Health,
		EquippedWeapon:           snapshot.EquippedWeapon,
		EquippedWeaponInstanceID: snapshot.EquippedWeaponInstanceID,
		EquippedArmour:           cloneEquippedArmour(snapshot.EquippedArmour),
		EquippedArmourInstances:  cloneEquippedArmourInstances(snapshot.EquippedArmourInstances),
		EquippedSpellbook:        snapshot.EquippedSpellbook,
		CombatStyle:              snapshot.CombatStyle,
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
	if state.EquippedWeapon != WeaponUnarmed && !inventory.contains(string(state.EquippedWeapon)) {
		return fmt.Errorf("equipped weapon %q is not in inventory", state.EquippedWeapon)
	}
	if state.EquippedWeapon != WeaponUnarmed {
		state.EquippedWeaponInstanceID = resolveOwnedInstanceID(inventory, string(state.EquippedWeapon), state.EquippedWeaponInstanceID)
		if state.EquippedWeaponInstanceID == "" {
			return fmt.Errorf("equipped weapon %q has no owned instance", state.EquippedWeapon)
		}
	}
	if err := validateEquippedArmour(state.EquippedArmour, inventory); err != nil {
		return fmt.Errorf("restore equipment: %w", err)
	}
	if state.EquippedSpellbook != "" {
		item, exists := itemDefinition(state.EquippedSpellbook)
		if !exists || item.Kind != ItemKindSpellbook || !inventory.contains(state.EquippedSpellbook) {
			return fmt.Errorf("restore spellbook: item %q is not an owned spellbook", state.EquippedSpellbook)
		}
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
		prayer: playerMaximumPrayer, maximumPrayer: playerMaximumPrayer,
		activePrayers: make(map[PrayerID]bool), prayerDrainProgress: make(map[PrayerID]uint64),
		equippedWeapon: state.EquippedWeapon, equippedWeaponInstanceID: state.EquippedWeaponInstanceID, equippedSpellbook: state.EquippedSpellbook,
		dodgeRate: playerDodgeRate, inventory: inventory,
		equippedArmour:          cloneEquippedArmour(state.EquippedArmour),
		equippedArmourInstances: resolveArmourInstanceIDs(inventory, state.EquippedArmour, state.EquippedArmourInstances),
		combatStyle:             state.CombatStyle,
		skills:                  skills,
	}
	return nil
}

func resolveOwnedInstanceID(inventory inventoryState, itemID, preferred string) string {
	if item, exists := inventory.itemByInstance(preferred); exists && item.ItemID == itemID {
		return preferred
	}
	for _, item := range inventory.items {
		if item.ItemID == itemID && item.InstanceID != "" {
			return item.InstanceID
		}
	}
	return ""
}

func resolveArmourInstanceIDs(inventory inventoryState, armour map[ArmourSlot]string, preferred map[ArmourSlot]string) map[ArmourSlot]string {
	resolved := make(map[ArmourSlot]string, len(armour))
	for slot, itemID := range armour {
		resolved[slot] = resolveOwnedInstanceID(inventory, itemID, preferred[slot])
	}
	return resolved
}
