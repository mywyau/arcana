package protocol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

const (
	MessageMoveTo   = "move_to"
	MessageWelcome  = "welcome"
	MessageSnapshot = "snapshot"
	MessageError    = "error"
)

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type ClientMessage struct {
	Type        string    `json:"type"`
	Sequence    uint64    `json:"sequence"`
	Destination *Position `json:"destination"`
}

type ServerMessage struct {
	Type     string         `json:"type"`
	PlayerID string         `json:"player_id,omitempty"`
	Snapshot *WorldSnapshot `json:"snapshot,omitempty"`
	Error    *ErrorMessage  `json:"error,omitempty"`
}

type WorldSnapshot struct {
	Tick    uint64           `json:"tick"`
	Players []PlayerSnapshot `json:"players"`
}

type PlayerSnapshot struct {
	ID                  string     `json:"id"`
	Position            Position   `json:"position"`
	Destination         *Position  `json:"destination,omitempty"`
	Path                []Position `json:"path"`
	Speed               float64    `json:"speed"`
	LastCommandSequence uint64     `json:"last_command_sequence"`
	LastCommandTick     uint64     `json:"last_command_tick"`
}

type ErrorMessage struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func DecodeClientMessage(data []byte) (ClientMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	var message ClientMessage
	if err := decoder.Decode(&message); err != nil {
		return ClientMessage{}, fmt.Errorf("decode message: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return ClientMessage{}, fmt.Errorf("message must contain one JSON value")
	}

	if message.Type != MessageMoveTo {
		return ClientMessage{}, fmt.Errorf("unsupported message type %q", message.Type)
	}
	if message.Sequence == 0 {
		return ClientMessage{}, fmt.Errorf("sequence must be positive")
	}
	if message.Destination == nil {
		return ClientMessage{}, fmt.Errorf("destination is required")
	}

	return message, nil
}
