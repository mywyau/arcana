package websockettransport

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/mywyau/arcana/server/internal/character"
	"github.com/mywyau/arcana/server/internal/identity"
	"github.com/mywyau/arcana/server/internal/protocol"
	"github.com/mywyau/arcana/server/internal/simulation"
)

const (
	connectionReadLimit   = 4 * 1024
	playerSpeed           = 6.0
	writeTimeout          = 2 * time.Second
	authenticationTimeout = 5 * time.Second
	characterSaveTimeout  = 2 * time.Second
	characterSaveInterval = 30 * time.Second
)

type Handler struct {
	world            *simulation.World
	snapshotInterval time.Duration
	guestIdentities  *identity.GuestManager
	characters       character.Repository
}

func NewHandler(world *simulation.World, snapshotInterval time.Duration) (*Handler, error) {
	guestIdentities, err := identity.NewGuestManager(identity.NewMemoryGuestRepository())
	if err != nil {
		return nil, err
	}
	return NewHandlerWithRepositories(
		world,
		snapshotInterval,
		guestIdentities,
		character.NewMemoryRepository(),
	)
}

func NewHandlerWithGuestManager(
	world *simulation.World,
	snapshotInterval time.Duration,
	guestIdentities *identity.GuestManager,
) (*Handler, error) {
	return NewHandlerWithRepositories(
		world,
		snapshotInterval,
		guestIdentities,
		character.NewMemoryRepository(),
	)
}

func NewHandlerWithRepositories(
	world *simulation.World,
	snapshotInterval time.Duration,
	guestIdentities *identity.GuestManager,
	characters character.Repository,
) (*Handler, error) {
	if world == nil {
		return nil, fmt.Errorf("world is required")
	}
	if snapshotInterval <= 0 {
		return nil, fmt.Errorf("snapshot interval must be positive")
	}
	if guestIdentities == nil {
		return nil, fmt.Errorf("guest identity manager is required")
	}
	if characters == nil {
		return nil, fmt.Errorf("character repository is required")
	}
	return &Handler{
		world: world, snapshotInterval: snapshotInterval,
		guestIdentities: guestIdentities, characters: characters,
	}, nil
}

func (h *Handler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	connection, err := websocket.Accept(response, request, nil)
	if err != nil {
		return
	}
	defer connection.CloseNow()
	connection.SetReadLimit(connectionReadLimit)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	guestIdentity, err := h.authenticate(ctx, connection)
	if err != nil {
		code := "authentication_failed"
		clientError := errors.New("authentication unavailable")
		if errors.Is(err, identity.ErrInvalidGuestToken) {
			code = "invalid_guest_token"
			clientError = identity.ErrInvalidGuestToken
		} else {
			log.Printf("authenticate WebSocket connection: %v", err)
		}
		_ = writeMessage(ctx, connection, errorMessage(code, clientError))
		_ = connection.Close(websocket.StatusPolicyViolation, "authentication failed")
		return
	}
	playerID := simulation.PlayerID(guestIdentity.PlayerID)
	characterState, found, err := h.characters.Load(ctx, string(playerID))
	if err != nil {
		log.Printf("load character %s: %v", playerID, err)
		_ = writeMessage(ctx, connection, errorMessage(
			"character_load_failed",
			errors.New("could not load character"),
		))
		_ = connection.Close(websocket.StatusInternalError, "could not load character")
		return
	}
	if !found {
		characterState = simulation.DefaultPersistentPlayerState()
		if err := h.characters.Save(ctx, string(playerID), characterState); err != nil {
			log.Printf("create character %s: %v", playerID, err)
			_ = writeMessage(ctx, connection, errorMessage(
				"character_save_failed",
				errors.New("could not create character"),
			))
			_ = connection.Close(websocket.StatusInternalError, "could not create character")
			return
		}
	}
	if err := h.world.QueueJoinWithPersistentState(playerID, characterState, playerSpeed); err != nil {
		log.Printf("join character %s: %v", playerID, err)
		_ = connection.Close(websocket.StatusInternalError, "could not join world")
		return
	}
	defer func() {
		h.saveCharacter(playerID)
		_ = h.world.QueueLeave(playerID)
	}()

	responses := make(chan protocol.ServerMessage, 8)

	if err := writeMessage(ctx, connection, protocol.ServerMessage{
		Type:       protocol.MessageWelcome,
		PlayerID:   string(playerID),
		GuestToken: guestIdentity.Token,
	}); err != nil {
		return
	}
	go h.readCommands(ctx, cancel, connection, playerID, responses)

	ticker := time.NewTicker(h.snapshotInterval)
	defer ticker.Stop()
	saveTicker := time.NewTicker(characterSaveInterval)
	defer saveTicker.Stop()

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
				Snapshot: snapshotMessage(h.world.Snapshot(), playerID),
			}
			if err := writeMessage(ctx, connection, message); err != nil {
				return
			}
		case <-saveTicker.C:
			h.saveCharacter(playerID)
		}
	}
}

