package simulation

import "testing"

func TestSwiftSetRequiresAllPiecesForAttackSpeedBonus(t *testing.T) {
	player := &playerState{equippedArmour: map[ArmourSlot]string{
		ArmourSlotHead: ItemSwiftHelm,
		ArmourSlotBody: ItemSwiftBody,
	}}
	sword, _ := weaponDefinition(WeaponSword)
	if got := effectiveWeaponAttackInterval(player, sword); got != swordAttackIntervalPulses {
		t.Fatalf("partial Swift set attack interval = %d, want %d", got, swordAttackIntervalPulses)
	}

	player.equippedArmour[ArmourSlotFeet] = ItemSwiftBoots
	if got, want := effectiveWeaponAttackInterval(player, sword), uint64(swordAttackIntervalPulses-1); got != want {
		t.Fatalf("full Swift set attack interval = %d, want %d", got, want)
	}
	dagger, _ := weaponDefinition(WeaponDagger)
	if got := effectiveWeaponAttackInterval(player, dagger); got != minimumAttackIntervalPulses {
		t.Fatalf("Swift dagger attack interval = %d, want minimum %d", got, minimumAttackIntervalPulses)
	}
}

func TestSwiftSetDefinitionsUseDistinctArmourSlots(t *testing.T) {
	tests := []struct {
		itemID string
		slot   ArmourSlot
	}{
		{itemID: ItemSwiftHelm, slot: ArmourSlotHead},
		{itemID: ItemSwiftBody, slot: ArmourSlotBody},
		{itemID: ItemSwiftBoots, slot: ArmourSlotFeet},
	}
	for _, test := range tests {
		if got, exists := armourSlot(test.itemID); !exists || got != test.slot {
			t.Errorf("armourSlot(%q) = %q, %t; want %q, true", test.itemID, got, exists, test.slot)
		}
	}
}

func TestMetalArmourProgressionCoversEveryImplementedSlot(t *testing.T) {
	tiers := [][]string{
		{ItemBronzeHelm, ItemBronzePlate, ItemBronzeShield, ItemBronzeLegs, ItemBronzeGloves, ItemBronzeBoots},
		{ItemIronHelm, ItemIronPlate, ItemIronShield, ItemIronLegs, ItemIronGloves, ItemIronBoots},
		{ItemSteelHelm, ItemSteelPlate, ItemSteelShield, ItemSteelLegs, ItemSteelGloves, ItemSteelBoots},
		{ItemMithrilHelm, ItemMithrilPlate, ItemMithrilShield, ItemMithrilLegs, ItemMithrilGloves, ItemMithrilBoots},
		{ItemAdamantiteHelm, ItemAdamantitePlate, ItemAdamantiteShield, ItemAdamantiteLegs, ItemAdamantiteGloves, ItemAdamantiteBoots},
	}
	wantSlots := []ArmourSlot{ArmourSlotHead, ArmourSlotBody, ArmourSlotLeftHand, ArmourSlotLegs, ArmourSlotHands, ArmourSlotFeet}
	for _, tier := range tiers {
		for index, itemID := range tier {
			definition, exists := armourDefinition(itemID)
			if !exists || definition.Slot != wantSlots[index] {
				t.Errorf("armourDefinition(%q) = %+v, %t; want slot %q", itemID, definition, exists, wantSlots[index])
			}
		}
	}
}

func TestNoviceCombatSetsProvideFocusedOffence(t *testing.T) {
	magic := equippedOffenceBonuses(map[ArmourSlot]string{
		ArmourSlotHead: ItemNoviceRobeHood, ArmourSlotBody: ItemNoviceRobeTop,
		ArmourSlotHands: ItemNoviceRobeGloves, ArmourSlotLegs: ItemNoviceRobeBottoms, ArmourSlotFeet: ItemNoviceRobeBoots,
	})
	if magic.MagicAccuracy != 9 || magic.MagicStrength != 5 || magic.RangedAccuracy != 0 {
		t.Fatalf("novice magic set offence = %+v", magic)
	}
	ranged := equippedOffenceBonuses(map[ArmourSlot]string{
		ArmourSlotHead: ItemNoviceRangerHood, ArmourSlotBody: ItemNoviceRangerBody,
		ArmourSlotHands: ItemNoviceRangerGloves, ArmourSlotLegs: ItemNoviceRangerLegs, ArmourSlotFeet: ItemNoviceRangerBoots,
	})
	if ranged.RangedAccuracy != 9 || ranged.RangedStrength != 5 || ranged.MagicAccuracy != 0 {
		t.Fatalf("novice ranged set offence = %+v", ranged)
	}
}

