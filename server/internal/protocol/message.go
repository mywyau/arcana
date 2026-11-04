package protocol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

const (
	MessageMoveTo       = "move_to"
	MessageEngageTarget = "engage_target"
	MessageDisengage    = "disengage"
	MessageWelcome      = "welcome"
	MessageSnapshot     = "snapshot"
	MessageError        = "error"
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
	TargetID    string    `json:"target_id,omitempty"`
}

type ServerMessage struct {
	Type     string         `json:"type"`
	PlayerID string         `json:"player_id,omitempty"`
	Snapshot *WorldSnapshot `json:"snapshot,omitempty"`
	Error    *ErrorMessage  `json:"error,omitempty"`
}

type WorldSnapshot struct {
	Tick        uint64           `json:"tick"`
	ActionPulse uint64           `json:"action_pulse"`
	Dummy       DummySnapshot    `json:"dummy"`
	Players     []PlayerSnapshot `json:"players"`
}

type DummySnapshot struct {
	Position      Position `json:"position"`
	Health        int      `json:"health"`
	MaximumHealth int      `json:"maximum_health"`
}

type PlayerSnapshot struct {
	ID                  string             `json:"id"`
	Position            Position           `json:"position"`
	Destination         *Position          `json:"destination,omitempty"`
	Path                []Position         `json:"path"`
	Speed               float64            `json:"speed"`
	LastCommandSequence uint64             `json:"last_command_sequence"`
	LastCommandTick     uint64             `json:"last_command_tick"`
	NavigationLocation  NavigationLocation `json:"navigation_location"`
	LastActionSequence  uint64             `json:"last_action_sequence"`
	LastActionPulse     uint64             `json:"last_action_pulse"`
	LastActionResult    string             `json:"last_action_result"`
	CombatTarget        string             `json:"combat_target"`
	NextAttackPulse     uint64             `json:"next_attack_pulse"`
	AutoApproaching     bool               `json:"auto_approaching"`
}

type NavigationLocation struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
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

	if message.Type != MessageMoveTo && message.Type != MessageEngageTarget && message.Type != MessageDisengage {
		return ClientMessage{}, fmt.Errorf("unsupported message type %q", message.Type)
	}
	if message.Sequence == 0 {
		return ClientMessage{}, fmt.Errorf("sequence must be positive")
	}
	if message.Type == MessageMoveTo && message.Destination == nil {
		return ClientMessage{}, fmt.Errorf("destination is required")
	}
	if message.Type == MessageEngageTarget && message.TargetID == "" {
		return ClientMessage{}, fmt.Errorf("target ID is required")
	}

	return message, nil
}
