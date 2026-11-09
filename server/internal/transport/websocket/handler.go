package websockettransport

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/mywyau/arcana/server/internal/account"
	"github.com/mywyau/arcana/server/internal/character"
	"github.com/mywyau/arcana/server/internal/identity"
	"github.com/mywyau/arcana/server/internal/protocol"
	"github.com/mywyau/arcana/server/internal/simulation"
)

const (
	connectionReadLimit          = 4 * 1024
	playerSpeed                  = 6.0
	writeTimeout                 = 2 * time.Second
	authenticationTimeout        = 5 * time.Second
	characterSaveTimeout         = 2 * time.Second
	characterSaveInterval        = 30 * time.Second
	connectionReplacementTimeout = 5 * time.Second
)

var errAccountRequired = errors.New("an account is required to enter this world")
var errIncompatibleClient = errors.New("this client is incompatible with the server; update required")
var errHTTPAccountRequired = errors.New("email and password authentication must use the account HTTP API")

type Handler struct {
	world            *simulation.World
	snapshotInterval time.Duration
	guestIdentities  *identity.GuestManager
	accounts         *account.Manager
	characters       character.Repository
	allowDebug       bool
	allowGuests      bool
	acceptOptions    websocket.AcceptOptions
	activeMu         sync.Mutex
	activeClients    map[simulation.PlayerID]*activeClient
}

type activeClient struct {
	replaced chan struct{}
	done     chan struct{}
}

type HandlerOptions struct {
	AllowDebugCommands bool
	AllowedOrigins     []string
	AccountManager     *account.Manager
	RequireAccounts    bool
}

type outboundResponse struct {
	message    protocol.ServerMessage
	closeAfter bool
}

func NewHandler(world *simulation.World, snapshotInterval time.Duration) (*Handler, error) {
	return NewHandlerWithOptions(
		world, snapshotInterval, HandlerOptions{AllowDebugCommands: true},
	)
}

