package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"

	"github.com/mywyau/arcana/server/internal/protocol"
)

const (
	scenarioClustered = "clustered"
	scenarioSpread    = "spread"
	loadTestReadLimit = 4 * 1024 * 1024
	connectTimeout    = 10 * time.Second
)

type loadTestConfig struct {
	URL          string
	Clients      int
	RampDuration time.Duration
	HoldDuration time.Duration
	MoveInterval time.Duration
	Scenario     string
	AllowRemote  bool
}

func (c loadTestConfig) validate() error {
	if c.Clients < 1 || c.Clients > 10_000 {
		return fmt.Errorf("clients must be between 1 and 10000")
	}
	if c.RampDuration < 0 {
		return fmt.Errorf("ramp cannot be negative")
	}
	if c.HoldDuration <= 0 {
		return fmt.Errorf("duration must be positive")
	}
	if c.MoveInterval < 0 {
		return fmt.Errorf("move-every cannot be negative")
	}
	if c.Scenario != scenarioClustered && c.Scenario != scenarioSpread {
		return fmt.Errorf("scenario must be %q or %q", scenarioClustered, scenarioSpread)
	}
	return validateEndpoint(c.URL, c.AllowRemote)
}

func validateEndpoint(rawURL string, allowRemote bool) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("parse url: %w", err)
	}
	if parsed.Scheme != "ws" && parsed.Scheme != "wss" {
		return fmt.Errorf("url must use ws or wss")
	}
	if parsed.Host == "" {
		return fmt.Errorf("url host is required")
	}
	if allowRemote {
		return nil
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "localhost" {
		return nil
	}
	ip := net.ParseIP(host)
	if ip != nil && ip.IsLoopback() {
		return nil
	}
	return fmt.Errorf("remote endpoint %q requires --allow-remote", parsed.Host)
}

type loadTestStats struct {
	started               atomic.Int64
	connected             atomic.Int64
	active                atomic.Int64
	peakActive            atomic.Int64
	connectFailures       atomic.Int64
	unexpectedDisconnects atomic.Int64
	messages              atomic.Int64
	snapshots             atomic.Int64
	bytes                 atomic.Int64
	serverErrors          atomic.Int64
	movesSent             atomic.Int64
	moveFailures          atomic.Int64
	connectNanos          atomic.Int64
	maximumConnectNanos   atomic.Int64
}

type loadTestReport struct {
	Elapsed               time.Duration
	Started               int64
	Connected             int64
	Active                int64
	PeakActive            int64
	ConnectFailures       int64
	UnexpectedDisconnects int64
	Messages              int64
	Snapshots             int64
	Bytes                 int64
	ServerErrors          int64
	MovesSent             int64
	MoveFailures          int64
	AverageConnect        time.Duration
	MaximumConnect        time.Duration
}

func (s *loadTestStats) report(elapsed time.Duration) loadTestReport {
	connected := s.connected.Load()
	averageConnect := time.Duration(0)
	if connected > 0 {
		averageConnect = time.Duration(s.connectNanos.Load() / connected)
	}
	return loadTestReport{
		Elapsed: elapsed, Started: s.started.Load(), Connected: connected,
		Active: s.active.Load(), PeakActive: s.peakActive.Load(),
		ConnectFailures: s.connectFailures.Load(), UnexpectedDisconnects: s.unexpectedDisconnects.Load(),
		Messages: s.messages.Load(), Snapshots: s.snapshots.Load(), Bytes: s.bytes.Load(),
		ServerErrors: s.serverErrors.Load(), MovesSent: s.movesSent.Load(), MoveFailures: s.moveFailures.Load(),
		AverageConnect: averageConnect, MaximumConnect: time.Duration(s.maximumConnectNanos.Load()),
	}
}

func (r loadTestReport) String() string {
	seconds := r.Elapsed.Seconds()
	messagesPerSecond := 0.0
	bytesPerSecond := 0.0
	if seconds > 0 {
		messagesPerSecond = float64(r.Messages) / seconds
		bytesPerSecond = float64(r.Bytes) / seconds
	}
	return fmt.Sprintf(
		"load test complete elapsed=%s started=%d connected=%d peak_active=%d connect_failures=%d unexpected_disconnects=%d messages=%d snapshots=%d messages_per_second=%.1f received=%s received_per_second=%s server_errors=%d moves_sent=%d move_failures=%d connect_avg=%s connect_max=%s",
		r.Elapsed.Round(time.Millisecond), r.Started, r.Connected, r.PeakActive, r.ConnectFailures,
		r.UnexpectedDisconnects, r.Messages, r.Snapshots, messagesPerSecond,
		formatBytes(r.Bytes), formatBytes(int64(bytesPerSecond)), r.ServerErrors,
		r.MovesSent, r.MoveFailures, r.AverageConnect.Round(time.Millisecond), r.MaximumConnect.Round(time.Millisecond),
	)
}

