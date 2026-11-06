package simulation

import "testing"

func TestWindBoltDefinition(t *testing.T) {
	spell, exists := spellDefinition(SpellWindBolt)
	if !exists || spell.Name != "Wind Bolt" || spell.RequiredMagicLevel != 1 ||
		len(spell.Costs) != 1 || spell.Costs[0] != (ItemCost{ItemID: ItemAirRune, Quantity: 1}) ||
		spell.BasePower != 4 || spell.Range != 10 || spell.DamageVariance != 0.20 {
		t.Fatalf("Wind Bolt definition = %+v, exists %t", spell, exists)
	}
}

func TestSpellSelectionDoesNotRequireStaff(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
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

func TestSpellSelectionRejectsUnknownAndStaleCommands(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
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
	for _, item := range snapshot.Items {
		if item.ItemID == itemID {
			return item.Quantity
		}
	}
	return 0
}
