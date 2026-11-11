package simulation

import (
	"fmt"
	"math"
	"sort"
	"time"
)

const (
	actionPulseInterval                  = 600 * time.Millisecond
	skillDamagePointsPerLevel            = 0.5
	swordAttackIntervalPulses            = 4
	swordRange                           = 1.55
	swordApproachRange                   = 1.25
	swordMeleeStrength                   = 10
	swordDamageVariance                  = 0.20
	steelSwordMeleeStrength              = 14
	runeSwordMeleeStrength               = 40
	daggerRange                          = 1.3
	daggerApproachRange                  = 1.05
	daggerMeleeStrength                  = 7
	daggerDamageVariance                 = 2.0 / 7.0
	daggerAttackIntervalPulses           = 3
	bowRange                             = 12.0
	bowApproachRange                     = 11.25
	bowRangedStrength                    = 4
	bowDamageVariance                    = 0.25
	bowAttackIntervalPulses              = 5
	maceRange                            = 1.45
	maceApproachRange                    = 1.15
	maceMeleeStrength                    = 12
	maceDamageVariance                   = 0.25
	maceAttackIntervalPulses             = 5
	battleAxeRange                       = 1.55
	battleAxeApproachRange               = 1.25
	battleAxeMeleeStrength               = 15
	battleAxeDamageVariance              = 4.0 / 15.0
	battleAxeAttackIntervalPulses        = 6
	staffRange                           = 10.0
	staffApproachRange                   = 9.25
	staffMagicStrength                   = 6
	staffDamageVariance                  = 1.0 / 3.0
	staffAttackIntervalPulses            = 5
	polearmRange                         = 2.8
	polearmApproachRange                 = 2.35
	spearMeleeStrength                   = 11
	spearAttackIntervalPulses            = 5
	halberdMeleeStrength                 = 16
	halberdAttackIntervalPulses          = 6
	crossbowRange                        = 11.0
	crossbowApproachRange                = 10.25
	crossbowRangedStrength               = 8
	crossbowAttackIntervalPulses         = 7
	longRangeStyleRangeBonus             = 4.0
	wideStyleDamageVarianceBonus         = 0.15
	accurateStyleDamageVarianceReduction = 0.10
	accurateStyleAccuracyBonus           = 2
	playerMaximumHealth                  = 100
	playerMaximumPrayer                  = 100
	playerDodgeRate                      = defaultBaseDodgeBonus
	dummyMaximumHealth                   = 100
	wildFowlMaximumHealth                = 30
	wildFowlAttackDamage                 = 4
	cowMaximumHealth                     = 45
	cowAttackDamage                      = 6
	dummyDodgeRate                       = defaultBaseDodgeBonus
	dummyAttackDamage                    = 15
	dummyAttackIntervalPulses            = 4
	dummyAttackRange                     = 1.55
	dummyLeashRange                      = 6.0
	roamingEnemySpeed                    = 3.0
	riftTyrantMovementSpeed              = DefaultPlayerMovementSpeed + 0.5
	rangedTyrantMovementSpeed            = DefaultPlayerMovementSpeed + 0.75
	mageTyrantMovementSpeed              = DefaultPlayerMovementSpeed + 0.5
	roamingRepathPulses                  = 1
	respawnDelayPulses                   = 5
	dummyTargetID                        = "dummy"
	upperDummyTargetID                   = "upper_dummy"
	overlookDummyTargetID                = "overlook_dummy"
	hillWardenTargetID                   = "hill_warden"
	wildFowlTargetID                     = "wild_fowl"
	cowTargetID                          = "cow"
	riftTyrantTargetID                   = "rift_tyrant"
	rangedTyrantTargetID                 = "ranged_tyrant"
	mageTyrantTargetID                   = "mage_tyrant"
	hillWardenEvasionScore               = 18
	wildFowlEvasionScore                 = 2
	cowEvasionScore                      = 4
	riftTyrantMaximumHealth              = 650
	riftTyrantAccuracyScore              = 40
	riftTyrantEvasionScore               = 30
	rangedTyrantMaximumHealth            = 600
	rangedTyrantAccuracyScore            = 44
	rangedTyrantEvasionScore             = 34
	mageTyrantMaximumHealth              = 575
	mageTyrantAccuracyScore              = 46
	mageTyrantEvasionScore               = 32
	combatEventHistoryLimit              = 32
)

const (
	targetKindDummy      = "dummy"
	targetKindCreature   = "creature"
	targetKindBird       = "bird"
	targetKindCow        = "cow"
	targetKindBoss       = "boss"
	targetKindRangedBoss = "ranged_boss"
	targetKindMageBoss   = "mage_boss"
)

type WeaponID string

const (
	WeaponSword          WeaponID = "sword"
	WeaponBronzeSword    WeaponID = "bronze_sword"
	WeaponMithrilSword   WeaponID = "mithril_sword"
	WeaponUnarmed        WeaponID = "unarmed"
	WeaponSteelSword     WeaponID = "steel_sword"
	WeaponRuneSword      WeaponID = "rune_sword"
	WeaponDagger         WeaponID = "dagger"
	WeaponBow            WeaponID = "bow"
	WeaponOakShortbow    WeaponID = "oak_shortbow"
	WeaponWillowShortbow WeaponID = "willow_shortbow"
	WeaponMapleShortbow  WeaponID = "maple_shortbow"
	WeaponYewShortbow    WeaponID = "yew_shortbow"
	WeaponMace           WeaponID = "mace"
	WeaponBattleAxe      WeaponID = "battle_axe"
	WeaponStaff          WeaponID = "staff"
	WeaponSpear          WeaponID = "spear"
	WeaponHalberd        WeaponID = "halberd"
	WeaponCrossbow       WeaponID = "crossbow"
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
)

