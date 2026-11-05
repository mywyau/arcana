package navigation

import (
	"container/heap"
	"fmt"
	"math"

	"github.com/mywyau/arcana/server/internal/simulation"
)

const (
	heightTolerance = 0.25
	linkTolerance   = 1.05
	gridCellSize    = 0.5
	agentRadius     = 0.4
	hillCenterX     = -12.0
	hillCenterZ     = 10.0
	hillRadius      = 4.0
	hillHeight      = 2.0
)

type Layer struct {
	ID           string
	Height       float64
	MinimumX     float64
	MaximumX     float64
	MinimumZ     float64
	MaximumZ     float64
	Elevation    func(x, z float64) float64
	MaximumSlope float64
}

type Link struct {
	ID        string
	FromLayer string
	From      simulation.Position
	ToLayer   string
	To        simulation.Position
}

type Obstacle struct {
	LayerID  string
	MinimumX float64
	MaximumX float64
	MinimumZ float64
	MaximumZ float64
}

type LayeredMap struct {
	layers    []Layer
	links     []Link
	obstacles []Obstacle
}

func NewLayeredMap(layers []Layer, links []Link) *LayeredMap {
	return NewLayeredMapWithObstacles(layers, links, nil)
}

func NewLayeredMapWithObstacles(layers []Layer, links []Link, obstacles []Obstacle) *LayeredMap {
	return &LayeredMap{
		layers:    append([]Layer(nil), layers...),
		links:     append([]Link(nil), links...),
		obstacles: append([]Obstacle(nil), obstacles...),
	}
}

// NewPrototypeMap matches the surfaces, ramp, and obstacles shown by the client.
func NewPrototypeMap() *LayeredMap {
	return NewLayeredMapWithObstacles(
		[]Layer{
			{
				ID:           "ground",
				MinimumX:     -20,
				MaximumX:     20,
				MinimumZ:     -20,
				MaximumZ:     20,
				Elevation:    prototypeGroundHeight,
				MaximumSlope: 1,
			},
			{ID: "upper", Height: 3, MinimumX: -6, MaximumX: 2, MinimumZ: -4, MaximumZ: 4},
		},
		[]Link{
			{
				ID:        "ground_upper_ramp",
				FromLayer: "ground",
				From:      simulation.Position{X: 4, Y: 0, Z: 0},
				ToLayer:   "upper",
				To:        simulation.Position{X: 2, Y: 3, Z: 0},
			},
		},
		[]Obstacle{
			{LayerID: "ground", MinimumX: 0.5, MaximumX: 1.5, MinimumZ: -1, MaximumZ: 1},
			{LayerID: "upper", MinimumX: -3, MaximumX: -2, MinimumZ: -1, MaximumZ: 2},
		},
	)
}

func (m *LayeredMap) Locate(position simulation.Position) (simulation.NavigationLocation, error) {
	if layer, err := m.layerAt(position); err == nil {
		return simulation.NavigationLocation{
			Kind: simulation.NavigationLocationLayer,
			ID:   layer.ID,
		}, nil
	}
	for _, link := range m.links {
		if positionOnLink(position, link) {
			id := link.ID
			if id == "" {
				id = link.FromLayer + "_" + link.ToLayer
			}
			return simulation.NavigationLocation{
				Kind: simulation.NavigationLocationConnector,
				ID:   id,
			}, nil
		}
	}
	return simulation.NavigationLocation{}, fmt.Errorf("position %+v has no navigation location", position)
}

func (m *LayeredMap) FindPath(start, destination simulation.Position) ([]simulation.Position, error) {
	destinationLayer, err := m.layerAt(destination)
	if err != nil {
		return m.pathToLink(start, destination, err)
	}
	if !m.positionWalkable(destinationLayer, destination) {
		return nil, fmt.Errorf("destination is blocked")
	}

	startLayer, err := m.layerAt(start)
	if err != nil {
		return m.pathFromLink(start, destination, destinationLayer, err)
	}
	if startLayer.ID == destinationLayer.ID {
		return m.findLayerPath(startLayer, start, destination)
	}

	for _, link := range m.links {
		var startEndpoint, destinationEndpoint simulation.Position
		switch {
		case link.FromLayer == startLayer.ID && link.ToLayer == destinationLayer.ID:
			startEndpoint, destinationEndpoint = link.From, link.To
		case link.ToLayer == startLayer.ID && link.FromLayer == destinationLayer.ID:
			startEndpoint, destinationEndpoint = link.To, link.From
		default:
			continue
		}

		first, firstErr := m.findLayerPath(startLayer, start, startEndpoint)
		second, secondErr := m.findLayerPath(destinationLayer, destinationEndpoint, destination)
		if firstErr != nil || secondErr != nil {
			continue
		}
		path := append(first, destinationEndpoint)
		path = appendWithoutDuplicate(path, second...)
		return path, nil
	}
	return nil, fmt.Errorf("no route from layer %q to %q", startLayer.ID, destinationLayer.ID)
}

