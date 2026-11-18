package websockettransport

import (
	"testing"

	"github.com/mywyau/arcana/server/internal/simulation"
)

func TestEnvironmentDefinitionsOnlyRepeatWhenVisibleLandmarksChange(t *testing.T) {
	state := environmentDefinitionState{}
	initial := simulation.Snapshot{
		GatheringSpots: []simulation.GatheringSpotSnapshot{{ID: "tree"}},
		BankChests:     []simulation.BankChestSnapshot{{ID: "bank"}},
		Merchants:      []simulation.MerchantSnapshot{{ID: "shopkeeper"}},
	}
	if !state.shouldInclude(initial, true) {
		t.Fatal("initial environment did not include definitions")
	}
	if state.shouldInclude(initial, false) {
		t.Fatal("realtime frame included environment definitions")
	}
	if state.shouldInclude(initial, true) {
		t.Fatal("unchanged environment repeated definitions")
	}

	changed := initial
	changed.GatheringSpots = append(
		[]simulation.GatheringSpotSnapshot(nil), initial.GatheringSpots...,
	)
	changed.GatheringSpots = append(changed.GatheringSpots, simulation.GatheringSpotSnapshot{ID: "rock"})
	if !state.shouldInclude(changed, true) {
		t.Fatal("changed visible landmarks did not include definitions")
	}
	changed.Merchants = append(
		[]simulation.MerchantSnapshot(nil), initial.Merchants...,
	)
	changed.Merchants = append(changed.Merchants, simulation.MerchantSnapshot{ID: "second-shop"})
	if !state.shouldInclude(changed, true) {
		t.Fatal("changed visible merchants did not include definitions")
	}
}
