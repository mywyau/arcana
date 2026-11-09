package simulation

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
	mathrand "math/rand"
	"sort"
	"sync"
	"time"
)

type PlayerID string

type RollSource interface {
	Float64() float64
}

type randomRollSource struct {
	random *mathrand.Rand
}

func (source *randomRollSource) Float64() float64 {
	return source.random.Float64()
}

func newRandomRollSource() RollSource {
	var seedBytes [8]byte
	if _, err := cryptorand.Read(seedBytes[:]); err == nil {
		return &randomRollSource{random: mathrand.New(mathrand.NewSource(int64(binary.LittleEndian.Uint64(seedBytes[:]))))}
	}
	return &randomRollSource{random: mathrand.New(mathrand.NewSource(time.Now().UnixNano()))}
}

// Position is a precise point in three-dimensional world space.
type Position struct {
	X float64
	Y float64
	Z float64
}

// MoveCommand is a player's request to move toward a world position.
type MoveCommand struct {
	PlayerID    PlayerID
	Sequence    uint64
	Destination Position
}

type Pathfinder interface {
	FindPath(start, destination Position) ([]Position, error)
	FindPathWithinRange(start, target Position, maximumDistance float64) ([]Position, error)
	Locate(position Position) (NavigationLocation, error)
}

type NavigationLocationKind string

const (
	NavigationLocationUnknown   NavigationLocationKind = "unknown"
	NavigationLocationLayer     NavigationLocationKind = "layer"
	NavigationLocationConnector NavigationLocationKind = "connector"
)

type NavigationLocation struct {
	Kind NavigationLocationKind
	ID   string
}

// PlayerSnapshot is a detached view of a player's authoritative state.
type PlayerSnapshot struct {
	ID                           PlayerID
	Position                     Position
	Destination                  *Position
	Path                         []Position
	NavigationLocation           NavigationLocation
	Speed                        float64
	LastCommandSequence          uint64
	LastCommandTick              uint64
	LastActionSequence           uint64
	LastActionPulse              uint64
	LastActionResult             string
	CombatTarget                 string
	NextAttackPulse              uint64
	AutoApproaching              bool
	AutoPickingUp                bool
	Health                       int
	MaximumHealth                int
	Prayer                       int
	MaximumPrayer                int
	ActivePrayers                []PrayerID
	ActivePrayerSet              PrayerSetID
	LastPrayerSequence           uint64
	SelectedSpell                SpellID
	LastSpellSequence            uint64
	Defeated                     bool
	RespawnAtPulse               uint64
	EquippedWeapon               WeaponID
	EquippedWeaponInstanceID     string
	EquippedWeaponReinforcement  int
	EquippedArmour               map[ArmourSlot]string
	EquippedArmourInstances      map[ArmourSlot]string
	EquippedArmourReinforcements map[ArmourSlot]int
	EquippedSpellbook            string
	ActiveSpellbook              SpellbookID
	LastEquipmentSequence        uint64
	LastEquipmentPulse           uint64
	CombatStyle                  CombatStyleID
	LastCombatStyleSequence      uint64
	DodgeRate                    float64
	CombatStats                  CombatStatsSnapshot
	CombatLevel                  int
	CombatMatchup                CombatMatchupSnapshot
	ExperienceLocked             bool
	LastExperienceLockSequence   uint64
	Inventory                    InventorySnapshot
	AmmunitionPouch              AmmunitionPouchSnapshot
	LastAmmunitionSequence       uint64
	RingRack                     RingRackSnapshot
	LastRingSequence             uint64
	RunePouch                    RunePouchSnapshot
	LastRuneSequence             uint64
	Bank                         BankSnapshot
	LastBankSequence             uint64
	LastMerchantSequence         uint64
	LastPickupSequence           uint64
	LastConsumeSequence          uint64
	NextConsumePulse             uint64
	LastBrewSequence             uint64
	BrewRemaining                int
	NextBrewPulse                uint64
	BrewingResult                string
	LastFletchSequence           uint64
	LastGemCutSequence           uint64
	LastFireSequence             uint64
	FletchRemaining              int
	NextFletchPulse              uint64
	FletchingResult              string
	GatheringSpot                string
	GatheringResult              string
	NextGatherPulse              uint64
	LastGatherSequence           uint64
	Skills                       []SkillSnapshot
	ActiveCombatBoosts           []ActiveCombatBoostSnapshot
}

