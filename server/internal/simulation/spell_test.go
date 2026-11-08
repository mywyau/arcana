package simulation

import "testing"

func TestWindBoltDefinition(t *testing.T) {
	spell, exists := spellDefinition(SpellWindBolt)
	if !exists || spell.Name != "Wind Bolt" || spell.RequiredMagicLevel != 1 ||
		spell.BookID != SpellbookStandard || spell.IconID != "wind_bolt" ||
		spell.CastMode != SpellCastAutocast || spell.TargetType != SpellTargetEnemy ||
		spell.EffectType != SpellEffectDirectDamage || spell.Element != MagicElementAir ||
		len(spell.Costs) != 1 || spell.Costs[0] != (ItemCost{ItemID: ItemAirRune, Quantity: 1}) ||
		spell.BasePower != 4 || spell.Range != 10 || spell.DamageVariance != 0.20 {
		t.Fatalf("Wind Bolt definition = %+v, exists %t", spell, exists)
	}
}

func TestStandardSpellbookContainsWindBolt(t *testing.T) {
	book, exists := spellbookDefinition(SpellbookStandard)
	if !exists || book.Name != "Standard" || book.EquipmentItemID != ItemStandardSpellbook {
		t.Fatalf("Standard spellbook = %+v, exists %t", book, exists)
	}
	spells := spellDefinitionsForBook(SpellbookStandard)
	wantIDs := []SpellID{SpellWindBolt, SpellWaterBolt, SpellEarthBolt, SpellFireBolt}
	if len(spells) != len(wantIDs) {
		t.Fatalf("Standard spells = %+v", spells)
	}
	for index, wantID := range wantIDs {
		if spells[index].ID != wantID {
			t.Fatalf("standard spell %d = %q, want %q", index, spells[index].ID, wantID)
		}
	}

	spells[0].Costs[0].Quantity = 99
	windBolt, _ := spellDefinition(SpellWindBolt)
	if windBolt.Costs[0].Quantity != 1 {
		t.Fatalf("catalogue costs were mutated through a lookup: %+v", windBolt.Costs)
	}
}

func TestSpecialistSpellbooksContainTheirElement(t *testing.T) {
	tests := []struct {
		bookID SpellbookID
		name   string
		itemID string
		spell  SpellID
	}{
		{SpellbookDark, "Skin Bible", ItemSkinBible, SpellDarkBolt},
		{SpellbookHoly, "Holy Book", ItemHolyBook, SpellHolyBolt},
	}
	for _, test := range tests {
		book, exists := spellbookDefinition(test.bookID)
		spells := spellDefinitionsForBook(test.bookID)
		if !exists || book.Name != test.name || book.EquipmentItemID != test.itemID ||
			len(spells) != 1 || spells[0].ID != test.spell || spells[0].BookID != test.bookID {
			t.Errorf("spellbook %q = %+v, spells %+v", test.bookID, book, spells)
		}
	}
	cleric, exists := spellbookDefinition(SpellbookCleric)
	if !exists || cleric.Name != "Cleric's Tome" || cleric.EquipmentItemID != ItemClericsTome ||
		len(spellDefinitionsForBook(SpellbookCleric)) != 0 {
		t.Errorf("Cleric spellbook = %+v", cleric)
	}
}

func TestElementalWeaknessModifiesSpellDamage(t *testing.T) {
	target := &TargetSnapshot{Elements: ElementalResistances{Fire: -15, Water: 20}}
	if got := elementalDamageMultiplier(target, MagicElementFire); got != 1.15 {
		t.Fatalf("fire weakness multiplier = %.2f, want 1.15", got)
	}
	if got := elementalDamageMultiplier(target, MagicElementWater); got != 0.80 {
		t.Fatalf("water resistance multiplier = %.2f, want 0.80", got)
	}
	if got := elementalDamageMultiplier(target, MagicElementHoly); got != 1 {
		t.Fatalf("neutral holy multiplier = %.2f, want 1", got)
	}
}

func TestSpellCatalogRejectsInvalidRelationships(t *testing.T) {
	validBook := SpellbookDefinition{ID: "test", Name: "Test", EquipmentItemID: "test_book"}
	validSpell := SpellDefinition{
		ID: "test_spell", BookID: validBook.ID, Name: "Test Spell", IconID: "test_spell",
		CastMode: SpellCastAutocast, TargetType: SpellTargetEnemy, EffectType: SpellEffectDirectDamage,
		Element:            MagicElementAir,
		RequiredMagicLevel: 1,
		Costs:              []ItemCost{{ItemID: ItemAirRune, Quantity: 1}},
		BasePower:          1,
		Range:              1,
		AttackInterval:     1,
	}
	unknownBookSpell := validSpell
	unknownBookSpell.BookID = "missing"
	duplicateCostSpell := validSpell
	duplicateCostSpell.Costs = append(duplicateCostSpell.Costs, duplicateCostSpell.Costs[0])

	tests := []struct {
		name   string
		books  []SpellbookDefinition
		spells []SpellDefinition
	}{
		{
			name: "one item grants two books",
			books: []SpellbookDefinition{
				validBook,
				{ID: "other", Name: "Other", EquipmentItemID: validBook.EquipmentItemID},
			},
		},
		{
			name:   "spell references an unknown book",
			books:  []SpellbookDefinition{validBook},
			spells: []SpellDefinition{unknownBookSpell},
		},
		{
			name:   "spell repeats a rune cost",
			books:  []SpellbookDefinition{validBook},
			spells: []SpellDefinition{duplicateCostSpell},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := buildSpellCatalog(test.books, test.spells); err == nil {
				t.Fatal("buildSpellCatalog() error = nil")
			}
		})
	}
}

