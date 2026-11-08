package simulation

import "fmt"

type SpellbookID string
type SpellCastMode string
type SpellTargetType string
type SpellEffectType string
type MagicElement string

const (
	MagicElementAir   MagicElement = "air"
	MagicElementWater MagicElement = "water"
	MagicElementEarth MagicElement = "earth"
	MagicElementFire  MagicElement = "fire"
	MagicElementHoly  MagicElement = "holy"
	MagicElementDark  MagicElement = "dark"
)

type ElementalResistances struct {
	Air   int
	Water int
	Earth int
	Fire  int
	Holy  int
	Dark  int
}

const (
	SpellbookStandard SpellbookID = "standard"
	SpellbookDark     SpellbookID = "dark"
	SpellbookHoly     SpellbookID = "holy"
	SpellbookCleric   SpellbookID = "cleric"

	SpellCastAutocast   SpellCastMode = "autocast"
	SpellCastInstant    SpellCastMode = "instant"
	SpellCastNextAttack SpellCastMode = "next_attack"

	SpellTargetEnemy  SpellTargetType = "enemy"
	SpellTargetSelf   SpellTargetType = "self"
	SpellTargetAlly   SpellTargetType = "ally"
	SpellTargetGround SpellTargetType = "ground"

	SpellEffectDirectDamage   SpellEffectType = "direct_damage"
	SpellEffectDamageOverTime SpellEffectType = "damage_over_time"
	SpellEffectHeal           SpellEffectType = "heal"
)

type SpellbookDefinition struct {
	ID   SpellbookID
	Name string

	// EquipmentItemID will identify the equippable item that grants this
	// spellbook. It is empty for a book that is available without an item.
	EquipmentItemID string
}

type SpellDefinition struct {
	ID                 SpellID
	BookID             SpellbookID
	Name               string
	IconID             string
	CastMode           SpellCastMode
	TargetType         SpellTargetType
	EffectType         SpellEffectType
	Element            MagicElement
	RequiredMagicLevel int
	Costs              []ItemCost
	BasePower          int
	AccuracyBonus      int
	Range              float64
	DamageVariance     float64
	AttackInterval     uint64
}

type spellCatalog struct {
	books        map[SpellbookID]SpellbookDefinition
	spells       map[SpellID]SpellDefinition
	spellsByBook map[SpellbookID][]SpellID
}

var configuredSpellCatalog = mustBuildSpellCatalog(
	[]SpellbookDefinition{
		{ID: SpellbookStandard, Name: "Standard", EquipmentItemID: ItemStandardSpellbook},
		{ID: SpellbookDark, Name: "Skin Bible", EquipmentItemID: ItemSkinBible},
		{ID: SpellbookHoly, Name: "Holy Book", EquipmentItemID: ItemHolyBook},
		{ID: SpellbookCleric, Name: "Cleric's Tome", EquipmentItemID: ItemClericsTome},
	},
	[]SpellDefinition{
		{
			ID: SpellWindBolt, BookID: SpellbookStandard, Name: "Wind Bolt", IconID: "wind_bolt",
			CastMode: SpellCastAutocast, TargetType: SpellTargetEnemy, EffectType: SpellEffectDirectDamage,
			Element:            MagicElementAir,
			RequiredMagicLevel: 1,
			Costs:              boltRuneCosts(ItemAirRune),
			BasePower:          4,
			AccuracyBonus:      10,
			Range:              10,
			DamageVariance:     0.20,
			AttackInterval:     windBoltAttackIntervalPulses,
		},
		basicBolt(SpellbookStandard, SpellWaterBolt, "Water Bolt", "water_bolt", MagicElementWater, 2, ItemWaterRune, 5),
		basicBolt(SpellbookStandard, SpellEarthBolt, "Earth Bolt", "earth_bolt", MagicElementEarth, 3, ItemEarthRune, 6),
		basicBolt(SpellbookStandard, SpellFireBolt, "Fire Bolt", "fire_bolt", MagicElementFire, 4, ItemFireRune, 7),
		basicBolt(SpellbookHoly, SpellHolyBolt, "Holy Bolt", "holy_bolt", MagicElementHoly, 5, ItemHolyRune, 8),
		basicBolt(SpellbookDark, SpellDarkBolt, "Dark Bolt", "dark_bolt", MagicElementDark, 6, ItemDarkRune, 8),
	},
)