// Snapshot is a detached view of the authoritative world at one tick.
type Snapshot struct {
	Tick           uint64
	Players        []PlayerSnapshot
	ActionPulse    uint64
	Targets        []TargetSnapshot
	Drops          []DropSnapshot
	Fires          []FireSnapshot
	CombatEvents   []CombatEvent
	GatheringSpots []GatheringSpotSnapshot
	BankChests     []BankChestSnapshot
	Merchants      []MerchantSnapshot
}

type playerState struct {
	id                           PlayerID
	position                     Position
	destination                  *Position
	path                         []Position
	navigationLocation           NavigationLocation
	speed                        float64
	baseSpeed                    float64
	lastCommandSequence          uint64
	lastCommandTick              uint64
	lastActionSequence           uint64
	lastActionPulse              uint64
	lastActionResult             string
	combatTarget                 string
	nextAttackPulse              uint64
	autoApproaching              bool
	autoPickingUp                bool
	spawnPosition                Position
	health                       int
	maximumHealth                int
	prayer                       int
	maximumPrayer                int
	activePrayers                map[PrayerID]bool
	activePrayerSet              PrayerSetID
	prayerDrainProgress          map[PrayerID]uint64
	lastPrayerSequence           uint64
	selectedSpell                SpellID
	lastSpellSequence            uint64
	defeated                     bool
	respawnAtPulse               uint64
	equippedWeapon               WeaponID
	equippedWeaponInstanceID     string
	equippedWeaponReinforcement  int
	equippedArmour               map[ArmourSlot]string
	equippedArmourInstances      map[ArmourSlot]string
	equippedArmourReinforcements map[ArmourSlot]int
	equippedSpellbook            string
	lastEquipmentSequence        uint64
	lastEquipmentPulse           uint64
	combatStyle                  CombatStyleID
	lastCombatStyleSequence      uint64
	experienceLocked             bool
	lastExperienceLockSequence   uint64
	dodgeRate                    float64
	inventory                    inventoryState
	ammunitionPouch              inventoryState
	lastAmmunitionSequence       uint64
	ringRack                     inventoryState
	lastRingSequence             uint64
	runePouch                    inventoryState
	lastRuneSequence             uint64
	bank                         bankState
	lastBankSequence             uint64
	lastMerchantSequence         uint64
	lastPickupSequence           uint64
	lastDropSequence             uint64
	lastInventoryMoveSequence    uint64
	lastConsumeSequence          uint64
	nextConsumePulse             uint64
	combatBoosts                 map[SkillID]activeCombatBoost
	lastBrewSequence             uint64
	brewRemaining                int
	nextBrewPulse                uint64
	brewingResult                string
	lastFletchSequence           uint64
	lastGemCutSequence           uint64
	lastFireSequence             uint64
	fletchRemaining              int
	nextFletchPulse              uint64
	fletchingResult              string
	gatheringSpot                string
	gatheringRecipe              string
	gatheringInstanceID          string
	productionBatch              productionBatch
	gatheringResult              string
	gatheringInteractionPosition Position
	gatheringCompletionPosition  Position
	gatheringTraversal           bool
	nextGatherPulse              uint64
	lastGatherSequence           uint64
	lastDebugSkillSequence       uint64
	skills                       skillState
}

