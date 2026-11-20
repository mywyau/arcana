package simulation

import "fmt"

const (
	maxAlchemyBatch     = 20
	manualAlchemyPulses = 2
	autoAlchemyPulses   = 5
)

// AlchemyValue is the coin payout for one ordinary, unreinforced piece.
func AlchemyValue(item InventoryItem) int {
	if item.Reinforcement != 0 {
		return 0
	}
	switch item.ItemID {
	case ItemBronzeArrow, ItemBronzeBolt:
		return 2
	case ItemIronArrow, ItemIronBolt:
		return 3
	case ItemBronzeBoots, ItemBronzeGloves:
		return 20
	case ItemBronzeHelm, ItemBronzeShield, ItemBronzeLegs, ItemIronBoots, ItemIronGloves:
		return 22
	case ItemBronzePlate, ItemIronHelm, ItemIronShield, ItemIronLegs:
		return 23
	case ItemIronPlate:
		return 24
	case string(WeaponBronzeSword):
		return 28
	case string(WeaponSword):
		return 30
	default:
		return 0
	}
}

type alchemyBatch struct {
	slot, remaining     int
	itemID, instance    string
	interval, nextPulse uint64
}

func alchemyReady(p *playerState, pulse uint64) error {
	if p.defeated {
		return fmt.Errorf("you cannot cast Alchemy while defeated")
	}
	if activeSpellbook(p) != alchemySpell.BookID {
		return fmt.Errorf("equip the Standard Spellbook to cast Alchemy")
	}
	if err := statusActionError(p, statusActionCast, pulse); err != nil {
		return err
	}
	if err := statusActionError(p, statusActionProduction, pulse); err != nil {
		return err
	}
	if SkillLevelForExperience(p.skills[SkillMagic]) < alchemySpell.RequiredMagicLevel {
		return fmt.Errorf("Alchemy requires Magic level %d", alchemySpell.RequiredMagicLevel)
	}
	if !p.runePouch.contains(ItemFocusRune) && !p.inventory.contains(ItemFocusRune) {
		return fmt.Errorf("Alchemy needs one Focus Rune per item")
	}
	return nil
}

func alchemyTarget(p *playerState, slot int, instance, itemID string) (InventoryItem, error) {
	item, ok := p.inventory.itemAtSlot(slot)
	if !ok || item.InstanceID != instance || item.ItemID != itemID || AlchemyValue(item) == 0 {
		return InventoryItem{}, fmt.Errorf("that item is unavailable for Alchemy")
	}
	if entry, ok := p.protectedInventory[item.Slot]; ok && protectedEntryMatchesItem(entry, item) {
		return InventoryItem{}, fmt.Errorf("that item is protected")
	}
	return item, nil
}

// QueueAlchemy binds a cast to one slot. Quantity zero cancels the pending cast.
func (w *World) QueueAlchemy(id PlayerID, sequence uint64, slot int, itemID, instance string, quantity int, auto bool) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := w.players[id]
	if p == nil || sequence == 0 || sequence <= p.lastAlchemySequence {
		return fmt.Errorf("invalid Alchemy request")
	}
	if quantity == 0 && itemID == "" && !auto {
		p.alchemy = alchemyBatch{}
		p.lastAlchemySequence = sequence
		p.alchemyResult = ""
		return nil
	}
	if quantity < 1 || quantity > maxAlchemyBatch || (!auto && quantity != 1) {
		return fmt.Errorf("choose one item for manual Alchemy or up to 20 for automatic casting")
	}
	if p.alchemy.remaining != 0 {
		return fmt.Errorf("an Alchemy cast is already in progress")
	}
	if err := alchemyReady(p, w.combat.pulse); err != nil {
		return err
	}
	item, err := alchemyTarget(p, slot, instance, itemID)
	if err != nil {
		return err
	}
	if item.Quantity < quantity {
		return fmt.Errorf("the selected slot contains only %d", item.Quantity)
	}
	// Stackable items have no persistent instance ID. A transient tag follows
	// this stack through inventory copies and vanishes if its slot is refilled.
	for i := range p.inventory.items {
		if p.inventory.items[i].Slot == slot {
			p.inventory.items[i].alchemyTag = sequence
			break
		}
	}
	interval := uint64(manualAlchemyPulses)
	if auto {
		interval = autoAlchemyPulses
	}
	p.alchemy = alchemyBatch{slot: slot, remaining: quantity, instance: instance, itemID: itemID, interval: interval, nextPulse: w.combat.pulse + interval}
	p.lastAlchemySequence = sequence
	p.alchemyResult = ""
	return nil
}

func (w *World) advanceAlchemy() {
	for _, p := range w.players {
		batch := &p.alchemy
		if batch.remaining == 0 {
			continue
		}
		if err := alchemyReady(p, w.combat.pulse); err != nil {
			p.alchemyResult = err.Error()
			p.alchemy = alchemyBatch{}
			continue
		}
		if w.combat.pulse < batch.nextPulse {
			continue
		}
		item, err := alchemyTarget(p, batch.slot, batch.instance, batch.itemID)
		if err == nil && item.alchemyTag != p.lastAlchemySequence {
			err = fmt.Errorf("the selected stack has changed")
		}
		if err == nil {
			err = convertAlchemyItem(p, item)
		}
		if err != nil {
			p.alchemyResult = err.Error()
			p.alchemy = alchemyBatch{}
			continue
		}
		batch.remaining--
		batch.nextPulse = w.combat.pulse + batch.interval
	}
}

func convertAlchemyItem(p *playerState, item InventoryItem) error {
	// Rune payment stages detached copies of both containers. Commit only after
	// the same transaction removes the selected instance and fits the reward.
	staged := playerState{inventory: p.inventory, runePouch: p.runePouch}
	if !paySpellCosts(&staged, alchemySpell.Costs) {
		return fmt.Errorf("Alchemy needs a Focus Rune")
	}
	if !staged.inventory.removeFromSlot(item.Slot, item.ItemID, 1) {
		return fmt.Errorf("the selected item is missing")
	}
	if !staged.inventory.add(ItemCoins, AlchemyValue(item)) {
		return fmt.Errorf("your inventory cannot hold the coins")
	}
	p.inventory, p.runePouch = staged.inventory, staged.runePouch
	return nil
}
