package simulation

import "testing"

func TestTimedActionDefinitionsUseExpectedKinds(t *testing.T) {
	spots := newGatheringSpots()
	tests := []struct {
		spotID string
		want   timedActionKind
	}{
		{spotID: starterFishingSpotID, want: timedActionGathering},
		{spotID: starterCopperVeinID, want: timedActionGathering},
		{spotID: starterNormalTreeID, want: timedActionGathering},
		{spotID: starterCookingRangeID, want: timedActionProduction},
		{spotID: starterFurnaceID, want: timedActionProduction},
		{spotID: starterAnvilID, want: timedActionProduction},
		{spotID: starterAirRuneAltarID, want: timedActionProduction},
		{spotID: starterAgilityObstacleID, want: timedActionWorldInteraction},
		{spotID: starterPrayerAltarID, want: timedActionWorldInteraction},
	}

	for _, test := range tests {
		spot, exists := spots[test.spotID]
		if !exists {
			t.Fatalf("spot %q is missing", test.spotID)
		}
		if spot.ActionKind != test.want {
			t.Errorf("spot %q action kind = %d, want %d", test.spotID, spot.ActionKind, test.want)
		}
	}
}