// World owns the active gameplay state. Gameplay state changes happen in Step.
type World struct {
	mu                         sync.RWMutex
	tick                       uint64
	players                    map[PlayerID]*playerState
	pendingJoins               map[PlayerID]*playerState
	pendingLeaves              map[PlayerID]struct{}
	pendingMoves               map[PlayerID]pendingMove
	pendingPickups             map[PlayerID]pendingPickup
	pendingItemDrops           map[PlayerID]DropItemCommand
	pendingInventoryMoves      map[PlayerID]MoveInventoryItemCommand
	pendingConsumables         map[PlayerID]ConsumeItemCommand
	pendingBrews               map[PlayerID]pendingBrew
	pendingFletches            map[PlayerID]pendingFletch
	pendingGathering           map[PlayerID]pendingGather
	pendingBankTransfers       map[PlayerID]BankTransferCommand
	pendingMerchantSales       map[PlayerID]MerchantSellCommand
	pendingAmmunitionTransfers map[PlayerID]AmmunitionTransferCommand
	pendingRingTransfers       map[PlayerID]RingTransferCommand
	pendingRuneTransfers       map[PlayerID]RuneTransferCommand
	pendingExperienceLocks     map[PlayerID]ExperienceLockCommand
	pendingDebugSkills         map[PlayerID]DebugSetSkillCommand
	drops                      map[string]DropSnapshot
	dropLifetimes              map[string]time.Duration
	fires                      map[string]FireSnapshot
	fireLifetimes              map[string]time.Duration
	nextFireID                 uint64
	nextDropID                 uint64
	pathfinder                 Pathfinder
	combat                     *combatState
	rollSource                 RollSource
	gatheringSpots             map[string]gatheringSpotDefinition
	merchantStock              []merchantStockEntry
	debugContentEnabled        bool
}

type pendingMove struct {
	command MoveCommand
	path    []Position
}

type pendingPickup struct {
	command PickupCommand
	path    []Position
	started bool
}

func NewWorld() *World {
	return NewWorldWithRollSource(newRandomRollSource())
}

// NewWorldWithRollSource allows deterministic combat rolls in tests.
func NewWorldWithRollSource(rollSource RollSource) *World {
	if rollSource == nil {
		rollSource = newRandomRollSource()
	}
	return &World{
		players:                    make(map[PlayerID]*playerState),
		pendingJoins:               make(map[PlayerID]*playerState),
		pendingLeaves:              make(map[PlayerID]struct{}),
		pendingMoves:               make(map[PlayerID]pendingMove),
		pendingPickups:             make(map[PlayerID]pendingPickup),
		pendingItemDrops:           make(map[PlayerID]DropItemCommand),
		pendingInventoryMoves:      make(map[PlayerID]MoveInventoryItemCommand),
		pendingConsumables:         make(map[PlayerID]ConsumeItemCommand),
		pendingBrews:               make(map[PlayerID]pendingBrew),
		pendingFletches:            make(map[PlayerID]pendingFletch),
		pendingGathering:           make(map[PlayerID]pendingGather),
		pendingBankTransfers:       make(map[PlayerID]BankTransferCommand),
		pendingMerchantSales:       make(map[PlayerID]MerchantSellCommand),
		pendingAmmunitionTransfers: make(map[PlayerID]AmmunitionTransferCommand),
		pendingRingTransfers:       make(map[PlayerID]RingTransferCommand),
		pendingRuneTransfers:       make(map[PlayerID]RuneTransferCommand),
		pendingExperienceLocks:     make(map[PlayerID]ExperienceLockCommand),
		pendingDebugSkills:         make(map[PlayerID]DebugSetSkillCommand),
		drops:                      make(map[string]DropSnapshot),
		dropLifetimes:              make(map[string]time.Duration),
		fires:                      make(map[string]FireSnapshot),
		fireLifetimes:              make(map[string]time.Duration),
		combat:                     newCombatState(),
		rollSource:                 rollSource,
		gatheringSpots:             newGatheringSpots(),
		debugContentEnabled:        true,
	}
}

func NewWorldWithPathfinder(pathfinder Pathfinder) *World {
	world := NewWorld()
	world.pathfinder = pathfinder
	return world
}

// SetDebugContentEnabled controls whether development-only world fixtures are
// exposed to clients. Local worlds default to enabled for existing workflows.
func (w *World) SetDebugContentEnabled(enabled bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.debugContentEnabled = enabled
}