func defaultCombatStyle(weapon WeaponID) CombatStyleID {
	definition, exists := weaponDefinition(weapon)
	if !exists {
		return CombatStyleSlash
	}
	return definition.DefaultStyle
}

func combatStyleValidForWeapon(style CombatStyleID, weapon WeaponID) bool {
	definition, exists := weaponDefinition(weapon)
	if !exists {
		return false
	}
	for _, allowed := range definition.CombatStyles {
		if style == allowed {
			return true
		}
	}
	return false
}

type WeaponDefinition struct {
	ID               WeaponID
	DefaultStyle     CombatStyleID
	CombatStyles     []CombatStyleID
	StyleDamageTypes map[CombatStyleID]DamageType
	TwoHanded        bool
	AccuracyBonus    int
	MeleeStrength    int
	RangedStrength   int
	MagicStrength    int
	DamageVariance   float64
	AttackInterval   uint64
	Range            float64
	ApproachRange    float64
	Ammunition       ItemCost
	AmmunitionFamily string
}

func weaponDefinition(id WeaponID) (WeaponDefinition, bool) {
	item, exists := itemDefinition(string(id))
	if !exists || item.Kind != ItemKindWeapon {
		return WeaponDefinition{}, false
	}
	definition := item.Weapon
	definition.CombatStyles = append([]CombatStyleID(nil), item.Weapon.CombatStyles...)
	definition.StyleDamageTypes = make(map[CombatStyleID]DamageType, len(item.Weapon.StyleDamageTypes))
	for style, damageType := range item.Weapon.StyleDamageTypes {
		definition.StyleDamageTypes[style] = damageType
	}
	return definition, true
}

func weaponDamageTypeForStyle(weapon WeaponDefinition, style CombatStyleID) (DamageType, bool) {
	if !combatStyleValidForWeapon(style, weapon.ID) {
		return "", false
	}
	damageType, exists := weapon.StyleDamageTypes[style]
	return damageType, exists
}

func weaponDefinitionForStyle(id WeaponID, style CombatStyleID) (WeaponDefinition, bool) {
	weapon, exists := weaponDefinition(id)
	if !exists {
		return WeaponDefinition{}, false
	}
	if weapon.RangedStrength > 0 || weapon.MagicStrength > 0 {
		switch style {
		case CombatStyleRapid:
			weapon.AttackInterval = max(2, weapon.AttackInterval-1)
			weapon.DamageVariance += wideStyleDamageVarianceBonus
		case CombatStyleAccurate:
			weapon.AccuracyBonus += accurateStyleAccuracyBonus
			weapon.DamageVariance = max(0, weapon.DamageVariance-accurateStyleDamageVarianceReduction)
		case CombatStyleLongRange:
			weapon.Range += longRangeStyleRangeBonus
			weapon.ApproachRange += longRangeStyleRangeBonus
			weapon.DamageVariance += wideStyleDamageVarianceBonus
		}
	}
	return weapon, true
}

func equippedWeaponDefinitionForStyle(player *playerState) (WeaponDefinition, bool) {
	if player == nil {
		return WeaponDefinition{}, false
	}
	weapon, exists := weaponDefinitionForStyle(player.equippedWeapon, player.combatStyle)
	if !exists {
		return WeaponDefinition{}, false
	}
	if reinforcement := equippedReinforcement(player, player.equippedWeaponInstanceID); reinforcement > 0 {
		weapon.AccuracyBonus += reinforcement
		if weapon.MeleeStrength > 0 {
			weapon.MeleeStrength += reinforcement
		} else if weapon.RangedStrength > 0 {
			weapon.RangedStrength += reinforcement
		} else if weapon.MagicStrength > 0 {
			weapon.MagicStrength += reinforcement
		}
	}
	return weapon, true
}

type CombatEventType string

const (
	CombatEventDamage CombatEventType = "damage"
	CombatEventMiss   CombatEventType = "missed"
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
	SpellID  SpellID
}

type TargetCommand struct {
	PlayerID PlayerID
	Sequence uint64
	TargetID string
}

type EquipCommand struct {
	PlayerID   PlayerID
	Sequence   uint64
	ItemID     WeaponID
	InstanceID string
	Unequip    bool
}

type UnequipCommand struct {
	PlayerID   PlayerID
	Sequence   uint64
	ItemID     string
	InstanceID string
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
	Accuracy        int
	Evasion         int
	AttackType      DamageType
	CombatLevel     int
	Defence         DefenceBonuses
	Elements        ElementalResistances
	attackSequence  uint64
}