func (m *LayeredMap) pathToLink(
	start simulation.Position,
	destination simulation.Position,
	layerError error,
) ([]simulation.Position, error) {
	for _, link := range m.links {
		if !positionOnLink(destination, link) {
			continue
		}
		if positionOnLink(start, link) {
			return []simulation.Position{destination}, nil
		}

		startLayer, err := m.layerAt(start)
		if err != nil {
			return nil, fmt.Errorf("start: %w", err)
		}
		var endpoint simulation.Position
		switch startLayer.ID {
		case link.FromLayer:
			endpoint = link.From
		case link.ToLayer:
			endpoint = link.To
		default:
			continue
		}
		path, err := m.findLayerPath(startLayer, start, endpoint)
		if err != nil {
			return nil, err
		}
		return appendWithoutDuplicate(path, destination), nil
	}
	return nil, fmt.Errorf("destination: %w", layerError)
}

func (m *LayeredMap) pathFromLink(
	start simulation.Position,
	destination simulation.Position,
	destinationLayer Layer,
	layerError error,
) ([]simulation.Position, error) {
	for _, link := range m.links {
		if !positionOnLink(start, link) {
			continue
		}
		var endpoint simulation.Position
		switch destinationLayer.ID {
		case link.FromLayer:
			endpoint = link.From
		case link.ToLayer:
			endpoint = link.To
		default:
			continue
		}
		path, err := m.findLayerPath(destinationLayer, endpoint, destination)
		if err != nil {
			return nil, err
		}
		return appendWithoutDuplicate([]simulation.Position{endpoint}, path...), nil
	}
	return nil, fmt.Errorf("start: %w", layerError)
}

func (m *LayeredMap) findLayerPath(layer Layer, start, destination simulation.Position) ([]simulation.Position, error) {
	startCell := layer.positionToCell(start)
	destinationCell := layer.positionToCell(destination)
	if !m.cellWalkable(layer, startCell) || !m.cellWalkable(layer, destinationCell) {
		return nil, fmt.Errorf("route endpoint is blocked on layer %q", layer.ID)
	}
	if m.lineWalkable(layer, start, destination) {
		return layer.surfaceSegment(start, destination), nil
	}

	cells, err := m.findCellPath(layer, startCell, destinationCell)
	if err != nil {
		return nil, err
	}
	raw := make([]simulation.Position, 0, len(cells)+1)
	raw = append(raw, start)
	for _, cell := range cells[1:] {
		raw = append(raw, layer.cellPosition(cell))
	}
	raw[len(raw)-1] = destination

	smoothed := m.smoothPath(layer, raw)
	return layer.expandSurfacePath(smoothed), nil
}

func (m *LayeredMap) findCellPath(layer Layer, start, destination gridCell) ([]gridCell, error) {
	frontier := &cellQueue{}
	heap.Init(frontier)
	heap.Push(frontier, &queueItem{cell: start})
	cameFrom := map[gridCell]gridCell{}
	cost := map[gridCell]int{start: 0}

	for frontier.Len() > 0 {
		current := heap.Pop(frontier).(*queueItem).cell
		if current == destination {
			return reconstructCells(cameFrom, start, destination), nil
		}

		for _, offset := range neighborOffsets {
			next := gridCell{X: current.X + offset.X, Z: current.Z + offset.Z}
			if !m.cellTraversable(layer, current, next) {
				continue
			}
			diagonal := offset.X != 0 && offset.Z != 0
			if diagonal {
				xSide := gridCell{X: current.X + offset.X, Z: current.Z}
				zSide := gridCell{X: current.X, Z: current.Z + offset.Z}
				if !m.cellTraversable(layer, current, xSide) || !m.cellTraversable(layer, current, zSide) ||
					!m.cellTraversable(layer, xSide, next) || !m.cellTraversable(layer, zSide, next) {
					continue
				}
			}

			stepCost := 10
			if diagonal {
				stepCost = 14
			}
			newCost := cost[current] + stepCost
			oldCost, visited := cost[next]
			if visited && newCost >= oldCost {
				continue
			}
			cost[next] = newCost
			cameFrom[next] = current
			heap.Push(frontier, &queueItem{
				cell:     next,
				priority: newCost + octileDistance(next, destination),
			})
		}
	}
	return nil, fmt.Errorf("destination is unreachable on layer %q", layer.ID)
}

