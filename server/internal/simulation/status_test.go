package simulation

import (
	"math"
	"strings"
	"testing"
)

func TestStunRefreshesAndExpires(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	player := world.players["player-one"]
	initialRevision := player.statusesRevision

	if applied := world.applyStatus(player, StatusStun, "tyrant", 3); !applied {
		t.Fatal("applyStatus(stun) = false")
	}
	if !statusActive(player, StatusStun, world.combat.pulse) {
		t.Fatal("stun was not active immediately after application")
	}
	if player.statusesRevision != initialRevision+1 {
		t.Fatalf("status revision = %d, want %d", player.statusesRevision, initialRevision+1)
	}
	statuses := activeStatusSnapshots(player)
	if len(statuses) != 1 || statuses[0].EffectID != StatusStun || statuses[0].SourceID != "tyrant" || statuses[0].ExpiresAtPulse != 3 {
		t.Fatalf("active status snapshots = %#v", statuses)
	}

	world.combat.pulse = 1
	world.applyStatus(player, StatusStun, "second-source", 1)
	if got := player.activeStatuses[StatusStun].ExpiresAtPulse; got != 3 {
		t.Fatalf("shorter refresh reduced expiry to %d, want 3", got)
	}
	if len(player.activeStatuses) != 1 {
		t.Fatalf("refresh created %d status entries, want 1", len(player.activeStatuses))
	}

	world.combat.pulse = 4
	world.expireStatuses()
	if statusActive(player, StatusStun, world.combat.pulse) || len(player.activeStatuses) != 0 {
		t.Fatalf("expired statuses = %#v", player.activeStatuses)
	}
}

func TestStunCancelsMovementAndTimedActions(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	player := world.players["player-one"]
	destination := Position{X: 5}
	player.destination = &destination
	player.path = []Position{destination}
	player.autoApproaching = true
	player.autoPickingUp = true
	player.gatheringSpot = "copper-rock"
	player.brewRemaining = 2
	player.nextBrewPulse = 4
	player.fletchRemaining = 2
	player.nextFletchPulse = 4

	world.applyStatus(player, StatusStun, "tyrant", 3)

	if player.destination != nil || len(player.path) != 0 || player.autoApproaching || player.autoPickingUp {
		t.Fatalf("movement was not cancelled: destination=%v path=%v approach=%v pickup=%v", player.destination, player.path, player.autoApproaching, player.autoPickingUp)
	}
	if player.gatheringSpot != "" || player.gatheringResult != "stunned" {
		t.Fatalf("gathering state = spot %q result %q", player.gatheringSpot, player.gatheringResult)
	}
	if player.brewRemaining != 0 || player.nextBrewPulse != 0 || player.brewingResult != "brewing_stopped" {
		t.Fatalf("brewing state = remaining %d next %d result %q", player.brewRemaining, player.nextBrewPulse, player.brewingResult)
	}
	if player.fletchRemaining != 0 || player.nextFletchPulse != 0 || player.fletchingResult != "fletching_stopped" {
		t.Fatalf("fletching state = remaining %d next %d result %q", player.fletchRemaining, player.nextFletchPulse, player.fletchingResult)
	}
}

func TestStunBlocksActionsButAllowsConsumablesAndPrayers(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	player := world.players["player-one"]
	world.applyStatus(player, StatusStun, "tyrant", 3)

	assertStunned := func(name string, err error) {
		t.Helper()
		if err == nil || !strings.Contains(err.Error(), "stunned") {
			t.Fatalf("%s error = %v, want stunned", name, err)
		}
	}
	assertStunned("move", world.QueueMove(MoveCommand{
		PlayerID: "player-one", Sequence: 1, Destination: Position{X: 1},
	}))
	assertStunned("attack", world.QueueTarget(TargetCommand{
		PlayerID: "player-one", Sequence: 1, TargetID: hillWardenTargetID,
	}))
	var spotID string
	for id := range world.gatheringSpots {
		spotID = id
		break
	}
	assertStunned("gather", world.QueueGather(GatherCommand{
		PlayerID: "player-one", Sequence: 1, SpotID: spotID,
	}))
	assertStunned("brew", world.QueueBrew(BrewCommand{
		PlayerID: "player-one", Sequence: 1, ItemIDs: []string{ItemVialOfWater, ItemSunleaf}, Quantity: 1,
	}))
	assertStunned("fletch", world.QueueFletch(FletchCommand{
		PlayerID: "player-one", Sequence: 1, RecipeID: FletchArrowShafts, Quantity: 1,
	}))

	if err := world.QueuePrayer(PrayerCommand{
		PlayerID: "player-one", Sequence: 1, PrayerID: PrayerAccuracy, Active: true,
	}); err != nil {
		t.Fatalf("QueuePrayer() while stunned returned %v", err)
	}
	player.health = player.maximumHealth - 1
	if !player.inventory.add(ItemCookedMinnow, 1) {
		t.Fatal("could not add cooked minnow")
	}
	foodSlot := -1
	for _, item := range player.inventory.items {
		if item.ItemID == ItemCookedMinnow {
			foodSlot = item.Slot
			break
		}
	}
	if err := world.QueueConsumeItem(ConsumeItemCommand{
		PlayerID: "player-one", Sequence: 1, ItemID: ItemCookedMinnow, InventorySlot: foodSlot,
	}); err != nil {
		t.Fatalf("QueueConsumeItem() while stunned returned %v", err)
	}
}

