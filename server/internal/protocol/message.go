package protocol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

const (
	MessageAuthenticate       = "authenticate"
	MessageRegisterAccount    = "register_account"
	MessageLogoutAccount      = "logout_account"
	MessageMoveTo             = "move_to"
	MessageEngageTarget       = "engage_target"
	MessageDisengage          = "disengage"
	MessageEquipItem          = "equip_item"
	MessageUnequipItem        = "unequip_item"
	MessageCombatStyle        = "set_combat_style"
	MessageExperienceLock     = "set_experience_lock"
	MessageDebugSetSkill      = "debug_set_skill"
	MessageSetPrayer          = "set_prayer"
	MessageSetSpell           = "set_spell"
	MessagePickupItem         = "pickup_item"
	MessageDropItem           = "drop_item"
	MessageMoveInventoryItem  = "move_inventory_item"
	MessageConsumeItem        = "consume_item"
	MessageBrew               = "brew"
	MessageFletch             = "fletch"
	MessageCutGem             = "cut_gem"
	MessageLightFire          = "light_fire"
	MessageGather             = "gather"
	MessageBankDeposit        = "bank_deposit"
	MessageBankWithdraw       = "bank_withdraw"
	MessageBankExpand         = "bank_expand"
	MessageMerchantSell       = "merchant_sell"
	MessageAmmunitionDeposit  = "ammunition_deposit"
	MessageAmmunitionWithdraw = "ammunition_withdraw"
	MessageRingDeposit        = "ring_deposit"
	MessageRingWithdraw       = "ring_withdraw"
	MessageRuneDeposit        = "rune_deposit"
	MessageRuneWithdraw       = "rune_withdraw"
	MessageWelcome            = "welcome"
	MessageAccountSession     = "account_session"
	MessageAccountLoggedOut   = "account_logged_out"
	MessageSnapshot           = "snapshot"
	MessageError              = "error"
)

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type ClientMessage struct {
	Type           string    `json:"type"`
	Sequence       uint64    `json:"sequence"`
	Destination    *Position `json:"destination"`
	TargetID       string    `json:"target_id,omitempty"`
	ItemID         string    `json:"item_id,omitempty"`
	ItemIDs        []string  `json:"item_ids,omitempty"`
	InstanceID     string    `json:"instance_id,omitempty"`
	DropID         string    `json:"drop_id,omitempty"`
	FromSlot       int       `json:"from_slot,omitempty"`
	ToSlot         int       `json:"to_slot,omitempty"`
	CombatStyle    string    `json:"combat_style,omitempty"`
	PrayerID       string    `json:"prayer_id,omitempty"`
	SpellID        string    `json:"spell_id,omitempty"`
	SpotID         string    `json:"spot_id,omitempty"`
	RecipeID       string    `json:"recipe_id,omitempty"`
	BankSlot       int       `json:"bank_slot,omitempty"`
	InventorySlot  int       `json:"inventory_slot,omitempty"`
	AmmunitionSlot int       `json:"ammunition_slot,omitempty"`
	RingSlot       int       `json:"ring_slot,omitempty"`
	RuneSlot       int       `json:"rune_slot,omitempty"`
	Quantity       int       `json:"quantity,omitempty"`
	Active         bool      `json:"active,omitempty"`
	SkillID        string    `json:"skill_id,omitempty"`
	Level          int       `json:"level,omitempty"`
	Experience     int64     `json:"experience,omitempty"`
	SetExperience  bool      `json:"set_experience,omitempty"`
	GuestToken     string    `json:"guest_token,omitempty"`
	SessionToken   string    `json:"session_token,omitempty"`
	Email          string    `json:"email,omitempty"`
	Password       string    `json:"password,omitempty"`
}

type ServerMessage struct {
	Type              string                     `json:"type"`
	PlayerID          string                     `json:"player_id,omitempty"`
	GuestToken        string                     `json:"guest_token,omitempty"`
	SessionToken      string                     `json:"session_token,omitempty"`
	AccountEmail      string                     `json:"account_email,omitempty"`
	ProductionRecipes []ProductionRecipeSnapshot `json:"production_recipes,omitempty"`
	Spellbooks        []SpellbookCatalogSnapshot `json:"spellbooks,omitempty"`
	Spells            []SpellCatalogSnapshot     `json:"spells,omitempty"`
	Prayers           []PrayerCatalogSnapshot    `json:"prayers,omitempty"`
	PrayerSets        []PrayerSetCatalogSnapshot `json:"prayer_sets,omitempty"`
	Snapshot          *WorldSnapshot             `json:"snapshot,omitempty"`
	Error             *ErrorMessage              `json:"error,omitempty"`
}

