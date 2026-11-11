package servermetrics

import (
	"fmt"
	"sync"
	"time"
)

// EntityCounts is the latest entity population observed while building a
// normal outbound world snapshot. Gathering it here avoids an extra snapshot.
type EntityCounts struct {
	Players        int
	Targets        int
	Drops          int
	Fires          int
	GatheringSpots int
}

type PayloadKind uint8

const (
	PayloadOther PayloadKind = iota
	PayloadRealtimeSnapshot
	PayloadEnvironmentSnapshot
)

type durationSamples struct {
	Count int64
	Total time.Duration
	Max   time.Duration
}

func (s *durationSamples) add(value time.Duration) {
	s.Count++
	s.Total += value
	if value > s.Max {
		s.Max = value
	}
}

// Window contains one completed reporting window. Durations are aggregated so
// recording a sample does not allocate or retain per-operation details.
type Window struct {
	StartedAt time.Time
	EndedAt   time.Time

	TickCount               int64
	TickTotal               time.Duration
	TickMax                 time.Duration
	SlowTickCount           int64
	SnapshotCount           int64
	SnapshotTotal           time.Duration
	SnapshotMax             time.Duration
	EncodingCount           int64
	EncodingTotal           time.Duration
	EncodingMax             time.Duration
	WriteCount              int64
	WriteTotal              time.Duration
	WriteMax                time.Duration
	WriteErrorCount         int64
	SaveCount               int64
	SaveTotal               time.Duration
	SaveMax                 time.Duration
	SaveErrorCount          int64
	BytesTotal              int64
	BytesMax                int64
	RealtimePayloadCount    int64
	RealtimePayloadBytes    int64
	EnvironmentPayloadCount int64
	EnvironmentPayloadBytes int64
	VisibilityCount         int64
	VisibilityTotal         EntityCounts
	InterestQueryCount      int64
	InterestQueryTotal      time.Duration
	InterestQueryMax        time.Duration
	CandidateTotal          EntityCounts
	EnvironmentQueries      int64
	DeltaSnapshotCount      int64
	FullReconciliationCount int64
	PlayerUpdateCount       int64
	TargetUpdateCount       int64
	PlayerRemovalCount      int64
	TargetRemovalCount      int64
	FullPayloadCount        int64
	FullPayloadBytes        int64
	DeltaPayloadCount       int64
	DeltaPayloadBytes       int64

	Connections int
	Entities    EntityCounts
}

func (w Window) TickAverage() time.Duration     { return average(w.TickTotal, w.TickCount) }
func (w Window) SnapshotAverage() time.Duration { return average(w.SnapshotTotal, w.SnapshotCount) }
func (w Window) EncodingAverage() time.Duration { return average(w.EncodingTotal, w.EncodingCount) }
func (w Window) WriteAverage() time.Duration    { return average(w.WriteTotal, w.WriteCount) }
func (w Window) SaveAverage() time.Duration     { return average(w.SaveTotal, w.SaveCount) }
func (w Window) BytesAverage() int64 {
	if w.EncodingCount == 0 {
		return 0
	}
	return w.BytesTotal / w.EncodingCount
}

func (w Window) RealtimePayloadAverage() int64 {
	return byteAverage(w.RealtimePayloadBytes, w.RealtimePayloadCount)
}

func (w Window) EnvironmentPayloadAverage() int64 {
	return byteAverage(w.EnvironmentPayloadBytes, w.EnvironmentPayloadCount)
}

func (w Window) FullPayloadAverage() int64 {
	return byteAverage(w.FullPayloadBytes, w.FullPayloadCount)
}

func (w Window) DeltaPayloadAverage() int64 {
	return byteAverage(w.DeltaPayloadBytes, w.DeltaPayloadCount)
}

func byteAverage(total, count int64) int64 {
	if count == 0 {
		return 0
	}
	return total / count
}

func (w Window) VisibilityAverage() EntityCounts {
	if w.VisibilityCount == 0 {
		return EntityCounts{}
	}
	count := int(w.VisibilityCount)
	environmentCount := int(w.EnvironmentQueries)
	if environmentCount == 0 {
		environmentCount = 1
	}
	return EntityCounts{
		Players: w.VisibilityTotal.Players / count, Targets: w.VisibilityTotal.Targets / count,
		Drops: w.VisibilityTotal.Drops / environmentCount, Fires: w.VisibilityTotal.Fires / environmentCount,
		GatheringSpots: w.VisibilityTotal.GatheringSpots / environmentCount,
	}
}