func basicBolt(bookID SpellbookID, id SpellID, name, iconID string, element MagicElement, level int, runeID string, power int) SpellDefinition {
	return SpellDefinition{
		ID: id, BookID: bookID, Name: name, IconID: iconID,
		CastMode: SpellCastAutocast, TargetType: SpellTargetEnemy, EffectType: SpellEffectDirectDamage,
		Element: element, RequiredMagicLevel: level,
		Costs: boltRuneCosts(runeID), BasePower: power, AccuracyBonus: 10,
		Range: 10, DamageVariance: 0.20, AttackInterval: windBoltAttackIntervalPulses,
	}
}

func boltRuneCosts(elementalRuneID string) []ItemCost {
	return []ItemCost{
		{ItemID: elementalRuneID, Quantity: 1},
		{ItemID: ItemFocusRune, Quantity: 1},
	}
}

func buildSpellCatalog(books []SpellbookDefinition, spells []SpellDefinition) (spellCatalog, error) {
	catalog := spellCatalog{
		books:        make(map[SpellbookID]SpellbookDefinition, len(books)),
		spells:       make(map[SpellID]SpellDefinition, len(spells)),
		spellsByBook: make(map[SpellbookID][]SpellID, len(books)),
	}
	equipmentItems := make(map[string]SpellbookID)
	for index, book := range books {
		if book.ID == "" || book.Name == "" {
			return spellCatalog{}, fmt.Errorf("spellbook definition %d requires an ID and name", index)
		}
		if _, exists := catalog.books[book.ID]; exists {
			return spellCatalog{}, fmt.Errorf("spellbook ID %q is duplicated", book.ID)
		}
		if book.EquipmentItemID != "" {
			if existing, exists := equipmentItems[book.EquipmentItemID]; exists {
				return spellCatalog{}, fmt.Errorf(
					"spellbook item %q grants both %q and %q",
					book.EquipmentItemID,
					existing,
					book.ID,
				)
			}
			equipmentItems[book.EquipmentItemID] = book.ID
		}
		catalog.books[book.ID] = book
		catalog.spellsByBook[book.ID] = nil
	}

	for index, spell := range spells {
		if err := validateSpellDefinition(index, spell, catalog.books); err != nil {
			return spellCatalog{}, err
		}
		if _, exists := catalog.spells[spell.ID]; exists {
			return spellCatalog{}, fmt.Errorf("spell ID %q is duplicated", spell.ID)
		}
		spell.Costs = append([]ItemCost(nil), spell.Costs...)
		catalog.spells[spell.ID] = spell
		catalog.spellsByBook[spell.BookID] = append(catalog.spellsByBook[spell.BookID], spell.ID)
	}
	return catalog, nil
}

