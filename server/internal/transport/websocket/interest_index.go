package websockettransport

import (
	"math"
	"sort"

	"github.com/mywyau/arcana/server/internal/servermetrics"
	"github.com/mywyau/arcana/server/internal/simulation"
)

const interestCellSize = 20.0

type interestCellKey struct {
	x int
	z int
}

type interestCell struct {
	players        []int
	targets        []int
	drops          []int
	fires          []int
	gatheringSpots []int
}

// interestIndex is immutable after construction and is shared by every
// connection reading the corresponding public frame.
type interestIndex struct {
	cellSize float64
	cells    map[interestCellKey]*interestCell

	playersByID      map[string]int
	playersTargeting map[string][]int
	targetsByID      map[string]int
	targetsTargeting map[simulation.PlayerID][]int
}

type interestCandidates struct {
	players        []int
	targets        []int
	drops          []int
	fires          []int
	gatheringSpots []int
}

func buildInterestIndex(snapshot simulation.Snapshot, cellSize float64) *interestIndex {
	index := &interestIndex{
		cellSize:         cellSize,
		cells:            make(map[interestCellKey]*interestCell),
		playersByID:      make(map[string]int, len(snapshot.Players)),
		playersTargeting: make(map[string][]int),
		targetsByID:      make(map[string]int, len(snapshot.Targets)),
		targetsTargeting: make(map[simulation.PlayerID][]int),
	}
	for position, player := range snapshot.Players {
		cell := index.cellFor(player.Position)
		cell.players = append(cell.players, position)
		index.playersByID[string(player.ID)] = position
		if player.CombatTarget != "" {
			index.playersTargeting[player.CombatTarget] = append(index.playersTargeting[player.CombatTarget], position)
		}
	}
	for position, target := range snapshot.Targets {
		cell := index.cellFor(target.Position)
		cell.targets = append(cell.targets, position)
		index.targetsByID[target.ID] = position
		if target.TargetID != "" {
			index.targetsTargeting[target.TargetID] = append(index.targetsTargeting[target.TargetID], position)
		}
	}
	for position, drop := range snapshot.Drops {
		cell := index.cellFor(drop.Position)
		cell.drops = append(cell.drops, position)
	}
	for position, fire := range snapshot.Fires {
		cell := index.cellFor(fire.Position)
		cell.fires = append(cell.fires, position)
	}
	for position, spot := range snapshot.GatheringSpots {
		cell := index.cellFor(spot.Position)
		cell.gatheringSpots = append(cell.gatheringSpots, position)
	}
	return index
}

func (i *interestIndex) cellFor(position simulation.Position) *interestCell {
	key := i.keyFor(position)
	cell := i.cells[key]
	if cell == nil {
		cell = &interestCell{}
		i.cells[key] = cell
	}
	return cell
}

func (i *interestIndex) keyFor(position simulation.Position) interestCellKey {
	return interestCellKey{
		x: int(math.Floor(position.X / i.cellSize)),
		z: int(math.Floor(position.Z / i.cellSize)),
	}
}

func (i *interestIndex) candidates(viewer simulation.PlayerSnapshot, radius float64) interestCandidates {
	candidates := interestCandidates{}
	minimumX := int(math.Floor((viewer.Position.X - radius) / i.cellSize))
	maximumX := int(math.Floor((viewer.Position.X + radius) / i.cellSize))
	minimumZ := int(math.Floor((viewer.Position.Z - radius) / i.cellSize))
	maximumZ := int(math.Floor((viewer.Position.Z + radius) / i.cellSize))
	for x := minimumX; x <= maximumX; x++ {
		for z := minimumZ; z <= maximumZ; z++ {
			cell := i.cells[interestCellKey{x: x, z: z}]
			if cell == nil {
				continue
			}
			candidates.players = append(candidates.players, cell.players...)
			candidates.targets = append(candidates.targets, cell.targets...)
			candidates.drops = append(candidates.drops, cell.drops...)
			candidates.fires = append(candidates.fires, cell.fires...)
			candidates.gatheringSpots = append(candidates.gatheringSpots, cell.gatheringSpots...)
		}
	}

	if position, found := i.playersByID[string(viewer.ID)]; found {
		candidates.players = append(candidates.players, position)
	}
	if viewer.CombatTarget != "" {
		if position, found := i.playersByID[viewer.CombatTarget]; found {
			candidates.players = append(candidates.players, position)
		}
		if position, found := i.targetsByID[viewer.CombatTarget]; found {
			candidates.targets = append(candidates.targets, position)
		}
	}
	candidates.players = append(candidates.players, i.playersTargeting[string(viewer.ID)]...)
	candidates.targets = append(candidates.targets, i.targetsTargeting[viewer.ID]...)
	return candidates
}

func (i *interestIndex) candidateSnapshot(
	snapshot simulation.Snapshot,
	viewer simulation.PlayerSnapshot,
	radius float64,
) (simulation.Snapshot, servermetrics.EntityCounts) {
	candidates := i.candidates(viewer, radius)
	result := snapshot
	result.Players = selectCandidates(snapshot.Players, candidates.players)
	result.Targets = selectCandidates(snapshot.Targets, candidates.targets)
	result.Drops = selectCandidates(snapshot.Drops, candidates.drops)
	result.Fires = selectCandidates(snapshot.Fires, candidates.fires)
	result.GatheringSpots = selectCandidates(snapshot.GatheringSpots, candidates.gatheringSpots)
	return result, snapshotEntityCounts(result)
}

func filterSnapshotForPlayerWithIndex(
	snapshot simulation.Snapshot,
	index *interestIndex,
	viewerID simulation.PlayerID,
	privateViewer *simulation.PlayerSnapshot,
	radius float64,
) (simulation.Snapshot, servermetrics.EntityCounts, servermetrics.EntityCounts) {
	viewer, found := snapshotPlayerByID(snapshot.Players, viewerID)
	if privateViewer != nil && privateViewer.ID == viewerID {
		viewer = *privateViewer
		found = true
	}
	if index == nil || !found {
		filtered, visible := filterSnapshotForPlayer(snapshot, viewerID, privateViewer, radius)
		return filtered, snapshotEntityCounts(snapshot), visible
	}
	candidateSnapshot, candidates := index.candidateSnapshot(snapshot, viewer, radius)
	filtered, visible := filterSnapshotForPlayer(candidateSnapshot, viewerID, privateViewer, radius)
	return filtered, candidates, visible
}

func selectCandidates[T any](values []T, candidates []int) []T {
	sort.Ints(candidates)
	selected := make([]T, 0, len(candidates))
	previous := -1
	for _, position := range candidates {
		if position == previous {
			continue
		}
		selected = append(selected, values[position])
		previous = position
	}
	return selected
}