type combatState struct {
	elapsed          time.Duration
	pulse            uint64
	targets          map[string]*TargetSnapshot
	pendingTargets   map[PlayerID]TargetCommand
	pendingEquipment map[PlayerID]EquipCommand
	pendingStyles    map[PlayerID]CombatStyleCommand
	pendingPrayers   map[PlayerID]PrayerCommand
	pendingSpells    map[PlayerID]SpellCommand
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
	wildFowlHome := Position{X: -20, Z: -5}
	cowHome := Position{X: -16, Z: -10}
	riftTyrantHome := Position{X: 38, Z: 34}
	rangedTyrantHome := Position{X: -40, Z: -35}
	mageTyrantHome := Position{X: 38, Z: -34}
	hillWardenProfile := combatProfileForEnemy(hillWardenTargetID)
	hillWarden := &TargetSnapshot{
		ID:            hillWardenTargetID,
		Name:          "Hill Warden",
		Kind:          targetKindCreature,
		Position:      hillWardenHome,
		Health:        dummyMaximumHealth,
		MaximumHealth: dummyMaximumHealth,
		DodgeRate:     dummyDodgeRate,
		Accuracy:      dummyAccuracyScore,
		Evasion:       hillWardenEvasionScore,
		AttackType:    hillWardenProfile.DamageType,
		Defence:       hillWardenProfile.Defence,
		Elements:      hillWardenProfile.Elements,
		CombatLevel:   hillWardenProfile.CombatLevel,
	}
	wildFowlProfile := combatProfileForEnemy(wildFowlTargetID)
	wildFowl := &TargetSnapshot{
		ID:            wildFowlTargetID,
		Name:          "Wild Fowl",
		Kind:          targetKindBird,
		Position:      wildFowlHome,
		Health:        wildFowlMaximumHealth,
		MaximumHealth: wildFowlMaximumHealth,
		DodgeRate:     defaultBaseDodgeBonus,
		Accuracy:      5,
		Evasion:       wildFowlEvasionScore,
		AttackType:    wildFowlProfile.DamageType,
		Defence:       wildFowlProfile.Defence,
		Elements:      wildFowlProfile.Elements,
		CombatLevel:   wildFowlProfile.CombatLevel,
	}
	cowProfile := combatProfileForEnemy(cowTargetID)
	cow := &TargetSnapshot{
		ID: cowTargetID, Name: "Cow", Kind: targetKindCow, Position: cowHome,
		Health: cowMaximumHealth, MaximumHealth: cowMaximumHealth,
		DodgeRate: defaultBaseDodgeBonus, Accuracy: 7, Evasion: cowEvasionScore,
		AttackType: cowProfile.DamageType, Defence: cowProfile.Defence, Elements: cowProfile.Elements,
		CombatLevel: cowProfile.CombatLevel,
	}
	riftTyrantProfile := combatProfileForEnemy(riftTyrantTargetID)
	riftTyrant := &TargetSnapshot{
		ID: riftTyrantTargetID, Name: "Rift Tyrant", Kind: targetKindBoss, Position: riftTyrantHome,
		Health: riftTyrantMaximumHealth, MaximumHealth: riftTyrantMaximumHealth,
		DodgeRate: defaultBaseDodgeBonus, Accuracy: riftTyrantAccuracyScore, Evasion: riftTyrantEvasionScore,
		AttackType: enemyAttackType(riftTyrantProfile, 0), Defence: riftTyrantProfile.Defence,
		Elements: riftTyrantProfile.Elements, CombatLevel: riftTyrantProfile.CombatLevel,
	}
	rangedTyrantProfile := combatProfileForEnemy(rangedTyrantTargetID)
	rangedTyrant := &TargetSnapshot{
		ID: rangedTyrantTargetID, Name: "Ranged Tyrant", Kind: targetKindRangedBoss, Position: rangedTyrantHome,
		Health: rangedTyrantMaximumHealth, MaximumHealth: rangedTyrantMaximumHealth,
		DodgeRate: defaultBaseDodgeBonus, Accuracy: rangedTyrantAccuracyScore, Evasion: rangedTyrantEvasionScore,
		AttackType: rangedTyrantProfile.DamageType, Defence: rangedTyrantProfile.Defence,
		Elements: rangedTyrantProfile.Elements, CombatLevel: rangedTyrantProfile.CombatLevel,
	}
	mageTyrantProfile := combatProfileForEnemy(mageTyrantTargetID)
	mageTyrant := &TargetSnapshot{
		ID: mageTyrantTargetID, Name: "Mage Tyrant", Kind: targetKindMageBoss, Position: mageTyrantHome,
		Health: mageTyrantMaximumHealth, MaximumHealth: mageTyrantMaximumHealth,
		DodgeRate: defaultBaseDodgeBonus, Accuracy: mageTyrantAccuracyScore, Evasion: mageTyrantEvasionScore,
		AttackType: mageTyrantProfile.DamageType, Defence: mageTyrantProfile.Defence,
		Elements: mageTyrantProfile.Elements, CombatLevel: mageTyrantProfile.CombatLevel,
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
				Accuracy:      dummyAccuracyScore,
				Evasion:       dummyEvasionScore,
				AttackType:    DamageTypeCrush,
				CombatLevel:   defaultEnemyCombatProfile.CombatLevel,
			},
			upperDummyTargetID: {
				ID:            upperDummyTargetID,
				Name:          "Upper Dummy",
				Kind:          targetKindDummy,
				Position:      Position{X: -4, Y: 3, Z: 2},
				Health:        dummyMaximumHealth,
				MaximumHealth: dummyMaximumHealth,
				DodgeRate:     dummyDodgeRate,
				Accuracy:      dummyAccuracyScore,
				Evasion:       dummyEvasionScore,
				AttackType:    DamageTypeCrush,
				CombatLevel:   defaultEnemyCombatProfile.CombatLevel,
			},
			overlookDummyTargetID: {
				ID:            overlookDummyTargetID,
				Name:          "Overlook Dummy",
				Kind:          targetKindDummy,
				Position:      Position{X: 2.75, Y: 0, Z: 3.5},
				Health:        dummyMaximumHealth,
				MaximumHealth: dummyMaximumHealth,
				DodgeRate:     dummyDodgeRate,
				Accuracy:      dummyAccuracyScore,
				Evasion:       dummyEvasionScore,
				AttackType:    DamageTypeCrush,
				CombatLevel:   defaultEnemyCombatProfile.CombatLevel,
			},
			hillWardenTargetID:   hillWarden,
			wildFowlTargetID:     wildFowl,
			cowTargetID:          cow,
			riftTyrantTargetID:   riftTyrant,
			rangedTyrantTargetID: rangedTyrant,
			mageTyrantTargetID:   mageTyrant,
		},
		pendingTargets:   make(map[PlayerID]TargetCommand),
		pendingEquipment: make(map[PlayerID]EquipCommand),
		pendingStyles:    make(map[PlayerID]CombatStyleCommand),
		pendingPrayers:   make(map[PlayerID]PrayerCommand),
		pendingSpells:    make(map[PlayerID]SpellCommand),
		roamingEnemies: map[string]*roamingEnemyState{
			hillWardenTargetID: {
				target: hillWarden,
				home:   hillWardenHome,
				speed:  roamingEnemySpeed,
			},
			wildFowlTargetID: {
				target: wildFowl,
				home:   wildFowlHome,
				speed:  roamingEnemySpeed * 0.8,
			},
			cowTargetID: {
				target: cow,
				home:   cowHome,
				speed:  roamingEnemySpeed * 0.65,
			},
			riftTyrantTargetID: {
				target: riftTyrant,
				home:   riftTyrantHome,
				speed:  riftTyrantMovementSpeed,
			},
			rangedTyrantTargetID: {
				target: rangedTyrant,
				home:   rangedTyrantHome,
				speed:  rangedTyrantMovementSpeed,
			},
			mageTyrantTargetID: {
				target: mageTyrant,
				home:   mageTyrantHome,
				speed:  mageTyrantMovementSpeed,
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

// QueueEquip records owned equipment to equip on the next action pulse.
func (w *World) QueueEquip(command EquipCommand) error {
	command.Unequip = false
	return w.queueEquipment(command)
}

// QueueUnequip records an equipped item to remove on the next action pulse.
// Weapons transition to the innate Unarmed profile; other equipment clears its slot.
func (w *World) QueueUnequip(command UnequipCommand) error {
	return w.queueEquipment(EquipCommand{
		PlayerID:   command.PlayerID,
		Sequence:   command.Sequence,
		ItemID:     WeaponID(command.ItemID),
		InstanceID: command.InstanceID,
		Unequip:    true,
	})
}

func (w *World) queueEquipment(command EquipCommand) error {
	if command.Sequence == 0 {
		return fmt.Errorf("equipment sequence must be positive")
	}
	item, itemExists := itemDefinition(string(command.ItemID))
	if !itemExists || item.Virtual ||
		(item.Kind != ItemKindWeapon && item.Kind != ItemKindArmour && item.Kind != ItemKindSpellbook) {
		return fmt.Errorf("item %q does not exist", command.ItemID)
	}
	if item.Kind == ItemKindArmour && item.Armour.Slot == ArmourSlotRing {
		return fmt.Errorf("rings must be managed through the ring rack")
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
	if itemIsInstanced(string(command.ItemID)) && command.InstanceID == "" {
		if command.Unequip && item.Kind == ItemKindWeapon {
			command.InstanceID = player.equippedWeaponInstanceID
		} else if command.Unequip && item.Kind == ItemKindArmour {
			command.InstanceID = player.equippedArmourInstances[item.Armour.Slot]
		} else {
			command.InstanceID = resolveOwnedInstanceID(player.inventory, string(command.ItemID), "")
		}
	}
	if command.Unequip {
		if item.Kind == ItemKindWeapon && player.equippedWeapon != command.ItemID {
			return fmt.Errorf("weapon %q is not equipped", command.ItemID)
		}
		if item.Kind == ItemKindWeapon && command.InstanceID != "" && player.equippedWeaponInstanceID != command.InstanceID {
			return fmt.Errorf("weapon instance %q is not equipped", command.InstanceID)
		}
		if item.Kind == ItemKindArmour && player.equippedArmour[item.Armour.Slot] != string(command.ItemID) {
			return fmt.Errorf("armour %q is not equipped", command.ItemID)
		}
		if item.Kind == ItemKindArmour && command.InstanceID != "" && player.equippedArmourInstances[item.Armour.Slot] != command.InstanceID {
			return fmt.Errorf("armour instance %q is not equipped", command.InstanceID)
		}
		if item.Kind == ItemKindSpellbook && player.equippedSpellbook != string(command.ItemID) {
			return fmt.Errorf("spellbook %q is not equipped", command.ItemID)
		}
		equippedItem, equipped := currentEquippedItem(player, item)
		candidate := cloneInventory(player.inventory)
		if !equipped || !addReturnedEquipment(&candidate, equippedItem) {
			return fmt.Errorf("inventory is full")
		}
	} else {
		if requirement, unmet := unmetEquipRequirement(player, item); unmet {
			return fmt.Errorf("item %q requires %s level %d", item.ID, requirement.Skill, requirement.Level)
		}
		ownedItem, ownsInstance := player.inventory.itemByInstance(command.InstanceID)
		if command.InstanceID != "" && (!ownsInstance || ownedItem.ItemID != string(command.ItemID)) {
			return fmt.Errorf("player %q does not own item instance %q", command.PlayerID, command.InstanceID)
		}
		if command.InstanceID == "" && !player.inventory.contains(string(command.ItemID)) {
			return fmt.Errorf("player %q does not own item %q", command.PlayerID, command.ItemID)
		}
		candidate := cloneInventory(player.inventory)
		if !removeInventoryEquipment(&candidate, item, command.InstanceID) {
			return fmt.Errorf("player %q does not own item %q", command.PlayerID, command.ItemID)
		}
		if outgoing, equipped := currentEquippedItem(player, item); equipped && !addReturnedEquipment(&candidate, outgoing) {
			return fmt.Errorf("inventory is full")
		}
		if conflict, _, exists := conflictingHandEquipment(player, item); exists && !addReturnedEquipment(&candidate, conflict) {
			return fmt.Errorf("inventory is full")
		}
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
	if command.TargetID != "" {
		delete(w.pendingPickups, command.PlayerID)
		delete(w.pendingGathering, command.PlayerID)
		clearGathering(player, "")
		player.autoPickingUp = false
	}

	w.combat.pendingTargets[command.PlayerID] = command
	return nil
}

func (w *World) advanceCombat(elapsed time.Duration) {
	w.combat.elapsed += elapsed
	for w.combat.elapsed >= actionPulseInterval {
		w.combat.elapsed -= actionPulseInterval
		w.combat.pulse++
		w.applyPendingBrews()
		w.applyPendingFletches()
		w.respawnDefeatedEntities()
		w.applyEquipmentCommands()
		w.applyCombatStyleCommands()
		w.applyPrayerCommands()
		w.applySpellCommands()
		w.resolveAutoAttacks()
		w.resolveGatheringActions()
		w.updateRoamingEnemies()
		w.resolveDummyAttack()
		w.drainActivePrayers()
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
			item, itemExists := itemDefinition(string(command.ItemID))
			if !itemExists {
				delete(w.combat.pendingEquipment, id)
				continue
			}
			applyEquipmentChange(player, item, command)
			player.lastEquipmentSequence = command.Sequence
			player.lastEquipmentPulse = w.combat.pulse
		}
		delete(w.combat.pendingEquipment, id)
	}
}

func applyEquipmentChange(player *playerState, item ItemDefinition, command EquipCommand) bool {
	if player == nil {
		return false
	}
	candidate := cloneInventory(player.inventory)
	if command.Unequip {
		outgoing, equipped := currentEquippedItem(player, item)
		if !equipped || !addReturnedEquipment(&candidate, outgoing) {
			return false
		}
		player.inventory = candidate
		clearEquippedItem(player, item)
		refreshPlayerMaximumHealth(player)
		return true
	}
	incoming, exists := inventoryEquipment(candidate, item, command.InstanceID)
	if !exists || !removeInventoryEquipment(&candidate, item, command.InstanceID) {
		return false
	}
	if outgoing, equipped := currentEquippedItem(player, item); equipped && !addReturnedEquipment(&candidate, outgoing) {
		return false
	}
	conflict, conflictDefinition, hasConflict := conflictingHandEquipment(player, item)
	if hasConflict && !addReturnedEquipment(&candidate, conflict) {
		return false
	}
	player.inventory = candidate
	if hasConflict {
		clearEquippedItem(player, conflictDefinition)
	}
	setEquippedItem(player, item, incoming)
	refreshPlayerMaximumHealth(player)
	return true
}

func conflictingHandEquipment(player *playerState, incoming ItemDefinition) (InventoryItem, ItemDefinition, bool) {
	if player == nil {
		return InventoryItem{}, ItemDefinition{}, false
	}
	if incoming.Kind == ItemKindWeapon && incoming.Weapon.TwoHanded {
		leftHandID := player.equippedArmour[ArmourSlotLeftHand]
		leftHand, exists := itemDefinition(leftHandID)
		if !exists {
			return InventoryItem{}, ItemDefinition{}, false
		}
		item, equipped := currentEquippedItem(player, leftHand)
		return item, leftHand, equipped
	}
	if incoming.Kind == ItemKindArmour && incoming.Armour.Slot == ArmourSlotLeftHand {
		weapon, exists := itemDefinition(string(player.equippedWeapon))
		if !exists || !weapon.Weapon.TwoHanded {
			return InventoryItem{}, ItemDefinition{}, false
		}
		item, equipped := currentEquippedItem(player, weapon)
		return item, weapon, equipped
	}
	return InventoryItem{}, ItemDefinition{}, false
}

func inventoryEquipment(inventory inventoryState, definition ItemDefinition, instanceID string) (InventoryItem, bool) {
	if itemIsInstanced(definition.ID) {
		item, exists := inventory.itemByInstance(instanceID)
		return item, exists && item.ItemID == definition.ID
	}
	if inventory.contains(definition.ID) {
		return InventoryItem{ItemID: definition.ID, Quantity: 1}, true
	}
	return InventoryItem{}, false
}

func removeInventoryEquipment(inventory *inventoryState, definition ItemDefinition, instanceID string) bool {
	if inventory == nil {
		return false
	}
	if itemIsInstanced(definition.ID) {
		item, removed := inventory.removeInstance(instanceID)
		return removed && item.ItemID == definition.ID
	}
	return inventory.remove(definition.ID, 1)
}

func addReturnedEquipment(inventory *inventoryState, item InventoryItem) bool {
	if inventory == nil || item.ItemID == "" {
		return false
	}
	if itemIsInstanced(item.ItemID) {
		item.Slot = 0
		return inventory.addInventoryItem(item)
	}
	return inventory.add(item.ItemID, item.Quantity)
}

func currentEquippedItem(player *playerState, definition ItemDefinition) (InventoryItem, bool) {
	if player == nil {
		return InventoryItem{}, false
	}
	switch definition.Kind {
	case ItemKindWeapon:
		if player.equippedWeapon == WeaponUnarmed {
			return InventoryItem{}, false
		}
		return InventoryItem{ItemID: string(player.equippedWeapon), Quantity: 1, InstanceID: player.equippedWeaponInstanceID, Reinforcement: player.equippedWeaponReinforcement}, true
	case ItemKindArmour:
		equippedItemID := player.equippedArmour[definition.Armour.Slot]
		if equippedItemID == "" {
			return InventoryItem{}, false
		}
		return InventoryItem{ItemID: equippedItemID, Quantity: 1, InstanceID: player.equippedArmourInstances[definition.Armour.Slot], Reinforcement: player.equippedArmourReinforcements[definition.Armour.Slot]}, true
	case ItemKindSpellbook:
		return InventoryItem{ItemID: player.equippedSpellbook, Quantity: 1}, player.equippedSpellbook != ""
	default:
		return InventoryItem{}, false
	}
}

func clearEquippedItem(player *playerState, definition ItemDefinition) {
	switch definition.Kind {
	case ItemKindSpellbook:
		player.equippedSpellbook = ""
		player.selectedSpell = ""
	case ItemKindArmour:
		delete(player.equippedArmour, definition.Armour.Slot)
		delete(player.equippedArmourInstances, definition.Armour.Slot)
		delete(player.equippedArmourReinforcements, definition.Armour.Slot)
	case ItemKindWeapon:
		player.equippedWeapon = WeaponUnarmed
		player.equippedWeaponInstanceID = ""
		player.equippedWeaponReinforcement = 0
		player.combatStyle = defaultCombatStyle(WeaponUnarmed)
	}
}

func setEquippedItem(player *playerState, definition ItemDefinition, item InventoryItem) {
	if player.equippedArmour == nil {
		player.equippedArmour = make(map[ArmourSlot]string)
	}
	if player.equippedArmourInstances == nil {
		player.equippedArmourInstances = make(map[ArmourSlot]string)
	}
	if player.equippedArmourReinforcements == nil {
		player.equippedArmourReinforcements = make(map[ArmourSlot]int)
	}
	switch definition.Kind {
	case ItemKindSpellbook:
		player.equippedSpellbook = definition.ID
		if selected, exists := spellDefinition(player.selectedSpell); exists && selected.BookID != definition.Spellbook {
			player.selectedSpell = ""
		}
	case ItemKindArmour:
		player.equippedArmour[definition.Armour.Slot] = definition.ID
		player.equippedArmourInstances[definition.Armour.Slot] = item.InstanceID
		player.equippedArmourReinforcements[definition.Armour.Slot] = item.Reinforcement
	case ItemKindWeapon:
		player.equippedWeapon = definition.Weapon.ID
		player.equippedWeaponInstanceID = item.InstanceID
		player.equippedWeaponReinforcement = item.Reinforcement
		player.combatStyle = defaultCombatStyle(definition.Weapon.ID)
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
			target.dummy.attackSequence = 0
			profile := combatProfileForEnemy(target.id)
			target.dummy.AttackType = enemyAttackType(profile, 0)
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
			player.autoApproaching = false
		} else {
			stopPlayer(player)
			player.autoApproaching = false
			player.lastActionResult = "engaged"
			if target, targetExists := w.combat.dummyByID(command.TargetID); targetExists && !target.Defeated {
				weapon, _ := equippedWeaponDefinitionForStyle(player)
				attackRange := weapon.Range
				approachRange := weapon.ApproachRange
				if spell, castingSpell := spellDefinition(player.selectedSpell); castingSpell {
					spell = spellDefinitionForStyle(player, spell)
					attackRange = spell.Range
					approachRange = max(0.1, spell.Range-0.75)
				}
				if distance(player.position, target.Position) > attackRange {
					if err := w.approachDummy(player, target.Position, approachRange); err != nil {
						player.lastActionResult = "target_unreachable"
					} else {
						player.lastActionResult = "approaching"
					}
				}
			}
		}
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
		weapon, _ := equippedWeaponDefinitionForStyle(player)
		spell, castingSpell := spellDefinition(player.selectedSpell)
		if castingSpell {
			spell = spellDefinitionForStyle(player, spell)
		}
		attackRange := weapon.Range
		approachRange := weapon.ApproachRange
		attackInterval := effectiveWeaponAttackInterval(player, weapon)
		if castingSpell {
			attackRange = spell.Range
			approachRange = max(0.1, spell.Range-0.75)
			attackInterval = spell.AttackInterval
		}
		if distance(player.position, target.Position) > attackRange {
			if err := w.approachDummy(player, target.Position, approachRange); err != nil {
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

		if castingSpell && !paySpellCosts(player, spell.Costs) {
			player.lastActionResult = "missing_reagents"
			continue
		}
		ammunitionStrength := 0
		if !castingSpell && weapon.Ammunition.Quantity > 0 {
			ammunitionCost, strength, available := ammunitionForWeapon(player, weapon)
			if !available || !payItemCosts(&player.ammunitionPouch, []ItemCost{ammunitionCost}) {
				player.lastActionResult = "missing_ammunition"
				continue
			}
			ammunitionStrength = strength
		}
		player.lastActionPulse = w.combat.pulse
		player.nextAttackPulse = w.combat.pulse + attackInterval
		accuracyChance := targetAccuracyChance(player, weapon)
		dodgeChance := targetDodgeChance(player, target, weapon)
		if castingSpell {
			accuracyChance = targetSpellAccuracyChance(player, spell)
			dodgeChance = targetSpellDodgeChance(player, target, spell)
		}
		outcomeRoll := w.rollSource.Float64()
		if outcomeRoll >= accuracyChance {
			player.lastActionResult = "missed"
			w.combat.recordEvent(CombatEvent{
				Pulse:    w.combat.pulse,
				Type:     CombatEventMiss,
				SourceID: string(player.id),
				TargetID: player.combatTarget,
				SpellID:  player.selectedSpell,
			})
			continue
		}
		if outcomeRoll < accuracyChance*dodgeChance {
			player.lastActionResult = "dodged"
			w.combat.recordEvent(CombatEvent{
				Pulse:    w.combat.pulse,
				Type:     CombatEventDodged,
				SourceID: string(player.id),
				TargetID: player.combatTarget,
				SpellID:  player.selectedSpell,
			})
			continue
		}

		poweredDamage := weaponPoweredDamageWithAmmunition(player, weapon, ammunitionStrength)
		damageVariance := weapon.DamageVariance
		damageType := DamageTypeCrush
		if styleDamageType, hasDamageType := weaponDamageTypeForStyle(weapon, player.combatStyle); hasDamageType {
			damageType = styleDamageType
		}
		if castingSpell {
			poweredDamage = spellPoweredDamage(player, spell)
			poweredDamage *= elementalDamageMultiplier(target, spell.Element)
			damageVariance = spell.DamageVariance
			damageType = DamageTypeMagic
		}
		poweredDamage *= 1 - targetDamageReduction(target, damageType)
		// Defence mitigation happens before variance so the weapon's damage feel
		// remains the final stage before overkill clamping.
		rolledDamage := rollDamageWithVariance(poweredDamage, damageVariance, w.rollSource.Float64())
		damageDealt := min(rolledDamage, target.Health)
		target.Health -= damageDealt
		if castingSpell {
			grantSkillExperience(player, SkillMagic, int64(damageDealt))
		} else if skill, trains := skillForCombatStyle(player.equippedWeapon, player.combatStyle); trains {
			grantSkillExperience(player, skill, int64(damageDealt))
		}
		player.lastActionResult = "hit"
		w.combat.recordEvent(CombatEvent{
			Pulse:    w.combat.pulse,
			Type:     CombatEventDamage,
			SourceID: string(player.id),
			TargetID: player.combatTarget,
			Amount:   damageDealt,
			SpellID:  player.selectedSpell,
		})
		if target.TargetID == "" {
			profile := combatProfileForEnemy(player.combatTarget)
			target.TargetID = player.id
			target.NextAttackPulse = w.combat.pulse + profile.AttackInterval
		}
		if target.Health == 0 {
			w.defeatDummy(player.combatTarget, target)
		}
	}
}

func weaponPoweredDamage(player *playerState, weapon WeaponDefinition) float64 {
	return weaponPoweredDamageWithAmmunition(player, weapon, 0)
}

func weaponPoweredDamageWithAmmunition(player *playerState, weapon WeaponDefinition, ammunitionStrength int) float64 {
	strength := weapon.MeleeStrength
	skill := SkillStrength
	if weapon.RangedStrength > 0 {
		strength = weapon.RangedStrength + ammunitionStrength
		skill = SkillRange
	} else if weapon.MagicStrength > 0 {
		strength = weapon.MagicStrength
		skill = SkillMagic
	}
	_, equipmentStrength := equipmentOffenceForWeapon(player, weapon)
	strength += equipmentStrength
	level := 1
	if player != nil {
		level = effectiveCombatSkillLevel(player, skill)
	}
	damage := float64(strength) + float64(level-1)*skillDamagePointsPerLevel
	if player != nil {
		bonuses := activePrayerBonuses(player)
		switch {
		case weapon.RangedStrength > 0:
			damage *= 1 + bonuses.RangedDamage
		case weapon.MagicStrength > 0:
			damage *= 1 + bonuses.MagicDamage
		case weapon.MeleeStrength > 0:
			damage *= 1 + bonuses.MeleeDamage
		}
	}
	return damage
}

func rollDamageWithVariance(poweredDamage, variance, roll float64) int {
	minimum, maximum := damageRange(poweredDamage, variance)
	if minimum == maximum {
		return minimum
	}
	roll = max(0, min(roll, math.Nextafter(1, 0)))
	return minimum + int(math.Floor(roll*float64(maximum-minimum+1)))
}

func damageRange(poweredDamage, variance float64) (int, int) {
	variance = max(0, variance)
	minimum := max(0, int(math.Round(poweredDamage*(1-variance))))
	maximum := max(minimum, int(math.Round(poweredDamage*(1+variance))))
	return minimum, maximum
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
		profile := combatProfileForEnemy(id)
		if enemy.target.Defeated {
			continue
		}

		hadTarget := enemy.target.TargetID != ""
		player, hasTarget := w.players[enemy.target.TargetID]
		if hadTarget && (!hasTarget || player.defeated || distance(player.position, enemy.home) > profile.LeashRange) {
			clearDummyTarget(enemy.target)
			enemy.path = nil
			hasTarget = false
		}
		if !hasTarget {
			player = w.nearestDetectablePlayer(id, enemy)
			if player != nil {
				enemy.target.TargetID = player.id
				enemy.target.NextAttackPulse = w.combat.pulse + profile.AttackInterval
				hasTarget = true
			}
		}

		if hasTarget {
			if distance(enemy.target.Position, player.position) <= profile.AttackRange {
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

func (w *World) nearestDetectablePlayer(enemyID string, enemy *roamingEnemyState) *playerState {
	profile := combatProfileForEnemy(enemyID)
	var nearest *playerState
	nearestDistance := math.Inf(1)
	for _, player := range w.players {
		if player.defeated || distance(player.position, enemy.home) > profile.LeashRange {
			continue
		}
		candidateDistance := distance(enemy.target.Position, player.position)
		if candidateDistance > profile.DetectionRange {
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
	profile := combatProfileForEnemy(dummyID)
	_, roaming := w.combat.roamingEnemies[dummyID]
	if (!roaming && distanceToPlayer > profile.LeashRange) ||
		(roaming && distance(player.position, w.combat.roamingEnemies[dummyID].home) > profile.LeashRange) {
		clearDummyTarget(dummy)
		return
	}
	if distanceToPlayer > profile.AttackRange || w.combat.pulse < dummy.NextAttackPulse {
		return
	}

	damageType := enemyAttackType(profile, dummy.attackSequence)
	dummy.attackSequence++
	dummy.AttackType = damageType
	dummy.NextAttackPulse = w.combat.pulse + profile.AttackInterval
	accuracyChance := calculatedAccuracyChance(dummy.Accuracy)
	dodgeChance := playerDodgeChance(dummy.Accuracy, player)
	outcomeRoll := w.rollSource.Float64()
	if outcomeRoll >= accuracyChance {
		w.combat.recordEvent(CombatEvent{
			Pulse:    w.combat.pulse,
			Type:     CombatEventMiss,
			SourceID: dummyID,
			TargetID: string(player.id),
		})
		return
	}
	if outcomeRoll < accuracyChance*dodgeChance {
		w.combat.recordEvent(CombatEvent{
			Pulse:    w.combat.pulse,
			Type:     CombatEventDodged,
			SourceID: dummyID,
			TargetID: string(player.id),
		})
		return
	}

	damageDealt := reduceIncomingDamage(player, profile.AttackDamage, damageType)
	player.health = max(0, player.health-damageDealt)
	w.combat.recordEvent(CombatEvent{
		Pulse:    w.combat.pulse,
		Type:     CombatEventDamage,
		SourceID: dummyID,
		TargetID: string(player.id),
		Amount:   damageDealt,
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
	delete(w.combat.pendingPrayers, player.id)
	clearActivePrayers(player)
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
	w.spawnTargetLoot(dummyID, dummy.Position)
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
	var destination Position
	var path []Position
	if w.pathfinder != nil {
		var err error
		path, err = w.pathfinder.FindPathWithinRange(player.position, target, approachRange)
		if err != nil {
			return fmt.Errorf("find combat path: %w", err)
		}
		if len(path) == 0 {
			return fmt.Errorf("find combat path: empty path")
		}
		destination = path[len(path)-1]
	} else {
		dx := player.position.X - target.X
		dz := player.position.Z - target.Z
		length := math.Hypot(dx, dz)
		if length == 0 {
			dx = 1
			length = 1
		}
		destination = Position{
			X: target.X + dx/length*approachRange,
			Y: target.Y,
			Z: target.Z + dz/length*approachRange,
		}
		path = []Position{destination}
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