// AddPlayer registers a player before it enters the active simulation.
func (w *World) AddPlayer(id PlayerID, position Position, speed float64) error {
	if err := validatePlayer(id, position, speed); err != nil {
		return err
	}

	location, err := w.locate(position)
	if err != nil {
		return fmt.Errorf("locate player: %w", err)
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if _, exists := w.players[id]; exists {
		return fmt.Errorf("player %q already exists", id)
	}

	w.players[id] = &playerState{
		id: id, position: position, spawnPosition: position, speed: speed, baseSpeed: speed,
		navigationLocation: location, health: playerMaximumHealth, maximumHealth: playerMaximumHealth,
		prayer: playerMaximumPrayer, maximumPrayer: playerMaximumPrayer,
		activePrayers: make(map[PrayerID]bool), activePrayerSet: PrayerSetCommon, prayerDrainProgress: make(map[PrayerID]uint64),
		equippedWeapon: WeaponUnarmed,
		dodgeRate:      playerDodgeRate, inventory: newEmptyInventory(), ammunitionPouch: newAmmunitionPouch(),
		ringRack: newRingRack(), runePouch: newRunePouch(), bank: newEmptyBank(),
		equippedArmour:               make(map[ArmourSlot]string),
		equippedArmourInstances:      make(map[ArmourSlot]string),
		equippedArmourReinforcements: make(map[ArmourSlot]int),
		combatStyle:                  defaultCombatStyle(WeaponUnarmed), skills: newDefaultSkills(), combatBoosts: make(map[SkillID]activeCombatBoost),
	}
	return nil
}

// QueueJoin records a player to be added on the next simulation step.
func (w *World) QueueJoin(id PlayerID, position Position, speed float64) error {
	if err := validatePlayer(id, position, speed); err != nil {
		return err
	}

	location, err := w.locate(position)
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
		id: id, position: position, spawnPosition: position, speed: speed, baseSpeed: speed,
		navigationLocation: location, health: playerMaximumHealth, maximumHealth: playerMaximumHealth,
		prayer: playerMaximumPrayer, maximumPrayer: playerMaximumPrayer,
		activePrayers: make(map[PrayerID]bool), activePrayerSet: PrayerSetCommon, prayerDrainProgress: make(map[PrayerID]uint64),
		equippedWeapon: WeaponUnarmed,
		dodgeRate:      playerDodgeRate, inventory: newEmptyInventory(), ammunitionPouch: newAmmunitionPouch(),
		ringRack: newRingRack(), runePouch: newRunePouch(), bank: newEmptyBank(),
		equippedArmour:               make(map[ArmourSlot]string),
		equippedArmourInstances:      make(map[ArmourSlot]string),
		equippedArmourReinforcements: make(map[ArmourSlot]int),
		combatStyle:                  defaultCombatStyle(WeaponUnarmed), skills: newDefaultSkills(), combatBoosts: make(map[SkillID]activeCombatBoost),
	}
	return nil
}

// QueueLeave records a player to be removed on the next simulation step.
func (w *World) QueueLeave(id PlayerID) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, active := w.players[id]; !active {
		if _, joining := w.pendingJoins[id]; !joining {
			return fmt.Errorf("player %q does not exist", id)
		}
	}

	w.pendingLeaves[id] = struct{}{}
	delete(w.pendingMoves, id)
	delete(w.pendingPickups, id)
	delete(w.pendingItemDrops, id)
	delete(w.pendingInventoryMoves, id)
	delete(w.pendingConsumables, id)
	delete(w.pendingBrews, id)
	delete(w.pendingFletches, id)
	delete(w.pendingGathering, id)
	delete(w.pendingBankTransfers, id)
	delete(w.pendingMerchantSales, id)
	delete(w.pendingAmmunitionTransfers, id)
	delete(w.pendingRingTransfers, id)
	delete(w.pendingRuneTransfers, id)
	delete(w.pendingExperienceLocks, id)
	delete(w.pendingDebugSkills, id)
	delete(w.combat.pendingTargets, id)
	delete(w.combat.pendingEquipment, id)
	delete(w.combat.pendingStyles, id)
	delete(w.combat.pendingPrayers, id)
	for _, target := range w.combat.dummyTargets() {
		if target.dummy.TargetID == id {
			clearDummyTarget(target.dummy)
		}
	}
	return nil
}

