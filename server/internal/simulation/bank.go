package simulation

import (
	"fmt"
	"sort"
)

const (
	StarterBankChestID       = "starter_bank_chest"
	TestMeleeChestID         = "test_melee_chest"
	TestRangedChestID        = "test_ranged_chest"
	TestMagicChestID         = "test_magic_chest"
	TestOresChestID          = "test_ores_chest"
	TestBarsChestID          = "test_bars_chest"
	TestToolsChestID         = "test_tools_chest"
	TestHerbsChestID         = "test_herbs_chest"
	TestCraftingChestID      = "test_crafting_chest"
	TestFoodChestID          = "test_food_chest"
	BankCapacity             = 100
	MaximumBankCapacity      = 1000
	BankExpansionSize        = 50
	BankExpansionBaseCost    = 1000
	MaximumBankStackQuantity = MaximumItemStackQuantity
	bankInteractionRange     = 2.5
)

var starterBankChest = BankChestSnapshot{
	ID:       StarterBankChestID,
	Name:     "Bank Chest",
	Position: Position{X: -18, Z: -21},
	Range:    bankInteractionRange,
}

var testSupplyChests = []BankChestSnapshot{
	{ID: TestMeleeChestID, Name: "Melee Test Chest", Position: Position{X: -6, Z: -35}, Range: bankInteractionRange},
	{ID: TestRangedChestID, Name: "Ranged Test Chest", Position: Position{X: 0, Z: -35}, Range: bankInteractionRange},
	{ID: TestMagicChestID, Name: "Magic Test Chest", Position: Position{X: 6, Z: -35}, Range: bankInteractionRange},
	{ID: TestOresChestID, Name: "Ores Test Chest", Position: Position{X: -6, Z: -41}, Range: bankInteractionRange},
	{ID: TestBarsChestID, Name: "Bars Test Chest", Position: Position{X: 0, Z: -41}, Range: bankInteractionRange},
	{ID: TestToolsChestID, Name: "Tools Test Chest", Position: Position{X: 6, Z: -41}, Range: bankInteractionRange},
	{ID: TestHerbsChestID, Name: "Herbs Test Chest", Position: Position{X: -6, Z: -47}, Range: bankInteractionRange},
	{ID: TestCraftingChestID, Name: "Crafting Test Chest", Position: Position{X: 0, Z: -47}, Range: bankInteractionRange},
	{ID: TestFoodChestID, Name: "Food Test Chest", Position: Position{X: 6, Z: -47}, Range: bankInteractionRange},
}

var (
	TestMeleeChestPosition    = testSupplyChests[0].Position
	TestRangedChestPosition   = testSupplyChests[1].Position
	TestMagicChestPosition    = testSupplyChests[2].Position
	TestOresChestPosition     = testSupplyChests[3].Position
	TestBarsChestPosition     = testSupplyChests[4].Position
	TestToolsChestPosition    = testSupplyChests[5].Position
	TestHerbsChestPosition    = testSupplyChests[6].Position
	TestCraftingChestPosition = testSupplyChests[7].Position
	TestFoodChestPosition     = testSupplyChests[8].Position
)

type BankChestSnapshot struct {
	ID       string
	Name     string
	Position Position
	Range    float64
}

type BankSnapshot struct {
	Capacity          int
	NextExpansionCost int
	Items             []InventoryItem
}

type bankState struct {
	capacity int
	items    []InventoryItem
}

type BankTransferCommand struct {
	PlayerID      PlayerID
	Sequence      uint64
	Deposit       bool
	ItemID        string
	InstanceID    string
	InventorySlot int
	BankSlot      int
	Quantity      int
}

func newEmptyBank() bankState {
	return bankState{capacity: BankCapacity}
}

func snapshotBank(bank bankState) BankSnapshot {
	return BankSnapshot{
		Capacity: bank.capacity, NextExpansionCost: bankExpansionCost(bank.capacity),
		Items: append([]InventoryItem(nil), bank.items...),
	}
}