func (h *Handler) saveCharacter(playerID simulation.PlayerID) {
	player, found := h.world.Player(playerID)
	if !found {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), characterSaveTimeout)
	defer cancel()
	if err := h.characters.Save(ctx, string(playerID), simulation.PersistentStateFromSnapshot(player)); err != nil {
		log.Printf("save character %s: %v", playerID, err)
	}
}

// SaveAll writes the durable state of every active player. It is used during
// graceful server shutdown in addition to periodic and disconnect saves.
func (h *Handler) SaveAll(ctx context.Context) error {
	var saveErrors []error
	for _, player := range h.world.Snapshot().Players {
		if err := h.characters.Save(
			ctx,
			string(player.ID),
			simulation.PersistentStateFromSnapshot(player),
		); err != nil {
			saveErrors = append(saveErrors, fmt.Errorf("save character %s: %w", player.ID, err))
		}
	}
	return errors.Join(saveErrors...)
}

func (h *Handler) authenticate(ctx context.Context, connection *websocket.Conn) (identity.GuestIdentity, error) {
	authCtx, cancel := context.WithTimeout(ctx, authenticationTimeout)
	defer cancel()
	_, data, err := connection.Read(authCtx)
	if err != nil {
		return identity.GuestIdentity{}, fmt.Errorf("read authentication: %w", err)
	}
	message, err := protocol.DecodeClientMessage(data)
	if err != nil {
		return identity.GuestIdentity{}, err
	}
	if message.Type != protocol.MessageAuthenticate {
		return identity.GuestIdentity{}, fmt.Errorf("authentication must be the first message")
	}
	return h.guestIdentities.Authenticate(ctx, message.GuestToken)
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
		case protocol.MessageAuthenticate:
			rejectionCode = "invalid_message"
			err = fmt.Errorf("connection is already authenticated")
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
				PlayerID:   playerID,
				Sequence:   message.Sequence,
				ItemID:     simulation.WeaponID(message.ItemID),
				InstanceID: message.InstanceID,
			})
		case protocol.MessageUnequipItem:
			rejectionCode = "equipment_rejected"
			err = h.world.QueueUnequip(simulation.UnequipCommand{
				PlayerID:   playerID,
				Sequence:   message.Sequence,
				ItemID:     message.ItemID,
				InstanceID: message.InstanceID,
			})
		case protocol.MessageCombatStyle:
			rejectionCode = "combat_style_rejected"
			err = h.world.QueueCombatStyle(simulation.CombatStyleCommand{
				PlayerID: playerID,
				Sequence: message.Sequence,
				Style:    simulation.CombatStyleID(message.CombatStyle),
			})
		case protocol.MessageSetPrayer:
			rejectionCode = "prayer_rejected"
			err = h.world.QueuePrayer(simulation.PrayerCommand{
				PlayerID: playerID,
				Sequence: message.Sequence,
				PrayerID: simulation.PrayerID(message.PrayerID),
				Active:   message.Active,
			})
		case protocol.MessageSetSpell:
			rejectionCode = "spell_rejected"
			err = h.world.QueueSpell(simulation.SpellCommand{
				PlayerID: playerID,
				Sequence: message.Sequence,
				SpellID:  simulation.SpellID(message.SpellID),
			})
		case protocol.MessageCutGem:
			rejectionCode = "gem_cut_rejected"
			err = h.world.CutGem(playerID, message.Sequence, message.ItemID)
		case protocol.MessagePickupItem:
			rejectionCode = "pickup_rejected"
			err = h.world.QueuePickup(simulation.PickupCommand{
				PlayerID: playerID,
				Sequence: message.Sequence,
				DropID:   message.DropID,
			})
		case protocol.MessageDropItem:
			rejectionCode = "drop_rejected"
			err = h.world.QueueDropItem(simulation.DropItemCommand{
				PlayerID:   playerID,
				Sequence:   message.Sequence,
				ItemID:     message.ItemID,
				InstanceID: message.InstanceID,
			})
		case protocol.MessageMoveInventoryItem:
			rejectionCode = "inventory_move_rejected"
			err = h.world.QueueMoveInventoryItem(simulation.MoveInventoryItemCommand{
				PlayerID: playerID,
				Sequence: message.Sequence,
				FromSlot: message.FromSlot,
				ToSlot:   message.ToSlot,
			})
		case protocol.MessageConsumeItem:
			rejectionCode = "consume_rejected"
			err = h.world.QueueConsumeItem(simulation.ConsumeItemCommand{
				PlayerID: playerID,
				Sequence: message.Sequence,
				ItemID:   message.ItemID,
			})
		case protocol.MessageBrew:
			rejectionCode = "brew_rejected"
			err = h.world.QueueBrew(simulation.BrewCommand{
				PlayerID: playerID,
				Sequence: message.Sequence,
				ItemIDs:  message.ItemIDs,
				Quantity: message.Quantity,
			})
		case protocol.MessageFletch:
			rejectionCode = "fletch_rejected"
			err = h.world.QueueFletch(simulation.FletchCommand{
				PlayerID: playerID,
				Sequence: message.Sequence,
				RecipeID: message.RecipeID,
				Quantity: message.Quantity,
			})
		case protocol.MessageGather:
			rejectionCode = "gather_rejected"
			err = h.world.QueueGather(simulation.GatherCommand{
				PlayerID:   playerID,
				Sequence:   message.Sequence,
				SpotID:     message.SpotID,
				RecipeID:   message.RecipeID,
				InstanceID: message.InstanceID,
			})
		case protocol.MessageBankDeposit:
			rejectionCode = "bank_rejected"
			err = h.world.QueueBankTransfer(simulation.BankTransferCommand{
				PlayerID: playerID, Sequence: message.Sequence, Deposit: true,
				ItemID: message.ItemID, InstanceID: message.InstanceID, InventorySlot: message.InventorySlot,
			})
		case protocol.MessageBankWithdraw:
			rejectionCode = "bank_rejected"
			err = h.world.QueueBankTransfer(simulation.BankTransferCommand{
				PlayerID: playerID, Sequence: message.Sequence, BankSlot: message.BankSlot, Quantity: message.Quantity,
			})
		case protocol.MessageAmmunitionDeposit:
			rejectionCode = "ammunition_rejected"
			err = h.world.QueueAmmunitionTransfer(simulation.AmmunitionTransferCommand{
				PlayerID: playerID, Sequence: message.Sequence, Deposit: true, InventorySlot: message.InventorySlot,
			})
		case protocol.MessageAmmunitionWithdraw:
			rejectionCode = "ammunition_rejected"
			err = h.world.QueueAmmunitionTransfer(simulation.AmmunitionTransferCommand{
				PlayerID: playerID, Sequence: message.Sequence, PouchSlot: message.AmmunitionSlot,
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

func snapshotMessage(snapshot simulation.Snapshot, localPlayerID simulation.PlayerID) *protocol.WorldSnapshot {
	message := &protocol.WorldSnapshot{
		Tick:           snapshot.Tick,
		ActionPulse:    snapshot.ActionPulse,
		Targets:        make([]protocol.TargetSnapshot, 0, len(snapshot.Targets)),
		Drops:          make([]protocol.DropSnapshot, 0, len(snapshot.Drops)),
		CombatEvents:   make([]protocol.CombatEvent, 0, len(snapshot.CombatEvents)),
		GatheringSpots: make([]protocol.GatheringSpotSnapshot, 0, len(snapshot.GatheringSpots)),
		BankChests:     make([]protocol.BankChestSnapshot, 0, len(snapshot.BankChests)),
		Players:        make([]protocol.PlayerSnapshot, 0, len(snapshot.Players)),
	}
	for _, target := range snapshot.Targets {
		message.Targets = append(message.Targets, targetSnapshotMessage(target))
	}
	for _, drop := range snapshot.Drops {
		message.Drops = append(message.Drops, protocol.DropSnapshot{
			ID:            drop.ID,
			Position:      protocol.Position{X: drop.Position.X, Y: drop.Position.Y, Z: drop.Position.Z},
			ItemID:        drop.ItemID,
			Quantity:      drop.Quantity,
			InstanceID:    drop.InstanceID,
			Reinforcement: drop.Reinforcement,
		})
	}
	for _, event := range snapshot.CombatEvents {
		message.CombatEvents = append(message.CombatEvents, protocol.CombatEvent{
			ID:       event.ID,
			Pulse:    event.Pulse,
			Type:     string(event.Type),
			SourceID: event.SourceID,
			TargetID: event.TargetID,
			Amount:   event.Amount,
			SpellID:  string(event.SpellID),
		})
	}
	for _, spot := range snapshot.GatheringSpots {
		message.GatheringSpots = append(message.GatheringSpots, protocol.GatheringSpotSnapshot{
			ID: spot.ID, Name: spot.Name, Skill: string(spot.Skill),
			Position: protocol.Position{X: spot.Position.X, Y: spot.Position.Y, Z: spot.Position.Z},
			Range:    spot.Range, RequiredLevel: spot.RequiredLevel,
			CurrentAmount: spot.CurrentAmount, MaximumAmount: spot.MaximumAmount,
			Station: string(spot.Station), RecipeID: spot.RecipeID,
		})
	}
	for _, chest := range snapshot.BankChests {
		message.BankChests = append(message.BankChests, protocol.BankChestSnapshot{
			ID: chest.ID, Name: chest.Name,
			Position: protocol.Position{X: chest.Position.X, Y: chest.Position.Y, Z: chest.Position.Z},
			Range:    chest.Range,
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
			LastActionSequence:           player.LastActionSequence,
			LastActionPulse:              player.LastActionPulse,
			LastActionResult:             player.LastActionResult,
			CombatTarget:                 player.CombatTarget,
			NextAttackPulse:              player.NextAttackPulse,
			AutoApproaching:              player.AutoApproaching,
			AutoPickingUp:                player.AutoPickingUp,
			Health:                       player.Health,
			MaximumHealth:                player.MaximumHealth,
			Prayer:                       player.Prayer,
			MaximumPrayer:                player.MaximumPrayer,
			Defeated:                     player.Defeated,
			RespawnAtPulse:               player.RespawnAtPulse,
			EquippedWeapon:               string(player.EquippedWeapon),
			EquippedWeaponInstanceID:     player.EquippedWeaponInstanceID,
			EquippedWeaponReinforcement:  player.EquippedWeaponReinforcement,
			EquippedArmour:               make(map[string]string, len(player.EquippedArmour)),
			EquippedArmourInstances:      make(map[string]string, len(player.EquippedArmourInstances)),
			EquippedArmourReinforcements: make(map[string]int, len(player.EquippedArmourReinforcements)),
			EquippedSpellbook:            player.EquippedSpellbook,
			ActiveSpellbook:              string(player.ActiveSpellbook),
			LastEquipmentSequence:        player.LastEquipmentSequence,
			LastEquipmentPulse:           player.LastEquipmentPulse,
			CombatStyle:                  string(player.CombatStyle),
			LastCombatStyleSequence:      player.LastCombatStyleSequence,
			ActivePrayers:                make([]string, 0, len(player.ActivePrayers)),
			ActivePrayerSet:              string(player.ActivePrayerSet),
			LastPrayerSequence:           player.LastPrayerSequence,
			SelectedSpell:                string(player.SelectedSpell),
			LastSpellSequence:            player.LastSpellSequence,
			DodgeRate:                    player.DodgeRate,
			CombatStats: protocol.CombatStatsSnapshot{
				Weapon:                 string(player.CombatStats.Weapon),
				Style:                  string(player.CombatStats.Style),
				Spell:                  string(player.CombatStats.Spell),
				AccuracySkill:          string(player.CombatStats.AccuracySkill),
				AccuracyLevel:          player.CombatStats.AccuracyLevel,
				WeaponAccuracyBonus:    player.CombatStats.WeaponAccuracyBonus,
				EquipmentAccuracyBonus: player.CombatStats.EquipmentAccuracyBonus,
				StyleAccuracyBonus:     player.CombatStats.StyleAccuracyBonus,
				SpellAccuracyBonus:     player.CombatStats.SpellAccuracyBonus,
				PrayerAccuracyBonus:    player.CombatStats.PrayerAccuracyBonus,
				TotalAccuracy:          player.CombatStats.TotalAccuracy,
				DamageSkill:            string(player.CombatStats.DamageSkill),
				DamageLevel:            player.CombatStats.DamageLevel,
				StrengthBonus:          player.CombatStats.StrengthBonus,
				EquipmentStrengthBonus: player.CombatStats.EquipmentStrengthBonus,
				SpellPower:             player.CombatStats.SpellPower,
				SkillDamageBonus:       player.CombatStats.SkillDamageBonus,
				PrayerDamageBonus:      player.CombatStats.PrayerDamageBonus,
				PrayerDamageReduction:  player.CombatStats.PrayerDamageReduction,
				BaseAttackInterval:     player.CombatStats.BaseAttackInterval,
				AttackInterval:         player.CombatStats.AttackInterval,
				AttackIntervalModifier: player.CombatStats.AttackIntervalModifier,
				MinimumDamage:          player.CombatStats.MinimumDamage,
				MaximumDamage:          player.CombatStats.MaximumDamage,
				DefenceLevel:           player.CombatStats.DefenceLevel,
				ArmourDamageReduction:  player.CombatStats.ArmourDamageReduction,
				MaximumHealthBonus:     player.CombatStats.MaximumHealthBonus,
				DefenceBonuses: protocol.DefenceBonuses{
					Stab: player.CombatStats.DefenceBonuses.Stab, Slash: player.CombatStats.DefenceBonuses.Slash,
					Crush: player.CombatStats.DefenceBonuses.Crush, Ranged: player.CombatStats.DefenceBonuses.Ranged,
					Magic: player.CombatStats.DefenceBonuses.Magic,
				},
			},
			Inventory: protocol.InventorySnapshot{
				Capacity: player.Inventory.Capacity,
				Items:    make([]protocol.InventoryItem, 0, len(player.Inventory.Items)),
			},
			LastBankSequence:    player.LastBankSequence,
			LastPickupSequence:  player.LastPickupSequence,
			LastConsumeSequence: player.LastConsumeSequence,
			NextConsumePulse:    player.NextConsumePulse,
			LastBrewSequence:    player.LastBrewSequence,
			BrewRemaining:       player.BrewRemaining,
			NextBrewPulse:       player.NextBrewPulse,
			BrewingResult:       player.BrewingResult,
			LastFletchSequence:  player.LastFletchSequence,
			FletchRemaining:     player.FletchRemaining,
			NextFletchPulse:     player.NextFletchPulse,
			FletchingResult:     player.FletchingResult,
			GatheringSpot:       player.GatheringSpot,
			GatheringResult:     player.GatheringResult,
			NextGatherPulse:     player.NextGatherPulse,
			LastGatherSequence:  player.LastGatherSequence,
			Skills:              make([]protocol.SkillSnapshot, 0, len(player.Skills)),
			Path:                make([]protocol.Position, 0, len(player.Path)),
		}
		for _, prayerID := range player.ActivePrayers {
			playerMessage.ActivePrayers = append(playerMessage.ActivePrayers, string(prayerID))
		}
		for slot, itemID := range player.EquippedArmour {
			playerMessage.EquippedArmour[string(slot)] = itemID
		}
		for slot, instanceID := range player.EquippedArmourInstances {
			playerMessage.EquippedArmourInstances[string(slot)] = instanceID
		}
		for slot, reinforcement := range player.EquippedArmourReinforcements {
			playerMessage.EquippedArmourReinforcements[string(slot)] = reinforcement
		}
		for _, item := range player.Inventory.Items {
			playerMessage.Inventory.Items = append(playerMessage.Inventory.Items, protocol.InventoryItem{
				Slot: item.Slot, ItemID: item.ItemID, Quantity: item.Quantity,
				InstanceID: item.InstanceID, Reinforcement: item.Reinforcement,
			})
		}
		if player.ID == localPlayerID {
			playerMessage.AmmunitionPouch = protocol.AmmunitionPouchSnapshot{
				Capacity: player.AmmunitionPouch.Capacity, UnlockedSlots: player.AmmunitionPouch.UnlockedSlots,
				Items: make([]protocol.InventoryItem, 0, len(player.AmmunitionPouch.Items)),
			}
			playerMessage.LastAmmunitionSequence = player.LastAmmunitionSequence
			for _, item := range player.AmmunitionPouch.Items {
				playerMessage.AmmunitionPouch.Items = append(playerMessage.AmmunitionPouch.Items, protocol.InventoryItem{
					Slot: item.Slot, ItemID: item.ItemID, Quantity: item.Quantity,
				})
			}
			playerMessage.Bank = &protocol.InventorySnapshot{
				Capacity: player.Bank.Capacity,
				Items:    make([]protocol.InventoryItem, 0, len(player.Bank.Items)),
			}
			for _, item := range player.Bank.Items {
				playerMessage.Bank.Items = append(playerMessage.Bank.Items, protocol.InventoryItem{
					Slot: item.Slot, ItemID: item.ItemID, Quantity: item.Quantity,
					InstanceID: item.InstanceID, Reinforcement: item.Reinforcement,
				})
			}
		}
		for _, skill := range player.Skills {
			playerMessage.Skills = append(playerMessage.Skills, protocol.SkillSnapshot{
				ID:                     string(skill.ID),
				Experience:             skill.Experience,
				Level:                  skill.Level,
				CurrentLevelExperience: skill.CurrentLevelExperience,
				NextLevelExperience:    skill.NextLevelExperience,
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

func targetSnapshotMessage(target simulation.TargetSnapshot) protocol.TargetSnapshot {
	return protocol.TargetSnapshot{
		ID:   target.ID,
		Name: target.Name,
		Kind: target.Kind,
		Position: protocol.Position{
			X: target.Position.X,
			Y: target.Position.Y,
			Z: target.Position.Z,
		},
		Health:          target.Health,
		MaximumHealth:   target.MaximumHealth,
		TargetID:        string(target.TargetID),
		NextAttackPulse: target.NextAttackPulse,
		Defeated:        target.Defeated,
		RespawnAtPulse:  target.RespawnAtPulse,
		DodgeRate:       target.DodgeRate,
		Accuracy:        target.Accuracy,
		Evasion:         target.Evasion,
		AttackType:      string(target.AttackType),
	}
}
