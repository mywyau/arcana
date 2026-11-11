package websockettransport

import (
	"reflect"
	"testing"

	"github.com/mywyau/arcana/server/internal/simulation"
)

func TestFilterSnapshotForPlayerKeepsNearbyDynamicEntities(t *testing.T) {
	viewer := simulation.PlayerSnapshot{ID: "viewer", Position: simulation.Position{}}
	snapshot := simulation.Snapshot{
		Players: []simulation.PlayerSnapshot{
			viewer,
			{ID: "boundary-player", Position: simulation.Position{X: 3, Z: 4}},
			{ID: "far-player", Position: simulation.Position{X: 6}},
		},
		Targets: []simulation.TargetSnapshot{
			{ID: "near-target", Position: simulation.Position{X: 5}},
			{ID: "far-target", Position: simulation.Position{X: 6}},
		},
		Drops: []simulation.DropSnapshot{
			{ID: "near-drop", Position: simulation.Position{Z: 4}},
			{ID: "far-drop", Position: simulation.Position{Z: 6}},
		},
		Fires: []simulation.FireSnapshot{
			{ID: "near-fire", Position: simulation.Position{X: -4}},
			{ID: "far-fire", Position: simulation.Position{X: -6}},
		},
		GatheringSpots: []simulation.GatheringSpotSnapshot{
			{ID: "near-spot", Position: simulation.Position{X: 1}},
			{ID: "far-spot", Position: simulation.Position{X: 10}},
		},
		BankChests: []simulation.BankChestSnapshot{{ID: "global-bank"}},
		Merchants:  []simulation.MerchantSnapshot{{ID: "global-merchant"}},
		CombatEvents: []simulation.CombatEvent{
			{ID: 1, SourceID: "viewer", TargetID: "near-target"},
			{ID: 2, SourceID: "far-player", TargetID: "far-target"},
		},
	}

	filtered, counts := filterSnapshotForPlayer(snapshot, viewer.ID, &viewer, 5)
	if len(filtered.Players) != 2 || filtered.Players[1].ID != "boundary-player" {
		t.Fatalf("visible players = %+v", filtered.Players)
	}
	if len(filtered.Targets) != 1 || filtered.Targets[0].ID != "near-target" ||
		len(filtered.Drops) != 1 || filtered.Drops[0].ID != "near-drop" ||
		len(filtered.Fires) != 1 || filtered.Fires[0].ID != "near-fire" ||
		len(filtered.GatheringSpots) != 1 || filtered.GatheringSpots[0].ID != "near-spot" {
		t.Fatalf("filtered dynamic entities = %+v", filtered)
	}
	if len(filtered.BankChests) != 1 || len(filtered.Merchants) != 1 {
		t.Fatalf("static landmarks were filtered: banks=%+v merchants=%+v", filtered.BankChests, filtered.Merchants)
	}
	if len(filtered.CombatEvents) != 1 || filtered.CombatEvents[0].ID != 1 {
		t.Fatalf("visible combat events = %+v", filtered.CombatEvents)
	}
	if counts.Players != 2 || counts.Targets != 1 || counts.Drops != 1 ||
		counts.Fires != 1 || counts.GatheringSpots != 1 {
		t.Fatalf("visible entity counts = %+v", counts)
	}
}

func TestFilterSnapshotForPlayerKeepsCombatParticipantsOutsideRadius(t *testing.T) {
	viewer := simulation.PlayerSnapshot{
		ID: "viewer", Position: simulation.Position{}, CombatTarget: "far-target",
	}
	snapshot := simulation.Snapshot{
		Players: []simulation.PlayerSnapshot{
			viewer,
			{ID: "attacker", Position: simulation.Position{X: 100}, CombatTarget: "viewer"},
			{ID: "unrelated", Position: simulation.Position{X: 100}},
		},
		Targets: []simulation.TargetSnapshot{
			{ID: "far-target", Position: simulation.Position{X: 100}},
			{ID: "enemy-attacker", Position: simulation.Position{X: 100}, TargetID: "viewer"},
			{ID: "unrelated-target", Position: simulation.Position{X: 100}},
		},
	}

	filtered, _ := filterSnapshotForPlayer(snapshot, viewer.ID, &viewer, 5)
	if len(filtered.Players) != 2 || filtered.Players[1].ID != "attacker" {
		t.Fatalf("combat-visible players = %+v", filtered.Players)
	}
	if len(filtered.Targets) != 2 || filtered.Targets[0].ID != "far-target" ||
		filtered.Targets[1].ID != "enemy-attacker" {
		t.Fatalf("combat-visible targets = %+v", filtered.Targets)
	}
}

