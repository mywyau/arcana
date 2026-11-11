package websockettransport

import (
	"time"

	"github.com/mywyau/arcana/server/internal/protocol"
)

const environmentSnapshotInterval = 500 * time.Millisecond

type environmentSchedule struct {
	interval time.Duration
	nextAt   time.Time
	pending  bool
}

func newEnvironmentSchedule(interval time.Duration) environmentSchedule {
	return environmentSchedule{interval: interval, pending: true}
}

func (s *environmentSchedule) request() {
	s.pending = true
}

func (s *environmentSchedule) due(now time.Time) bool {
	if !s.pending && !s.nextAt.IsZero() && now.Before(s.nextAt) {
		return false
	}
	s.pending = false
	s.nextAt = now.Add(s.interval)
	return true
}

func commandMayChangeEnvironment(messageType string) bool {
	switch messageType {
	case protocol.MessageLightFire,
		protocol.MessagePickupItem,
		protocol.MessageDropItem,
		protocol.MessageMerchantSell,
		protocol.MessageMerchantBuy:
		return true
	default:
		return false
	}
}