// QueueMove validates and records a click-to-move intention. For each player,
// only the highest sequence received before the next Step is retained.
func (w *World) QueueMove(command MoveCommand) error {
	if command.Sequence == 0 {
		return fmt.Errorf("command sequence must be positive")
	}
	if !command.Destination.valid() {
		return fmt.Errorf("destination must be finite")
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
	if command.Sequence <= player.lastCommandSequence {
		return fmt.Errorf("command sequence %d is not newer than %d", command.Sequence, player.lastCommandSequence)
	}
	if pending, exists := w.pendingMoves[command.PlayerID]; exists && command.Sequence <= pending.command.Sequence {
		return fmt.Errorf("command sequence %d is not newer than pending sequence %d", command.Sequence, pending.command.Sequence)
	}

	path := []Position{command.Destination}
	if w.pathfinder != nil {
		var err error
		path, err = w.pathfinder.FindPath(player.position, command.Destination)
		if err != nil {
			return fmt.Errorf("find path: %w", err)
		}
	}
	if len(path) == 0 {
		return fmt.Errorf("find path: empty path")
	}

	w.pendingMoves[command.PlayerID] = pendingMove{command: command, path: clonePositions(path)}
	delete(w.pendingPickups, command.PlayerID)
	delete(w.pendingGathering, command.PlayerID)
	clearGathering(player, "")
	player.autoPickingUp = false
	return nil
}

// Step advances all active gameplay state by one fixed simulation step.
func (w *World) Step(elapsed time.Duration) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.tick++
	if elapsed <= 0 {
		return
	}
	w.advanceDropLifetimes(elapsed)
	w.advanceFireLifetimes(elapsed)
	w.advanceMerchantStock(elapsed)

	for id, player := range w.pendingJoins {
		w.players[id] = player
		delete(w.pendingJoins, id)
	}
	w.applyPendingExperienceLocks()
	w.applyPendingDebugSkills()

	for id, move := range w.pendingMoves {
		player := w.players[id]
		player.destination = positionPointer(move.command.Destination)
		player.path = clonePositions(move.path)
		player.lastCommandSequence = move.command.Sequence
		player.lastCommandTick = w.tick
		player.autoApproaching = false
		delete(w.pendingMoves, id)
	}
	w.startPendingPickupApproaches()
	w.startPendingGathering()
	w.applyTargetCommands()

	seconds := elapsed.Seconds()
	for _, player := range w.players {
		movePlayer(player, seconds)
		if location, err := w.locate(player.position); err == nil {
			player.navigationLocation = location
		}
	}
	w.moveRoamingEnemies(seconds)
	w.applyPendingInventoryMoves()
	w.applyPendingAmmunitionTransfers()
	w.applyPendingRingTransfers()
	w.applyPendingRuneTransfers()
	w.applyPendingBankTransfers()
	w.applyPendingMerchantSales()
	w.applyPendingPickups()
	w.applyPendingConsumables()
	w.expireCombatBoosts()
	w.applyPendingItemDrops()
	w.advanceCombat(elapsed)

	for id := range w.pendingLeaves {
		delete(w.players, id)
		delete(w.pendingLeaves, id)
	}
}

func (w *World) moveRoamingEnemies(seconds float64) {
	for id, enemy := range w.combat.roamingEnemies {
		if enemy.target.Defeated || len(enemy.path) == 0 {
			continue
		}
		if player, exists := w.players[enemy.target.TargetID]; exists &&
			distance(enemy.target.Position, player.position) <= combatProfileForEnemy(id).AttackRange {
			enemy.path = nil
			continue
		}

		remainingMovement := enemy.speed * seconds
		for remainingMovement > 0 && len(enemy.path) > 0 {
			waypoint := enemy.path[0]
			distanceToWaypoint := distance(enemy.target.Position, waypoint)
			if distanceToWaypoint <= remainingMovement {
				enemy.target.Position = waypoint
				enemy.path = enemy.path[1:]
				remainingMovement -= distanceToWaypoint
				continue
			}
			weight := remainingMovement / distanceToWaypoint
			enemy.target.Position.X += (waypoint.X - enemy.target.Position.X) * weight
			enemy.target.Position.Y += (waypoint.Y - enemy.target.Position.Y) * weight
			enemy.target.Position.Z += (waypoint.Z - enemy.target.Position.Z) * weight
			remainingMovement = 0
		}
	}
}

// Player returns a detached copy of the current player state.
func (w *World) Player(id PlayerID) (PlayerSnapshot, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	player, exists := w.players[id]
	if !exists {
		return PlayerSnapshot{}, false
	}

	return snapshotPlayer(player), true
}