func bankFromSnapshot(snapshot BankSnapshot) (bankState, error) {
	if snapshot.Capacity == 0 && len(snapshot.Items) == 0 {
		return newEmptyBank(), nil
	}
	if snapshot.Capacity < BankCapacity || snapshot.Capacity > MaximumBankCapacity ||
		(snapshot.Capacity-BankCapacity)%BankExpansionSize != 0 {
		return bankState{}, fmt.Errorf("bank capacity must be between %d and %d in increments of %d", BankCapacity, MaximumBankCapacity, BankExpansionSize)
	}
	items := append([]InventoryItem(nil), snapshot.Items...)
	occupied := make(map[int]struct{}, len(items))
	instances := make(map[string]struct{}, len(items))
	resources := make(map[string]struct{}, len(items))
	for _, item := range items {
		definition, exists := itemDefinition(item.ItemID)
		if item.Slot < 0 || item.Slot >= snapshot.Capacity || !exists || definition.Virtual ||
			item.Quantity <= 0 || item.Quantity > MaximumBankStackQuantity {
			return bankState{}, fmt.Errorf("bank slot %d contains an invalid item", item.Slot)
		}
		if _, duplicate := occupied[item.Slot]; duplicate {
			return bankState{}, fmt.Errorf("bank slot %d is duplicated", item.Slot)
		}
		occupied[item.Slot] = struct{}{}
		if itemIsInstanced(item.ItemID) {
			if item.Quantity != 1 || item.InstanceID == "" || item.Reinforcement < 0 || item.Reinforcement > maximumReinforcementLevel {
				return bankState{}, fmt.Errorf("bank equipment %q requires one instance", item.ItemID)
			}
			if _, duplicate := instances[item.InstanceID]; duplicate {
				return bankState{}, fmt.Errorf("bank instance %q is duplicated", item.InstanceID)
			}
			instances[item.InstanceID] = struct{}{}
		} else if item.InstanceID != "" || item.Reinforcement != 0 {
			return bankState{}, fmt.Errorf("bank resource %q cannot have equipment metadata", item.ItemID)
		} else if _, duplicate := resources[item.ItemID]; duplicate {
			return bankState{}, fmt.Errorf("bank resource %q is split across multiple slots", item.ItemID)
		} else {
			resources[item.ItemID] = struct{}{}
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Slot < items[j].Slot })
	return bankState{capacity: snapshot.Capacity, items: items}, nil
}

func bankExpansionCost(capacity int) int {
	if capacity < BankCapacity || capacity >= MaximumBankCapacity {
		return 0
	}
	upgradeNumber := (capacity-BankCapacity)/BankExpansionSize + 1
	cost := BankExpansionBaseCost
	for tier := 1; tier < upgradeNumber; tier++ {
		// Compound by 1.65 and round upward to a clean hundred. This keeps
		// early storage attainable while putting the final tier near 5m.
		cost = ((cost*165 + 9_999) / 10_000) * 100
	}
	return cost
}

// BankExpansionCost exposes the authoritative next-tier price for persistence
// and presentation boundaries without duplicating the economy curve.
func BankExpansionCost(capacity int) int {
	return bankExpansionCost(capacity)
}

