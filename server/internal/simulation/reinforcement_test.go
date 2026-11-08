package simulation

import "testing"

func TestEquipmentInstancesRemainDistinctAndReinforcementTargetsOne(t *testing.T) {
	player := &playerState{
		inventory: newEmptyInventory(), skills: newDefaultSkills(),
		equippedArmour: make(map[ArmourSlot]string), equippedArmourInstances: make(map[ArmourSlot]string),
	}
	if !player.inventory.add(string(WeaponBronzeSword), 2) || !player.inventory.add(ItemBronzeBar, 3) {
		t.Fatal("failed to build test inventory")
	}
	first := player.inventory.items[0]
	second := player.inventory.items[1]
	if first.InstanceID == "" || second.InstanceID == "" || first.InstanceID == second.InstanceID || first.Quantity != 1 || second.Quantity != 1 {
		t.Fatalf("equipment instances = %+v", player.inventory.items)
	}
	player.skills[SkillSmithing] = SkillExperienceForLevel(2)
	recipe, _ := reinforcementRecipe("reinforce_bronze_sword")
	if err := applyReinforcement(player, recipe, first.InstanceID); err != nil {
		t.Fatalf("applyReinforcement() returned an unexpected error: %v", err)
	}
	first, _ = player.inventory.itemByInstance(first.InstanceID)
	second, _ = player.inventory.itemByInstance(second.InstanceID)
	if first.Reinforcement != 1 || second.Reinforcement != 0 || player.inventory.quantity(ItemBronzeBar) != 0 {
		t.Fatalf("inventory after reinforcement = %+v", player.inventory.items)
	}
}

func TestReinforcedEquippedBronzeSwordImprovesCombatStats(t *testing.T) {
	player := &playerState{
		equippedWeapon: WeaponBronzeSword, combatStyle: CombatStyleStab,
		inventory: newEmptyInventory(), skills: newDefaultSkills(),
		equippedArmour: make(map[ArmourSlot]string), equippedArmourInstances: make(map[ArmourSlot]string),
	}
	player.inventory.add(string(WeaponBronzeSword), 1)
	player.equippedWeaponInstanceID = player.inventory.items[0].InstanceID
	player.equippedWeaponReinforcement = 0
	player.inventory.removeInstance(player.equippedWeaponInstanceID)
	base := derivedCombatStats(player)
	player.equippedWeaponReinforcement = 1
	reinforced := derivedCombatStats(player)
	if reinforced.WeaponAccuracyBonus != base.WeaponAccuracyBonus+1 || reinforced.StrengthBonus != base.StrengthBonus+1 || reinforced.MaximumDamage <= base.MaximumDamage {
		t.Fatalf("base stats = %+v, reinforced stats = %+v", base, reinforced)
	}
}

func TestEquipmentCanReachPlusFiveButNoHigher(t *testing.T) {
	player := &playerState{inventory: newEmptyInventory(), skills: newDefaultSkills()}
	player.inventory.add(ItemBronzeHelm, 1)
	player.inventory.add(ItemBronzeBar, 12)
	player.inventory.items[0].Reinforcement = 4
	player.skills[SkillSmithing] = SkillExperienceForLevel(2)
	recipe, _ := reinforcementRecipe("reinforce_bronze_helm")
	if err := applyReinforcement(player, recipe, player.inventory.items[0].InstanceID); err != nil {
		t.Fatalf("applyReinforcement(+5) returned an unexpected error: %v", err)
	}
	if player.inventory.items[0].Reinforcement != maximumReinforcementLevel {
		t.Fatalf("reinforcement = %d, want +%d", player.inventory.items[0].Reinforcement, maximumReinforcementLevel)
	}
	remainingBars := player.inventory.quantity(ItemBronzeBar)
	if err := applyReinforcement(player, recipe, player.inventory.items[0].InstanceID); err == nil {
		t.Fatal("applyReinforcement() error = nil above +5")
	}
	if player.inventory.quantity(ItemBronzeBar) != remainingBars {
		t.Fatal("failed reinforcement consumed bars")
	}
}