func runLoadTest(parent context.Context, config loadTestConfig, logger *log.Logger) loadTestReport {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	startedAt := time.Now()
	stats := &loadTestStats{}
	var clients sync.WaitGroup

	progressDone := make(chan struct{})
	go reportProgress(ctx, logger, stats, startedAt, progressDone)

	spacing := time.Duration(0)
	if config.Clients > 1 {
		spacing = config.RampDuration / time.Duration(config.Clients-1)
	}
rampClients:
	for clientID := 0; clientID < config.Clients; clientID++ {
		if clientID > 0 && spacing > 0 {
			timer := time.NewTimer(spacing)
			select {
			case <-ctx.Done():
				timer.Stop()
				break rampClients
			case <-timer.C:
			}
		}
		if ctx.Err() != nil {
			break
		}
		stats.started.Add(1)
		clients.Add(1)
		go func(id int) {
			defer clients.Done()
			runLoadClient(ctx, config, id, stats)
		}(clientID)
	}

	if ctx.Err() == nil {
		logger.Printf("ramp complete started=%d; holding for %s", stats.started.Load(), config.HoldDuration)
		timer := time.NewTimer(config.HoldDuration)
		select {
		case <-ctx.Done():
			timer.Stop()
		case <-timer.C:
			cancel()
		}
	}
	cancel()
	clients.Wait()
	<-progressDone
	return stats.report(time.Since(startedAt))
}

func reportProgress(
	ctx context.Context,
	logger *log.Logger,
	stats *loadTestStats,
	startedAt time.Time,
	done chan<- struct{},
) {
	defer close(done)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			report := stats.report(time.Since(startedAt))
			logger.Printf(
				"progress started=%d connected=%d active=%d peak_active=%d failures=%d disconnects=%d snapshots=%d received=%s",
				report.Started, report.Connected, report.Active, report.PeakActive, report.ConnectFailures,
				report.UnexpectedDisconnects, report.Snapshots, formatBytes(report.Bytes),
			)
		}
	}
}

func runLoadClient(ctx context.Context, config loadTestConfig, clientID int, stats *loadTestStats) {
	connectStartedAt := time.Now()
	dialCtx, cancelDial := context.WithTimeout(ctx, connectTimeout)
	connection, _, err := websocket.Dial(dialCtx, config.URL, nil)
	cancelDial()
	if err != nil {
		stats.connectFailures.Add(1)
		return
	}
	defer connection.CloseNow()
	connection.SetReadLimit(loadTestReadLimit)

	authentication, err := json.Marshal(protocol.ClientMessage{
		Type: protocol.MessageAuthenticate, ProtocolVersion: protocol.CurrentVersion,
		ClientVersion: "loadtest",
	})
	if err != nil || connection.Write(ctx, websocket.MessageText, authentication) != nil {
		stats.connectFailures.Add(1)
		return
	}

	handshakeCtx, cancelHandshake := context.WithTimeout(ctx, connectTimeout)
	connected := false
	for !connected {
		_, data, readErr := connection.Read(handshakeCtx)
		if readErr != nil {
			cancelHandshake()
			stats.connectFailures.Add(1)
			return
		}
		stats.recordMessage(data)
		envelope, decodeErr := decodeServerEnvelope(data)
		if decodeErr != nil {
			cancelHandshake()
			stats.connectFailures.Add(1)
			return
		}
		if envelope.Type == protocol.MessageError {
			cancelHandshake()
			stats.serverErrors.Add(1)
			stats.connectFailures.Add(1)
			return
		}
		connected = envelope.Type == protocol.MessageWelcome
	}
	cancelHandshake()
	connectDuration := time.Since(connectStartedAt)
	stats.connected.Add(1)
	active := stats.active.Add(1)
	recordMaximum(&stats.peakActive, active)
	defer stats.active.Add(-1)
	stats.connectNanos.Add(int64(connectDuration))
	recordMaximum(&stats.maximumConnectNanos, int64(connectDuration))

	moveDone := make(chan struct{})
	if config.MoveInterval > 0 {
		go sendMovement(ctx, connection, config, clientID, stats, moveDone)
	} else {
		close(moveDone)
	}

	for {
		_, data, readErr := connection.Read(ctx)
		if readErr != nil {
			if ctx.Err() == nil && !errors.Is(readErr, context.Canceled) {
				stats.unexpectedDisconnects.Add(1)
			}
			<-moveDone
			return
		}
		stats.recordMessage(data)
		envelope, decodeErr := decodeServerEnvelope(data)
		if decodeErr != nil {
			stats.serverErrors.Add(1)
			continue
		}
		switch envelope.Type {
		case protocol.MessageSnapshot:
			stats.snapshots.Add(1)
		case protocol.MessageError:
			stats.serverErrors.Add(1)
		}
	}
}

