package simulation

import (
	"fmt"
	"math"
	"time"
)

const (
	MerchantID               = "starter_shopkeeper"
	merchantInteractionRange = 3.0
	merchantStockLifetime    = 2 * time.Minute
)

var starterMerchant = MerchantSnapshot{
	ID: MerchantID, Name: "Travelling Shopkeeper", Position: Position{X: 0, Z: -24},
	Range: merchantInteractionRange,
}

type MerchantSnapshot struct {
	ID       string
	Name     string
	Position Position
	Range    float64
	Stock    []MerchantStockSnapshot
}

type MerchantStockSnapshot struct {
	ItemID        string
	Quantity      int
	InstanceID    string
	Reinforcement int
	BuyPrice      int
}

type MerchantSellCommand struct {
	PlayerID      PlayerID
	Sequence      uint64
	InventorySlot int
}

type MerchantBuyCommand struct {
	PlayerID   PlayerID
	Sequence   uint64
	ItemID     string
	InstanceID string
}

type merchantTransaction struct {
	PlayerID      PlayerID
	Sequence      uint64
	InventorySlot int
	ItemID        string
	InstanceID    string
	Buy           bool
}

type merchantStockEntry struct {
	Item      InventoryItem
	Remaining time.Duration
}

func MerchantSellPrice(item InventoryItem) int {
	definition, exists := itemDefinition(item.ItemID)
	if !exists || definition.Virtual || item.ItemID == ItemCoins || item.Quantity <= 0 {
		return 0
	}
	unitPrice := 1
	switch definition.Kind {
	case ItemKindWeapon:
		unitPrice = 25
	case ItemKindArmour:
		unitPrice = 20
	case ItemKindSpellbook:
		unitPrice = 10
	case ItemKindConsumable:
		unitPrice = 3
	case ItemKindResource:
		unitPrice = 1
	}
	unitPrice += item.Reinforcement * 10
	if item.Quantity > math.MaxInt/unitPrice {
		return 0
	}
	return unitPrice * item.Quantity
}

// MerchantBuyPrice returns the per-item buy-back price. The small markup makes
// reversing an accidental sale possible without creating a free trade loop.
func MerchantBuyPrice(item InventoryItem) int {
	item.Quantity = 1
	basePrice := MerchantSellPrice(item)
	if basePrice <= 0 {
		return 0
	}
	return basePrice + max(1, (basePrice+4)/5)
}