type PrayerCatalogSnapshot struct {
	ID                   string  `json:"id"`
	Name                 string  `json:"name"`
	RequiredLevel        int     `json:"required_level"`
	MeleeDamageBonus     float64 `json:"melee_damage_bonus"`
	RangedDamageBonus    float64 `json:"ranged_damage_bonus"`
	MagicDamageBonus     float64 `json:"magic_damage_bonus"`
	HealthRegeneration   float64 `json:"health_regeneration"`
	MeleeHitChanceBonus  float64 `json:"melee_hit_chance_bonus"`
	DamageReduction      float64 `json:"damage_reduction"`
	DodgeBonus           float64 `json:"dodge_bonus"`
	DrainAmount          int     `json:"drain_amount"`
	DrainIntervalPulses  uint64  `json:"drain_interval_pulses"`
	DrainIntervalSeconds float64 `json:"drain_interval_seconds"`
	ExclusiveGroup       string  `json:"exclusive_group,omitempty"`
}

type PrayerSetCatalogSnapshot struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	PrayerIDs []string `json:"prayer_ids"`
}

type SpellbookCatalogSnapshot struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	EquipmentItemID string `json:"equipment_item_id,omitempty"`
}

type SpellCatalogSnapshot struct {
	ID                 string     `json:"id"`
	BookID             string     `json:"book_id"`
	Name               string     `json:"name"`
	IconID             string     `json:"icon_id"`
	CastMode           string     `json:"cast_mode"`
	TargetType         string     `json:"target_type"`
	EffectType         string     `json:"effect_type"`
	Element            string     `json:"element"`
	RequiredMagicLevel int        `json:"required_magic_level"`
	Costs              []ItemCost `json:"costs"`
	BasePower          int        `json:"base_power"`
	AccuracyBonus      int        `json:"accuracy_bonus"`
	Range              float64    `json:"range"`
	DamageVariance     float64    `json:"damage_variance"`
	AttackInterval     uint64     `json:"attack_interval"`
}

type ProductionRecipeSnapshot struct {
	Station        string     `json:"station"`
	Group          string     `json:"group,omitempty"`
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	RequiredLevel  int        `json:"required_level"`
	Costs          []ItemCost `json:"costs"`
	RewardItem     string     `json:"reward_item"`
	RewardQuantity int        `json:"reward_quantity"`
	RewardXP       int64      `json:"reward_xp,omitempty"`
	Interval       uint64     `json:"interval"`
	RequiredTool   string     `json:"required_tool,omitempty"`
}

type ItemCost struct {
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
}

type WorldSnapshot struct {
	Tick           uint64                  `json:"tick"`
	ActionPulse    uint64                  `json:"action_pulse"`
	Targets        []TargetSnapshot        `json:"targets"`
	Drops          []DropSnapshot          `json:"drops"`
	Fires          []FireSnapshot          `json:"fires"`
	CombatEvents   []CombatEvent           `json:"combat_events"`
	GatheringSpots []GatheringSpotSnapshot `json:"gathering_spots"`
	BankChests     []BankChestSnapshot     `json:"bank_chests"`
	Merchants      []MerchantSnapshot      `json:"merchants"`
	Players        []PlayerSnapshot        `json:"players"`
}

type CombatEvent struct {
	ID       uint64 `json:"id"`
	Pulse    uint64 `json:"pulse"`
	Type     string `json:"type"`
	SourceID string `json:"source_id"`
	TargetID string `json:"target_id"`
	Amount   int    `json:"amount"`
	SpellID  string `json:"spell_id,omitempty"`
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
	Accuracy        int      `json:"accuracy"`
	Evasion         int      `json:"evasion"`
	AttackType      string   `json:"attack_type"`
	CombatLevel     int      `json:"combat_level"`
}