type serverEnvelope struct {
	Type  string                 `json:"type"`
	Error *protocol.ErrorMessage `json:"error,omitempty"`
}

func decodeServerEnvelope(data []byte) (serverEnvelope, error) {
	var envelope serverEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return serverEnvelope{}, err
	}
	if envelope.Type == "" {
		return serverEnvelope{}, fmt.Errorf("server message type is empty")
	}
	return envelope, nil
}

func (s *loadTestStats) recordMessage(data []byte) {
	s.messages.Add(1)
	s.bytes.Add(int64(len(data)))
}

func sendMovement(
	ctx context.Context,
	connection *websocket.Conn,
	config loadTestConfig,
	clientID int,
	stats *loadTestStats,
	done chan<- struct{},
) {
	defer close(done)
	random := rand.New(rand.NewSource(int64(clientID + 1)))
	ticker := time.NewTicker(config.MoveInterval)
	defer ticker.Stop()
	sequence := uint64(1)
	for {
		destination := movementDestination(config.Scenario, clientID, config.Clients, random)
		message, err := json.Marshal(protocol.ClientMessage{
			Type: protocol.MessageMoveTo, Sequence: sequence, Destination: &destination,
		})
		if err != nil || connection.Write(ctx, websocket.MessageText, message) != nil {
			if ctx.Err() == nil {
				stats.moveFailures.Add(1)
			}
			return
		}
		stats.movesSent.Add(1)
		sequence++
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func movementDestination(
	scenario string,
	clientID int,
	clientCount int,
	random *rand.Rand,
) protocol.Position {
	if scenario == scenarioClustered {
		return protocol.Position{X: random.Float64()*10 + 5, Z: random.Float64()*10 + 5}
	}
	side := int(math.Ceil(math.Sqrt(float64(clientCount))))
	row := clientID / side
	column := clientID % side
	spacing := 90.0 / math.Max(1, float64(side-1))
	x := -45.0 + float64(column)*spacing
	z := -45.0 + float64(row)*spacing
	// Small movement around each client's assigned point keeps the spread while
	// still exercising command processing and snapshot position updates.
	x = clamp(x+(random.Float64()*4-2), -45, 45)
	z = clamp(z+(random.Float64()*4-2), -45, 45)
	if math.Hypot(x+12, z-10) < 5 {
		x += 6
	}
	return protocol.Position{X: x, Z: z}
}

func clamp(value, minimum, maximum float64) float64 {
	return math.Max(minimum, math.Min(maximum, value))
}

func recordMaximum(target *atomic.Int64, candidate int64) {
	for {
		current := target.Load()
		if candidate <= current || target.CompareAndSwap(current, candidate) {
			return
		}
	}
}

func formatBytes(value int64) string {
	const (
		kibibyte = 1024
		mebibyte = 1024 * kibibyte
		gibibyte = 1024 * mebibyte
	)
	switch {
	case value >= gibibyte:
		return fmt.Sprintf("%.1fGiB", float64(value)/gibibyte)
	case value >= mebibyte:
		return fmt.Sprintf("%.1fMiB", float64(value)/mebibyte)
	case value >= kibibyte:
		return fmt.Sprintf("%.1fKiB", float64(value)/kibibyte)
	default:
		return fmt.Sprintf("%dB", value)
	}
}