func (w Window) InterestQueryAverage() time.Duration {
	return average(w.InterestQueryTotal, w.InterestQueryCount)
}

func (w Window) CandidateAverage() EntityCounts {
	if w.InterestQueryCount == 0 {
		return EntityCounts{}
	}
	count := int(w.InterestQueryCount)
	environmentCount := int(w.EnvironmentQueries)
	if environmentCount == 0 {
		environmentCount = 1
	}
	return EntityCounts{
		Players: w.CandidateTotal.Players / count, Targets: w.CandidateTotal.Targets / count,
		Drops: w.CandidateTotal.Drops / environmentCount, Fires: w.CandidateTotal.Fires / environmentCount,
		GatheringSpots: w.CandidateTotal.GatheringSpots / environmentCount,
	}
}

func (w Window) PlayerUpdatesAverage() float64 {
	return countAverage(w.PlayerUpdateCount, w.DeltaSnapshotCount)
}

func (w Window) TargetUpdatesAverage() float64 {
	return countAverage(w.TargetUpdateCount, w.DeltaSnapshotCount)
}

func countAverage(total, count int64) float64 {
	if count == 0 {
		return 0
	}
	return float64(total) / float64(count)
}

func average(total time.Duration, count int64) time.Duration {
	if count == 0 {
		return 0
	}
	return total / time.Duration(count)
}

// Collector aggregates inexpensive process-local measurements. It is safe for
// the simulation and all connection goroutines to use concurrently.
type Collector struct {
	mu sync.Mutex

	startedAt               time.Time
	ticks                   durationSamples
	slowTicks               int64
	snapshots               durationSamples
	encodings               durationSamples
	writes                  durationSamples
	writeErrs               int64
	saves                   durationSamples
	saveErrs                int64
	bytes                   int64
	maxBytes                int64
	realtimePayloadCount    int64
	realtimePayloadBytes    int64
	environmentPayloadCount int64
	environmentPayloadBytes int64
	visibilityCount         int64
	visibilityTotal         EntityCounts
	interestQueries         durationSamples
	candidateTotal          EntityCounts
	environmentQueries      int64
	deltaSnapshots          int64
	fullReconciliations     int64
	playerUpdates           int64
	targetUpdates           int64
	playerRemovals          int64
	targetRemovals          int64
	fullPayloadCount        int64
	fullPayloadBytes        int64
	deltaPayloadCount       int64
	deltaPayloadBytes       int64

	connections int
	entities    EntityCounts
}

func NewCollector(now time.Time) *Collector {
	return &Collector{startedAt: now}
}

func (c *Collector) RecordTick(elapsed, budget time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ticks.add(elapsed)
	if elapsed > budget {
		c.slowTicks++
	}
}

func (c *Collector) RecordSnapshot(elapsed time.Duration, entities EntityCounts) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.snapshots.add(elapsed)
	c.entities = entities
}

func (c *Collector) RecordEncoding(elapsed time.Duration, size int, kind PayloadKind) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.encodings.add(elapsed)
	c.bytes += int64(size)
	if int64(size) > c.maxBytes {
		c.maxBytes = int64(size)
	}
	switch kind {
	case PayloadRealtimeSnapshot:
		c.realtimePayloadCount++
		c.realtimePayloadBytes += int64(size)
	case PayloadEnvironmentSnapshot:
		c.environmentPayloadCount++
		c.environmentPayloadBytes += int64(size)
	}
}

func (c *Collector) RecordInterestQuery(
	elapsed time.Duration,
	candidates, visible EntityCounts,
	includeEnvironment bool,
) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.interestQueries.add(elapsed)
	c.candidateTotal.Players += candidates.Players
	c.candidateTotal.Targets += candidates.Targets
	c.candidateTotal.Drops += candidates.Drops
	c.candidateTotal.Fires += candidates.Fires
	c.candidateTotal.GatheringSpots += candidates.GatheringSpots
	c.visibilityCount++
	c.visibilityTotal.Players += visible.Players
	c.visibilityTotal.Targets += visible.Targets
	c.visibilityTotal.Drops += visible.Drops
	c.visibilityTotal.Fires += visible.Fires
	c.visibilityTotal.GatheringSpots += visible.GatheringSpots
	if includeEnvironment {
		c.environmentQueries++
	}
}

