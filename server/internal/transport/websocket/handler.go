package websockettransport

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/mywyau/arcana/server/internal/protocol"
	"github.com/mywyau/arcana/server/internal/simulation"
)

const (
	connectionReadLimit = 4 * 1024
	playerSpeed         = 4.0
	writeTimeout        = 2 * time.Second
)

type Handler struct {
	world            *simulation.World
	snapshotInterval time.Duration
	nextPlayer       atomic.Uint64
}

func NewHandler(world *simulation.World, snapshotInterval time.Duration) (*Handler, error) {
	if world == nil {
		return nil, fmt.Errorf("world is required")
	}
	if snapshotInterval <= 0 {
		return nil, fmt.Errorf("snapshot interval must be positive")
	}
	return &Handler{world: world, snapshotInterval: snapshotInterval}, nil
}

func (h *Handler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	connection, err := websocket.Accept(response, request, nil)
	if err != nil {
		return
	}
	defer connection.CloseNow()
	connection.SetReadLimit(connectionReadLimit)

	playerID := simulation.PlayerID(fmt.Sprintf("player-%d", h.nextPlayer.Add(1)))
	if err := h.world.QueueJoin(playerID, simulation.Position{}, playerSpeed); err != nil {
		_ = connection.Close(websocket.StatusInternalError, "could not join world")
		return
	}
	defer func() {
		_ = h.world.QueueLeave(playerID)
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	responses := make(chan protocol.ServerMessage, 8)
	go h.readCommands(ctx, cancel, connection, playerID, responses)

	if err := writeMessage(ctx, connection, protocol.ServerMessage{
		Type:     protocol.MessageWelcome,
		PlayerID: string(playerID),
	}); err != nil {
		return
	}

	ticker := time.NewTicker(h.snapshotInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case message := <-responses:
			if err := writeMessage(ctx, connection, message); err != nil {
				return
			}
		case <-ticker.C:
			message := protocol.ServerMessage{
				Type:     protocol.MessageSnapshot,
				Snapshot: snapshotMessage(h.world.Snapshot()),
			}
			if err := writeMessage(ctx, connection, message); err != nil {
				return
			}
		}
	}
}

func (h *Handler) readCommands(
	ctx context.Context,
	cancel context.CancelFunc,
	connection *websocket.Conn,
	playerID simulation.PlayerID,
	responses chan<- protocol.ServerMessage,
) {
	defer cancel()

	for {
		_, data, err := connection.Read(ctx)
		if err != nil {
			return
		}

		message, err := protocol.DecodeClientMessage(data)
		if err != nil {
			sendResponse(ctx, responses, errorMessage("invalid_message", err))
			continue
		}

		destination := simulation.Position{
			X: message.Destination.X,
			Y: message.Destination.Y,
			Z: message.Destination.Z,
		}
		err = h.world.QueueMove(simulation.MoveCommand{
			PlayerID:    playerID,
			Sequence:    message.Sequence,
			Destination: destination,
		})
		if err != nil {
			sendResponse(ctx, responses, errorMessage("command_rejected", err))
		}
	}
}

func writeMessage(ctx context.Context, connection *websocket.Conn, message protocol.ServerMessage) error {
	writeCtx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()
	return wsjson.Write(writeCtx, connection, message)
}

func sendResponse(ctx context.Context, responses chan<- protocol.ServerMessage, message protocol.ServerMessage) {
	select {
	case responses <- message:
	case <-ctx.Done():
	}
}

func errorMessage(code string, err error) protocol.ServerMessage {
	return protocol.ServerMessage{
		Type: protocol.MessageError,
		Error: &protocol.ErrorMessage{
			Code:    code,
			Message: err.Error(),
		},
	}
}

func snapshotMessage(snapshot simulation.Snapshot) *protocol.WorldSnapshot {
	message := &protocol.WorldSnapshot{
		Tick:    snapshot.Tick,
		Players: make([]protocol.PlayerSnapshot, 0, len(snapshot.Players)),
	}
	for _, player := range snapshot.Players {
		playerMessage := protocol.PlayerSnapshot{
			ID: string(player.ID),
			Position: protocol.Position{
				X: player.Position.X,
				Y: player.Position.Y,
				Z: player.Position.Z,
			},
			Speed:               player.Speed,
			LastCommandSequence: player.LastCommandSequence,
			LastCommandTick:     player.LastCommandTick,
		}
		if player.Destination != nil {
			playerMessage.Destination = &protocol.Position{
				X: player.Destination.X,
				Y: player.Destination.Y,
				Z: player.Destination.Z,
			}
		}
		message.Players = append(message.Players, playerMessage)
	}
	return message
}
