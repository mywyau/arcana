package simulation

import (
	"math"
	"testing"
	"time"
)

func TestWorldMovesPlayerContinuouslyTowardDestination(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	mustQueueMove(t, world, "player-one", 1, Position{X: 3, Z: 4})

	world.Step(500 * time.Millisecond)

	player := mustPlayer(t, world, "player-one")
	assertPosition(t, player.Position, Position{X: 1.2, Z: 1.6})
	if player.Destination == nil {
		t.Fatal("Destination is nil before the player arrives")
	}
}

func TestWorldMovesPlayerContinuouslyInThreeDimensions(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 7)
	mustQueueMove(t, world, "player-one", 1, Position{X: 2, Y: 3, Z: 6})

	world.Step(500 * time.Millisecond)

	player := mustPlayer(t, world, "player-one")
	assertPosition(t, player.Position, Position{X: 1, Y: 1.5, Z: 3})
}

func TestWorldStopsPlayerAtDestinationWithoutOvershooting(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	mustQueueMove(t, world, "player-one", 1, Position{X: 1})

	world.Step(time.Second)

	player := mustPlayer(t, world, "player-one")
	assertPosition(t, player.Position, Position{X: 1})
	if player.Destination != nil {
		t.Fatalf("Destination = %+v after arrival, want nil", player.Destination)
	}
}

func TestWorldMovesThroughPathfinderWaypoints(t *testing.T) {
	pathfinder := fixedPathfinder{path: []Position{{X: 3}, {X: 3, Y: 4}}}
	world := NewWorldWithPathfinder(pathfinder)
	mustAddPlayer(t, world, "player-one", Position{}, 5)
	mustQueueMove(t, world, "player-one", 1, Position{X: 3, Y: 4})

	world.Step(time.Second)

	player := mustPlayer(t, world, "player-one")
	assertPosition(t, player.Position, Position{X: 3, Y: 2})
	if got, want := len(player.Path), 1; got != want {
		t.Fatalf("len(Player().Path) = %d, want %d", got, want)
	}
}

func TestWorldUpdatesNavigationLocationWhileMoving(t *testing.T) {
	pathfinder := transitionPathfinder{}
	world := NewWorldWithPathfinder(pathfinder)
	mustAddPlayer(t, world, "player-one", Position{}, 1)
	mustQueueMove(t, world, "player-one", 1, Position{Y: 3})

	if got := mustPlayer(t, world, "player-one").NavigationLocation; got.ID != "ground" {
		t.Fatalf("initial NavigationLocation = %+v, want ground", got)
	}
	world.Step(time.Second)
	if got := mustPlayer(t, world, "player-one").NavigationLocation; got.Kind != NavigationLocationConnector {
		t.Fatalf("mid-route NavigationLocation = %+v, want connector", got)
	}
	world.Step(2 * time.Second)
	if got := mustPlayer(t, world, "player-one").NavigationLocation; got.ID != "upper" {
		t.Fatalf("final NavigationLocation = %+v, want upper", got)
	}
}

func TestWorldUsesLatestDestinationIntention(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 2)
	mustQueueMove(t, world, "player-one", 1, Position{X: 10})
	mustQueueMove(t, world, "player-one", 2, Position{Z: 10})

	world.Step(500 * time.Millisecond)

	player := mustPlayer(t, world, "player-one")
	assertPosition(t, player.Position, Position{Z: 1})
	if got, want := player.LastCommandSequence, uint64(2); got != want {
		t.Fatalf("LastCommandSequence = %d, want %d", got, want)
	}
}

func TestWorldRejectsInvalidMovementInput(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 2)

	if err := world.QueueMove(MoveCommand{PlayerID: "missing", Sequence: 1}); err == nil {
		t.Fatal("QueueMove() error = nil for missing player")
	}
	if err := world.QueueMove(MoveCommand{PlayerID: "player-one", Sequence: 1, Destination: Position{X: math.NaN()}}); err == nil {
		t.Fatal("QueueMove() error = nil for non-finite destination")
	}
	if err := world.QueueMove(MoveCommand{PlayerID: "player-one", Sequence: 1, Destination: Position{Y: math.Inf(1)}}); err == nil {
		t.Fatal("QueueMove() error = nil for non-finite destination height")
	}
	if err := world.QueueMove(MoveCommand{PlayerID: "player-one", Sequence: 0}); err == nil {
		t.Fatal("QueueMove() error = nil for zero sequence")
	}
}