func (c *Collector) RecordSnapshotDelta(
	full bool,
	playerUpdates, targetUpdates, playerRemovals, targetRemovals int,
) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.deltaSnapshots++
	if full {
		c.fullReconciliations++
	}
	c.playerUpdates += int64(playerUpdates)
	c.targetUpdates += int64(targetUpdates)
	c.playerRemovals += int64(playerRemovals)
	c.targetRemovals += int64(targetRemovals)
}

func (c *Collector) RecordSnapshotPayload(size int, full bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if full {
		c.fullPayloadCount++
		c.fullPayloadBytes += int64(size)
		return
	}
	c.deltaPayloadCount++
	c.deltaPayloadBytes += int64(size)
}

func (c *Collector) RecordWrite(elapsed time.Duration, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.writes.add(elapsed)
	if err != nil {
		c.writeErrs++
	}
}

func (c *Collector) RecordSave(elapsed time.Duration, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.saves.add(elapsed)
	if err != nil {
		c.saveErrs++
	}
}

func (c *Collector) SetConnections(count int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connections = count
}

// SnapshotAndReset returns the completed interval and starts a new one. Gauge
// values such as connection and entity counts carry into the next window.
func (c *Collector) SnapshotAndReset(now time.Time) Window {
	c.mu.Lock()
	defer c.mu.Unlock()

	window := Window{
		StartedAt: c.startedAt,
		EndedAt:   now,

		TickCount:               c.ticks.Count,
		TickTotal:               c.ticks.Total,
		TickMax:                 c.ticks.Max,
		SlowTickCount:           c.slowTicks,
		SnapshotCount:           c.snapshots.Count,
		SnapshotTotal:           c.snapshots.Total,
		SnapshotMax:             c.snapshots.Max,
		EncodingCount:           c.encodings.Count,
		EncodingTotal:           c.encodings.Total,
		EncodingMax:             c.encodings.Max,
		WriteCount:              c.writes.Count,
		WriteTotal:              c.writes.Total,
		WriteMax:                c.writes.Max,
		WriteErrorCount:         c.writeErrs,
		SaveCount:               c.saves.Count,
		SaveTotal:               c.saves.Total,
		SaveMax:                 c.saves.Max,
		SaveErrorCount:          c.saveErrs,
		BytesTotal:              c.bytes,
		BytesMax:                c.maxBytes,
		RealtimePayloadCount:    c.realtimePayloadCount,
		RealtimePayloadBytes:    c.realtimePayloadBytes,
		EnvironmentPayloadCount: c.environmentPayloadCount,
		EnvironmentPayloadBytes: c.environmentPayloadBytes,
		VisibilityCount:         c.visibilityCount,
		VisibilityTotal:         c.visibilityTotal,
		InterestQueryCount:      c.interestQueries.Count,
		InterestQueryTotal:      c.interestQueries.Total,
		InterestQueryMax:        c.interestQueries.Max,
		CandidateTotal:          c.candidateTotal,
		EnvironmentQueries:      c.environmentQueries,
		DeltaSnapshotCount:      c.deltaSnapshots,
		FullReconciliationCount: c.fullReconciliations,
		PlayerUpdateCount:       c.playerUpdates,
		TargetUpdateCount:       c.targetUpdates,
		PlayerRemovalCount:      c.playerRemovals,
		TargetRemovalCount:      c.targetRemovals,
		FullPayloadCount:        c.fullPayloadCount,
		FullPayloadBytes:        c.fullPayloadBytes,
		DeltaPayloadCount:       c.deltaPayloadCount,
		DeltaPayloadBytes:       c.deltaPayloadBytes,
		Connections:             c.connections,
		Entities:                c.entities,
	}
	c.startedAt = now
	c.ticks = durationSamples{}
	c.slowTicks = 0
	c.snapshots = durationSamples{}
	c.encodings = durationSamples{}
	c.writes = durationSamples{}
	c.writeErrs = 0
	c.saves = durationSamples{}
	c.saveErrs = 0
	c.bytes = 0
	c.maxBytes = 0
	c.realtimePayloadCount = 0
	c.realtimePayloadBytes = 0
	c.environmentPayloadCount = 0
	c.environmentPayloadBytes = 0
	c.visibilityCount = 0
	c.visibilityTotal = EntityCounts{}
	c.interestQueries = durationSamples{}
	c.candidateTotal = EntityCounts{}
	c.environmentQueries = 0
	c.deltaSnapshots = 0
	c.fullReconciliations = 0
	c.playerUpdates = 0
	c.targetUpdates = 0
	c.playerRemovals = 0
	c.targetRemovals = 0
	c.fullPayloadCount = 0
	c.fullPayloadBytes = 0
	c.deltaPayloadCount = 0
	c.deltaPayloadBytes = 0
	return window
}

