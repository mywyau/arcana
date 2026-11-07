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

func TestMetalArmourUsesHeadAndBodySlots(t *testing.T) {
	tests := []struct {
		itemID string
		slot   ArmourSlot
	}{
		{ItemBronzeHelm, ArmourSlotHead},
		{ItemBronzePlate, ArmourSlotBody},
		{ItemIronHelm, ArmourSlotHead},
		{ItemIronPlate, ArmourSlotBody},
		{ItemMithrilHelm, ArmourSlotHead},
		{ItemMithrilPlate, ArmourSlotBody},
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