func TestWorldRejectsStaleAndDuplicateMovementCommands(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 2)
	mustQueueMove(t, world, "player-one", 2, Position{X: 10})

	if err := world.QueueMove(MoveCommand{PlayerID: "player-one", Sequence: 1, Destination: Position{Z: 10}}); err == nil {
		t.Fatal("QueueMove() error = nil for command older than pending command")
	}
	if err := world.QueueMove(MoveCommand{PlayerID: "player-one", Sequence: 2, Destination: Position{Z: 10}}); err == nil {
		t.Fatal("QueueMove() error = nil for duplicate pending command")
	}

	world.Step(50 * time.Millisecond)

	if err := world.QueueMove(MoveCommand{PlayerID: "player-one", Sequence: 2, Destination: Position{Z: 10}}); err == nil {
		t.Fatal("QueueMove() error = nil for already applied command")
	}
}

func TestWorldSnapshotIsOrderedAndDetached(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-b", Position{X: 2}, 2)
	mustAddPlayer(t, world, "player-a", Position{}, 2)
	mustQueueMove(t, world, "player-a", 7, Position{Z: 10})
	world.Step(500 * time.Millisecond)

	snapshot := world.Snapshot()
	if got, want := snapshot.Tick, uint64(1); got != want {
		t.Fatalf("Snapshot().Tick = %d, want %d", got, want)
	}
	if got, want := len(snapshot.Players), 2; got != want {
		t.Fatalf("len(Snapshot().Players) = %d, want %d", got, want)
	}
	if snapshot.Players[0].ID != "player-a" || snapshot.Players[1].ID != "player-b" {
		t.Fatalf("Snapshot().Players order = [%q, %q], want [player-a, player-b]", snapshot.Players[0].ID, snapshot.Players[1].ID)
	}
	if got, want := snapshot.Players[0].LastCommandSequence, uint64(7); got != want {
		t.Fatalf("LastCommandSequence = %d, want %d", got, want)
	}
	if got, want := snapshot.Players[0].LastCommandTick, uint64(1); got != want {
		t.Fatalf("LastCommandTick = %d, want %d", got, want)
	}
	player := snapshot.Players[0]
	if player.Prayer != 100 || player.MaximumPrayer != 100 {
		t.Fatalf("starting prayer = %d/%d, want 100/100", player.Prayer, player.MaximumPrayer)
	}

	snapshot.Players[0].Position = Position{X: 999}
	*snapshot.Players[0].Destination = Position{Z: 999}
	snapshot.Players[0].Path[0] = Position{Y: 999}
	fresh := world.Snapshot()
	assertPosition(t, fresh.Players[0].Position, Position{Z: 1})
	if fresh.Players[0].Destination == nil {
		t.Fatal("fresh snapshot destination is nil")
	}
	assertPosition(t, *fresh.Players[0].Destination, Position{Z: 10})
	assertPosition(t, fresh.Players[0].Path[0], Position{Z: 10})
}

func TestWorldAppliesQueuedPlayerLifecycleOnStep(t *testing.T) {
	world := NewWorld()
	if err := world.QueueJoin("player-one", Position{X: 2}, 4); err != nil {
		t.Fatalf("QueueJoin() returned an unexpected error: %v", err)
	}
	mustQueueMove(t, world, "player-one", 1, Position{X: 6})

	if _, exists := world.Player("player-one"); exists {
		t.Fatal("player exists before queued join is applied")
	}

	world.Step(500 * time.Millisecond)
	player := mustPlayer(t, world, "player-one")
	assertPosition(t, player.Position, Position{X: 4})

	if err := world.QueueLeave("player-one"); err != nil {
		t.Fatalf("QueueLeave() returned an unexpected error: %v", err)
	}
	world.Step(50 * time.Millisecond)
	if _, exists := world.Player("player-one"); exists {
		t.Fatal("player exists after queued leave is applied")
	}
}