func TestArrowsCannotBeReinforced(t *testing.T) {
	if _, exists := reinforcementRecipe("reinforce_iron_arrow"); exists {
		t.Fatal("iron arrows unexpectedly have a reinforcement recipe")
	}
	if _, exists := reinforcementRecipe("reinforce_iron_arrow_tips"); exists {
		t.Fatal("iron arrow tips unexpectedly have a reinforcement recipe")
	}
	for _, itemID := range []string{ItemNoviceRobeTop, ItemNoviceRangerBody} {
		if _, exists := reinforcementRecipe("reinforce_" + itemID); exists {
			t.Fatalf("non-metal starter armour %q unexpectedly has a reinforcement recipe", itemID)
		}
	}
}

func TestPlusFiveEquipmentAppliesAllFiveStatPoints(t *testing.T) {
	player := &playerState{
		equippedWeapon: WeaponMithrilSword, combatStyle: CombatStyleStab,
		equippedWeaponInstanceID: "weapon-instance", equippedWeaponReinforcement: 5,
		equippedArmour:               map[ArmourSlot]string{ArmourSlotLeftHand: ItemMithrilShield},
		equippedArmourInstances:      map[ArmourSlot]string{ArmourSlotLeftHand: "shield-instance"},
		equippedArmourReinforcements: map[ArmourSlot]int{ArmourSlotLeftHand: 5},
		skills:                       newDefaultSkills(),
	}
	weapon, _ := equippedWeaponDefinitionForStyle(player)
	baseWeapon, _ := weaponDefinitionForStyle(WeaponMithrilSword, CombatStyleStab)
	if weapon.AccuracyBonus != baseWeapon.AccuracyBonus+5 || weapon.MeleeStrength != baseWeapon.MeleeStrength+5 {
		t.Fatalf("+5 weapon = %+v, base = %+v", weapon, baseWeapon)
	}
	baseShield, _ := armourDefinition(ItemMithrilShield)
	defence := equippedDefenceBonusesForPlayer(player)
	if defence.Stab != baseShield.Defence.Stab+5 || defence.Ranged != baseShield.Defence.Ranged+5 || defence.Magic != baseShield.Defence.Magic {
		t.Fatalf("+5 shield defence = %+v, base = %+v", defence, baseShield.Defence)
	}
}

func TestDroppedReinforcedEquipmentKeepsItsInstanceMetadata(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	player := world.players["player-one"]
	player.inventory.add(string(WeaponBronzeSword), 1)
	player.inventory.items[0].Reinforcement = 1
	instanceID := player.inventory.items[0].InstanceID
	if err := world.QueueDropItem(DropItemCommand{
		PlayerID: "player-one", Sequence: 1, ItemID: string(WeaponBronzeSword), InstanceID: instanceID,
	}); err != nil {
		t.Fatalf("QueueDropItem() returned an unexpected error: %v", err)
	}
	world.applyPendingItemDrops()
	var drop DropSnapshot
	for _, candidate := range world.drops {
		drop = candidate
	}
	if drop.InstanceID != instanceID || drop.Reinforcement != 1 {
		t.Fatalf("drop metadata = %+v", drop)
	}
	if err := world.QueuePickup(PickupCommand{PlayerID: "player-one", Sequence: 1, DropID: drop.ID}); err != nil {
		t.Fatalf("QueuePickup() returned an unexpected error: %v", err)
	}
	world.startPendingPickupApproaches()
	world.applyPendingPickups()
	pickedUp, exists := player.inventory.itemByInstance(instanceID)
	if !exists || pickedUp.Reinforcement != 1 {
		t.Fatalf("picked-up item = %+v, exists %t", pickedUp, exists)
	}
}