func (w *World) QueueMerchantSale(command MerchantSellCommand) error {
	if command.Sequence == 0 {
		return fmt.Errorf("merchant sequence must be positive")
	}
	if command.InventorySlot < 0 {
		return fmt.Errorf("inventory slot cannot be negative")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	player, exists := w.players[command.PlayerID]
	if !exists || player.defeated {
		return fmt.Errorf("player %q cannot use the shop", command.PlayerID)
	}
	if distance(player.position, starterMerchant.Position) > starterMerchant.Range {
		return fmt.Errorf("player is too far from the shopkeeper")
	}
	if w.playerIsInCombat(player) {
		return fmt.Errorf("player cannot use the shop during combat")
	}
	if command.Sequence <= player.lastMerchantSequence {
		return fmt.Errorf("merchant sequence %d is not newer than %d", command.Sequence, player.lastMerchantSequence)
	}
	if pending, exists := w.pendingMerchantTransactions[command.PlayerID]; exists && command.Sequence <= pending.Sequence {
		return fmt.Errorf("merchant sequence %d is not newer than pending sequence %d", command.Sequence, pending.Sequence)
	}
	item, exists := inventoryItemAt(player.inventory, command.InventorySlot)
	if !exists || MerchantSellPrice(item) <= 0 {
		return fmt.Errorf("that item cannot be sold")
	}
	w.pendingMerchantTransactions[command.PlayerID] = merchantTransaction{
		PlayerID: command.PlayerID, Sequence: command.Sequence, InventorySlot: command.InventorySlot,
	}
	return nil
}

func (w *World) QueueMerchantPurchase(command MerchantBuyCommand) error {
	if command.Sequence == 0 {
		return fmt.Errorf("merchant sequence must be positive")
	}
	if command.ItemID == "" {
		return fmt.Errorf("merchant item ID is required")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	player, exists := w.players[command.PlayerID]
	if !exists || player.defeated {
		return fmt.Errorf("player %q cannot use the shop", command.PlayerID)
	}
	if distance(player.position, starterMerchant.Position) > starterMerchant.Range {
		return fmt.Errorf("player is too far from the shopkeeper")
	}
	if w.playerIsInCombat(player) {
		return fmt.Errorf("player cannot use the shop during combat")
	}
	if command.Sequence <= player.lastMerchantSequence {
		return fmt.Errorf("merchant sequence %d is not newer than %d", command.Sequence, player.lastMerchantSequence)
	}
	if pending, found := w.pendingMerchantTransactions[command.PlayerID]; found && command.Sequence <= pending.Sequence {
		return fmt.Errorf("merchant sequence %d is not newer than pending sequence %d", command.Sequence, pending.Sequence)
	}
	stockIndex := w.merchantStockIndex(command.ItemID, command.InstanceID)
	if stockIndex < 0 || MerchantBuyPrice(w.merchantStock[stockIndex].Item) <= 0 {
		return fmt.Errorf("that item is not available to buy")
	}
	stock := w.merchantStock[stockIndex]
	price := MerchantBuyPrice(stock.Item)
	candidate := cloneInventory(player.inventory)
	if !candidate.remove(ItemCoins, price) {
		return fmt.Errorf("you need %d coins to buy that item", price)
	}
	purchased := stock.Item
	purchased.Quantity = 1
	if purchased.InstanceID != "" {
		if !candidate.addInventoryItem(purchased) {
			return fmt.Errorf("your inventory is full")
		}
	} else if !candidate.add(purchased.ItemID, 1) {
		return fmt.Errorf("your inventory is full")
	}
	w.pendingMerchantTransactions[command.PlayerID] = merchantTransaction{
		PlayerID: command.PlayerID, Sequence: command.Sequence,
		ItemID: command.ItemID, InstanceID: command.InstanceID, Buy: true,
	}
	return nil
}

func (w *World) applyPendingMerchantTransactions() {
	for playerID, command := range w.pendingMerchantTransactions {
		delete(w.pendingMerchantTransactions, playerID)
		player, exists := w.players[playerID]
		if !exists || player.defeated || distance(player.position, starterMerchant.Position) > starterMerchant.Range ||
			w.playerIsInCombat(player) {
			continue
		}
		if command.Buy {
			w.applyMerchantPurchase(player, command)
			continue
		}
		item, exists := inventoryItemAt(player.inventory, command.InventorySlot)
		price := MerchantSellPrice(item)
		if !exists || price <= 0 {
			continue
		}
		candidate := cloneInventory(player.inventory)
		sold, removed := removeInventorySlot(&candidate, command.InventorySlot)
		if !removed || !candidate.add(ItemCoins, price) {
			continue
		}
		player.inventory = candidate
		player.lastMerchantSequence = command.Sequence
		if !itemIsInstanced(sold.ItemID) {
			merged := false
			for index := range w.merchantStock {
				if w.merchantStock[index].Item.ItemID == sold.ItemID && w.merchantStock[index].Item.InstanceID == "" {
					w.merchantStock[index].Item.Quantity += sold.Quantity
					merged = true
					break
				}
			}
			if merged {
				continue
			}
		}
		w.merchantStock = append(w.merchantStock, merchantStockEntry{Item: sold, Remaining: merchantStockLifetime})
	}
}

func (w *World) applyMerchantPurchase(player *playerState, command merchantTransaction) {
	stockIndex := w.merchantStockIndex(command.ItemID, command.InstanceID)
	if stockIndex < 0 {
		return
	}
	stock := w.merchantStock[stockIndex]
	price := MerchantBuyPrice(stock.Item)
	if price <= 0 {
		return
	}
	candidate := cloneInventory(player.inventory)
	if !candidate.remove(ItemCoins, price) {
		return
	}
	purchased := stock.Item
	purchased.Quantity = 1
	added := false
	if purchased.InstanceID != "" {
		added = candidate.addInventoryItem(purchased)
	} else {
		added = candidate.add(purchased.ItemID, 1)
	}
	if !added {
		return
	}
	player.inventory = candidate
	player.lastMerchantSequence = command.Sequence
	if stock.Item.Quantity <= 1 {
		w.merchantStock = append(w.merchantStock[:stockIndex], w.merchantStock[stockIndex+1:]...)
		return
	}
	w.merchantStock[stockIndex].Item.Quantity--
}

func (w *World) merchantStockIndex(itemID, instanceID string) int {
	for index, stock := range w.merchantStock {
		if stock.Item.ItemID == itemID && stock.Item.InstanceID == instanceID {
			return index
		}
	}
	return -1
}

func (w *World) merchantSnapshot() MerchantSnapshot {
	snapshot := starterMerchant
	snapshot.Stock = make([]MerchantStockSnapshot, 0, len(w.merchantStock))
	for _, stock := range w.merchantStock {
		snapshot.Stock = append(snapshot.Stock, MerchantStockSnapshot{
			ItemID: stock.Item.ItemID, Quantity: stock.Item.Quantity,
			InstanceID: stock.Item.InstanceID, Reinforcement: stock.Item.Reinforcement,
			BuyPrice: MerchantBuyPrice(stock.Item),
		})
	}
	return snapshot
}

func (w *World) advanceMerchantStock(elapsed time.Duration) {
	remaining := w.merchantStock[:0]
	for _, stock := range w.merchantStock {
		stock.Remaining -= elapsed
		destroyed := false
		for stock.Remaining <= 0 {
			if stock.Item.Quantity < 5 {
				destroyed = true
				break
			}
			stock.Item.Quantity -= max(1, stock.Item.Quantity/5)
			stock.Remaining += merchantStockLifetime
		}
		if !destroyed {
			remaining = append(remaining, stock)
		}
	}
	w.merchantStock = remaining
}
