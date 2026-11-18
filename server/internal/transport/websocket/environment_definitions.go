package websockettransport

import "github.com/mywyau/arcana/server/internal/simulation"

// environmentDefinitionState tracks the static landmark membership last
// prepared for one connection. Definitions are resent when the player moves
// into a view containing a different set of gathering spots, bank chests, or
// merchants.
type environmentDefinitionState struct {
	initialized      bool
	gatheringSpotIDs map[string]struct{}
	bankChestIDs     map[string]struct{}
	merchantIDs      map[string]struct{}
}

func (s *environmentDefinitionState) shouldInclude(
	snapshot simulation.Snapshot,
	includeEnvironment bool,
) bool {
	if !includeEnvironment {
		return false
	}
	nextGatheringSpotIDs := gatheringSpotIDSet(snapshot.GatheringSpots)
	nextBankChestIDs := bankChestIDSet(snapshot.BankChests)
	nextMerchantIDs := merchantIDSet(snapshot.Merchants)
	changed := !s.initialized || !sameIDSet(s.gatheringSpotIDs, nextGatheringSpotIDs) ||
		!sameIDSet(s.bankChestIDs, nextBankChestIDs) || !sameIDSet(s.merchantIDs, nextMerchantIDs)
	if !changed {
		return false
	}
	s.initialized = true
	s.gatheringSpotIDs = nextGatheringSpotIDs
	s.bankChestIDs = nextBankChestIDs
	s.merchantIDs = nextMerchantIDs
	return true
}

func merchantIDSet(merchants []simulation.MerchantSnapshot) map[string]struct{} {
	result := make(map[string]struct{}, len(merchants))
	for _, merchant := range merchants {
		result[merchant.ID] = struct{}{}
	}
	return result
}

func gatheringSpotIDSet(spots []simulation.GatheringSpotSnapshot) map[string]struct{} {
	result := make(map[string]struct{}, len(spots))
	for _, spot := range spots {
		result[spot.ID] = struct{}{}
	}
	return result
}

func bankChestIDSet(chests []simulation.BankChestSnapshot) map[string]struct{} {
	result := make(map[string]struct{}, len(chests))
	for _, chest := range chests {
		result[chest.ID] = struct{}{}
	}
	return result
}

func sameIDSet(first, second map[string]struct{}) bool {
	if len(first) != len(second) {
		return false
	}
	for id := range first {
		if _, found := second[id]; !found {
			return false
		}
	}
	return true
}
