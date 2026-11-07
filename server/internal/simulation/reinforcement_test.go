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
	base := derivedCombatStats(player)
	player.inventory.items[0].Reinforcement = 1
	reinforced := derivedCombatStats(player)
	if reinforced.WeaponAccuracyBonus != base.WeaponAccuracyBonus+1 || reinforced.StrengthBonus != base.StrengthBonus+1 || reinforced.MaximumDamage <= base.MaximumDamage {
		t.Fatalf("base stats = %+v, reinforced stats = %+v", base, reinforced)
	}
}

func TestReinforcedEquipmentCannotBeReinforcedTwice(t *testing.T) {
	player := &playerState{inventory: newEmptyInventory(), skills: newDefaultSkills()}
	player.inventory.add(ItemBronzeHelm, 1)
	player.inventory.add(ItemBronzeBar, 4)
	player.inventory.items[0].Reinforcement = 1
	player.skills[SkillSmithing] = SkillExperienceForLevel(2)
	recipe, _ := reinforcementRecipe("reinforce_bronze_helm")
	if err := applyReinforcement(player, recipe, player.inventory.items[0].InstanceID); err == nil {
		t.Fatal("applyReinforcement() error = nil for already reinforced equipment")
	}
	if player.inventory.quantity(ItemBronzeBar) != 4 {
		t.Fatal("failed reinforcement consumed bars")
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
