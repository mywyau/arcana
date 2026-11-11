package websockettransport

import (
	"math"

	"github.com/mywyau/arcana/server/internal/servermetrics"
	"github.com/mywyau/arcana/server/internal/simulation"
)

// filterSnapshotForPlayer creates a connection-local view from the shared
// immutable public frame. Static interaction landmarks remain global in this
// first slice; moving and frequently changing entities are distance filtered.
func filterSnapshotForPlayer(
	snapshot simulation.Snapshot,
	viewerID simulation.PlayerID,
	privateViewer *simulation.PlayerSnapshot,
	radius float64,
) (simulation.Snapshot, servermetrics.EntityCounts) {
	viewer, found := snapshotPlayerByID(snapshot.Players, viewerID)
	if privateViewer != nil && privateViewer.ID == viewerID {
		viewer = *privateViewer
		found = true
	}
	if !found || radius <= 0 || math.IsNaN(radius) || math.IsInf(radius, 0) {
		return snapshot, snapshotEntityCounts(snapshot)
	}

	filtered := snapshot
	filtered.Players = make([]simulation.PlayerSnapshot, 0, len(snapshot.Players))
	filtered.Targets = make([]simulation.TargetSnapshot, 0, len(snapshot.Targets))
	filtered.Drops = make([]simulation.DropSnapshot, 0, len(snapshot.Drops))
	filtered.Fires = make([]simulation.FireSnapshot, 0, len(snapshot.Fires))
	filtered.GatheringSpots = make([]simulation.GatheringSpotSnapshot, 0, len(snapshot.GatheringSpots))
	filtered.CombatEvents = make([]simulation.CombatEvent, 0, len(snapshot.CombatEvents))

	radiusSquared := radius * radius
	visibleIDs := map[string]struct{}{string(viewerID): {}}
	combatTarget := viewer.CombatTarget
	viewerIncluded := false
	for _, player := range snapshot.Players {
		visible := player.ID == viewerID || string(player.ID) == combatTarget ||
			player.CombatTarget == string(viewerID) ||
			positionWithinRadius(viewer.Position, player.Position, radiusSquared)
		if !visible {
			continue
		}
		if player.ID == viewerID && privateViewer != nil {
			player = *privateViewer
			viewerIncluded = true
		}
		filtered.Players = append(filtered.Players, player)
		visibleIDs[string(player.ID)] = struct{}{}
	}
	// A newly joined player may not be present in a public frame cached just
	// before its simulation join. Include it immediately for its own connection.
	if !viewerIncluded && privateViewer != nil {
		filtered.Players = append(filtered.Players, *privateViewer)
		visibleIDs[string(viewerID)] = struct{}{}
	}

	for _, target := range snapshot.Targets {
		if target.ID != combatTarget && target.TargetID != viewerID &&
			!positionWithinRadius(viewer.Position, target.Position, radiusSquared) {
			continue
		}
		filtered.Targets = append(filtered.Targets, target)
		visibleIDs[target.ID] = struct{}{}
	}
	for _, drop := range snapshot.Drops {
		if positionWithinRadius(viewer.Position, drop.Position, radiusSquared) {
			filtered.Drops = append(filtered.Drops, drop)
			visibleIDs[drop.ID] = struct{}{}
		}
	}
	for _, fire := range snapshot.Fires {
		if positionWithinRadius(viewer.Position, fire.Position, radiusSquared) {
			filtered.Fires = append(filtered.Fires, fire)
			visibleIDs[fire.ID] = struct{}{}
		}
	}
	for _, spot := range snapshot.GatheringSpots {
		if positionWithinRadius(viewer.Position, spot.Position, radiusSquared) {
			filtered.GatheringSpots = append(filtered.GatheringSpots, spot)
			visibleIDs[spot.ID] = struct{}{}
		}
	}
	for _, event := range snapshot.CombatEvents {
		_, sourceVisible := visibleIDs[event.SourceID]
		_, targetVisible := visibleIDs[event.TargetID]
		if sourceVisible || targetVisible {
			filtered.CombatEvents = append(filtered.CombatEvents, event)
		}
	}

	return filtered, snapshotEntityCounts(filtered)
}

func snapshotPlayerByID(players []simulation.PlayerSnapshot, id simulation.PlayerID) (simulation.PlayerSnapshot, bool) {
	for _, player := range players {
		if player.ID == id {
			return player, true
		}
	}
	return simulation.PlayerSnapshot{}, false
}

func positionWithinRadius(first, second simulation.Position, radiusSquared float64) bool {
	deltaX := first.X - second.X
	deltaY := first.Y - second.Y
	deltaZ := first.Z - second.Z
	return deltaX*deltaX+deltaY*deltaY+deltaZ*deltaZ <= radiusSquared
}

func snapshotEntityCounts(snapshot simulation.Snapshot) servermetrics.EntityCounts {
	return servermetrics.EntityCounts{
		Players: len(snapshot.Players), Targets: len(snapshot.Targets),
		Drops: len(snapshot.Drops), Fires: len(snapshot.Fires),
		GatheringSpots: len(snapshot.GatheringSpots),
	}
}