func NewHandlerWithOptions(
	world *simulation.World,
	snapshotInterval time.Duration,
	options HandlerOptions,
) (*Handler, error) {
	guestIdentities, err := identity.NewGuestManager(identity.NewMemoryGuestRepository())
	if err != nil {
		return nil, err
	}
	return NewHandlerWithRepositoriesAndOptions(
		world,
		snapshotInterval,
		guestIdentities,
		character.NewMemoryRepository(),
		options,
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
	return NewHandlerWithRepositoriesAndOptions(
		world, snapshotInterval, guestIdentities, characters,
		HandlerOptions{AllowDebugCommands: true},
	)
}

func NewHandlerWithRepositoriesAndOptions(
	world *simulation.World,
	snapshotInterval time.Duration,
	guestIdentities *identity.GuestManager,
	characters character.Repository,
	options HandlerOptions,
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
	accountManager := options.AccountManager
	if accountManager == nil {
		accountRepository := account.NewMemoryRepository()
		var err error
		accountManager, err = account.NewManager(accountRepository, accountRepository)
		if err != nil {
			return nil, err
		}
	}
	return &Handler{
		world: world, snapshotInterval: snapshotInterval,
		guestIdentities: guestIdentities, accounts: accountManager, characters: characters,
		allowDebug:  options.AllowDebugCommands,
		allowGuests: !options.RequireAccounts,
		acceptOptions: websocket.AcceptOptions{
			OriginPatterns: append([]string(nil), options.AllowedOrigins...),
		},
		activeClients: make(map[simulation.PlayerID]*activeClient),
	}, nil
}

func (h *Handler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	connection, err := websocket.Accept(response, request, &h.acceptOptions)
	if err != nil {
		return
	}
	defer connection.CloseNow()
	connection.SetReadLimit(connectionReadLimit)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	authentication, err := h.authenticate(ctx, connection)
	if err != nil {
		code := "authentication_failed"
		clientError := errors.New("authentication unavailable")
		if errors.Is(err, identity.ErrInvalidGuestToken) {
			code = "invalid_guest_token"
			clientError = identity.ErrInvalidGuestToken
		} else if errors.Is(err, account.ErrInvalidSession) {
			code = "invalid_account_session"
			clientError = account.ErrInvalidSession
		} else if errors.Is(err, account.ErrInvalidCredentials) {
			code = "invalid_credentials"
			clientError = account.ErrInvalidCredentials
		} else if errors.Is(err, errAccountRequired) {
			code = "account_required"
			clientError = errAccountRequired
		} else if errors.Is(err, errIncompatibleClient) {
			code = "incompatible_client"
			clientError = errIncompatibleClient
		} else if errors.Is(err, errHTTPAccountRequired) {
			code = "account_http_required"
			clientError = errHTTPAccountRequired
		} else {
			log.Printf("authenticate WebSocket connection: %v", err)
		}
		_ = writeMessage(ctx, connection, errorMessage(code, clientError))
		closeReason := "authentication failed"
		if errors.Is(err, errIncompatibleClient) {
			closeReason = "incompatible client"
		}
		_ = connection.Close(websocket.StatusPolicyViolation, closeReason)
		return
	}
	playerID := simulation.PlayerID(authentication.PlayerID)
	client, previous := h.claimActiveClient(playerID)
	joinedWorld := false
	defer func() {
		if joinedWorld {
			h.saveCharacter(playerID)
			_ = h.world.QueueLeave(playerID)
			h.waitForPlayerDeparture(playerID)
		}
		h.releaseActiveClient(playerID, client)
	}()
	if previous != nil {
		select {
		case <-previous.done:
		case <-time.After(connectionReplacementTimeout):
			_ = writeMessage(ctx, connection, errorMessage(
				"connection_replacement_failed",
				errors.New("the previous connection did not close in time"),
			))
			_ = connection.Close(websocket.StatusInternalError, "replacement timed out")
			return
		}
	}
	if isClientReplaced(client) {
		h.closeReplacedClient(ctx, connection)
		return
	}
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
	joinedWorld = true

	responses := make(chan outboundResponse, 8)

	if err := writeMessage(ctx, connection, protocol.ServerMessage{
		Type:              protocol.MessageWelcome,
		PlayerID:          string(playerID),
		GuestToken:        authentication.GuestToken,
		SessionToken:      authentication.SessionToken,
		AccountEmail:      authentication.AccountEmail,
		ProductionRecipes: productionRecipeMessages(simulation.ProductionRecipeCatalog()),
		Spellbooks:        spellbookCatalogMessages(simulation.SpellbookCatalog()),
		Spells:            spellCatalogMessages(simulation.SpellCatalog()),
		Prayers:           prayerCatalogMessages(simulation.PrayerCatalog()),
		PrayerSets:        prayerSetCatalogMessages(simulation.PrayerSetCatalog()),
	}); err != nil {
		return
	}
	go h.readCommands(ctx, cancel, connection, playerID, authentication.SessionToken, responses)

	ticker := time.NewTicker(h.snapshotInterval)
	defer ticker.Stop()
	saveTicker := time.NewTicker(characterSaveInterval)
	defer saveTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-client.replaced:
			h.closeReplacedClient(ctx, connection)
			return
		case response := <-responses:
			if err := writeMessage(ctx, connection, response.message); err != nil {
				return
			}
			if response.closeAfter {
				_ = connection.Close(websocket.StatusNormalClosure, "logged out")
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

func productionRecipeMessages(recipes []simulation.ProductionRecipeSnapshot) []protocol.ProductionRecipeSnapshot {
	messages := make([]protocol.ProductionRecipeSnapshot, 0, len(recipes))
	for _, recipe := range recipes {
		message := protocol.ProductionRecipeSnapshot{
			Station: recipe.Station, Group: recipe.Group, ID: recipe.ID, Name: recipe.Name,
			RequiredLevel: recipe.RequiredLevel, RewardItem: recipe.RewardItem,
			RewardQuantity: recipe.RewardQuantity, RewardXP: recipe.RewardXP, Interval: recipe.Interval,
			RequiredTool: recipe.RequiredTool,
			Costs:        make([]protocol.ItemCost, 0, len(recipe.Costs)),
		}
		for _, cost := range recipe.Costs {
			message.Costs = append(message.Costs, protocol.ItemCost{ItemID: cost.ItemID, Quantity: cost.Quantity})
		}
		messages = append(messages, message)
	}
	return messages
}

func spellbookCatalogMessages(books []simulation.SpellbookCatalogSnapshot) []protocol.SpellbookCatalogSnapshot {
	messages := make([]protocol.SpellbookCatalogSnapshot, 0, len(books))
	for _, book := range books {
		messages = append(messages, protocol.SpellbookCatalogSnapshot{
			ID: book.ID, Name: book.Name, EquipmentItemID: book.EquipmentItemID,
		})
	}
	return messages
}

func spellCatalogMessages(spells []simulation.SpellCatalogSnapshot) []protocol.SpellCatalogSnapshot {
	messages := make([]protocol.SpellCatalogSnapshot, 0, len(spells))
	for _, spell := range spells {
		message := protocol.SpellCatalogSnapshot{
			ID: spell.ID, BookID: spell.BookID, Name: spell.Name, IconID: spell.IconID,
			CastMode: spell.CastMode, TargetType: spell.TargetType, EffectType: spell.EffectType,
			Element: spell.Element, RequiredMagicLevel: spell.RequiredMagicLevel,
			BasePower: spell.BasePower, AccuracyBonus: spell.AccuracyBonus, Range: spell.Range,
			DamageVariance: spell.DamageVariance, AttackInterval: spell.AttackInterval,
			Costs: make([]protocol.ItemCost, 0, len(spell.Costs)),
		}
		for _, cost := range spell.Costs {
			message.Costs = append(message.Costs, protocol.ItemCost{ItemID: cost.ItemID, Quantity: cost.Quantity})
		}
		messages = append(messages, message)
	}
	return messages
}

func prayerCatalogMessages(prayers []simulation.PrayerCatalogSnapshot) []protocol.PrayerCatalogSnapshot {
	messages := make([]protocol.PrayerCatalogSnapshot, 0, len(prayers))
	for _, prayer := range prayers {
		messages = append(messages, protocol.PrayerCatalogSnapshot{
			ID: prayer.ID, Name: prayer.Name, RequiredLevel: prayer.RequiredLevel,
			MeleeDamageBonus: prayer.MeleeDamageBonus, RangedDamageBonus: prayer.RangedDamageBonus,
			MagicDamageBonus: prayer.MagicDamageBonus, HealthRegeneration: prayer.HealthRegeneration,
			MeleeHitChanceBonus: prayer.MeleeHitChanceBonus, DamageReduction: prayer.DamageReduction,
			DodgeBonus: prayer.DodgeBonus, DrainAmount: prayer.DrainAmount,
			DrainIntervalPulses:  prayer.DrainIntervalPulses,
			DrainIntervalSeconds: prayer.DrainIntervalSeconds,
			ExclusiveGroup:       prayer.ExclusiveGroup,
		})
	}
	return messages
}

func prayerSetCatalogMessages(sets []simulation.PrayerSetCatalogSnapshot) []protocol.PrayerSetCatalogSnapshot {
	messages := make([]protocol.PrayerSetCatalogSnapshot, 0, len(sets))
	for _, set := range sets {
		messages = append(messages, protocol.PrayerSetCatalogSnapshot{
			ID: set.ID, Name: set.Name, PrayerIDs: append([]string(nil), set.PrayerIDs...),
		})
	}
	return messages
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

type authenticatedIdentity struct {
	PlayerID     string
	GuestToken   string
	SessionToken string
	AccountEmail string
}

func (h *Handler) claimActiveClient(playerID simulation.PlayerID) (*activeClient, *activeClient) {
	client := &activeClient{replaced: make(chan struct{}), done: make(chan struct{})}
	h.activeMu.Lock()
	previous := h.activeClients[playerID]
	h.activeClients[playerID] = client
	if previous != nil {
		close(previous.replaced)
	}
	h.activeMu.Unlock()
	return client, previous
}

func (h *Handler) releaseActiveClient(playerID simulation.PlayerID, client *activeClient) {
	h.activeMu.Lock()
	if h.activeClients[playerID] == client {
		delete(h.activeClients, playerID)
	}
	close(client.done)
	h.activeMu.Unlock()
}

func (h *Handler) waitForPlayerDeparture(playerID simulation.PlayerID) {
	deadline := time.Now().Add(connectionReplacementTimeout)
	for time.Now().Before(deadline) {
		if !h.world.HasPlayerOrPending(playerID) {
			return
		}
		time.Sleep(h.snapshotInterval)
	}
}

func isClientReplaced(client *activeClient) bool {
	select {
	case <-client.replaced:
		return true
	default:
		return false
	}
}

func (h *Handler) closeReplacedClient(ctx context.Context, connection *websocket.Conn) {
	_ = writeMessage(ctx, connection, errorMessage(
		"session_replaced",
		errors.New("this character was logged in from another client"),
	))
	_ = connection.Close(websocket.StatusPolicyViolation, "session replaced")
}

func (h *Handler) authenticate(ctx context.Context, connection *websocket.Conn) (authenticatedIdentity, error) {
	authCtx, cancel := context.WithTimeout(ctx, authenticationTimeout)
	defer cancel()
	_, data, err := connection.Read(authCtx)
	if err != nil {
		return authenticatedIdentity{}, fmt.Errorf("read authentication: %w", err)
	}
	message, err := protocol.DecodeClientMessage(data)
	if err != nil {
		return authenticatedIdentity{}, err
	}
	if message.Type != protocol.MessageAuthenticate {
		return authenticatedIdentity{}, fmt.Errorf("authentication must be the first message")
	}
	if message.ProtocolVersion != protocol.CurrentVersion {
		return authenticatedIdentity{}, errIncompatibleClient
	}
	usesSession := message.SessionToken != ""
	usesCredentials := message.Email != "" || message.Password != ""
	usesGuest := message.GuestToken != ""
	methodCount := 0
	for _, used := range []bool{usesSession, usesCredentials, usesGuest} {
		if used {
			methodCount++
		}
	}
	if methodCount > 1 || (usesCredentials && (message.Email == "" || message.Password == "")) {
		return authenticatedIdentity{}, fmt.Errorf("provide exactly one authentication method")
	}
	if usesCredentials {
		return authenticatedIdentity{}, errHTTPAccountRequired
	}
	if usesSession {
		value, err := h.accounts.AuthenticateSession(ctx, message.SessionToken)
		if err != nil {
			return authenticatedIdentity{}, err
		}
		return authenticatedIdentity{
			PlayerID: value.PlayerID, SessionToken: message.SessionToken, AccountEmail: value.Email,
		}, nil
	}
	if !h.allowGuests {
		return authenticatedIdentity{}, errAccountRequired
	}
	guestIdentity, err := h.guestIdentities.Authenticate(ctx, message.GuestToken)
	if err != nil {
		return authenticatedIdentity{}, err
	}
	return authenticatedIdentity{PlayerID: guestIdentity.PlayerID, GuestToken: guestIdentity.Token}, nil
}

func (h *Handler) readCommands(
	ctx context.Context,
	cancel context.CancelFunc,
	connection *websocket.Conn,
	playerID simulation.PlayerID,
	currentSessionToken string,
	responses chan<- outboundResponse,
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
		case protocol.MessageRegisterAccount:
			rejectionCode = "invalid_message"
			err = errors.New("account registration must use the account HTTP API")
		case protocol.MessageLogoutAccount:
			rejectionCode = "account_logout_failed"
			if currentSessionToken == "" {
				err = errors.New("this connection is not logged in")
			} else {
				err = h.accounts.Logout(ctx, currentSessionToken)
				if err == nil {
					currentSessionToken = ""
					sendFinalResponse(ctx, responses, protocol.ServerMessage{
						Type: protocol.MessageAccountLoggedOut,
					})
				}
			}
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
		case protocol.MessageExperienceLock:
			rejectionCode = "experience_lock_rejected"
			err = h.world.QueueExperienceLock(simulation.ExperienceLockCommand{
				PlayerID: playerID,
				Sequence: message.Sequence,
				Locked:   message.Active,
			})
		case protocol.MessageDebugSetSkill:
			rejectionCode = "debug_skill_rejected"
			if !h.allowDebug {
				err = fmt.Errorf("debug commands are disabled")
			} else {
				err = h.world.QueueDebugSetSkill(simulation.DebugSetSkillCommand{
					PlayerID:      playerID,
					Sequence:      message.Sequence,
					Skill:         simulation.SkillID(message.SkillID),
					Level:         message.Level,
					Experience:    message.Experience,
					SetExperience: message.SetExperience,
				})
			}
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
		case protocol.MessageLightFire:
			rejectionCode = "firemaking_rejected"
			err = h.world.LightFire(playerID, message.Sequence, message.ItemID)
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
				PlayerID:      playerID,
				Sequence:      message.Sequence,
				ItemID:        message.ItemID,
				InventorySlot: message.InventorySlot,
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
				Quantity:   message.Quantity,
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
		case protocol.MessageBankExpand:
			rejectionCode = "bank_rejected"
			err = h.world.PurchaseBankExpansion(playerID, message.Sequence)
		case protocol.MessageMerchantSell:
			rejectionCode = "merchant_rejected"
			err = h.world.QueueMerchantSale(simulation.MerchantSellCommand{
				PlayerID: playerID, Sequence: message.Sequence, InventorySlot: message.InventorySlot,
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
		case protocol.MessageRingDeposit:
			rejectionCode = "ring_rejected"
			err = h.world.QueueRingTransfer(simulation.RingTransferCommand{
				PlayerID: playerID, Sequence: message.Sequence, Deposit: true,
				ItemID: message.ItemID, InstanceID: message.InstanceID,
			})
		case protocol.MessageRingWithdraw:
			rejectionCode = "ring_rejected"
			err = h.world.QueueRingTransfer(simulation.RingTransferCommand{
				PlayerID: playerID, Sequence: message.Sequence, RackSlot: message.RingSlot,
			})
		case protocol.MessageRuneDeposit:
			rejectionCode = "rune_rejected"
			err = h.world.QueueRuneTransfer(simulation.RuneTransferCommand{
				PlayerID: playerID, Sequence: message.Sequence, Deposit: true, InventorySlot: message.InventorySlot,
			})
		case protocol.MessageRuneWithdraw:
			rejectionCode = "rune_rejected"
			err = h.world.QueueRuneTransfer(simulation.RuneTransferCommand{
				PlayerID: playerID, Sequence: message.Sequence, PouchSlot: message.RuneSlot,
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

func sendResponse(ctx context.Context, responses chan<- outboundResponse, message protocol.ServerMessage) {
	select {
	case responses <- outboundResponse{message: message}:
	case <-ctx.Done():
	}
}

func sendFinalResponse(ctx context.Context, responses chan<- outboundResponse, message protocol.ServerMessage) {
	select {
	case responses <- outboundResponse{message: message, closeAfter: true}:
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
		Fires:          make([]protocol.FireSnapshot, 0, len(snapshot.Fires)),
		CombatEvents:   make([]protocol.CombatEvent, 0, len(snapshot.CombatEvents)),
		GatheringSpots: make([]protocol.GatheringSpotSnapshot, 0, len(snapshot.GatheringSpots)),
		BankChests:     make([]protocol.BankChestSnapshot, 0, len(snapshot.BankChests)),
		Merchants:      make([]protocol.MerchantSnapshot, 0, len(snapshot.Merchants)),
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
			Infinite:      drop.Infinite,
		})
	}
	for _, fire := range snapshot.Fires {
		message.Fires = append(message.Fires, protocol.FireSnapshot{
			ID: fire.ID, LogItem: fire.LogItem,
			Position: protocol.Position{X: fire.Position.X, Y: fire.Position.Y, Z: fire.Position.Z},
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
	for _, merchant := range snapshot.Merchants {
		message.Merchants = append(message.Merchants, protocol.MerchantSnapshot{
			ID: merchant.ID, Name: merchant.Name,
			Position: protocol.Position{X: merchant.Position.X, Y: merchant.Position.Y, Z: merchant.Position.Z},
			Range:    merchant.Range,
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
			CombatLevel:                  player.CombatLevel,
			ExperienceLocked:             player.ExperienceLocked,
			LastExperienceLockSequence:   player.LastExperienceLockSequence,
			CombatMatchup: protocol.CombatMatchupSnapshot{
				TargetID:            player.CombatMatchup.TargetID,
				PlayerHitChance:     player.CombatMatchup.PlayerHitChance,
				PlayerMinimumDamage: player.CombatMatchup.PlayerMinimumDamage,
				PlayerMaximumDamage: player.CombatMatchup.PlayerMaximumDamage,
				PlayerAttackRange:   player.CombatMatchup.PlayerAttackRange,
				TargetHitChance:     player.CombatMatchup.TargetHitChance,
				TargetMinimumDamage: player.CombatMatchup.TargetMinimumDamage,
				TargetMaximumDamage: player.CombatMatchup.TargetMaximumDamage,
				TargetAttackRange:   player.CombatMatchup.TargetAttackRange,
			},
			CombatStats: protocol.CombatStatsSnapshot{
				Weapon:                  string(player.CombatStats.Weapon),
				Style:                   string(player.CombatStats.Style),
				Spell:                   string(player.CombatStats.Spell),
				AccuracySkill:           string(player.CombatStats.AccuracySkill),
				AccuracyLevel:           player.CombatStats.AccuracyLevel,
				WeaponAccuracyBonus:     player.CombatStats.WeaponAccuracyBonus,
				EquipmentAccuracyBonus:  player.CombatStats.EquipmentAccuracyBonus,
				StyleAccuracyBonus:      player.CombatStats.StyleAccuracyBonus,
				SpellAccuracyBonus:      player.CombatStats.SpellAccuracyBonus,
				PrayerHitChanceBonus:    player.CombatStats.PrayerHitChanceBonus,
				TotalAccuracy:           player.CombatStats.TotalAccuracy,
				DamageSkill:             string(player.CombatStats.DamageSkill),
				DamageLevel:             player.CombatStats.DamageLevel,
				StrengthBonus:           player.CombatStats.StrengthBonus,
				EquipmentStrengthBonus:  player.CombatStats.EquipmentStrengthBonus,
				AmmunitionStrengthBonus: player.CombatStats.AmmunitionStrengthBonus,
				SpellPower:              player.CombatStats.SpellPower,
				SkillDamageBonus:        player.CombatStats.SkillDamageBonus,
				PrayerDamageBonus:       player.CombatStats.PrayerDamageBonus,
				PrayerDamageReduction:   player.CombatStats.PrayerDamageReduction,
				BaseAttackInterval:      player.CombatStats.BaseAttackInterval,
				AttackInterval:          player.CombatStats.AttackInterval,
				AttackIntervalModifier:  player.CombatStats.AttackIntervalModifier,
				MinimumDamage:           player.CombatStats.MinimumDamage,
				MaximumDamage:           player.CombatStats.MaximumDamage,
				DefenceLevel:            player.CombatStats.DefenceLevel,
				ArmourDamageReduction:   player.CombatStats.ArmourDamageReduction,
				MaximumHealthBonus:      player.CombatStats.MaximumHealthBonus,
				DefenceBonuses: protocol.DefenceBonuses{
					Stab: player.CombatStats.DefenceBonuses.Stab, Slash: player.CombatStats.DefenceBonuses.Slash,
					Crush: player.CombatStats.DefenceBonuses.Crush, Ranged: player.CombatStats.DefenceBonuses.Ranged,
					Magic: player.CombatStats.DefenceBonuses.Magic,
				},
				AgilityLevel:          player.CombatStats.AgilityLevel,
				EquipmentEvasionBonus: player.CombatStats.EquipmentEvasionBonus,
				TotalEvasion:          player.CombatStats.TotalEvasion,
			},
			Inventory: protocol.InventorySnapshot{
				Capacity: player.Inventory.Capacity,
				Items:    make([]protocol.InventoryItem, 0, len(player.Inventory.Items)),
			},
			LastBankSequence:     player.LastBankSequence,
			LastMerchantSequence: player.LastMerchantSequence,
			LastPickupSequence:   player.LastPickupSequence,
			LastConsumeSequence:  player.LastConsumeSequence,
			NextConsumePulse:     player.NextConsumePulse,
			LastBrewSequence:     player.LastBrewSequence,
			BrewRemaining:        player.BrewRemaining,
			NextBrewPulse:        player.NextBrewPulse,
			BrewingResult:        player.BrewingResult,
			LastFletchSequence:   player.LastFletchSequence,
			FletchRemaining:      player.FletchRemaining,
			NextFletchPulse:      player.NextFletchPulse,
			FletchingResult:      player.FletchingResult,
			GatheringSpot:        player.GatheringSpot,
			GatheringResult:      player.GatheringResult,
			NextGatherPulse:      player.NextGatherPulse,
			LastGatherSequence:   player.LastGatherSequence,
			Skills:               make([]protocol.SkillSnapshot, 0, len(player.Skills)),
			ActiveCombatBoosts:   make([]protocol.ActiveCombatBoost, 0, len(player.ActiveCombatBoosts)),
			Path:                 make([]protocol.Position, 0, len(player.Path)),
		}
		for _, prayerID := range player.ActivePrayers {
			playerMessage.ActivePrayers = append(playerMessage.ActivePrayers, string(prayerID))
		}
		for _, boost := range player.ActiveCombatBoosts {
			playerMessage.ActiveCombatBoosts = append(playerMessage.ActiveCombatBoosts, protocol.ActiveCombatBoost{
				Skill: string(boost.Skill), Levels: boost.Levels, ExpiresAtPulse: boost.ExpiresAtPulse,
			})
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
				SellPrice: simulation.MerchantSellPrice(item),
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
			playerMessage.RingRack = protocol.RingRackSnapshot{
				Capacity: player.RingRack.Capacity,
				Items:    make([]protocol.InventoryItem, 0, len(player.RingRack.Items)),
			}
			playerMessage.LastRingSequence = player.LastRingSequence
			for _, item := range player.RingRack.Items {
				playerMessage.RingRack.Items = append(playerMessage.RingRack.Items, protocol.InventoryItem{
					Slot: item.Slot, ItemID: item.ItemID, Quantity: item.Quantity,
					InstanceID: item.InstanceID, Reinforcement: item.Reinforcement,
				})
			}
			playerMessage.RunePouch = protocol.RunePouchSnapshot{
				Capacity: player.RunePouch.Capacity, UnlockedSlots: player.RunePouch.UnlockedSlots,
				Items: make([]protocol.InventoryItem, 0, len(player.RunePouch.Items)),
			}
			playerMessage.LastRuneSequence = player.LastRuneSequence
			for _, item := range player.RunePouch.Items {
				playerMessage.RunePouch.Items = append(playerMessage.RunePouch.Items, protocol.InventoryItem{
					Slot: item.Slot, ItemID: item.ItemID, Quantity: item.Quantity,
				})
			}
			playerMessage.Bank = &protocol.InventorySnapshot{
				Capacity: player.Bank.Capacity, NextExpansionCost: player.Bank.NextExpansionCost,
				Items: make([]protocol.InventoryItem, 0, len(player.Bank.Items)),
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
		CombatLevel:     target.CombatLevel,
	}
}