type DropSnapshot struct {
	ID            string   `json:"id"`
	Position      Position `json:"position"`
	ItemID        string   `json:"item_id"`
	Quantity      int      `json:"quantity"`
	InstanceID    string   `json:"instance_id,omitempty"`
	Reinforcement int      `json:"reinforcement,omitempty"`
	Infinite      bool     `json:"infinite,omitempty"`
}

type FireSnapshot struct {
	ID       string   `json:"id"`
	Position Position `json:"position"`
	LogItem  string   `json:"log_item"`
}

type GatheringSpotSnapshot struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Skill         string   `json:"skill"`
	Position      Position `json:"position"`
	Range         float64  `json:"range"`
	RequiredLevel int      `json:"required_level"`
	CurrentAmount int      `json:"current_amount"`
	MaximumAmount int      `json:"maximum_amount"`
	Station       string   `json:"station,omitempty"`
	RecipeID      string   `json:"recipe_id,omitempty"`
}

type BankChestSnapshot struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Position Position `json:"position"`
	Range    float64  `json:"range"`
}

type MerchantSnapshot struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Position Position `json:"position"`
	Range    float64  `json:"range"`
}

type PlayerSnapshot struct {
	ID                           string                  `json:"id"`
	Position                     Position                `json:"position"`
	Destination                  *Position               `json:"destination,omitempty"`
	Path                         []Position              `json:"path"`
	Speed                        float64                 `json:"speed"`
	LastCommandSequence          uint64                  `json:"last_command_sequence"`
	LastCommandTick              uint64                  `json:"last_command_tick"`
	NavigationLocation           NavigationLocation      `json:"navigation_location"`
	LastActionSequence           uint64                  `json:"last_action_sequence"`
	LastActionPulse              uint64                  `json:"last_action_pulse"`
	LastActionResult             string                  `json:"last_action_result"`
	CombatTarget                 string                  `json:"combat_target"`
	NextAttackPulse              uint64                  `json:"next_attack_pulse"`
	AutoApproaching              bool                    `json:"auto_approaching"`
	AutoPickingUp                bool                    `json:"auto_picking_up"`
	Health                       int                     `json:"health"`
	MaximumHealth                int                     `json:"maximum_health"`
	Prayer                       int                     `json:"prayer"`
	MaximumPrayer                int                     `json:"maximum_prayer"`
	Defeated                     bool                    `json:"defeated"`
	RespawnAtPulse               uint64                  `json:"respawn_at_pulse"`
	EquippedWeapon               string                  `json:"equipped_weapon"`
	EquippedWeaponInstanceID     string                  `json:"equipped_weapon_instance_id,omitempty"`
	EquippedWeaponReinforcement  int                     `json:"equipped_weapon_reinforcement,omitempty"`
	EquippedArmour               map[string]string       `json:"equipped_armour"`
	EquippedArmourInstances      map[string]string       `json:"equipped_armour_instances"`
	EquippedArmourReinforcements map[string]int          `json:"equipped_armour_reinforcements"`
	EquippedSpellbook            string                  `json:"equipped_spellbook"`
	ActiveSpellbook              string                  `json:"active_spellbook"`
	LastEquipmentSequence        uint64                  `json:"last_equipment_sequence"`
	LastEquipmentPulse           uint64                  `json:"last_equipment_pulse"`
	CombatStyle                  string                  `json:"combat_style"`
	LastCombatStyleSequence      uint64                  `json:"last_combat_style_sequence"`
	ActivePrayers                []string                `json:"active_prayers"`
	ActivePrayerSet              string                  `json:"active_prayer_set"`
	LastPrayerSequence           uint64                  `json:"last_prayer_sequence"`
	SelectedSpell                string                  `json:"selected_spell"`
	LastSpellSequence            uint64                  `json:"last_spell_sequence"`
	DodgeRate                    float64                 `json:"dodge_rate"`
	CombatLevel                  int                     `json:"combat_level"`
	CombatStats                  CombatStatsSnapshot     `json:"combat_stats"`
	CombatMatchup                CombatMatchupSnapshot   `json:"combat_matchup"`
	ExperienceLocked             bool                    `json:"experience_locked"`
	LastExperienceLockSequence   uint64                  `json:"last_experience_lock_sequence"`
	Inventory                    InventorySnapshot       `json:"inventory"`
	Bank                         *InventorySnapshot      `json:"bank,omitempty"`
	AmmunitionPouch              AmmunitionPouchSnapshot `json:"ammunition_pouch"`
	LastAmmunitionSequence       uint64                  `json:"last_ammunition_sequence,omitempty"`
	RingRack                     RingRackSnapshot        `json:"ring_rack"`
	LastRingSequence             uint64                  `json:"last_ring_sequence,omitempty"`
	RunePouch                    RunePouchSnapshot       `json:"rune_pouch"`
	LastRuneSequence             uint64                  `json:"last_rune_sequence,omitempty"`
	LastBankSequence             uint64                  `json:"last_bank_sequence,omitempty"`
	LastMerchantSequence         uint64                  `json:"last_merchant_sequence,omitempty"`
	LastPickupSequence           uint64                  `json:"last_pickup_sequence"`
	LastConsumeSequence          uint64                  `json:"last_consume_sequence"`
	NextConsumePulse             uint64                  `json:"next_consume_pulse"`
	LastBrewSequence             uint64                  `json:"last_brew_sequence"`
	BrewRemaining                int                     `json:"brew_remaining"`
	NextBrewPulse                uint64                  `json:"next_brew_pulse"`
	BrewingResult                string                  `json:"brewing_result"`
	LastFletchSequence           uint64                  `json:"last_fletch_sequence"`
	FletchRemaining              int                     `json:"fletch_remaining"`
	NextFletchPulse              uint64                  `json:"next_fletch_pulse"`
	FletchingResult              string                  `json:"fletching_result"`
	GatheringSpot                string                  `json:"gathering_spot"`
	GatheringResult              string                  `json:"gathering_result"`
	NextGatherPulse              uint64                  `json:"next_gather_pulse"`
	LastGatherSequence           uint64                  `json:"last_gather_sequence"`
	Skills                       []SkillSnapshot         `json:"skills"`
	ActiveCombatBoosts           []ActiveCombatBoost     `json:"active_combat_boosts"`
}

