package simulation

import (
	"sort"
	"strings"
)

// Each entry is one selectable item, including a single stack inside a pouch.
type equippedItem struct {
	InventoryItem
	key       string
	slot      string
	container *inventoryState
}

func equippedItems(p *playerState) []equippedItem {
	var items []equippedItem
	if p.equippedWeapon != WeaponUnarmed && p.equippedWeapon != "" {
		items = append(items, equippedItem{
			InventoryItem: InventoryItem{
				ItemID: string(p.equippedWeapon), Quantity: 1,
				InstanceID: p.equippedWeaponInstanceID, Reinforcement: p.equippedWeaponReinforcement,
			},
			key: "gear:" + p.equippedWeaponInstanceID, slot: "main_hand",
		})
	}
	for slot, id := range p.equippedArmour {
		instance := p.equippedArmourInstances[slot]
		items = append(items, equippedItem{
			InventoryItem: InventoryItem{ItemID: id, Quantity: 1, InstanceID: instance, Reinforcement: p.equippedArmourReinforcements[slot]},
			key:           "gear:" + instance, slot: string(slot),
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].slot < items[j].slot })
	if p.equippedSpellbook != "" {
		items = append(items, equippedItem{InventoryItem: InventoryItem{ItemID: p.equippedSpellbook, Quantity: 1}, key: "book:" + p.equippedSpellbook, slot: "spellbook"})
	}
	for _, group := range []struct {
		prefix    string
		container *inventoryState
	}{{"gear:", &p.ringRack}, {"ammo:", &p.ammunitionPouch}, {"rune:", &p.runePouch}} {
		for _, item := range group.container.items {
			id := item.ItemID
			if item.InstanceID != "" {
				id = item.InstanceID
			}
			items = append(items, equippedItem{InventoryItem: item, key: group.prefix + id, container: group.container})
		}
	}
	return items
}

func (item equippedItem) remove(p *playerState) {
	if item.container != nil {
		item.container.removeFromSlot(item.Slot, item.ItemID, item.Quantity)
		return
	}
	definition, _ := itemDefinition(item.ItemID)
	clearEquippedItem(p, definition)
}

func hasEquippedKey(p *playerState, key string) bool {
	group, id, _ := strings.Cut(key, ":")
	if id == "" {
		return false
	}
	switch group {
	case "gear":
		if p.equippedWeaponInstanceID == id {
			return true
		}
		for _, instance := range p.equippedArmourInstances {
			if instance == id {
				return true
			}
		}
		_, ok := p.ringRack.itemByInstance(id)
		return ok
	case "book":
		return p.equippedSpellbook == id
	case "ammo":
		return p.ammunitionPouch.contains(id)
	case "rune":
		return p.runePouch.contains(id)
	}
	return false
}