func (w *World) PurchaseBankExpansion(playerID PlayerID, sequence uint64) error {
	if sequence == 0 {
		return fmt.Errorf("bank sequence must be positive")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	player, exists := w.players[playerID]
	if !exists || player.defeated {
		return fmt.Errorf("player %q cannot use the bank", playerID)
	}
	if distance(player.position, starterBankChest.Position) > starterBankChest.Range {
		return fmt.Errorf("player is too far from the bank chest")
	}
	if w.playerIsInCombat(player) {
		return fmt.Errorf("player cannot use the bank during combat")
	}
	if sequence <= player.lastBankSequence {
		return fmt.Errorf("bank sequence %d is not newer than %d", sequence, player.lastBankSequence)
	}
	if _, pending := w.pendingBankTransfers[playerID]; pending {
		return fmt.Errorf("finish the pending bank transfer first")
	}
	cost := bankExpansionCost(player.bank.capacity)
	if cost == 0 {
		return fmt.Errorf("bank capacity is already at the maximum")
	}
	inventory := cloneInventory(player.inventory)
	bank := cloneBank(player.bank)
	remaining := cost
	fromInventory := min(remaining, inventory.quantity(ItemCoins))
	if fromInventory > 0 {
		inventory.remove(ItemCoins, fromInventory)
		remaining -= fromInventory
	}
	if remaining > 0 {
		for _, item := range bank.items {
			if item.ItemID != ItemCoins || item.Quantity < remaining {
				continue
			}
			bank.removeQuantity(item.Slot, remaining)
			remaining = 0
			break
		}
	}
	if remaining > 0 {
		return fmt.Errorf("%d coins are required", cost)
	}
	bank.capacity += BankExpansionSize
	player.inventory = inventory
	player.bank = bank
	player.lastBankSequence = sequence
	return nil
}

func (bank bankState) itemAt(slot int) (InventoryItem, bool) {
	for _, item := range bank.items {
		if item.Slot == slot {
			return item, true
		}
	}
	return InventoryItem{}, false
}

func (bank *bankState) add(item InventoryItem) bool {
	if item.ItemID == "" || item.Quantity <= 0 {
		return false
	}
	if item.Quantity > MaximumBankStackQuantity {
		return false
	}
	if !itemIsInstanced(item.ItemID) {
		for index := range bank.items {
			if bank.items[index].ItemID == item.ItemID && bank.items[index].InstanceID == "" {
				if bank.items[index].Quantity > MaximumBankStackQuantity-item.Quantity {
					return false
				}
				bank.items[index].Quantity += item.Quantity
				return true
			}
		}
	}
	if len(bank.items) >= bank.capacity {
		return false
	}
	occupied := make(map[int]struct{}, len(bank.items))
	for _, existing := range bank.items {
		occupied[existing.Slot] = struct{}{}
	}
	for slot := 0; slot < bank.capacity; slot++ {
		if _, used := occupied[slot]; used {
			continue
		}
		item.Slot = slot
		bank.items = append(bank.items, item)
		sort.Slice(bank.items, func(i, j int) bool { return bank.items[i].Slot < bank.items[j].Slot })
		return true
	}
	return false
}

func (bank *bankState) removeQuantity(slot, quantity int) (InventoryItem, bool) {
	if quantity <= 0 {
		return InventoryItem{}, false
	}
	for index := range bank.items {
		if bank.items[index].Slot != slot {
			continue
		}
		if bank.items[index].Quantity < quantity || itemIsInstanced(bank.items[index].ItemID) && quantity != 1 {
			return InventoryItem{}, false
		}
		removed := bank.items[index]
		removed.Quantity = quantity
		if bank.items[index].Quantity == quantity {
			bank.items = append(bank.items[:index], bank.items[index+1:]...)
		} else {
			bank.items[index].Quantity -= quantity
		}
		return removed, true
	}
	return InventoryItem{}, false
}

func cloneBank(bank bankState) bankState {
	return bankState{capacity: bank.capacity, items: append([]InventoryItem(nil), bank.items...)}
}

func (w *World) QueueBankTransfer(command BankTransferCommand) error {
	if command.Sequence == 0 {
		return fmt.Errorf("bank sequence must be positive")
	}
	if !command.Deposit && command.Quantity == 0 {
		command.Quantity = 1
	}
	if !command.Deposit && command.Quantity < 1 {
		return fmt.Errorf("withdraw quantity must be positive")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	player, exists := w.players[command.PlayerID]
	if !exists || player.defeated {
		return fmt.Errorf("player %q cannot use the bank", command.PlayerID)
	}
	if distance(player.position, starterBankChest.Position) > starterBankChest.Range {
		return fmt.Errorf("player is too far from the bank chest")
	}
	if w.playerIsInCombat(player) {
		return fmt.Errorf("player cannot use the bank during combat")
	}
	if command.Sequence <= player.lastBankSequence {
		return fmt.Errorf("bank sequence %d is not newer than %d", command.Sequence, player.lastBankSequence)
	}
	if pending, exists := w.pendingBankTransfers[command.PlayerID]; exists && command.Sequence <= pending.Sequence {
		return fmt.Errorf("bank sequence %d is not newer than pending sequence %d", command.Sequence, pending.Sequence)
	}
	if command.Deposit {
		if err := validateBankDeposit(player, command); err != nil {
			return err
		}
	} else if item, exists := player.bank.itemAt(command.BankSlot); !exists {
		return fmt.Errorf("bank slot %d is empty", command.BankSlot)
	} else if item.Quantity < command.Quantity || itemIsInstanced(item.ItemID) && command.Quantity != 1 {
		return fmt.Errorf("bank slot %d does not contain %d items", command.BankSlot, command.Quantity)
	}
	if _, _, err := bankTransferAfter(player, command); err != nil {
		return err
	}
	w.pendingBankTransfers[command.PlayerID] = command
	return nil
}

func validateBankDeposit(player *playerState, command BankTransferCommand) error {
	item, exists := inventoryItemAt(player.inventory, command.InventorySlot)
	if command.ItemID == "" || !exists || item.ItemID != command.ItemID {
		return fmt.Errorf("player does not own item %q", command.ItemID)
	}
	if itemIsInstanced(command.ItemID) {
		if item.InstanceID != command.InstanceID {
			return fmt.Errorf("player does not own item instance %q", command.InstanceID)
		}
	}
	if (command.InstanceID != "" && command.InstanceID == player.equippedWeaponInstanceID) ||
		command.ItemID == player.equippedSpellbook {
		return fmt.Errorf("item %q is currently equipped", command.ItemID)
	}
	for slot, instanceID := range player.equippedArmourInstances {
		if command.InstanceID != "" && command.InstanceID == instanceID ||
			command.InstanceID == "" && player.equippedArmour[slot] == command.ItemID {
			return fmt.Errorf("item %q is currently equipped", command.ItemID)
		}
	}
	return nil
}

func (w *World) applyPendingBankTransfers() {
	playerIDs := make([]PlayerID, 0, len(w.pendingBankTransfers))
	for id := range w.pendingBankTransfers {
		playerIDs = append(playerIDs, id)
	}
	sort.Slice(playerIDs, func(i, j int) bool { return playerIDs[i] < playerIDs[j] })
	for _, id := range playerIDs {
		command := w.pendingBankTransfers[id]
		delete(w.pendingBankTransfers, id)
		player, exists := w.players[id]
		if !exists || player.defeated || distance(player.position, starterBankChest.Position) > starterBankChest.Range ||
			w.playerIsInCombat(player) {
			continue
		}
		inventory, bank, err := bankTransferAfter(player, command)
		if err != nil {
			continue
		}
		player.inventory = inventory
		player.bank = bank
		player.lastBankSequence = command.Sequence
	}
}

func (w *World) playerIsInCombat(player *playerState) bool {
	if player == nil || player.combatTarget != "" {
		return player != nil
	}
	for _, target := range w.combat.dummyTargets() {
		if !target.dummy.Defeated && target.dummy.TargetID == player.id {
			return true
		}
	}
	return false
}

func bankTransferAfter(player *playerState, command BankTransferCommand) (inventoryState, bankState, error) {
	inventory := cloneInventory(player.inventory)
	bank := cloneBank(player.bank)
	if command.Deposit {
		item, removed := removeInventorySlot(&inventory, command.InventorySlot)
		if !removed {
			return inventoryState{}, bankState{}, fmt.Errorf("item %q is no longer available", command.ItemID)
		}
		if !bank.add(item) {
			return inventoryState{}, bankState{}, fmt.Errorf("bank is full")
		}
		return inventory, bank, nil
	}
	item, removed := bank.removeQuantity(command.BankSlot, command.Quantity)
	if !removed {
		return inventoryState{}, bankState{}, fmt.Errorf("bank slot %d is empty", command.BankSlot)
	}
	if itemIsInstanced(item.ItemID) {
		if !inventory.addInventoryItem(item) {
			return inventoryState{}, bankState{}, fmt.Errorf("inventory is full")
		}
	} else if !inventory.add(item.ItemID, command.Quantity) {
		return inventoryState{}, bankState{}, fmt.Errorf("inventory is full")
	}
	return inventory, bank, nil
}

func inventoryItemAt(inventory inventoryState, slot int) (InventoryItem, bool) {
	for _, item := range inventory.items {
		if item.Slot == slot {
			return item, true
		}
	}
	return InventoryItem{}, false
}

func removeInventorySlot(inventory *inventoryState, slot int) (InventoryItem, bool) {
	if inventory == nil {
		return InventoryItem{}, false
	}
	for index, item := range inventory.items {
		if item.Slot != slot {
			continue
		}
		inventory.items = append(inventory.items[:index], inventory.items[index+1:]...)
		return item, true
	}
	return InventoryItem{}, false
}
