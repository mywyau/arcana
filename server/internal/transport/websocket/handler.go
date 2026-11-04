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

		rejectionCode := "command_rejected"
		switch message.Type {
		case protocol.MessageMoveTo:
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
		case protocol.MessageEngageTarget:
			rejectionCode = "action_rejected"
			err = h.world.QueueTarget(simulation.TargetCommand{
				PlayerID: playerID,
				Sequence: message.Sequence,
				TargetID: message.TargetID,
			})
		case protocol.MessageDisengage:
			rejectionCode = "action_rejected"
			err = h.world.QueueTarget(simulation.TargetCommand{
				PlayerID: playerID,
				Sequence: message.Sequence,
			})
		case protocol.MessageEquipItem:
			rejectionCode = "equipment_rejected"
			err = h.world.QueueEquip(simulation.EquipCommand{
				PlayerID: playerID,
				Sequence: message.Sequence,
				ItemID:   simulation.WeaponID(message.ItemID),
			})
		}
		if err != nil {
			sendResponse(ctx, responses, errorMessage(rejectionCode, err))
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
		Tick:        snapshot.Tick,
		ActionPulse: snapshot.ActionPulse,
		Dummy: protocol.DummySnapshot{
			Position: protocol.Position{
				X: snapshot.Dummy.Position.X,
				Y: snapshot.Dummy.Position.Y,
				Z: snapshot.Dummy.Position.Z,
			},
			Health:          snapshot.Dummy.Health,
			MaximumHealth:   snapshot.Dummy.MaximumHealth,
			TargetID:        string(snapshot.Dummy.TargetID),
			NextAttackPulse: snapshot.Dummy.NextAttackPulse,
			Defeated:        snapshot.Dummy.Defeated,
			RespawnAtPulse:  snapshot.Dummy.RespawnAtPulse,
			DodgeRate:       snapshot.Dummy.DodgeRate,
		},
		CombatEvents: make([]protocol.CombatEvent, 0, len(snapshot.CombatEvents)),
		Players:      make([]protocol.PlayerSnapshot, 0, len(snapshot.Players)),
	}
	for _, event := range snapshot.CombatEvents {
		message.CombatEvents = append(message.CombatEvents, protocol.CombatEvent{
			ID:       event.ID,
			Pulse:    event.Pulse,
			Type:     string(event.Type),
			SourceID: event.SourceID,
			TargetID: event.TargetID,
			Amount:   event.Amount,
		})
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
			NavigationLocation: protocol.NavigationLocation{
				Kind: string(player.NavigationLocation.Kind),
				ID:   player.NavigationLocation.ID,
			},
			LastActionSequence:    player.LastActionSequence,
			LastActionPulse:       player.LastActionPulse,
			LastActionResult:      player.LastActionResult,
			CombatTarget:          player.CombatTarget,
			NextAttackPulse:       player.NextAttackPulse,
			AutoApproaching:       player.AutoApproaching,
			Health:                player.Health,
			MaximumHealth:         player.MaximumHealth,
			Defeated:              player.Defeated,
			RespawnAtPulse:        player.RespawnAtPulse,
			EquippedWeapon:        string(player.EquippedWeapon),
			LastEquipmentSequence: player.LastEquipmentSequence,
			LastEquipmentPulse:    player.LastEquipmentPulse,
			DodgeRate:             player.DodgeRate,
			Inventory: protocol.InventorySnapshot{
				Capacity: player.Inventory.Capacity,
				Items:    make([]protocol.InventoryItem, 0, len(player.Inventory.Items)),
			},
			Path: make([]protocol.Position, 0, len(player.Path)),
		}
		for _, item := range player.Inventory.Items {
			playerMessage.Inventory.Items = append(playerMessage.Inventory.Items, protocol.InventoryItem{
				Slot: item.Slot, ItemID: item.ItemID, Quantity: item.Quantity,
			})
		}
		for _, waypoint := range player.Path {
			playerMessage.Path = append(playerMessage.Path, protocol.Position{
				X: waypoint.X,
				Y: waypoint.Y,
				Z: waypoint.Z,
			})
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
