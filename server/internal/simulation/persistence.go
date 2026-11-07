package simulation

import "fmt"

// PersistentPlayerState is the durable subset of an active player's state.
// Transient movement, targeting, cooldown, and sequence state starts fresh on join.
type PersistentPlayerState struct {
	Position                     Position
	Health                       int
	EquippedWeapon               WeaponID
	EquippedWeaponInstanceID     string
	EquippedWeaponReinforcement  int
	EquippedArmour               map[ArmourSlot]string
	EquippedArmourInstances      map[ArmourSlot]string
	EquippedArmourReinforcements map[ArmourSlot]int
	EquippedSpellbook            string
	CombatStyle                  CombatStyleID
	Inventory                    InventorySnapshot
	AmmunitionPouch              AmmunitionPouchSnapshot
	Bank                         BankSnapshot
	Skills                       []PersistentSkillState
}

// DefaultPersistentPlayerState is used for a newly created character.
func DefaultPersistentPlayerState() PersistentPlayerState {
	return PersistentPlayerState{
		Health:                       playerMaximumHealth,
		EquippedWeapon:               WeaponUnarmed,
		EquippedArmour:               make(map[ArmourSlot]string),
		EquippedArmourInstances:      make(map[ArmourSlot]string),
		EquippedArmourReinforcements: make(map[ArmourSlot]int),
		CombatStyle:                  defaultCombatStyle(WeaponUnarmed),
		Inventory:                    snapshotInventory(newEmptyInventory()),
		AmmunitionPouch:              snapshotAmmunitionPouch(newAmmunitionPouch()),
		Bank:                         snapshotBank(newEmptyBank()),
		Skills:                       persistentSkills(newDefaultSkills()),
	}
}