func (m *LayeredMap) smoothPath(layer Layer, path []simulation.Position) []simulation.Position {
	if len(path) < 3 {
		return path
	}
	smoothed := []simulation.Position{path[0]}
	anchor := 0
	for anchor < len(path)-1 {
		furthest := anchor + 1
		for candidate := len(path) - 1; candidate > anchor+1; candidate-- {
			if m.lineWalkable(layer, path[anchor], path[candidate]) {
				furthest = candidate
				break
			}
		}
		smoothed = append(smoothed, path[furthest])
		anchor = furthest
	}
	return smoothed
}

func (m *LayeredMap) lineWalkable(layer Layer, start, end simulation.Position) bool {
	distance := math.Hypot(end.X-start.X, end.Z-start.Z)
	steps := max(1, int(math.Ceil(distance/(gridCellSize/2))))
	previous := layer.positionToCell(start)
	previousPosition := start
	for index := 0; index <= steps; index++ {
		progress := float64(index) / float64(steps)
		position := simulation.Position{
			X: start.X + (end.X-start.X)*progress,
			Z: start.Z + (end.Z-start.Z)*progress,
		}
		position.Y = layer.surfaceHeight(position.X, position.Z)
		if !m.positionWalkable(layer, position) {
			return false
		}
		if index > 0 && !layer.slopeWalkable(previousPosition, position) {
			return false
		}
		cell := layer.positionToCell(position)
		if !m.cellWalkable(layer, cell) {
			return false
		}
		if cell.X != previous.X && cell.Z != previous.Z {
			if !m.cellWalkable(layer, gridCell{X: cell.X, Z: previous.Z}) ||
				!m.cellWalkable(layer, gridCell{X: previous.X, Z: cell.Z}) {
				return false
			}
		}
		previous = cell
		previousPosition = position
	}
	return true
}

func (m *LayeredMap) cellTraversable(layer Layer, from, to gridCell) bool {
	if !m.cellWalkable(layer, from) || !m.cellWalkable(layer, to) {
		return false
	}
	return layer.slopeWalkable(layer.cellPosition(from), layer.cellPosition(to))
}

func (m *LayeredMap) cellWalkable(layer Layer, cell gridCell) bool {
	if cell.X < 0 || cell.Z < 0 || cell.X > layer.maximumCellX() || cell.Z > layer.maximumCellZ() {
		return false
	}
	return m.positionWalkable(layer, layer.cellPosition(cell))
}

func (m *LayeredMap) positionWalkable(layer Layer, position simulation.Position) bool {
	if position.X < layer.MinimumX || position.X > layer.MaximumX ||
		position.Z < layer.MinimumZ || position.Z > layer.MaximumZ {
		return false
	}
	for _, obstacle := range m.obstacles {
		if obstacle.LayerID == layer.ID &&
			position.X >= obstacle.MinimumX-agentRadius &&
			position.X <= obstacle.MaximumX+agentRadius &&
			position.Z >= obstacle.MinimumZ-agentRadius &&
			position.Z <= obstacle.MaximumZ+agentRadius {
			return false
		}
	}
	return true
}

func (m *LayeredMap) layerAt(position simulation.Position) (Layer, error) {
	for _, layer := range m.layers {
		if math.Abs(position.Y-layer.surfaceHeight(position.X, position.Z)) <= heightTolerance &&
			position.X >= layer.MinimumX && position.X <= layer.MaximumX &&
			position.Z >= layer.MinimumZ && position.Z <= layer.MaximumZ {
			return layer, nil
		}
	}
	return Layer{}, fmt.Errorf("position %+v is not on a walkable layer", position)
}

type gridCell struct {
	X int
	Z int
}

var neighborOffsets = []gridCell{
	{X: -1}, {X: 1}, {Z: -1}, {Z: 1},
	{X: -1, Z: -1}, {X: -1, Z: 1}, {X: 1, Z: -1}, {X: 1, Z: 1},
}

func (l Layer) positionToCell(position simulation.Position) gridCell {
	return gridCell{
		X: int(math.Round((position.X - l.MinimumX) / gridCellSize)),
		Z: int(math.Round((position.Z - l.MinimumZ) / gridCellSize)),
	}
}

func (l Layer) cellPosition(cell gridCell) simulation.Position {
	x := l.MinimumX + float64(cell.X)*gridCellSize
	z := l.MinimumZ + float64(cell.Z)*gridCellSize
	return simulation.Position{
		X: x,
		Y: l.surfaceHeight(x, z),
		Z: z,
	}
}

func (l Layer) surfaceHeight(x, z float64) float64 {
	if l.Elevation != nil {
		return l.Elevation(x, z)
	}
	return l.Height
}

