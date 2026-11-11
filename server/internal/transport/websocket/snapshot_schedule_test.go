package websockettransport

import (
	"testing"
	"time"

	"github.com/mywyau/arcana/server/internal/protocol"
)

func TestEnvironmentScheduleIncludesFirstAndPeriodicSnapshots(t *testing.T) {
	schedule := newEnvironmentSchedule(500 * time.Millisecond)
	startedAt := time.Unix(100, 0)
	if !schedule.due(startedAt) {
		t.Fatal("first environment snapshot was not due")
	}
	if schedule.due(startedAt.Add(499 * time.Millisecond)) {
		t.Fatal("environment snapshot was due before its interval")
	}
	if !schedule.due(startedAt.Add(500 * time.Millisecond)) {
		t.Fatal("periodic environment snapshot was not due")
	}
}

func TestEnvironmentScheduleCanBeRequestedEarly(t *testing.T) {
	schedule := newEnvironmentSchedule(time.Second)
	startedAt := time.Unix(100, 0)
	schedule.due(startedAt)
	schedule.request()
	if !schedule.due(startedAt.Add(100 * time.Millisecond)) {
		t.Fatal("requested environment snapshot was not due")
	}
	if schedule.due(startedAt.Add(200 * time.Millisecond)) {
		t.Fatal("request did not restart the periodic interval")
	}
}

func TestCommandMayChangeEnvironment(t *testing.T) {
	for _, messageType := range []string{
		protocol.MessageLightFire, protocol.MessagePickupItem, protocol.MessageDropItem,
		protocol.MessageMerchantSell, protocol.MessageMerchantBuy,
	} {
		if !commandMayChangeEnvironment(messageType) {
			t.Errorf("commandMayChangeEnvironment(%q) = false", messageType)
		}
	}
	if commandMayChangeEnvironment(protocol.MessageMoveTo) {
		t.Fatal("movement unexpectedly requested an environment refresh")
	}
}
