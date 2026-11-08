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
}

type MerchantSellCommand struct {
	PlayerID      PlayerID
	Sequence      uint64
	InventorySlot int
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
	if pending, exists := w.pendingMerchantSales[command.PlayerID]; exists && command.Sequence <= pending.Sequence {
		return fmt.Errorf("merchant sequence %d is not newer than pending sequence %d", command.Sequence, pending.Sequence)
	}
	item, exists := inventoryItemAt(player.inventory, command.InventorySlot)
	if !exists || MerchantSellPrice(item) <= 0 {
		return fmt.Errorf("that item cannot be sold")
	}
	w.pendingMerchantSales[command.PlayerID] = command
	return nil
}

func (w *World) applyPendingMerchantSales() {
	for playerID, command := range w.pendingMerchantSales {
		delete(w.pendingMerchantSales, playerID)
		player, exists := w.players[playerID]
		if !exists || player.defeated || distance(player.position, starterMerchant.Position) > starterMerchant.Range ||
			w.playerIsInCombat(player) {
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
