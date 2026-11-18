package websockettransport

import (
	"context"
	"testing"
	"time"

	"github.com/mywyau/arcana/server/internal/protocol"
)

func TestOutboundQueueReplacesPendingSnapshot(t *testing.T) {
	queue := newOutboundQueue()
	first := protocol.ServerMessage{Type: protocol.MessageSnapshot, PlayerID: "first"}
	latest := protocol.ServerMessage{Type: protocol.MessageSnapshot, PlayerID: "latest"}
	if replaced := queue.replaceSnapshot(first, time.Unix(1, 0)); replaced {
		t.Fatal("first snapshot unexpectedly replaced another frame")
	}
	if replaced := queue.replaceSnapshot(latest, time.Unix(2, 0)); !replaced {
		t.Fatal("latest snapshot did not replace the pending frame")
	}

	message, ok := queue.next(context.Background())
	if !ok || message.message.PlayerID != "latest" || !message.replaceable {
		t.Fatalf("dequeued snapshot = %+v, available %t", message, ok)
	}
}

func TestOutboundQueuePreservesEnvironmentDefinitionsAcrossReplacement(t *testing.T) {
	queue := newOutboundQueue()
	definitions := &protocol.WorldSnapshot{
		EnvironmentIncluded:            true,
		EnvironmentDefinitionsIncluded: true,
		Drops:                          []protocol.DropSnapshot{{ID: "drop"}},
		GatheringSpots:                 []protocol.GatheringSpotSnapshot{{ID: "tree"}},
		GatheringSpotStates:            []protocol.GatheringSpotStateSnapshot{{ID: "tree", CurrentAmount: 3}},
		BankChests:                     []protocol.BankChestSnapshot{{ID: "bank"}},
		Merchants:                      []protocol.MerchantSnapshot{{ID: "shopkeeper"}},
		MerchantStocks:                 []protocol.MerchantStockStateSnapshot{{ID: "shopkeeper"}},
	}
	latest := &protocol.WorldSnapshot{Tick: 2}
	queue.replaceSnapshot(
		protocol.ServerMessage{Type: protocol.MessageSnapshot, Snapshot: definitions}, time.Unix(1, 0),
	)
	queue.replaceSnapshot(
		protocol.ServerMessage{Type: protocol.MessageSnapshot, Snapshot: latest}, time.Unix(2, 0),
	)

	queued, ok := queue.next(context.Background())
	if !ok || queued.message.Snapshot == nil {
		t.Fatal("merged snapshot was not available")
	}
	got := queued.message.Snapshot
	if !got.EnvironmentIncluded || !got.EnvironmentDefinitionsIncluded || got.Tick != 2 ||
		len(got.Drops) != 1 || len(got.GatheringSpots) != 1 ||
		len(got.GatheringSpotStates) != 1 || len(got.BankChests) != 1 ||
		len(got.Merchants) != 1 || len(got.MerchantStocks) != 1 {
		t.Fatalf("merged snapshot = %+v", got)
	}
}

func TestOutboundQueuePreservesReliableResponsesAheadOfSnapshots(t *testing.T) {
	queue := newOutboundQueue()
	queue.replaceSnapshot(protocol.ServerMessage{Type: protocol.MessageSnapshot}, time.Now())
	if !queue.enqueueReliable(queuedOutbound{message: protocol.ServerMessage{Type: protocol.MessageError}}) {
		t.Fatal("reliable response was rejected")
	}

	first, ok := queue.next(context.Background())
	if !ok || first.message.Type != protocol.MessageError || first.replaceable {
		t.Fatalf("first dequeued message = %+v, available %t", first, ok)
	}
	second, ok := queue.next(context.Background())
	if !ok || second.message.Type != protocol.MessageSnapshot || !second.replaceable {
		t.Fatalf("second dequeued message = %+v, available %t", second, ok)
	}
}

func TestOutboundQueueBoundsReliableResponses(t *testing.T) {
	queue := newOutboundQueue()
	for index := 0; index < outboundReliableCapacity; index++ {
		if !queue.enqueueReliable(queuedOutbound{message: protocol.ServerMessage{Type: protocol.MessageError}}) {
			t.Fatalf("reliable response %d was rejected before capacity", index)
		}
	}
	if queue.enqueueReliable(queuedOutbound{message: protocol.ServerMessage{Type: protocol.MessageError}}) {
		t.Fatal("reliable queue accepted a response beyond its capacity")
	}
}

func TestOutboundQueueCloseUnblocksReader(t *testing.T) {
	queue := newOutboundQueue()
	queue.close()
	if message, ok := queue.next(context.Background()); ok {
		t.Fatalf("closed queue returned message %+v", message)
	}
}

func TestReplacedSnapshotDeltasFromLastDequeuedState(t *testing.T) {
	tracker := newSnapshotDeltaTracker()
	now := time.Unix(10, 0)
	initial := deltaTestSnapshot()
	tracker.apply(initial, "local", now)
	tracker.applyPrivateState(initial, "local", true)

	queue := newOutboundQueue()
	stale := deltaTestSnapshot()
	stale.Players[1].Position.X = 3
	latest := deltaTestSnapshot()
	latest.Players[1].Position.X = 9
	queue.replaceSnapshot(protocol.ServerMessage{Type: protocol.MessageSnapshot, Snapshot: stale}, now.Add(time.Second))
	queue.replaceSnapshot(protocol.ServerMessage{Type: protocol.MessageSnapshot, Snapshot: latest}, now.Add(2*time.Second))

	queued, ok := queue.next(context.Background())
	if !ok {
		t.Fatal("latest snapshot was not available")
	}
	stats := tracker.apply(queued.message.Snapshot, "local", queued.createdAt)
	if stats.PlayerUpdates != 1 || len(queued.message.Snapshot.PlayerUpdates) != 1 ||
		queued.message.Snapshot.PlayerUpdates[0].Position == nil ||
		queued.message.Snapshot.PlayerUpdates[0].Position.X != 9 {
		t.Fatalf("latest delta = %#v, stats %#v", queued.message.Snapshot.PlayerUpdates, stats)
	}
}