func validateSpellDefinition(
	index int,
	spell SpellDefinition,
	books map[SpellbookID]SpellbookDefinition,
) error {
	if spell.ID == "" || spell.Name == "" || spell.IconID == "" {
		return fmt.Errorf("spell definition %d requires an ID, name, and icon ID", index)
	}
	if _, exists := books[spell.BookID]; !exists {
		return fmt.Errorf("spell %q references unknown spellbook %q", spell.ID, spell.BookID)
	}
	if !knownSpellCastMode(spell.CastMode) {
		return fmt.Errorf("spell %q has unknown cast mode %q", spell.ID, spell.CastMode)
	}
	if !knownSpellTargetType(spell.TargetType) {
		return fmt.Errorf("spell %q has unknown target type %q", spell.ID, spell.TargetType)
	}
	if !knownSpellEffectType(spell.EffectType) {
		return fmt.Errorf("spell %q has unknown effect type %q", spell.ID, spell.EffectType)
	}
	if !knownMagicElement(spell.Element) {
		return fmt.Errorf("spell %q has unknown magic element %q", spell.ID, spell.Element)
	}
	if spell.RequiredMagicLevel < 1 {
		return fmt.Errorf("spell %q requires a positive Magic level", spell.ID)
	}
	if spell.Range < 0 || (spell.TargetType != SpellTargetSelf && spell.Range == 0) ||
		spell.DamageVariance < 0 || spell.DamageVariance > 1 {
		return fmt.Errorf("spell %q has invalid range or damage variance", spell.ID)
	}
	if spell.CastMode == SpellCastAutocast && spell.AttackInterval == 0 {
		return fmt.Errorf("spell %q requires an attack interval", spell.ID)
	}
	if spell.BasePower <= 0 {
		return fmt.Errorf("spell %q requires positive power", spell.ID)
	}
	if len(spell.Costs) == 0 {
		return fmt.Errorf("spell %q requires at least one item cost", spell.ID)
	}
	costItems := make(map[string]struct{}, len(spell.Costs))
	for _, cost := range spell.Costs {
		if cost.ItemID == "" || cost.Quantity <= 0 {
			return fmt.Errorf("spell %q has an invalid item cost", spell.ID)
		}
		item, exists := itemDefinition(cost.ItemID)
		if !exists || item.Virtual {
			return fmt.Errorf("spell %q costs unknown item %q", spell.ID, cost.ItemID)
		}
		if _, duplicate := costItems[cost.ItemID]; duplicate {
			return fmt.Errorf("spell %q repeats item cost %q", spell.ID, cost.ItemID)
		}
		costItems[cost.ItemID] = struct{}{}
	}
	return nil
}

func knownMagicElement(element MagicElement) bool {
	switch element {
	case MagicElementAir, MagicElementWater, MagicElementEarth, MagicElementFire, MagicElementHoly, MagicElementDark:
		return true
	default:
		return false
	}
}

func knownSpellCastMode(mode SpellCastMode) bool {
	switch mode {
	case SpellCastAutocast, SpellCastInstant, SpellCastNextAttack:
		return true
	default:
		return false
	}
}

func knownSpellTargetType(targetType SpellTargetType) bool {
	switch targetType {
	case SpellTargetEnemy, SpellTargetSelf, SpellTargetAlly, SpellTargetGround:
		return true
	default:
		return false
	}
}

func knownSpellEffectType(effectType SpellEffectType) bool {
	switch effectType {
	case SpellEffectDirectDamage, SpellEffectDamageOverTime, SpellEffectHeal:
		return true
	default:
		return false
	}
}

func mustBuildSpellCatalog(books []SpellbookDefinition, spells []SpellDefinition) spellCatalog {
	catalog, err := buildSpellCatalog(books, spells)
	if err != nil {
		panic(err)
	}
	return catalog
}

func spellbookDefinition(id SpellbookID) (SpellbookDefinition, bool) {
	definition, exists := configuredSpellCatalog.books[id]
	return definition, exists
}

func spellDefinition(id SpellID) (SpellDefinition, bool) {
	definition, exists := configuredSpellCatalog.spells[id]
	definition.Costs = append([]ItemCost(nil), definition.Costs...)
	return definition, exists
}

func spellDefinitionsForBook(id SpellbookID) []SpellDefinition {
	spellIDs := configuredSpellCatalog.spellsByBook[id]
	definitions := make([]SpellDefinition, 0, len(spellIDs))
	for _, spellID := range spellIDs {
		definition, exists := spellDefinition(spellID)
		if exists {
			definitions = append(definitions, definition)
		}
	}
	return definitions
}
