package simulation

const BaseInventoryCapacity = 8

// InventoryItem is one occupied slot in a player's authoritative inventory.
type InventoryItem struct {
	Slot     int
	ItemID   string
	Quantity int
}

// InventorySnapshot is a detached view of a player's inventory.
type InventorySnapshot struct {
	Capacity int
	Items    []InventoryItem
}

type inventoryState struct {
	capacity int
	items    []InventoryItem
}

func newStarterInventory() inventoryState {
	return inventoryState{
		capacity: BaseInventoryCapacity,
		items: []InventoryItem{
			{Slot: 0, ItemID: string(WeaponSword), Quantity: 1},
			{Slot: 1, ItemID: string(WeaponDagger), Quantity: 1},
			{Slot: 2, ItemID: string(WeaponBow), Quantity: 1},
		},
	}
}

func (inventory inventoryState) contains(itemID string) bool {
	for _, item := range inventory.items {
		if item.ItemID == itemID && item.Quantity > 0 {
			return true
		}
	}
	return false
}

func snapshotInventory(inventory inventoryState) InventorySnapshot {
	return InventorySnapshot{
		Capacity: inventory.capacity,
		Items:    append([]InventoryItem(nil), inventory.items...),
	}
}