type ActiveCombatBoost struct {
	Skill          string `json:"skill"`
	Levels         int    `json:"levels"`
	ExpiresAtPulse uint64 `json:"expires_at_pulse"`
}

type CombatStatsSnapshot struct {
	Weapon                  string         `json:"weapon"`
	Style                   string         `json:"style"`
	Spell                   string         `json:"spell"`
	AccuracySkill           string         `json:"accuracy_skill"`
	AccuracyLevel           int            `json:"accuracy_level"`
	WeaponAccuracyBonus     int            `json:"weapon_accuracy_bonus"`
	EquipmentAccuracyBonus  int            `json:"equipment_accuracy_bonus"`
	StyleAccuracyBonus      int            `json:"style_accuracy_bonus"`
	SpellAccuracyBonus      int            `json:"spell_accuracy_bonus"`
	PrayerHitChanceBonus    float64        `json:"prayer_hit_chance_bonus"`
	TotalAccuracy           int            `json:"total_accuracy"`
	DamageSkill             string         `json:"damage_skill"`
	DamageLevel             int            `json:"damage_level"`
	StrengthBonus           int            `json:"strength_bonus"`
	EquipmentStrengthBonus  int            `json:"equipment_strength_bonus"`
	AmmunitionStrengthBonus int            `json:"ammunition_strength_bonus"`
	SpellPower              int            `json:"spell_power"`
	SkillDamageBonus        float64        `json:"skill_damage_bonus"`
	PrayerDamageBonus       float64        `json:"prayer_damage_bonus"`
	PrayerDamageReduction   float64        `json:"prayer_damage_reduction"`
	BaseAttackInterval      uint64         `json:"base_attack_interval"`
	AttackInterval          uint64         `json:"attack_interval"`
	AttackIntervalModifier  int            `json:"attack_interval_modifier"`
	MinimumDamage           int            `json:"minimum_damage"`
	MaximumDamage           int            `json:"maximum_damage"`
	DefenceLevel            int            `json:"defence_level"`
	DefenceBonuses          DefenceBonuses `json:"defence_bonuses"`
	AgilityLevel            int            `json:"agility_level"`
	EquipmentEvasionBonus   int            `json:"equipment_evasion_bonus"`
	TotalEvasion            int            `json:"total_evasion"`
	ArmourDamageReduction   float64        `json:"armour_damage_reduction"`
	MaximumHealthBonus      int            `json:"maximum_health_bonus"`
}