// HasPlayerOrPending reports whether a character still occupies or is waiting
// to enter this world. Connection replacement uses it to avoid racing a new
// join against the previous connection's queued leave.
func (w *World) HasPlayerOrPending(id PlayerID) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if _, exists := w.players[id]; exists {
		return true
	}
	_, pending := w.pendingJoins[id]
	return pending
}

// Snapshot returns a detached, deterministically ordered view of the world.
func (w *World) Snapshot() Snapshot {
	w.mu.RLock()
	defer w.mu.RUnlock()

	bankChests := []BankChestSnapshot{starterBankChest}
	if w.debugContentEnabled {
		bankChests = append(bankChests, testSupplyChests...)
	}
	snapshot := Snapshot{
		Tick:           w.tick,
		Players:        make([]PlayerSnapshot, 0, len(w.players)),
		ActionPulse:    w.combat.pulse,
		Targets:        make([]TargetSnapshot, 0, len(w.combat.targets)),
		Drops:          make([]DropSnapshot, 0, len(w.drops)),
		Fires:          w.fireSnapshots(),
		GatheringSpots: w.gatheringSpotSnapshots(),
		BankChests:     bankChests,
		Merchants:      []MerchantSnapshot{starterMerchant},
		CombatEvents:   append([]CombatEvent(nil), w.combat.events...),
	}
	for _, target := range w.combat.dummyTargets() {
		snapshot.Targets = append(snapshot.Targets, *target.dummy)
	}
	dropIDs := make([]string, 0, len(w.drops))
	for id := range w.drops {
		dropIDs = append(dropIDs, id)
	}
	sort.Strings(dropIDs)
	for _, id := range dropIDs {
		snapshot.Drops = append(snapshot.Drops, w.drops[id])
	}
	for _, player := range w.players {
		playerSnapshot := snapshotPlayer(player)
		playerSnapshot.CombatMatchup = w.combatMatchupForPlayer(player)
		snapshot.Players = append(snapshot.Players, playerSnapshot)
	}
	sort.Slice(snapshot.Players, func(i, j int) bool {
		return snapshot.Players[i].ID < snapshot.Players[j].ID
	})
	return snapshot
}

// Tick returns the number of completed simulation steps.
func (w *World) Tick() uint64 {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.tick
}

func movePlayer(player *playerState, seconds float64) {
	if player.destination == nil || len(player.path) == 0 {
		return
	}

	remainingMovement := player.speed * seconds
	for remainingMovement > 0 && len(player.path) > 0 {
		waypoint := player.path[0]
		dx := waypoint.X - player.position.X
		dy := waypoint.Y - player.position.Y
		dz := waypoint.Z - player.position.Z
		distance := math.Sqrt(dx*dx + dy*dy + dz*dz)

		if distance <= remainingMovement {
			player.position = waypoint
			player.path = player.path[1:]
			remainingMovement -= distance
			continue
		}

		player.position.X += dx / distance * remainingMovement
		player.position.Y += dy / distance * remainingMovement
		player.position.Z += dz / distance * remainingMovement
		remainingMovement = 0
	}

	if len(player.path) == 0 {
		player.destination = nil
	}
}