// Summary renders a compact, single-line value suitable for ordinary server
// logs. Memory figures are supplied by the caller to keep runtime concerns out
// of the collector.
func (w Window) Summary(heapBytes, gcCount uint64) string {
	visible := w.VisibilityAverage()
	candidates := w.CandidateAverage()
	return fmt.Sprintf(
		"performance window=%s connections=%d players=%d targets=%d drops=%d fires=%d gathering_spots=%d environment_snapshots=%d full_reconciliations=%d player_updates_avg=%.2f target_updates_avg=%.2f player_removals=%d target_removals=%d candidate_players_avg=%d candidate_targets_avg=%d candidate_drops_avg=%d candidate_fires_avg=%d candidate_spots_avg=%d visible_players_avg=%d visible_targets_avg=%d visible_drops_avg=%d visible_fires_avg=%d visible_spots_avg=%d interest_query_avg=%s interest_query_max=%s tick_avg=%s tick_max=%s slow_ticks=%d snapshot_avg=%s snapshot_max=%s encode_avg=%s payload_avg=%s payload_max=%s realtime_payload_avg=%s environment_payload_avg=%s full_payload_avg=%s delta_payload_avg=%s write_avg=%s write_max=%s write_errors=%d save_avg=%s save_max=%s save_errors=%d heap=%s gc=%d",
		w.EndedAt.Sub(w.StartedAt).Round(time.Millisecond),
		w.Connections, w.Entities.Players, w.Entities.Targets, w.Entities.Drops,
		w.Entities.Fires, w.Entities.GatheringSpots,
		w.EnvironmentQueries,
		w.FullReconciliationCount, w.PlayerUpdatesAverage(), w.TargetUpdatesAverage(),
		w.PlayerRemovalCount, w.TargetRemovalCount,
		candidates.Players, candidates.Targets, candidates.Drops, candidates.Fires, candidates.GatheringSpots,
		visible.Players, visible.Targets, visible.Drops, visible.Fires, visible.GatheringSpots,
		shortDuration(w.InterestQueryAverage()), shortDuration(w.InterestQueryMax),
		shortDuration(w.TickAverage()), shortDuration(w.TickMax), w.SlowTickCount,
		shortDuration(w.SnapshotAverage()), shortDuration(w.SnapshotMax),
		shortDuration(w.EncodingAverage()), byteSize(w.BytesAverage()), byteSize(w.BytesMax),
		byteSize(w.RealtimePayloadAverage()), byteSize(w.EnvironmentPayloadAverage()),
		byteSize(w.FullPayloadAverage()), byteSize(w.DeltaPayloadAverage()),
		shortDuration(w.WriteAverage()), shortDuration(w.WriteMax), w.WriteErrorCount,
		shortDuration(w.SaveAverage()), shortDuration(w.SaveMax), w.SaveErrorCount,
		byteSize(int64(heapBytes)), gcCount,
	)
}

func shortDuration(value time.Duration) time.Duration {
	return value.Round(time.Microsecond)
}

func byteSize(value int64) string {
	const kibibyte = 1024
	const mebibyte = 1024 * kibibyte
	if value >= mebibyte {
		return fmt.Sprintf("%.1fMiB", float64(value)/mebibyte)
	}
	if value >= kibibyte {
		return fmt.Sprintf("%.1fKiB", float64(value)/kibibyte)
	}
	return fmt.Sprintf("%dB", value)
}