type CombatMatchupSnapshot struct {
	TargetID            string  `json:"target_id"`
	PlayerHitChance     float64 `json:"player_hit_chance"`
	PlayerMinimumDamage int     `json:"player_minimum_damage"`
	PlayerMaximumDamage int     `json:"player_maximum_damage"`
	PlayerAttackRange   float64 `json:"player_attack_range"`
	TargetHitChance     float64 `json:"target_hit_chance"`
	TargetMinimumDamage int     `json:"target_minimum_damage"`
	TargetMaximumDamage int     `json:"target_maximum_damage"`
	TargetAttackRange   float64 `json:"target_attack_range"`
}

type DefenceBonuses struct {
	Stab   int `json:"stab"`
	Slash  int `json:"slash"`
	Crush  int `json:"crush"`
	Ranged int `json:"ranged"`
	Magic  int `json:"magic"`
}

type SkillSnapshot struct {
	ID                     string `json:"id"`
	Experience             int64  `json:"experience"`
	Level                  int    `json:"level"`
	CurrentLevelExperience int64  `json:"current_level_experience"`
	NextLevelExperience    int64  `json:"next_level_experience"`
}

type InventorySnapshot struct {
	Capacity          int             `json:"capacity"`
	NextExpansionCost int             `json:"next_expansion_cost,omitempty"`
	Items             []InventoryItem `json:"items"`
}

type AmmunitionPouchSnapshot struct {
	Capacity      int             `json:"capacity"`
	UnlockedSlots int             `json:"unlocked_slots"`
	Items         []InventoryItem `json:"items"`
}

type RingRackSnapshot struct {
	Capacity int             `json:"capacity"`
	Items    []InventoryItem `json:"items"`
}

type RunePouchSnapshot struct {
	Capacity      int             `json:"capacity"`
	UnlockedSlots int             `json:"unlocked_slots"`
	Items         []InventoryItem `json:"items"`
}