func snapshotPlayer(player *playerState) PlayerSnapshot {
	snapshot := PlayerSnapshot{
		ID:                           player.id,
		Position:                     player.position,
		Speed:                        player.speed,
		LastCommandSequence:          player.lastCommandSequence,
		LastCommandTick:              player.lastCommandTick,
		NavigationLocation:           player.navigationLocation,
		LastActionSequence:           player.lastActionSequence,
		LastActionPulse:              player.lastActionPulse,
		LastActionResult:             player.lastActionResult,
		CombatTarget:                 player.combatTarget,
		NextAttackPulse:              player.nextAttackPulse,
		AutoApproaching:              player.autoApproaching,
		AutoPickingUp:                player.autoPickingUp,
		Health:                       player.health,
		MaximumHealth:                player.maximumHealth,
		Prayer:                       player.prayer,
		MaximumPrayer:                player.maximumPrayer,
		ActivePrayers:                activePrayerIDs(player),
		ActivePrayerSet:              player.activePrayerSet,
		LastPrayerSequence:           player.lastPrayerSequence,
		SelectedSpell:                player.selectedSpell,
		LastSpellSequence:            player.lastSpellSequence,
		Defeated:                     player.defeated,
		RespawnAtPulse:               player.respawnAtPulse,
		EquippedWeapon:               player.equippedWeapon,
		EquippedWeaponInstanceID:     player.equippedWeaponInstanceID,
		EquippedWeaponReinforcement:  player.equippedWeaponReinforcement,
		EquippedArmour:               cloneEquippedArmour(player.equippedArmour),
		EquippedArmourInstances:      cloneEquippedArmourInstances(player.equippedArmourInstances),
		EquippedArmourReinforcements: cloneEquippedArmourReinforcements(player.equippedArmourReinforcements),
		EquippedSpellbook:            player.equippedSpellbook,
		ActiveSpellbook:              activeSpellbook(player),
		LastEquipmentSequence:        player.lastEquipmentSequence,
		LastEquipmentPulse:           player.lastEquipmentPulse,
		CombatStyle:                  player.combatStyle,
		LastCombatStyleSequence:      player.lastCombatStyleSequence,
		DodgeRate:                    playerDodgeChance(dummyAccuracyScore, player),
		CombatStats:                  derivedCombatStats(player),
		CombatLevel:                  playerCombatLevel(player),
		ExperienceLocked:             player.experienceLocked,
		LastExperienceLockSequence:   player.lastExperienceLockSequence,
		Inventory:                    snapshotInventory(player.inventory),
		AmmunitionPouch:              snapshotAmmunitionPouch(player.ammunitionPouch),
		LastAmmunitionSequence:       player.lastAmmunitionSequence,
		RingRack:                     snapshotRingRack(player.ringRack),
		LastRingSequence:             player.lastRingSequence,
		RunePouch:                    snapshotRunePouch(player.runePouch),
		LastRuneSequence:             player.lastRuneSequence,
		Bank:                         snapshotBank(player.bank),
		LastBankSequence:             player.lastBankSequence,
		LastMerchantSequence:         player.lastMerchantSequence,
		LastPickupSequence:           player.lastPickupSequence,
		LastConsumeSequence:          player.lastConsumeSequence,
		NextConsumePulse:             player.nextConsumePulse,
		LastBrewSequence:             player.lastBrewSequence,
		BrewRemaining:                player.brewRemaining,
		NextBrewPulse:                player.nextBrewPulse,
		BrewingResult:                player.brewingResult,
		LastFletchSequence:           player.lastFletchSequence,
		LastGemCutSequence:           player.lastGemCutSequence,
		LastFireSequence:             player.lastFireSequence,
		FletchRemaining:              player.fletchRemaining,
		NextFletchPulse:              player.nextFletchPulse,
		FletchingResult:              player.fletchingResult,
		GatheringSpot:                player.gatheringSpot,
		GatheringResult:              player.gatheringResult,
		NextGatherPulse:              player.nextGatherPulse,
		LastGatherSequence:           player.lastGatherSequence,
		Skills:                       snapshotSkills(player.skills),
		ActiveCombatBoosts:           activeCombatBoostSnapshots(player),
	}
	if player.destination != nil {
		snapshot.Destination = positionPointer(*player.destination)
	}
	snapshot.Path = clonePositions(player.path)
	return snapshot
}

func validatePlayer(id PlayerID, position Position, speed float64) error {
	if id == "" {
		return fmt.Errorf("player ID is required")
	}
	if !position.valid() {
		return fmt.Errorf("player position must be finite")
	}
	if math.IsNaN(speed) || math.IsInf(speed, 0) || speed <= 0 {
		return fmt.Errorf("player speed must be finite and positive")
	}
	return nil
}

func (p Position) valid() bool {
	return !math.IsNaN(p.X) && !math.IsInf(p.X, 0) &&
		!math.IsNaN(p.Y) && !math.IsInf(p.Y, 0) &&
		!math.IsNaN(p.Z) && !math.IsInf(p.Z, 0)
}

func positionPointer(position Position) *Position {
	return &position
}

func clonePositions(positions []Position) []Position {
	return append([]Position(nil), positions...)
}

func (w *World) locate(position Position) (NavigationLocation, error) {
	if w.pathfinder == nil {
		return NavigationLocation{Kind: NavigationLocationUnknown}, nil
	}
	return w.pathfinder.Locate(position)
}
