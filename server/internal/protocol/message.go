package protocol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

const (
	MessageAuthenticate      = "authenticate"
	MessageMoveTo            = "move_to"
	MessageEngageTarget      = "engage_target"
	MessageDisengage         = "disengage"
	MessageEquipItem         = "equip_item"
	MessageCombatStyle       = "set_combat_style"
	MessageSetPrayer         = "set_prayer"
	MessagePickupItem        = "pickup_item"
	MessageDropItem          = "drop_item"
	MessageMoveInventoryItem = "move_inventory_item"
	MessageWelcome           = "welcome"
	MessageSnapshot          = "snapshot"
	MessageError             = "error"
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
	ItemID      string    `json:"item_id,omitempty"`
	DropID      string    `json:"drop_id,omitempty"`
	FromSlot    int       `json:"from_slot,omitempty"`
	ToSlot      int       `json:"to_slot,omitempty"`
	CombatStyle string    `json:"combat_style,omitempty"`
	PrayerID    string    `json:"prayer_id,omitempty"`
	Active      bool      `json:"active,omitempty"`
	GuestToken  string    `json:"guest_token,omitempty"`
}

type ServerMessage struct {
	Type       string         `json:"type"`
	PlayerID   string         `json:"player_id,omitempty"`
	GuestToken string         `json:"guest_token,omitempty"`
	Snapshot   *WorldSnapshot `json:"snapshot,omitempty"`
	Error      *ErrorMessage  `json:"error,omitempty"`
}

type WorldSnapshot struct {
	Tick         uint64           `json:"tick"`
	ActionPulse  uint64           `json:"action_pulse"`
	Targets      []TargetSnapshot `json:"targets"`
	Drops        []DropSnapshot   `json:"drops"`
	CombatEvents []CombatEvent    `json:"combat_events"`
	Players      []PlayerSnapshot `json:"players"`
}

type CombatEvent struct {
	ID       uint64 `json:"id"`
	Pulse    uint64 `json:"pulse"`
	Type     string `json:"type"`
	SourceID string `json:"source_id"`
	TargetID string `json:"target_id"`
	Amount   int    `json:"amount"`
}

type TargetSnapshot struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Kind            string   `json:"kind"`
	Position        Position `json:"position"`
	Health          int      `json:"health"`
	MaximumHealth   int      `json:"maximum_health"`
	TargetID        string   `json:"target_id"`
	NextAttackPulse uint64   `json:"next_attack_pulse"`
	Defeated        bool     `json:"defeated"`
	RespawnAtPulse  uint64   `json:"respawn_at_pulse"`
	DodgeRate       float64  `json:"dodge_rate"`
}

type DropSnapshot struct {
	ID       string   `json:"id"`
	Position Position `json:"position"`
	ItemID   string   `json:"item_id"`
	Quantity int      `json:"quantity"`
}

type PlayerSnapshot struct {
	ID                      string             `json:"id"`
	Position                Position           `json:"position"`
	Destination             *Position          `json:"destination,omitempty"`
	Path                    []Position         `json:"path"`
	Speed                   float64            `json:"speed"`
	LastCommandSequence     uint64             `json:"last_command_sequence"`
	LastCommandTick         uint64             `json:"last_command_tick"`
	NavigationLocation      NavigationLocation `json:"navigation_location"`
	LastActionSequence      uint64             `json:"last_action_sequence"`
	LastActionPulse         uint64             `json:"last_action_pulse"`
	LastActionResult        string             `json:"last_action_result"`
	CombatTarget            string             `json:"combat_target"`
	NextAttackPulse         uint64             `json:"next_attack_pulse"`
	AutoApproaching         bool               `json:"auto_approaching"`
	AutoPickingUp           bool               `json:"auto_picking_up"`
	Health                  int                `json:"health"`
	MaximumHealth           int                `json:"maximum_health"`
	Mana                    int                `json:"mana"`
	MaximumMana             int                `json:"maximum_mana"`
	Prayer                  int                `json:"prayer"`
	MaximumPrayer           int                `json:"maximum_prayer"`
	Defeated                bool               `json:"defeated"`
	RespawnAtPulse          uint64             `json:"respawn_at_pulse"`
	EquippedWeapon          string             `json:"equipped_weapon"`
	EquippedArmour          map[string]string  `json:"equipped_armour"`
	LastEquipmentSequence   uint64             `json:"last_equipment_sequence"`
	LastEquipmentPulse      uint64             `json:"last_equipment_pulse"`
	CombatStyle             string             `json:"combat_style"`
	LastCombatStyleSequence uint64             `json:"last_combat_style_sequence"`
	ActivePrayers           []string           `json:"active_prayers"`
	LastPrayerSequence      uint64             `json:"last_prayer_sequence"`
	DodgeRate               float64            `json:"dodge_rate"`
	Inventory               InventorySnapshot  `json:"inventory"`
	LastPickupSequence      uint64             `json:"last_pickup_sequence"`
	Skills                  []SkillSnapshot    `json:"skills"`
}

type SkillSnapshot struct {
	ID                     string `json:"id"`
	Experience             int64  `json:"experience"`
	Level                  int    `json:"level"`
	CurrentLevelExperience int64  `json:"current_level_experience"`
	NextLevelExperience    int64  `json:"next_level_experience"`
}

type InventorySnapshot struct {
	Capacity int             `json:"capacity"`
	Items    []InventoryItem `json:"items"`
}

type InventoryItem struct {
	Slot     int    `json:"slot"`
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
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

	if message.Type != MessageAuthenticate && message.Type != MessageMoveTo && message.Type != MessageEngageTarget &&
		message.Type != MessageDisengage && message.Type != MessageEquipItem &&
		message.Type != MessagePickupItem && message.Type != MessageDropItem &&
		message.Type != MessageMoveInventoryItem && message.Type != MessageCombatStyle &&
		message.Type != MessageSetPrayer {
		return ClientMessage{}, fmt.Errorf("unsupported message type %q", message.Type)
	}
	if message.Type != MessageAuthenticate && message.Sequence == 0 {
		return ClientMessage{}, fmt.Errorf("sequence must be positive")
	}
	if message.Type == MessageMoveTo && message.Destination == nil {
		return ClientMessage{}, fmt.Errorf("destination is required")
	}
	if message.Type == MessageEngageTarget && message.TargetID == "" {
		return ClientMessage{}, fmt.Errorf("target ID is required")
	}
	if message.Type == MessageEquipItem && message.ItemID == "" {
		return ClientMessage{}, fmt.Errorf("item ID is required")
	}
	if message.Type == MessagePickupItem && message.DropID == "" {
		return ClientMessage{}, fmt.Errorf("drop ID is required")
	}
	if message.Type == MessageDropItem && message.ItemID == "" {
		return ClientMessage{}, fmt.Errorf("item ID is required")
	}
	if message.Type == MessageMoveInventoryItem && message.FromSlot == message.ToSlot {
		return ClientMessage{}, fmt.Errorf("inventory slots must be different")
	}
	if message.Type == MessageCombatStyle && message.CombatStyle == "" {
		return ClientMessage{}, fmt.Errorf("combat style is required")
	}
	if message.Type == MessageSetPrayer && message.PrayerID == "" {
		return ClientMessage{}, fmt.Errorf("prayer ID is required")
	}

	return message, nil
}