type InventoryItem struct {
	Slot          int    `json:"slot"`
	ItemID        string `json:"item_id"`
	Quantity      int    `json:"quantity"`
	InstanceID    string `json:"instance_id,omitempty"`
	Reinforcement int    `json:"reinforcement,omitempty"`
	SellPrice     int    `json:"sell_price,omitempty"`
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

	if message.Type != MessageAuthenticate && message.Type != MessageRegisterAccount && message.Type != MessageLogoutAccount && message.Type != MessageMoveTo && message.Type != MessageEngageTarget &&
		message.Type != MessageDisengage && message.Type != MessageEquipItem && message.Type != MessageUnequipItem &&
		message.Type != MessagePickupItem && message.Type != MessageDropItem &&
		message.Type != MessageMoveInventoryItem && message.Type != MessageCombatStyle &&
		message.Type != MessageExperienceLock && message.Type != MessageDebugSetSkill &&
		message.Type != MessageSetPrayer && message.Type != MessageSetSpell &&
		message.Type != MessageConsumeItem && message.Type != MessageBrew && message.Type != MessageFletch && message.Type != MessageCutGem && message.Type != MessageLightFire && message.Type != MessageGather &&
		message.Type != MessageBankDeposit && message.Type != MessageBankWithdraw && message.Type != MessageBankExpand && message.Type != MessageMerchantSell &&
		message.Type != MessageAmmunitionDeposit && message.Type != MessageAmmunitionWithdraw &&
		message.Type != MessageRingDeposit && message.Type != MessageRingWithdraw &&
		message.Type != MessageRuneDeposit && message.Type != MessageRuneWithdraw {
		return ClientMessage{}, fmt.Errorf("unsupported message type %q", message.Type)
	}
	if message.Type != MessageAuthenticate && message.Sequence == 0 {
		return ClientMessage{}, fmt.Errorf("sequence must be positive")
	}
	if message.Type == MessageRegisterAccount && (message.Email == "" || message.Password == "") {
		return ClientMessage{}, fmt.Errorf("email and password are required")
	}
	if message.Type == MessageMoveTo && message.Destination == nil {
		return ClientMessage{}, fmt.Errorf("destination is required")
	}
	if message.Type == MessageEngageTarget && message.TargetID == "" {
		return ClientMessage{}, fmt.Errorf("target ID is required")
	}
	if (message.Type == MessageEquipItem || message.Type == MessageUnequipItem ||
		message.Type == MessageConsumeItem || message.Type == MessageCutGem || message.Type == MessageLightFire) && message.ItemID == "" {
		return ClientMessage{}, fmt.Errorf("item ID is required")
	}
	if message.Type == MessageConsumeItem && message.InventorySlot < 0 {
		return ClientMessage{}, fmt.Errorf("inventory slot cannot be negative")
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
	if message.Type == MessageDebugSetSkill && message.SkillID == "" {
		return ClientMessage{}, fmt.Errorf("skill ID is required")
	}
	if message.Type == MessageDebugSetSkill && !message.SetExperience && (message.Level < 1 || message.Level > 100) {
		return ClientMessage{}, fmt.Errorf("skill level must be between 1 and 100")
	}
	if message.Type == MessageDebugSetSkill && message.SetExperience && message.Experience < 0 {
		return ClientMessage{}, fmt.Errorf("skill experience cannot be negative")
	}
	if message.Type == MessageGather && message.SpotID == "" {
		return ClientMessage{}, fmt.Errorf("gathering spot ID is required")
	}
	if message.Type == MessageGather && message.Quantity < 0 {
		return ClientMessage{}, fmt.Errorf("gather quantity cannot be negative")
	}
	if message.Type == MessageBrew {
		if message.Quantity < 0 {
			return ClientMessage{}, fmt.Errorf("brew quantity cannot be negative")
		}
		if len(message.ItemIDs) < 2 || len(message.ItemIDs) > 3 {
			return ClientMessage{}, fmt.Errorf("brewing requires two or three items")
		}
		for _, itemID := range message.ItemIDs {
			if itemID == "" {
				return ClientMessage{}, fmt.Errorf("brewing item IDs cannot be empty")
			}
		}
	}
	if message.Type == MessageFletch {
		if message.RecipeID == "" {
			return ClientMessage{}, fmt.Errorf("fletching recipe ID is required")
		}
		if message.Quantity < 0 {
			return ClientMessage{}, fmt.Errorf("fletching quantity cannot be negative")
		}
	}
	if message.Type == MessageBankDeposit && message.ItemID == "" {
		return ClientMessage{}, fmt.Errorf("item ID is required")
	}
	if message.Type == MessageBankDeposit && message.InventorySlot < 0 {
		return ClientMessage{}, fmt.Errorf("inventory slot cannot be negative")
	}
	if message.Type == MessageBankWithdraw && message.BankSlot < 0 {
		return ClientMessage{}, fmt.Errorf("bank slot cannot be negative")
	}
	if message.Type == MessageBankWithdraw && message.Quantity < 0 {
		return ClientMessage{}, fmt.Errorf("withdraw quantity cannot be negative")
	}
	if message.Type == MessageMerchantSell && message.InventorySlot < 0 {
		return ClientMessage{}, fmt.Errorf("inventory slot cannot be negative")
	}
	if message.Type == MessageAmmunitionDeposit && message.InventorySlot < 0 {
		return ClientMessage{}, fmt.Errorf("inventory slot cannot be negative")
	}
	if message.Type == MessageAmmunitionWithdraw && message.AmmunitionSlot < 0 {
		return ClientMessage{}, fmt.Errorf("ammunition slot cannot be negative")
	}
	if message.Type == MessageRingDeposit && (message.ItemID == "" || message.InstanceID == "") {
		return ClientMessage{}, fmt.Errorf("ring item and instance IDs are required")
	}
	if message.Type == MessageRingWithdraw && message.RingSlot < 0 {
		return ClientMessage{}, fmt.Errorf("ring slot cannot be negative")
	}
	if message.Type == MessageRuneDeposit && message.InventorySlot < 0 {
		return ClientMessage{}, fmt.Errorf("inventory slot cannot be negative")
	}
	if message.Type == MessageRuneWithdraw && message.RuneSlot < 0 {
		return ClientMessage{}, fmt.Errorf("rune slot cannot be negative")
	}
	return message, nil
}
