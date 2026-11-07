package navigation

import (
	"math"
	"testing"
	"time"

	"github.com/mywyau/arcana/server/internal/simulation"
)

func TestHillWardenMeleeEngagementStartsApproach(t *testing.T) {
	world := simulation.NewWorldWithPathfinder(NewPrototypeMap())
	start := simulation.Position{X: hillCenterX - 6, Z: hillCenterZ}
	if err := world.AddPlayer("player-one", start, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	if err := world.QueueTarget(simulation.TargetCommand{
		PlayerID: "player-one",
		Sequence: 1,
		TargetID: "hill_warden",
	}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}

	world.Step(600 * time.Millisecond)
	player, exists := world.Player("player-one")
	if !exists {
		t.Fatal("Player() did not return the engaged player")
	}
	if player.LastActionResult != "approaching" || !player.AutoApproaching || player.Destination == nil {
		t.Fatalf("Hill Warden melee engagement did not start an approach: %+v", player)
	}
}

func TestStarterFishingSpotCanBeApproachedOnPrototypeMap(t *testing.T) {
	world := simulation.NewWorldWithPathfinder(NewPrototypeMap())
	if err := world.AddPlayer("player-one", simulation.Position{}, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	spots := world.Snapshot().GatheringSpots
	if len(spots) == 0 {
		t.Fatal("starter gathering spot is missing")
	}
	if err := world.QueueGather(simulation.GatherCommand{
		PlayerID: "player-one", Sequence: 1, SpotID: spots[0].ID,
	}); err != nil {
		t.Fatalf("QueueGather() returned an unexpected pathfinding error: %v", err)
	}
	world.Step(50 * time.Millisecond)
	player, _ := world.Player("player-one")
	if player.GatheringSpot != spots[0].ID || player.Destination == nil {
		t.Fatalf("fishing approach did not start: %+v", player)
	}
}

func TestPrototypeMapFindsMeleeRangePathOnHill(t *testing.T) {
	world := NewPrototypeMap()
	start := simulation.Position{X: hillCenterX - 6, Z: hillCenterZ}
	target := simulation.Position{X: hillCenterX, Y: hillHeight, Z: hillCenterZ}
	const meleeApproachRange = 1.25

	path, err := world.FindPathWithinRange(start, target, meleeApproachRange)
	if err != nil {
		t.Fatalf("FindPathWithinRange() returned an unexpected error: %v", err)
	}
	if len(path) == 0 {
		t.Fatal("FindPathWithinRange() returned an empty path")
	}
	destination := path[len(path)-1]
	if got := positionDistance(destination, target); got > meleeApproachRange {
		t.Fatalf("destination distance = %.3f, want at most %.3f", got, meleeApproachRange)
	}
	if math.Abs(destination.Y-target.Y) <= heightTolerance {
		t.Fatalf("destination height = %.3f, want projected hill surface below target height %.3f", destination.Y, target.Y)
	}
	if _, err := world.Locate(destination); err != nil {
		t.Fatalf("combat destination is not on navigation surface: %v", err)
	}
}

func TestFindPathWithinRangeRoutesAroundBlockedApproach(t *testing.T) {
	world := NewLayeredMapWithObstacles(
		[]Layer{{ID: "ground", MinimumX: -4, MaximumX: 4, MinimumZ: -4, MaximumZ: 4}},
		nil,
		[]Obstacle{{LayerID: "ground", MinimumX: 0.25, MaximumX: 1.5, MinimumZ: -0.4, MaximumZ: 0.4}},
	)
	start := simulation.Position{X: -2}
	target := simulation.Position{X: 2}
	const approachRange = 1.25

	path, err := world.FindPathWithinRange(start, target, approachRange)
	if err != nil {
		t.Fatalf("FindPathWithinRange() returned an unexpected error: %v", err)
	}
	if len(path) == 0 {
		t.Fatal("FindPathWithinRange() returned an empty path")
	}
	if got := positionDistance(path[len(path)-1], target); got > approachRange {
		t.Fatalf("destination distance = %.3f, want at most %.3f", got, approachRange)
	}
	if len(path) < 2 {
		t.Fatalf("FindPathWithinRange() = %+v, want route around obstacle", path)
	}
}

func TestPrototypeMapRoutesBetweenOverlappingLayers(t *testing.T) {
	path, err := NewPrototypeMap().FindPath(
		simulation.Position{},
		simulation.Position{X: -4, Y: 3, Z: 3},
	)
	if err != nil {
		t.Fatalf("FindPath() returned an unexpected error: %v", err)
	}
	groundLink := indexOf(path, simulation.Position{X: 4, Y: 0})
	upperLink := indexOf(path, simulation.Position{X: 2, Y: 3})
	if groundLink < 0 || upperLink <= groundLink {
		t.Fatalf("FindPath() = %+v, want ordered ramp endpoints", path)
	}
	if got, want := path[len(path)-1], (simulation.Position{X: -4, Y: 3, Z: 3}); got != want {
		t.Fatalf("last waypoint = %+v, want %+v", got, want)
	}
}

func TestPrototypeMapRejectsUnwalkableDestination(t *testing.T) {
	_, err := NewPrototypeMap().FindPath(simulation.Position{}, simulation.Position{Y: 10})
	if err == nil {
		t.Fatal("FindPath() error = nil for unwalkable destination")
	}
}

func TestPrototypeMapUsesExpandedGroundBounds(t *testing.T) {
	world := NewPrototypeMap()
	inside := simulation.Position{X: 19.5, Z: -19.5}
	location, err := world.Locate(inside)
	if err != nil {
		t.Fatalf("Locate(%+v) returned an unexpected error: %v", inside, err)
	}
	if location.Kind != simulation.NavigationLocationLayer || location.ID != "ground" {
		t.Fatalf("Locate(%+v) = %+v, want ground layer", inside, location)
	}

	outside := simulation.Position{X: 20.5}
	if _, err := world.Locate(outside); err == nil {
		t.Fatalf("Locate(%+v) error = nil outside expanded ground", outside)
	}
}

func TestPrototypeMapRoutesUpAndDownHill(t *testing.T) {
	world := NewPrototypeMap()
	bottom := simulation.Position{X: hillCenterX - hillRadius, Z: hillCenterZ}
	top := simulation.Position{X: hillCenterX, Y: hillHeight, Z: hillCenterZ}

	uphill, err := world.FindPath(bottom, top)
	if err != nil {
		t.Fatalf("uphill FindPath() returned an unexpected error: %v", err)
	}
	if len(uphill) < 2 {
		t.Fatalf("uphill FindPath() = %+v, want surface-following waypoints", uphill)
	}
	for index := 1; index < len(uphill); index++ {
		if uphill[index].Y < uphill[index-1].Y {
			t.Fatalf("uphill waypoint heights decreased: %+v", uphill)
		}
	}
	if got := uphill[len(uphill)-1]; got != top {
		t.Fatalf("uphill final waypoint = %+v, want %+v", got, top)
	}

	downhill, err := world.FindPath(top, bottom)
	if err != nil {
		t.Fatalf("downhill FindPath() returned an unexpected error: %v", err)
	}
	for index := 1; index < len(downhill); index++ {
		if downhill[index].Y > downhill[index-1].Y {
			t.Fatalf("downhill waypoint heights increased: %+v", downhill)
		}
	}
}

func TestLayeredMapRejectsRouteAcrossExcessiveSlope(t *testing.T) {
	steepHeight := func(x, _ float64) float64 {
		if x == 0 {
			return 2
		}
		return 0
	}
	world := NewLayeredMap([]Layer{{
		ID:           "steep",
		MinimumX:     -1,
		MaximumX:     1,
		MinimumZ:     0,
		MaximumZ:     0,
		Elevation:    steepHeight,
		MaximumSlope: 0.5,
	}}, nil)

	_, err := world.FindPath(simulation.Position{X: -1}, simulation.Position{X: 1})
	if err == nil {
		t.Fatal("FindPath() error = nil across an excessive slope")
	}
}

func TestPrototypeMapLocatesLayersAndConnector(t *testing.T) {
	world := NewPrototypeMap()
	tests := []struct {
		name     string
		position simulation.Position
		kind     simulation.NavigationLocationKind
		id       string
	}{
		{name: "ground", position: simulation.Position{}, kind: simulation.NavigationLocationLayer, id: "ground"},
		{name: "connector", position: simulation.Position{X: 3, Y: 1.5}, kind: simulation.NavigationLocationConnector, id: "ground_upper_ramp"},
		{name: "upper", position: simulation.Position{X: -1, Y: 3}, kind: simulation.NavigationLocationLayer, id: "upper"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			location, err := world.Locate(test.position)
			if err != nil {
				t.Fatalf("Locate() returned an unexpected error: %v", err)
			}
			if location.Kind != test.kind || location.ID != test.id {
				t.Fatalf("Locate() = %+v, want kind %q and ID %q", location, test.kind, test.id)
			}
		})
	}
}

func TestPrototypeMapRejectsObstacleTopAsConnector(t *testing.T) {
	_, err := NewPrototypeMap().FindPath(
		simulation.Position{},
		simulation.Position{X: 1, Y: 2},
	)
	if err == nil {
		t.Fatal("FindPath() error = nil for obstacle top")
	}
}

func TestPrototypeMapReroutesFromConnector(t *testing.T) {
	path, err := NewPrototypeMap().FindPath(
		simulation.Position{X: 3, Y: 1.5},
		simulation.Position{X: -4, Y: 3, Z: 3},
	)
	if err != nil {
		t.Fatalf("FindPath() returned an unexpected error: %v", err)
	}
	if got, want := path[0], (simulation.Position{X: 2, Y: 3}); got != want {
		t.Fatalf("FindPath()[0] = %+v, want %+v", got, want)
	}
}

func TestPrototypeMapRoutesToAndAlongConnector(t *testing.T) {
	world := NewPrototypeMap()
	destination := simulation.Position{X: 3, Y: 1.5, Z: 0.5}
	path, err := world.FindPath(simulation.Position{}, destination)
	if err != nil {
		t.Fatalf("FindPath() returned an unexpected error: %v", err)
	}
	if indexOf(path, simulation.Position{X: 4}) < 0 {
		t.Fatalf("FindPath() = %+v, want lower ramp endpoint", path)
	}
	if got := path[len(path)-1]; got != destination {
		t.Fatalf("last waypoint = %+v, want %+v", got, destination)
	}

	next := simulation.Position{X: 2.5, Y: 2.25, Z: -0.5}
	path, err = world.FindPath(destination, next)
	if err != nil {
		t.Fatalf("FindPath() along connector returned an unexpected error: %v", err)
	}
	if len(path) != 1 || path[0] != next {
		t.Fatalf("FindPath() along connector = %+v, want [%+v]", path, next)
	}
}

func TestPrototypeMapRoutesAroundInflatedObstacle(t *testing.T) {
	path, err := NewPrototypeMap().FindPath(simulation.Position{}, simulation.Position{X: 4})
	if err != nil {
		t.Fatalf("FindPath() returned an unexpected error: %v", err)
	}
	if len(path) < 2 {
		t.Fatalf("FindPath() = %+v, want route around obstacle", path)
	}
	foundClearance := false
	for _, waypoint := range path[:len(path)-1] {
		if waypoint.Z < -1-agentRadius || waypoint.Z > 1+agentRadius {
			foundClearance = true
		}
	}
	if !foundClearance {
		t.Fatalf("FindPath() = %+v, want waypoint outside inflated obstacle", path)
	}
}

func TestLayeredMapReportsUnreachableDestination(t *testing.T) {
	world := NewLayeredMapWithObstacles(
		[]Layer{{ID: "ground", MinimumX: -2, MaximumX: 2, MinimumZ: -2, MaximumZ: 2}},
		nil,
		[]Obstacle{{LayerID: "ground", MinimumX: -0.25, MaximumX: 0.25, MinimumZ: -2, MaximumZ: 2}},
	)
	_, err := world.FindPath(simulation.Position{X: -1}, simulation.Position{X: 1})
	if err == nil {
		t.Fatal("FindPath() error = nil for unreachable destination")
	}
}

func indexOf(path []simulation.Position, target simulation.Position) int {
	for index, position := range path {
		if position == target {
			return index
		}
	}
	return -1
}