func TestFilterSnapshotForPlayerAddsViewerMissingFromCachedFrame(t *testing.T) {
	viewer := simulation.PlayerSnapshot{ID: "viewer", Position: simulation.Position{X: 20}}
	filtered, _ := filterSnapshotForPlayer(simulation.Snapshot{
		Players: []simulation.PlayerSnapshot{{ID: "someone-else", Position: simulation.Position{X: 100}}},
	}, viewer.ID, &viewer, 5)
	if len(filtered.Players) != 1 || filtered.Players[0].ID != viewer.ID {
		t.Fatalf("filtered players = %+v", filtered.Players)
	}
}

func TestInterestIndexMatchesFullScanAcrossCellBoundaries(t *testing.T) {
	viewer := simulation.PlayerSnapshot{
		ID: "viewer", Position: simulation.Position{X: -0.5}, CombatTarget: "far-target",
	}
	snapshot := simulation.Snapshot{
		Players: []simulation.PlayerSnapshot{
			viewer,
			{ID: "near-across-boundary", Position: simulation.Position{X: 0.5}},
			{ID: "far-attacker", Position: simulation.Position{X: 90}, CombatTarget: "viewer"},
			{ID: "far-unrelated", Position: simulation.Position{X: -90}},
		},
		Targets: []simulation.TargetSnapshot{
			{ID: "far-target", Position: simulation.Position{Z: 90}},
			{ID: "far-enemy-attacker", Position: simulation.Position{Z: -90}, TargetID: "viewer"},
			{ID: "near-target", Position: simulation.Position{X: -4}},
		},
		Drops: []simulation.DropSnapshot{
			{ID: "near-drop", Position: simulation.Position{X: -2}},
			{ID: "far-drop", Position: simulation.Position{X: 80}},
		},
		Fires: []simulation.FireSnapshot{
			{ID: "near-fire", Position: simulation.Position{Z: 3}},
			{ID: "far-fire", Position: simulation.Position{Z: 80}},
		},
		GatheringSpots: []simulation.GatheringSpotSnapshot{
			{ID: "near-spot", Position: simulation.Position{X: 2}},
			{ID: "far-spot", Position: simulation.Position{X: -80}},
		},
		CombatEvents: []simulation.CombatEvent{
			{ID: 1, SourceID: "viewer", TargetID: "far-target"},
			{ID: 2, SourceID: "far-unrelated", TargetID: "unknown"},
		},
	}

	want, wantCounts := filterSnapshotForPlayer(snapshot, viewer.ID, &viewer, 5)
	index := buildInterestIndex(snapshot, interestCellSize)
	got, candidates, gotCounts := filterSnapshotForPlayerWithIndex(snapshot, index, viewer.ID, &viewer, 5)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("indexed snapshot differs from full scan\n got: %+v\nwant: %+v", got, want)
	}
	if gotCounts != wantCounts {
		t.Fatalf("indexed visible counts = %+v, want %+v", gotCounts, wantCounts)
	}
	if candidates.Players >= len(snapshot.Players) || candidates.Drops >= len(snapshot.Drops) ||
		candidates.Fires >= len(snapshot.Fires) || candidates.GatheringSpots >= len(snapshot.GatheringSpots) {
		t.Fatalf("spatial query did not reduce candidates: %+v", candidates)
	}
}

func TestInterestIndexSkipsEnvironmentWorkForRealtimeSnapshot(t *testing.T) {
	viewer := simulation.PlayerSnapshot{ID: "viewer"}
	snapshot := simulation.Snapshot{
		Players:        []simulation.PlayerSnapshot{viewer},
		Targets:        []simulation.TargetSnapshot{{ID: "target", Position: simulation.Position{X: 1}}},
		Drops:          []simulation.DropSnapshot{{ID: "drop", Position: simulation.Position{X: 1}}},
		Fires:          []simulation.FireSnapshot{{ID: "fire", Position: simulation.Position{X: 1}}},
		GatheringSpots: []simulation.GatheringSpotSnapshot{{ID: "spot", Position: simulation.Position{X: 1}}},
		BankChests:     []simulation.BankChestSnapshot{{ID: "bank"}},
		Merchants:      []simulation.MerchantSnapshot{{ID: "merchant"}},
	}
	filtered, candidates, visible := filterSnapshotForPlayerWithIndexOptions(
		snapshot, buildInterestIndex(snapshot, interestCellSize), viewer.ID, &viewer, 5, false,
	)
	if len(filtered.Players) != 1 || len(filtered.Targets) != 1 {
		t.Fatalf("realtime state was filtered out: %+v", filtered)
	}
	if filtered.Drops != nil || filtered.Fires != nil || filtered.GatheringSpots != nil ||
		filtered.BankChests != nil || filtered.Merchants != nil {
		t.Fatalf("environment state survived realtime filtering: %+v", filtered)
	}
	if candidates.Drops != 0 || candidates.Fires != 0 || candidates.GatheringSpots != 0 ||
		visible.Drops != 0 || visible.Fires != 0 || visible.GatheringSpots != 0 {
		t.Fatalf("realtime environment metrics: candidates=%+v visible=%+v", candidates, visible)
	}
}
