package websockettransport

import (
	"context"
	"sync"
	"time"

	"github.com/coder/websocket"

	"github.com/mywyau/arcana/server/internal/protocol"
)

const outboundReliableCapacity = 8

type queuedOutbound struct {
	message     protocol.ServerMessage
	closeAfter  bool
	closeStatus websocket.StatusCode
	closeReason string
	createdAt   time.Time
	replaceable bool
}

// outboundQueue keeps reliable responses in order and at most one pending
// world snapshot. Together with the snapshot currently being written, this
// bounds a connection to roughly two snapshot frames.
type outboundQueue struct {
	mu       sync.Mutex
	notify   chan struct{}
	reliable []queuedOutbound
	snapshot *queuedOutbound
	closed   bool
}

func newOutboundQueue() *outboundQueue {
	return &outboundQueue{notify: make(chan struct{}, 1)}
}

func (q *outboundQueue) enqueueReliable(message queuedOutbound) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed || len(q.reliable) >= outboundReliableCapacity {
		return false
	}
	message.replaceable = false
	q.reliable = append(q.reliable, message)
	q.signal()
	return true
}

func (q *outboundQueue) replaceSnapshot(message protocol.ServerMessage, createdAt time.Time) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return false
	}
	replaced := q.snapshot != nil
	if q.snapshot != nil {
		preserveEnvironmentPayload(q.snapshot.message.Snapshot, message.Snapshot)
	}
	queued := queuedOutbound{message: message, createdAt: createdAt, replaceable: true}
	q.snapshot = &queued
	q.signal()
	return replaced
}

// preserveEnvironmentPayload prevents a lower-frequency environment frame or
// one-time landmark definition from being lost when a newer realtime frame
// replaces it before the writer dequeues it.
func preserveEnvironmentPayload(previous, next *protocol.WorldSnapshot) {
	if previous == nil || next == nil {
		return
	}
	if previous.EnvironmentIncluded && !next.EnvironmentIncluded {
		next.EnvironmentIncluded = true
		next.Drops = previous.Drops
		next.Fires = previous.Fires
		next.GatheringSpotStates = previous.GatheringSpotStates
		next.MerchantStocks = previous.MerchantStocks
	}
	if previous.EnvironmentDefinitionsIncluded && !next.EnvironmentDefinitionsIncluded {
		next.EnvironmentDefinitionsIncluded = true
		next.GatheringSpots = previous.GatheringSpots
		next.BankChests = previous.BankChests
		next.Merchants = previous.Merchants
	}
}

func (q *outboundQueue) next(ctx context.Context) (queuedOutbound, bool) {
	for {
		q.mu.Lock()
		if len(q.reliable) > 0 {
			message := q.reliable[0]
			q.reliable[0] = queuedOutbound{}
			q.reliable = q.reliable[1:]
			q.mu.Unlock()
			return message, true
		}
		if q.snapshot != nil {
			message := *q.snapshot
			q.snapshot = nil
			q.mu.Unlock()
			return message, true
		}
		closed := q.closed
		q.mu.Unlock()
		if closed {
			return queuedOutbound{}, false
		}
		select {
		case <-q.notify:
		case <-ctx.Done():
			return queuedOutbound{}, false
		}
	}
}

func (q *outboundQueue) close() {
	q.mu.Lock()
	q.closed = true
	q.mu.Unlock()
	q.signal()
}

func (q *outboundQueue) signal() {
	select {
	case q.notify <- struct{}{}:
	default:
	}
}