// PersistentStateFromSnapshot extracts the state that survives a disconnect.
func PersistentStateFromSnapshot(snapshot PlayerSnapshot) PersistentPlayerState {
	state := PersistentPlayerState{
		Position:                     snapshot.Position,
		Health:                       snapshot.Health,
		EquippedWeapon:               snapshot.EquippedWeapon,
		EquippedWeaponInstanceID:     snapshot.EquippedWeaponInstanceID,
		EquippedWeaponReinforcement:  snapshot.EquippedWeaponReinforcement,
		EquippedArmour:               cloneEquippedArmour(snapshot.EquippedArmour),
		EquippedArmourInstances:      cloneEquippedArmourInstances(snapshot.EquippedArmourInstances),
		EquippedArmourReinforcements: cloneEquippedArmourReinforcements(snapshot.EquippedArmourReinforcements),
		EquippedSpellbook:            snapshot.EquippedSpellbook,
		CombatStyle:                  snapshot.CombatStyle,
		Inventory: InventorySnapshot{
			Capacity: snapshot.Inventory.Capacity,
			Items:    append([]InventoryItem(nil), snapshot.Inventory.Items...),
		},
		AmmunitionPouch: AmmunitionPouchSnapshot{
			Capacity: snapshot.AmmunitionPouch.Capacity, UnlockedSlots: snapshot.AmmunitionPouch.UnlockedSlots,
			Items: append([]InventoryItem(nil), snapshot.AmmunitionPouch.Items...),
		},
		Bank: BankSnapshot{
			Capacity: snapshot.Bank.Capacity,
			Items:    append([]InventoryItem(nil), snapshot.Bank.Items...),
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
		state.Health = max(playerMaximumHealth, snapshot.MaximumHealth)
	}
	return state
}

// QueueJoinWithPersistentState validates durable state and adds the character
// on the next simulation step.
func (w *World) QueueJoinWithPersistentState(id PlayerID, state PersistentPlayerState, speed float64) error {
	if err := validatePlayer(id, state.Position, speed); err != nil {
		return err
	}
	if state.Health <= 0 {
		return fmt.Errorf("player health must be positive")
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
	ammunitionPouch, err := ammunitionPouchFromSnapshot(state.AmmunitionPouch)
	if err != nil {
		return fmt.Errorf("restore ammunition pouch: %w", err)
	}
	bank, err := bankFromSnapshot(state.Bank)
	if err != nil {
		return fmt.Errorf("restore bank: %w", err)
	}
	if state.EquippedWeapon != WeaponUnarmed {
		state.EquippedWeaponInstanceID, state.EquippedWeaponReinforcement = extractLegacyEquippedInstance(
			&inventory, string(state.EquippedWeapon), state.EquippedWeaponInstanceID, state.EquippedWeaponReinforcement,
		)
		if state.EquippedWeaponInstanceID == "" {
			return fmt.Errorf("equipped weapon %q has no owned instance", state.EquippedWeapon)
		}
	}
	if state.EquippedArmourReinforcements == nil {
		state.EquippedArmourReinforcements = make(map[ArmourSlot]int)
	}
	if state.EquippedArmourInstances == nil {
		state.EquippedArmourInstances = make(map[ArmourSlot]string)
	}
	for slot, itemID := range state.EquippedArmour {
		definedSlot, exists := armourSlot(itemID)
		if !exists || definedSlot != slot {
			return fmt.Errorf("restore equipment: armour item %q is invalid for slot %q", itemID, slot)
		}
		instanceID, reinforcement := extractLegacyEquippedInstance(
			&inventory, itemID, state.EquippedArmourInstances[slot], state.EquippedArmourReinforcements[slot],
		)
		if instanceID == "" {
			return fmt.Errorf("restore equipment: armour %q has no owned instance", itemID)
		}
		state.EquippedArmourInstances[slot] = instanceID
		state.EquippedArmourReinforcements[slot] = reinforcement
	}
	maximumHealth := playerMaximumHealth + equippedHealthBonus(state.EquippedArmour)
	if state.Health > maximumHealth {
		return fmt.Errorf("player health must not exceed equipped maximum of %d", maximumHealth)
	}
	if state.EquippedSpellbook != "" {
		item, exists := itemDefinition(state.EquippedSpellbook)
		if !exists || item.Kind != ItemKindSpellbook {
			return fmt.Errorf("restore spellbook: item %q is not an owned spellbook", state.EquippedSpellbook)
		}
		// Older saves kept equipped spellbooks in inventory. Normalize them when loaded.
		inventory.remove(state.EquippedSpellbook, 1)
	}
	ownedInstances := make(map[string]string)
	for _, item := range inventory.items {
		if err := recordOwnedInstance(ownedInstances, item.InstanceID, "inventory"); err != nil {
			return fmt.Errorf("restore inventory: %w", err)
		}
	}
	for _, item := range bank.items {
		if err := recordOwnedInstance(ownedInstances, item.InstanceID, "bank"); err != nil {
			return fmt.Errorf("restore bank: %w", err)
		}
	}
	if err := recordOwnedInstance(ownedInstances, state.EquippedWeaponInstanceID, "equipped weapon"); err != nil {
		return fmt.Errorf("restore equipment: %w", err)
	}
	for slot, instanceID := range state.EquippedArmourInstances {
		if err := recordOwnedInstance(ownedInstances, instanceID, "equipped "+string(slot)); err != nil {
			return fmt.Errorf("restore equipment: %w", err)
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
		id: id, position: state.Position, spawnPosition: Position{},
		baseSpeed: speed, speed: movementSpeedForAgility(speed, SkillLevelForExperience(skills[SkillAgility])),
		navigationLocation: location, health: state.Health, maximumHealth: maximumHealth,
		prayer: playerMaximumPrayer, maximumPrayer: playerMaximumPrayer,
		activePrayers: make(map[PrayerID]bool), prayerDrainProgress: make(map[PrayerID]uint64),
		equippedWeapon: state.EquippedWeapon, equippedWeaponInstanceID: state.EquippedWeaponInstanceID,
		equippedWeaponReinforcement: state.EquippedWeaponReinforcement, equippedSpellbook: state.EquippedSpellbook,
		dodgeRate: playerDodgeRate, inventory: inventory, ammunitionPouch: ammunitionPouch, bank: bank,
		equippedArmour:               cloneEquippedArmour(state.EquippedArmour),
		equippedArmourInstances:      cloneEquippedArmourInstances(state.EquippedArmourInstances),
		equippedArmourReinforcements: cloneEquippedArmourReinforcements(state.EquippedArmourReinforcements),
		combatStyle:                  state.CombatStyle,
		skills:                       skills,
	}
	return nil
}

func recordOwnedInstance(owned map[string]string, instanceID, location string) error {
	if instanceID == "" {
		return nil
	}
	if previous, exists := owned[instanceID]; exists {
		return fmt.Errorf("instance %q appears in both %s and %s", instanceID, previous, location)
	}
	owned[instanceID] = location
	return nil
}

func extractLegacyEquippedInstance(inventory *inventoryState, itemID, preferred string, reinforcement int) (string, int) {
	if inventory == nil {
		return preferred, reinforcement
	}
	if item, exists := inventory.itemByInstance(preferred); exists && item.ItemID == itemID {
		inventory.removeInstance(preferred)
		return preferred, item.Reinforcement
	}
	for _, item := range inventory.items {
		if item.ItemID == itemID && item.InstanceID != "" {
			inventory.removeInstance(item.InstanceID)
			return item.InstanceID, item.Reinforcement
		}
	}
	return preferred, reinforcement
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