func TestWorldStepAdvancesTick(t *testing.T) {
	world := NewWorld()

	world.Step(time.Second)
	world.Step(time.Second)

	if got, want := world.Tick(), uint64(2); got != want {
		t.Fatalf("Tick() = %d, want %d", got, want)
	}
}

func mustAddPlayer(t *testing.T, world *World, id PlayerID, position Position, speed float64) {
	t.Helper()
	if err := world.AddPlayer(id, position, speed); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	// Most simulation tests exercise combat rather than character creation.
	// Seed their required testing loadout explicitly now that production
	// characters start unarmed with an empty inventory.
	player := world.players[id]
	for _, itemID := range []string{
		string(WeaponSword), string(WeaponDagger), string(WeaponBow), string(WeaponMace),
		string(WeaponBattleAxe), string(WeaponStaff), ItemAirRune,
	} {
		quantity := 1
		if itemID == ItemAirRune {
			quantity = 50
		}
		if !player.inventory.add(itemID, quantity) {
			t.Fatalf("could not add test item %q", itemID)
		}
	}
	sword, _ := itemDefinition(string(WeaponSword))
	swordItem, _ := inventoryEquipment(player.inventory, sword, resolveOwnedInstanceID(player.inventory, string(WeaponSword), ""))
	player.inventory.removeInstance(swordItem.InstanceID)
	setEquippedItem(player, sword, swordItem)
}

func mustEquipStandardSpellbook(t *testing.T, world *World, id PlayerID) {
	t.Helper()
	player, exists := world.players[id]
	if !exists {
		t.Fatalf("player %q does not exist", id)
	}
	if !player.inventory.add(ItemStandardSpellbook, 1) {
		t.Fatalf("could not add %q to player %q", ItemStandardSpellbook, id)
	}
	player.equippedSpellbook = ItemStandardSpellbook
}

func mustQueueMove(t *testing.T, world *World, id PlayerID, sequence uint64, destination Position) {
	t.Helper()
	command := MoveCommand{PlayerID: id, Sequence: sequence, Destination: destination}
	if err := world.QueueMove(command); err != nil {
		t.Fatalf("QueueMove() returned an unexpected error: %v", err)
	}
}

func mustPlayer(t *testing.T, world *World, id PlayerID) PlayerSnapshot {
	t.Helper()
	player, exists := world.Player(id)
	if !exists {
		t.Fatalf("Player(%q) does not exist", id)
	}
	return player
}

func assertPosition(t *testing.T, got, want Position) {
	t.Helper()
	const tolerance = 0.000_001
	if math.Abs(got.X-want.X) > tolerance ||
		math.Abs(got.Y-want.Y) > tolerance ||
		math.Abs(got.Z-want.Z) > tolerance {
		t.Fatalf("position = %+v, want %+v", got, want)
	}
}

type fixedPathfinder struct {
	path []Position
}

func (f fixedPathfinder) FindPath(_, _ Position) ([]Position, error) {
	return f.path, nil
}

func (f fixedPathfinder) FindPathWithinRange(_, _ Position, _ float64) ([]Position, error) {
	return f.path, nil
}

func (f fixedPathfinder) Locate(_ Position) (NavigationLocation, error) {
	return NavigationLocation{Kind: NavigationLocationUnknown}, nil
}

type transitionPathfinder struct{}

func (transitionPathfinder) FindPath(_, destination Position) ([]Position, error) {
	return []Position{destination}, nil
}

func (transitionPathfinder) FindPathWithinRange(_, target Position, _ float64) ([]Position, error) {
	return []Position{target}, nil
}

func (transitionPathfinder) Locate(position Position) (NavigationLocation, error) {
	switch {
	case position.Y <= 0:
		return NavigationLocation{Kind: NavigationLocationLayer, ID: "ground"}, nil
	case position.Y >= 3:
		return NavigationLocation{Kind: NavigationLocationLayer, ID: "upper"}, nil
	default:
		return NavigationLocation{Kind: NavigationLocationConnector, ID: "ramp"}, nil
	}
}
