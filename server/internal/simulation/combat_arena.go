package simulation

import "fmt"

// combatArena is intentionally small for the prototype. It keeps encounter
// targeting inside the same bounds represented by the client walls and server
// navigation obstacles without introducing instanced encounters yet.
type combatArena struct {
	Center   Position
	MinimumX float64
	MaximumX float64
	MinimumZ float64
	MaximumZ float64
	Entrance Position
}

const (
	tyrantArenaCenterZ  = 55.0
	tyrantArenaHalfSize = 17.5
	tyrantArenaEntryZ   = 38.0
)

var tyrantCombatArenas = map[string]combatArena{
	riftTyrantTargetID:   newTyrantCombatArena(-52),
	rangedTyrantTargetID: newTyrantCombatArena(0),
	mageTyrantTargetID:   newTyrantCombatArena(52),
}

func newTyrantCombatArena(centerX float64) combatArena {
	return combatArena{
		Center:   Position{X: centerX, Z: tyrantArenaCenterZ},
		MinimumX: centerX - tyrantArenaHalfSize,
		MaximumX: centerX + tyrantArenaHalfSize,
		MinimumZ: tyrantArenaCenterZ - tyrantArenaHalfSize,
		MaximumZ: tyrantArenaCenterZ + tyrantArenaHalfSize,
		Entrance: Position{X: centerX, Z: tyrantArenaEntryZ},
	}
}

func combatArenaForTarget(targetID string) (combatArena, bool) {
	arena, exists := tyrantCombatArenas[targetID]
	return arena, exists
}

func (arena combatArena) contains(position Position) bool {
	return position.X >= arena.MinimumX && position.X <= arena.MaximumX &&
		position.Z >= arena.MinimumZ && position.Z <= arena.MaximumZ
}

func (w *World) approachCombatArena(player *playerState, arena combatArena) error {
	if err := statusActionError(player, statusActionMove, w.combat.pulse); err != nil {
		return err
	}
	path := []Position{arena.Entrance}
	if w.pathfinder != nil {
		var err error
		path, err = w.pathfinder.FindPath(player.position, arena.Entrance)
		if err != nil {
			return err
		}
	}
	if len(path) == 0 {
		return fmt.Errorf("empty arena approach path")
	}
	player.destination = positionPointer(path[len(path)-1])
	player.path = clonePositions(path)
	player.autoApproaching = true
	return nil
}