func TestRangedArmourProgressionRequiresRangeAndImprovesRangedOffence(t *testing.T) {
	tiers := []struct {
		level        int
		items        []string
		wantAccuracy int
		wantStrength int
	}{
		{level: 5, items: []string{ItemScoutHood, ItemScoutBody, ItemScoutGloves, ItemScoutLegs, ItemScoutBoots}, wantAccuracy: 12, wantStrength: 7},
		{level: 10, items: []string{ItemTrackerHood, ItemTrackerBody, ItemTrackerGloves, ItemTrackerLegs, ItemTrackerBoots}, wantAccuracy: 15, wantStrength: 9},
		{level: 15, items: []string{ItemHunterHood, ItemHunterBody, ItemHunterGloves, ItemHunterLegs, ItemHunterBoots}, wantAccuracy: 18, wantStrength: 11},
		{level: 20, items: []string{ItemMarksmanHood, ItemMarksmanBody, ItemMarksmanGloves, ItemMarksmanLegs, ItemMarksmanBoots}, wantAccuracy: 22, wantStrength: 14},
	}
	bow, _ := weaponDefinition(WeaponBow)
	sword, _ := weaponDefinition(WeaponSword)
	for _, tier := range tiers {
		equipped := make(map[ArmourSlot]string, len(tier.items))
		for _, itemID := range tier.items {
			item, exists := itemDefinition(itemID)
			if !exists || len(item.EquipRequirements) != 1 ||
				item.EquipRequirements[0] != (SkillRequirement{Skill: SkillRange, Level: tier.level}) {
				t.Fatalf("%s requirements = %+v, want Range %d", itemID, item.EquipRequirements, tier.level)
			}
			equipped[item.Armour.Slot] = itemID
		}
		player := &playerState{equippedArmour: equipped}
		bonuses := equippedOffenceBonuses(equipped)
		if bonuses.RangedAccuracy != tier.wantAccuracy || bonuses.RangedStrength != tier.wantStrength ||
			bonuses.MeleeAccuracy != 0 || bonuses.MagicAccuracy != 0 {
			t.Errorf("level %d ranged bonuses = %+v", tier.level, bonuses)
		}
		if got := effectiveWeaponAttackInterval(player, bow); got != bow.AttackInterval {
			t.Errorf("level %d set bow interval = %d, want unchanged %d", tier.level, got, bow.AttackInterval)
		}
		if got := effectiveWeaponAttackInterval(player, sword); got != sword.AttackInterval {
			t.Errorf("level %d set sword interval = %d, want %d", tier.level, got, sword.AttackInterval)
		}
	}
}

func TestMetalArmourProgressesMeleeAccuracyAndRangedDefence(t *testing.T) {
	tiers := []struct {
		items             []string
		wantMeleeAccuracy int
		wantRangedDefence int
	}{
		{[]string{ItemBronzeHelm, ItemBronzePlate, ItemBronzeShield, ItemBronzeLegs, ItemBronzeGloves, ItemBronzeBoots}, 2, 18},
		{[]string{ItemIronHelm, ItemIronPlate, ItemIronShield, ItemIronLegs, ItemIronGloves, ItemIronBoots}, 4, 33},
		{[]string{ItemSteelHelm, ItemSteelPlate, ItemSteelShield, ItemSteelLegs, ItemSteelGloves, ItemSteelBoots}, 6, 45},
		{[]string{ItemMithrilHelm, ItemMithrilPlate, ItemMithrilShield, ItemMithrilLegs, ItemMithrilGloves, ItemMithrilBoots}, 8, 58},
		{[]string{ItemAdamantiteHelm, ItemAdamantitePlate, ItemAdamantiteShield, ItemAdamantiteLegs, ItemAdamantiteGloves, ItemAdamantiteBoots}, 10, 75},
	}
	previousAccuracy := 0
	previousDefence := 0
	for _, tier := range tiers {
		equipped := make(map[ArmourSlot]string, len(tier.items))
		for _, itemID := range tier.items {
			definition, _ := armourDefinition(itemID)
			equipped[definition.Slot] = itemID
		}
		offence := equippedOffenceBonuses(equipped)
		defence := equippedDefenceBonuses(equipped)
		if offence.MeleeAccuracy != tier.wantMeleeAccuracy || defence.Ranged != tier.wantRangedDefence {
			t.Errorf("metal tier offence=%+v defence=%+v", offence, defence)
		}
		if offence.MeleeAccuracy <= previousAccuracy || defence.Ranged <= previousDefence {
			t.Errorf("metal tier did not progress: offence=%+v defence=%+v", offence, defence)
		}
		previousAccuracy = offence.MeleeAccuracy
		previousDefence = defence.Ranged
	}
}

func TestCombatCapesUseCapeSlotAndFocusedBonuses(t *testing.T) {
	tests := []struct {
		itemID  string
		offence OffenceBonuses
	}{
		{ItemVanguardCape, OffenceBonuses{MeleeAccuracy: 3, MeleeStrength: 2}},
		{ItemWayfarerCape, OffenceBonuses{RangedAccuracy: 3, RangedStrength: 2}},
		{ItemArcanistCape, OffenceBonuses{MagicAccuracy: 3, MagicStrength: 2}},
	}
	for _, test := range tests {
		definition, exists := armourDefinition(test.itemID)
		if !exists || definition.Slot != ArmourSlotCape || definition.Offence != test.offence {
			t.Errorf("armourDefinition(%q) = %+v, %t", test.itemID, definition, exists)
		}
	}
}

