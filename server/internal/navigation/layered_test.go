package navigation

import (
	"testing"

	"github.com/mywyau/arcana/server/internal/simulation"
)

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
