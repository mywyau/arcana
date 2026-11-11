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

	TickCount       int64
	TickTotal       time.Duration
	TickMax         time.Duration
	SlowTickCount   int64
	SnapshotCount   int64
	SnapshotTotal   time.Duration
	SnapshotMax     time.Duration
	EncodingCount   int64
	EncodingTotal   time.Duration
	EncodingMax     time.Duration
	WriteCount      int64
	WriteTotal      time.Duration
	WriteMax        time.Duration
	WriteErrorCount int64
	SaveCount       int64
	SaveTotal       time.Duration
	SaveMax         time.Duration
	SaveErrorCount  int64
	BytesTotal      int64
	BytesMax        int64

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

	startedAt time.Time
	ticks     durationSamples
	slowTicks int64
	snapshots durationSamples
	encodings durationSamples
	writes    durationSamples
	writeErrs int64
	saves     durationSamples
	saveErrs  int64
	bytes     int64
	maxBytes  int64

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

func (c *Collector) RecordEncoding(elapsed time.Duration, size int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.encodings.add(elapsed)
	c.bytes += int64(size)
	if int64(size) > c.maxBytes {
		c.maxBytes = int64(size)
	}
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

		TickCount:       c.ticks.Count,
		TickTotal:       c.ticks.Total,
		TickMax:         c.ticks.Max,
		SlowTickCount:   c.slowTicks,
		SnapshotCount:   c.snapshots.Count,
		SnapshotTotal:   c.snapshots.Total,
		SnapshotMax:     c.snapshots.Max,
		EncodingCount:   c.encodings.Count,
		EncodingTotal:   c.encodings.Total,
		EncodingMax:     c.encodings.Max,
		WriteCount:      c.writes.Count,
		WriteTotal:      c.writes.Total,
		WriteMax:        c.writes.Max,
		WriteErrorCount: c.writeErrs,
		SaveCount:       c.saves.Count,
		SaveTotal:       c.saves.Total,
		SaveMax:         c.saves.Max,
		SaveErrorCount:  c.saveErrs,
		BytesTotal:      c.bytes,
		BytesMax:        c.maxBytes,
		Connections:     c.connections,
		Entities:        c.entities,
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
	return window
}

// Summary renders a compact, single-line value suitable for ordinary server
// logs. Memory figures are supplied by the caller to keep runtime concerns out
// of the collector.
func (w Window) Summary(heapBytes, gcCount uint64) string {
	return fmt.Sprintf(
		"performance window=%s connections=%d players=%d targets=%d drops=%d fires=%d gathering_spots=%d tick_avg=%s tick_max=%s slow_ticks=%d snapshot_avg=%s snapshot_max=%s encode_avg=%s payload_avg=%s payload_max=%s write_avg=%s write_max=%s write_errors=%d save_avg=%s save_max=%s save_errors=%d heap=%s gc=%d",
		w.EndedAt.Sub(w.StartedAt).Round(time.Millisecond),
		w.Connections, w.Entities.Players, w.Entities.Targets, w.Entities.Drops,
		w.Entities.Fires, w.Entities.GatheringSpots,
		shortDuration(w.TickAverage()), shortDuration(w.TickMax), w.SlowTickCount,
		shortDuration(w.SnapshotAverage()), shortDuration(w.SnapshotMax),
		shortDuration(w.EncodingAverage()), byteSize(w.BytesAverage()), byteSize(w.BytesMax),
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