func TestMetalArmourUsesHeadAndBodySlots(t *testing.T) {
	tests := []struct {
		itemID string
		slot   ArmourSlot
	}{
		{ItemBronzeHelm, ArmourSlotHead},
		{ItemBronzePlate, ArmourSlotBody},
		{ItemIronHelm, ArmourSlotHead},
		{ItemIronPlate, ArmourSlotBody},
		{ItemSteelHelm, ArmourSlotHead},
		{ItemSteelPlate, ArmourSlotBody},
		{ItemMithrilHelm, ArmourSlotHead},
		{ItemMithrilPlate, ArmourSlotBody},
		{ItemAdamantiteHelm, ArmourSlotHead},
		{ItemAdamantitePlate, ArmourSlotBody},
	}
	for _, test := range tests {
		if got, exists := armourSlot(test.itemID); !exists || got != test.slot {
			t.Errorf("armourSlot(%q) = %q, %t; want %q, true", test.itemID, got, exists, test.slot)
		}
	}
}

func TestSwiftSetChangesAuthoritativeAttackCadence(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{X: -4, Z: -1}, 4)
	playerState := world.players["player-one"]
	playerState.equippedArmour = map[ArmourSlot]string{
		ArmourSlotHead: ItemSwiftHelm,
		ArmourSlotBody: ItemSwiftBody,
		ArmourSlotFeet: ItemSwiftBoots,
	}
	if err := world.QueueTarget(TargetCommand{
		PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID,
	}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}

	world.Step(actionPulseInterval)
	player := mustPlayer(t, world, "player-one")
	if got, want := player.NextAttackPulse, uint64(1+swordAttackIntervalPulses-1); got != want {
		t.Fatalf("NextAttackPulse with Swift set = %d, want %d", got, want)
	}
	if player.CombatStats.BaseAttackInterval != swordAttackIntervalPulses ||
		player.CombatStats.AttackInterval != swordAttackIntervalPulses-1 ||
		player.CombatStats.AttackIntervalModifier != -1 {
		t.Fatalf("Swift set combat stats = %+v", player.CombatStats)
	}

	world.Step(2 * actionPulseInterval)
	if got, want := mustTarget(t, world.Snapshot(), dummyTargetID).Health, dummyMaximumHealth-swordDamage; got != want {
		t.Fatalf("dummy health before faster cadence = %d, want %d", got, want)
	}
	world.Step(actionPulseInterval)
	if got, want := mustTarget(t, world.Snapshot(), dummyTargetID).Health, dummyMaximumHealth-2*swordDamage; got != want {
		t.Fatalf("dummy health after faster cadence = %d, want %d", got, want)
	}
}

func TestArmourReducesMatchingDamageWithDiminishingReturns(t *testing.T) {
	player := &playerState{equippedArmour: map[ArmourSlot]string{
		ArmourSlotHead: ItemBronzeHelm,
		ArmourSlotBody: ItemBronzePlate,
	}}

	// Bronze armour supplies 7 crush defence, so it reduces 15 damage by
	// 7 / (100 + 7), rounded only after the complete calculation.
	if got, want := reduceIncomingDamage(player, 15, DamageTypeCrush), 14; got != want {
		t.Fatalf("incoming crush damage = %d, want %d", got, want)
	}
	if got := armourDamageReduction(player, DamageTypeMagic); got != 0 {
		t.Fatalf("negative magic defence produced %.3f damage reduction, want 0", got)
	}
}

func TestArmourHealthBonusRaisesMaximumAndUnequipClampsHealth(t *testing.T) {
	player := &playerState{
		health:        playerMaximumHealth,
		maximumHealth: playerMaximumHealth,
		equippedArmour: map[ArmourSlot]string{
			ArmourSlotHead: ItemBronzeHelm,
			ArmourSlotBody: ItemBronzePlate,
		},
	}

	refreshPlayerMaximumHealth(player)
	if got, want := player.maximumHealth, 108; got != want {
		t.Fatalf("maximum health with bronze armour = %d, want %d", got, want)
	}
	if got, want := player.health, playerMaximumHealth; got != want {
		t.Fatalf("equipping armour healed to %d, want unchanged %d", got, want)
	}

	player.health = player.maximumHealth
	delete(player.equippedArmour, ArmourSlotBody)
	refreshPlayerMaximumHealth(player)
	if got, want := player.maximumHealth, 102; got != want {
		t.Fatalf("maximum health after removing plate = %d, want %d", got, want)
	}
	if got, want := player.health, 102; got != want {
		t.Fatalf("health after removing plate = %d, want clamped %d", got, want)
	}
}