func TestSpellCatalogAllowsInstantAndNextAttackTiming(t *testing.T) {
	book := SpellbookDefinition{ID: "test", Name: "Test"}
	spells := []SpellDefinition{
		{
			ID: "instant_heal", BookID: book.ID, Name: "Instant Heal", IconID: "instant_heal",
			CastMode: SpellCastInstant, TargetType: SpellTargetSelf, EffectType: SpellEffectHeal,
			Element:            MagicElementHoly,
			RequiredMagicLevel: 1, Costs: []ItemCost{{ItemID: ItemAirRune, Quantity: 1}}, BasePower: 1,
		},
		{
			ID: "charged_strike", BookID: book.ID, Name: "Charged Strike", IconID: "charged_strike",
			CastMode: SpellCastNextAttack, TargetType: SpellTargetEnemy, EffectType: SpellEffectDirectDamage,
			Element:            MagicElementDark,
			RequiredMagicLevel: 1, Costs: []ItemCost{{ItemID: ItemAirRune, Quantity: 1}}, BasePower: 1, Range: 1,
		},
	}
	if _, err := buildSpellCatalog([]SpellbookDefinition{book}, spells); err != nil {
		t.Fatalf("buildSpellCatalog() rejected future cast modes: %v", err)
	}
}

func TestSpellSelectionDoesNotRequireStaff(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	mustEquipStandardSpellbook(t, world, "player-one")
	if got := mustPlayer(t, world, "player-one").EquippedWeapon; got != WeaponSword {
		t.Fatalf("starting weapon = %q, want %q", got, WeaponSword)
	}
	if err := world.QueueSpell(SpellCommand{
		PlayerID: "player-one", Sequence: 1, SpellID: SpellWindBolt,
	}); err != nil {
		t.Fatalf("QueueSpell() returned an unexpected error without a staff: %v", err)
	}

	world.Step(actionPulseInterval)
	player := mustPlayer(t, world, "player-one")
	if player.SelectedSpell != SpellWindBolt || player.LastSpellSequence != 1 {
		t.Fatalf("selected spell state = %+v", player)
	}
	for _, item := range player.Inventory.Items {
		if item.ItemID == ItemAirRune && item.Quantity != 50 {
			t.Fatalf("air runes after spell selection = %d, want 50", item.Quantity)
		}
	}
	if err := world.QueueSpell(SpellCommand{PlayerID: "player-one", Sequence: 2}); err != nil {
		t.Fatalf("QueueSpell(clear) returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	player = mustPlayer(t, world, "player-one")
	if player.SelectedSpell != "" || player.LastSpellSequence != 2 {
		t.Fatalf("cleared spell state = %+v", player)
	}
}

func TestWindBoltRequiresEquippedStandardSpellbook(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	world.players["player-one"].inventory.add(ItemAirRune, 10)

	err := world.QueueSpell(SpellCommand{PlayerID: "player-one", Sequence: 1, SpellID: SpellWindBolt})
	if err == nil {
		t.Fatal("QueueSpell() error = nil without an equipped spellbook")
	}
}

func TestStandardSpellbookCanEquipDuringCombatAndControlsAutocast(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	player := world.players["player-one"]
	player.inventory.add(ItemStandardSpellbook, 1)
	player.inventory.add(ItemAirRune, 10)

	if err := world.QueueTarget(TargetCommand{
		PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID,
	}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	if err := world.QueueEquip(EquipCommand{
		PlayerID: "player-one", Sequence: 1, ItemID: WeaponID(ItemStandardSpellbook),
	}); err != nil {
		t.Fatalf("QueueEquip(spellbook) during combat returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)

	snapshot := mustPlayer(t, world, "player-one")
	if snapshot.CombatTarget != dummyTargetID || snapshot.EquippedSpellbook != ItemStandardSpellbook ||
		snapshot.ActiveSpellbook != SpellbookStandard {
		t.Fatalf("spellbook combat state = %+v", snapshot)
	}
	if err := world.QueueSpell(SpellCommand{
		PlayerID: "player-one", Sequence: 1, SpellID: SpellWindBolt,
	}); err != nil {
		t.Fatalf("QueueSpell() returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	if selected := mustPlayer(t, world, "player-one").SelectedSpell; selected != SpellWindBolt {
		t.Fatalf("selected spell = %q, want %q", selected, SpellWindBolt)
	}

	if err := world.QueueUnequip(UnequipCommand{
		PlayerID: "player-one", Sequence: 2, ItemID: ItemStandardSpellbook,
	}); err != nil {
		t.Fatalf("QueueUnequip(spellbook) returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	snapshot = mustPlayer(t, world, "player-one")
	if snapshot.EquippedSpellbook != "" || snapshot.ActiveSpellbook != "" || snapshot.SelectedSpell != "" {
		t.Fatalf("state after unequipping spellbook = %+v", snapshot)
	}
}

func TestSpellSelectionRejectsUnknownAndStaleCommands(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	mustEquipStandardSpellbook(t, world, "player-one")
	if err := world.QueueSpell(SpellCommand{
		PlayerID: "player-one", Sequence: 1, SpellID: "unknown",
	}); err == nil {
		t.Fatal("QueueSpell() error = nil for an unknown spell")
	}
	if err := world.QueueSpell(SpellCommand{
		PlayerID: "player-one", Sequence: 1, SpellID: SpellWindBolt,
	}); err != nil {
		t.Fatalf("QueueSpell() returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	if err := world.QueueSpell(SpellCommand{
		PlayerID: "player-one", Sequence: 1, SpellID: SpellWindBolt,
	}); err == nil {
		t.Fatal("QueueSpell() error = nil for a stale sequence")
	}
}

func TestWindBoltConsumesRuneAndReplacesWeaponAttack(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.5))
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	mustEquipStandardSpellbook(t, world, "player-one")
	if err := world.QueueSpell(SpellCommand{PlayerID: "player-one", Sequence: 1, SpellID: SpellWindBolt}); err != nil {
		t.Fatalf("QueueSpell() returned an unexpected error: %v", err)
	}
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}

	world.Step(actionPulseInterval)
	snapshot := world.Snapshot()
	player := mustPlayer(t, world, "player-one")
	if player.EquippedWeapon != WeaponSword || player.AutoApproaching || player.Destination != nil {
		t.Fatalf("sword wielder did not cast at spell range: %+v", player)
	}
	if got := player.Inventory.itemsQuantity(ItemAirRune); got != 49 {
		t.Fatalf("air rune quantity after cast = %d, want 49", got)
	}
	if got, want := mustTarget(t, snapshot, dummyTargetID).Health, dummyMaximumHealth-4; got != want {
		t.Fatalf("dummy health after Wind Bolt = %d, want %d", got, want)
	}
	if player.NextAttackPulse != 1+windBoltAttackIntervalPulses {
		t.Fatalf("next spell attack pulse = %d, want %d", player.NextAttackPulse, 1+windBoltAttackIntervalPulses)
	}
	if magic := mustSkill(t, player, SkillMagic); magic.Experience != 4 {
		t.Fatalf("Magic XP after Wind Bolt = %d, want 4", magic.Experience)
	}
	event := snapshot.CombatEvents[len(snapshot.CombatEvents)-1]
	if event.Type != CombatEventDamage || event.SpellID != SpellWindBolt || event.Amount != 4 {
		t.Fatalf("Wind Bolt combat event = %+v", event)
	}
}

func TestWindBoltConsumesRuneWhenTargetDodges(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0))
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	mustEquipStandardSpellbook(t, world, "player-one")
	if err := world.QueueSpell(SpellCommand{PlayerID: "player-one", Sequence: 1, SpellID: SpellWindBolt}); err != nil {
		t.Fatalf("QueueSpell() returned an unexpected error: %v", err)
	}
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)

	player := mustPlayer(t, world, "player-one")
	if got := player.Inventory.itemsQuantity(ItemAirRune); got != 49 {
		t.Fatalf("air rune quantity after dodged cast = %d, want 49", got)
	}
	events := world.Snapshot().CombatEvents
	if len(events) != 1 || events[0].Type != CombatEventDodged || events[0].SpellID != SpellWindBolt {
		t.Fatalf("dodged Wind Bolt events = %+v", events)
	}
}

func TestWindBoltPausesWithoutRunes(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.5))
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	mustEquipStandardSpellbook(t, world, "player-one")
	world.players["player-one"].inventory.remove(ItemAirRune, 50)
	if err := world.QueueSpell(SpellCommand{PlayerID: "player-one", Sequence: 1, SpellID: SpellWindBolt}); err != nil {
		t.Fatalf("QueueSpell() returned an unexpected error without runes: %v", err)
	}
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)

	player := mustPlayer(t, world, "player-one")
	if player.LastActionResult != "missing_reagents" || player.NextAttackPulse != 0 {
		t.Fatalf("spell state without runes = %+v", player)
	}
	if got := mustTarget(t, world.Snapshot(), dummyTargetID).Health; got != dummyMaximumHealth {
		t.Fatalf("dummy health without runes = %d, want %d", got, dummyMaximumHealth)
	}
	if len(world.Snapshot().CombatEvents) != 0 {
		t.Fatalf("combat events without runes = %+v", world.Snapshot().CombatEvents)
	}
}

func (snapshot InventorySnapshot) itemsQuantity(itemID string) int {
	total := 0
	for _, item := range snapshot.Items {
		if item.ItemID == itemID {
			total += item.Quantity
		}
	}
	return total
}