func TestDefeatClearsStun(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	player := world.players["player-one"]
	world.applyStatus(player, StatusStun, "tyrant", 3)

	world.defeatPlayer(player)

	if len(player.activeStatuses) != 0 || statusActive(player, StatusStun, world.combat.pulse) {
		t.Fatalf("statuses survived defeat: %#v", player.activeStatuses)
	}
}

func TestRootBlocksOnlyMovementAndPreservesTimedActions(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", world.combat.targets[dummyTargetID].Position, 4)
	player := world.players["player-one"]
	destination := Position{X: player.position.X + 5, Y: player.position.Y, Z: player.position.Z}
	player.destination = &destination
	player.path = []Position{destination}
	player.autoApproaching = true
	player.gatheringSpot = "nearby-resource"
	player.gatheringResult = "gathering"
	player.brewRemaining = 2
	player.fletchRemaining = 2

	world.applyStatus(player, StatusRoot, "test", 3)

	if player.destination != nil || len(player.path) != 0 || player.autoApproaching {
		t.Fatalf("root did not cancel movement: destination=%v path=%v approaching=%v", player.destination, player.path, player.autoApproaching)
	}
	if player.gatheringSpot != "nearby-resource" || player.gatheringResult != "gathering" ||
		player.brewRemaining != 2 || player.fletchRemaining != 2 {
		t.Fatalf("root cancelled non-movement actions: gather=%q result=%q brew=%d fletch=%d", player.gatheringSpot, player.gatheringResult, player.brewRemaining, player.fletchRemaining)
	}
	if err := world.QueueMove(MoveCommand{PlayerID: "player-one", Sequence: 1, Destination: destination}); err == nil || !strings.Contains(err.Error(), "rooted") {
		t.Fatalf("QueueMove() while rooted error = %v", err)
	}
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID}); err != nil {
		t.Fatalf("QueueTarget() while rooted returned %v", err)
	}
	if err := statusActionError(player, statusActionCast, world.combat.pulse); err != nil {
		t.Fatalf("casting while rooted returned %v", err)
	}
}

func TestSlowUsesConfiguredMovementMultiplierAndRestoresCleanly(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	player := world.players["player-one"]
	baseSpeed := player.speed

	world.applyStatus(player, StatusSlow, "test", 3)
	if got, want := effectivePlayerMovementSpeed(player), baseSpeed*slowMovementSpeedMultiplier; math.Abs(got-want) > 0.000_001 {
		t.Fatalf("effective slowed speed = %.2f, want %.2f", got, want)
	}
	if got, want := snapshotPublicPlayer(player).Speed, baseSpeed*slowMovementSpeedMultiplier; math.Abs(got-want) > 0.000_001 {
		t.Fatalf("public slowed speed = %.2f, want %.2f", got, want)
	}
	destination := Position{X: 10}
	player.destination = &destination
	player.path = []Position{destination}
	movePlayer(player, 1)
	if got, want := player.position.X, baseSpeed*slowMovementSpeedMultiplier; math.Abs(got-want) > 0.000_001 {
		t.Fatalf("distance moved while slowed = %.2f, want %.2f", got, want)
	}

	world.combat.pulse = 4
	world.expireStatuses()
	if got := effectivePlayerMovementSpeed(player); math.Abs(got-baseSpeed) > 0.000_001 {
		t.Fatalf("speed after slow expired = %.2f, want %.2f", got, baseSpeed)
	}
	if err := statusActionError(player, statusActionMove, world.combat.pulse); err != nil {
		t.Fatalf("slow incorrectly blocked movement: %v", err)
	}
}

func TestSilenceBlocksSpellsButFallsBackToWeaponAttacks(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.5))
	mustAddPlayer(t, world, "player-one", world.combat.targets[dummyTargetID].Position, 4)
	mustEquipStandardSpellbook(t, world, "player-one")
	player := world.players["player-one"]
	world.applyStatus(player, StatusSilence, "test", 3)

	if err := world.QueueSpell(SpellCommand{PlayerID: "player-one", Sequence: 1, SpellID: SpellWindBolt}); err == nil || !strings.Contains(err.Error(), "silenced") {
		t.Fatalf("QueueSpell() while silenced error = %v", err)
	}
	player.selectedSpell = SpellWindBolt
	if err := world.QueueSpell(SpellCommand{PlayerID: "player-one", Sequence: 1}); err != nil {
		t.Fatalf("clearing a spell while silenced returned %v", err)
	}
	delete(world.combat.pendingSpells, player.id)
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID}); err != nil {
		t.Fatalf("weapon target while silenced returned %v", err)
	}
	delete(world.combat.pendingTargets, player.id)
	player.combatTarget = dummyTargetID
	player.nextAttackPulse = 0
	world.resolveAutoAttacks()
	if len(world.combat.events) == 0 {
		t.Fatal("silenced player did not perform a weapon attack")
	}
	event := world.combat.events[len(world.combat.events)-1]
	if event.SourceID != string(player.id) || event.TargetID != dummyTargetID || event.SpellID != "" {
		t.Fatalf("silenced fallback attack event = %+v", event)
	}
	if player.selectedSpell != SpellWindBolt {
		t.Fatalf("silence permanently cleared selected spell %q", player.selectedSpell)
	}
}
