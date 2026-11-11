package servermetrics

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestCollectorAggregatesAndResetsCounters(t *testing.T) {
	start := time.Date(2026, time.November, 10, 12, 0, 0, 0, time.UTC)
	collector := NewCollector(start)
	collector.RecordTick(10*time.Millisecond, 50*time.Millisecond)
	collector.RecordTick(60*time.Millisecond, 50*time.Millisecond)
	collector.RecordSnapshot(4*time.Millisecond, EntityCounts{Players: 2, Targets: 3, Drops: 4})
	collector.RecordEncoding(time.Millisecond, 1024)
	collector.RecordEncoding(3*time.Millisecond, 3072)
	collector.RecordInterestQuery(
		time.Millisecond,
		EntityCounts{Players: 4, Targets: 5},
		EntityCounts{Players: 2, Targets: 3, Drops: 4, Fires: 1, GatheringSpots: 6},
	)
	collector.RecordInterestQuery(
		3*time.Millisecond,
		EntityCounts{Players: 2, Targets: 3},
		EntityCounts{Players: 4, Targets: 1, Drops: 2, Fires: 1, GatheringSpots: 4},
	)
	collector.RecordWrite(2*time.Millisecond, nil)
	collector.RecordWrite(8*time.Millisecond, errors.New("closed"))
	collector.RecordSave(5*time.Millisecond, errors.New("database unavailable"))
	collector.SetConnections(2)

	window := collector.SnapshotAndReset(start.Add(30 * time.Second))
	if window.TickCount != 2 || window.TickAverage() != 35*time.Millisecond || window.TickMax != 60*time.Millisecond {
		t.Fatalf("tick metrics = count %d average %s max %s", window.TickCount, window.TickAverage(), window.TickMax)
	}
	if window.SlowTickCount != 1 {
		t.Fatalf("SlowTickCount = %d, want 1", window.SlowTickCount)
	}
	if window.BytesAverage() != 2048 || window.BytesMax != 3072 {
		t.Fatalf("payload metrics = average %d max %d", window.BytesAverage(), window.BytesMax)
	}
	if visible := window.VisibilityAverage(); visible.Players != 3 || visible.Targets != 2 ||
		visible.Drops != 3 || visible.Fires != 1 || visible.GatheringSpots != 5 {
		t.Fatalf("visibility metrics = %+v", visible)
	}
	if window.InterestQueryAverage() != 2*time.Millisecond || window.InterestQueryMax != 3*time.Millisecond {
		t.Fatalf("interest query metrics = average %s max %s", window.InterestQueryAverage(), window.InterestQueryMax)
	}
	if candidates := window.CandidateAverage(); candidates.Players != 3 || candidates.Targets != 4 {
		t.Fatalf("candidate metrics = %+v", candidates)
	}
	if window.WriteErrorCount != 1 || window.SaveErrorCount != 1 {
		t.Fatalf("error metrics = writes %d saves %d", window.WriteErrorCount, window.SaveErrorCount)
	}
	if window.Connections != 2 || window.Entities.Players != 2 {
		t.Fatalf("gauges = connections %d entities %+v", window.Connections, window.Entities)
	}

	next := collector.SnapshotAndReset(start.Add(time.Minute))
	if next.TickCount != 0 || next.SnapshotCount != 0 || next.BytesTotal != 0 ||
		next.VisibilityCount != 0 || next.InterestQueryCount != 0 {
		t.Fatalf("counter values survived reset: %+v", next)
	}
	if next.Connections != 2 || next.Entities.Players != 2 {
		t.Fatalf("gauge values did not survive reset: %+v", next)
	}
}

func TestWindowSummaryIsCompactAndIncludesOperationalSignals(t *testing.T) {
	window := Window{
		StartedAt:          time.Unix(0, 0),
		EndedAt:            time.Unix(30, 0),
		TickCount:          2,
		TickTotal:          20 * time.Millisecond,
		TickMax:            12 * time.Millisecond,
		SnapshotCount:      1,
		SnapshotTotal:      3 * time.Millisecond,
		EncodingCount:      1,
		EncodingTotal:      time.Millisecond,
		WriteCount:         1,
		WriteTotal:         2 * time.Millisecond,
		SaveCount:          1,
		SaveTotal:          4 * time.Millisecond,
		BytesTotal:         2048,
		BytesMax:           2048,
		Connections:        4,
		Entities:           EntityCounts{Players: 4, Targets: 2},
		VisibilityCount:    2,
		VisibilityTotal:    EntityCounts{Players: 6, Targets: 4},
		InterestQueryCount: 2,
		InterestQueryTotal: 4 * time.Millisecond,
		InterestQueryMax:   3 * time.Millisecond,
		CandidateTotal:     EntityCounts{Players: 8, Targets: 6},
		SlowTickCount:      1,
		WriteErrorCount:    1,
	}
	summary := window.Summary(4*1024*1024, 3)
	for _, wanted := range []string{
		"performance window=30s", "connections=4", "players=4", "tick_avg=10ms",
		"candidate_players_avg=4", "candidate_targets_avg=3",
		"visible_players_avg=3", "visible_targets_avg=2", "interest_query_avg=2ms", "slow_ticks=1",
		"payload_avg=2.0KiB", "write_errors=1", "heap=4.0MiB", "gc=3",
	} {
		if !strings.Contains(summary, wanted) {
			t.Fatalf("Summary() = %q, missing %q", summary, wanted)
		}
	}
}