func (l Layer) slopeWalkable(from, to simulation.Position) bool {
	if l.MaximumSlope <= 0 {
		return true
	}
	horizontalDistance := math.Hypot(to.X-from.X, to.Z-from.Z)
	if horizontalDistance == 0 {
		return math.Abs(to.Y-from.Y) <= heightTolerance
	}
	return math.Abs(to.Y-from.Y)/horizontalDistance <= l.MaximumSlope
}

// surfaceSegment returns waypoints that follow an elevated surface rather than
// cutting a straight chord through it. Flat layers retain their compact path.
func (l Layer) surfaceSegment(start, destination simulation.Position) []simulation.Position {
	if l.Elevation == nil {
		return []simulation.Position{destination}
	}
	distance := math.Hypot(destination.X-start.X, destination.Z-start.Z)
	steps := max(1, int(math.Ceil(distance/gridCellSize)))
	path := make([]simulation.Position, 0, steps)
	for index := 1; index <= steps; index++ {
		progress := float64(index) / float64(steps)
		position := simulation.Position{
			X: start.X + (destination.X-start.X)*progress,
			Z: start.Z + (destination.Z-start.Z)*progress,
		}
		position.Y = l.surfaceHeight(position.X, position.Z)
		path = append(path, position)
	}
	path[len(path)-1] = destination
	return path
}

func (l Layer) expandSurfacePath(path []simulation.Position) []simulation.Position {
	if len(path) < 2 {
		return nil
	}
	expanded := make([]simulation.Position, 0, len(path)-1)
	for index := 1; index < len(path); index++ {
		expanded = appendWithoutDuplicate(expanded, l.surfaceSegment(path[index-1], path[index])...)
	}
	return expanded
}

func (l Layer) maximumCellX() int {
	return int(math.Round((l.MaximumX - l.MinimumX) / gridCellSize))
}

func (l Layer) maximumCellZ() int {
	return int(math.Round((l.MaximumZ - l.MinimumZ) / gridCellSize))
}

func prototypeGroundHeight(x, z float64) float64 {
	distance := math.Hypot(x-hillCenterX, z-hillCenterZ)
	if distance >= hillRadius {
		return 0
	}
	return hillHeight * 0.5 * (1 + math.Cos(math.Pi*distance/hillRadius))
}

func positionOnLink(position simulation.Position, link Link) bool {
	ax := position.X - link.From.X
	ay := position.Y - link.From.Y
	az := position.Z - link.From.Z
	bx := link.To.X - link.From.X
	by := link.To.Y - link.From.Y
	bz := link.To.Z - link.From.Z
	lengthSquared := bx*bx + by*by + bz*bz
	if lengthSquared == 0 {
		return false
	}
	progress := (ax*bx + ay*by + az*bz) / lengthSquared
	if progress < 0 || progress > 1 {
		return false
	}
	closest := simulation.Position{
		X: link.From.X + bx*progress,
		Y: link.From.Y + by*progress,
		Z: link.From.Z + bz*progress,
	}
	dx := position.X - closest.X
	dy := position.Y - closest.Y
	dz := position.Z - closest.Z
	return math.Sqrt(dx*dx+dy*dy+dz*dz) <= linkTolerance
}

func appendWithoutDuplicate(path []simulation.Position, positions ...simulation.Position) []simulation.Position {
	for _, position := range positions {
		if len(path) == 0 || path[len(path)-1] != position {
			path = append(path, position)
		}
	}
	return path
}

func reconstructCells(cameFrom map[gridCell]gridCell, start, destination gridCell) []gridCell {
	path := []gridCell{destination}
	for current := destination; current != start; {
		current = cameFrom[current]
		path = append(path, current)
	}
	for left, right := 0, len(path)-1; left < right; left, right = left+1, right-1 {
		path[left], path[right] = path[right], path[left]
	}
	return path
}

func octileDistance(from, to gridCell) int {
	dx := abs(from.X - to.X)
	dz := abs(from.Z - to.Z)
	return 10*(dx+dz) - 6*min(dx, dz)
}

type queueItem struct {
	cell     gridCell
	priority int
	index    int
}

type cellQueue []*queueItem

func (q cellQueue) Len() int           { return len(q) }
func (q cellQueue) Less(i, j int) bool { return q[i].priority < q[j].priority }
func (q cellQueue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].index, q[j].index = i, j
}
func (q *cellQueue) Push(value any) {
	item := value.(*queueItem)
	item.index = len(*q)
	*q = append(*q, item)
}
func (q *cellQueue) Pop() any {
	old := *q
	item := old[len(old)-1]
	*q = old[:len(old)-1]
	return item
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
