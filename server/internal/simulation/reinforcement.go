package simulation

import (
	"fmt"
	"strings"
)

const maximumReinforcementLevel = 5

type reinforcementRecipeDefinition struct {
	ID            string
	Name          string
	RequiredLevel int
	TargetItem    string
	Costs         []ItemCost
	RewardXP      int64
	Interval      uint64
}

type reinforcementTier struct {
	Bar           string
	RequiredLevel int
	XPPerBar      int64
}

var bronzeReinforcement = reinforcementTier{Bar: ItemBronzeBar, RequiredLevel: 2, XPPerBar: 10}
var ironReinforcement = reinforcementTier{Bar: ItemIronBar, RequiredLevel: 5, XPPerBar: 18}
var mithrilReinforcement = reinforcementTier{Bar: ItemMithrilBar, RequiredLevel: 8, XPPerBar: 30}

var reinforcementTiers = map[string]reinforcementTier{
	string(WeaponBronzeSword): bronzeReinforcement,
	ItemBronzeHelm:            bronzeReinforcement, ItemBronzePlate: bronzeReinforcement,
	ItemBronzeShield: bronzeReinforcement, ItemBronzeLegs: bronzeReinforcement,
	ItemBronzeGloves: bronzeReinforcement, ItemBronzeBoots: bronzeReinforcement,
	string(WeaponSword): ironReinforcement,
	ItemIronHelm:        ironReinforcement, ItemIronPlate: ironReinforcement,
	ItemIronShield: ironReinforcement, ItemIronLegs: ironReinforcement,
	ItemIronGloves: ironReinforcement, ItemIronBoots: ironReinforcement,
	string(WeaponMithrilSword): mithrilReinforcement,
	ItemMithrilHelm:            mithrilReinforcement, ItemMithrilPlate: mithrilReinforcement,
	ItemMithrilShield: mithrilReinforcement, ItemMithrilLegs: mithrilReinforcement,
	ItemMithrilGloves: mithrilReinforcement, ItemMithrilBoots: mithrilReinforcement,
}

func reinforcementRecipe(id string) (reinforcementRecipeDefinition, bool) {
	const prefix = "reinforce_"
	if !strings.HasPrefix(id, prefix) {
		return reinforcementRecipeDefinition{}, false
	}
	itemID := strings.TrimPrefix(id, prefix)
	tier, exists := reinforcementTiers[itemID]
	definition, defined := itemDefinition(itemID)
	if !exists || !defined || (definition.Kind != ItemKindWeapon && definition.Kind != ItemKindArmour) {
		return reinforcementRecipeDefinition{}, false
	}
	barCost := reinforcementBarCost(definition)
	return reinforcementRecipeDefinition{
		ID: id, Name: "Reinforce " + definition.Name, RequiredLevel: tier.RequiredLevel,
		TargetItem: itemID, Costs: []ItemCost{{ItemID: tier.Bar, Quantity: barCost}},
		RewardXP: tier.XPPerBar * int64(barCost), Interval: uint64(3 + barCost/2),
	}, true
}

func reinforcementBarCost(definition ItemDefinition) int {
	if definition.Kind == ItemKindWeapon {
		return 3
	}
	switch definition.Armour.Slot {
	case ArmourSlotBody:
		return 4
	case ArmourSlotLegs, ArmourSlotLeftHand:
		return 3
	default:
		return 2
	}
}

func validateReinforcement(player *playerState, recipe reinforcementRecipeDefinition, instanceID string) error {
	if player == nil {
		return fmt.Errorf("player is required")
	}
	if SkillLevelForExperience(player.skills[SkillSmithing]) < recipe.RequiredLevel {
		return fmt.Errorf("smithing level %d is required", recipe.RequiredLevel)
	}
	item, exists := player.inventory.itemByInstance(instanceID)
	if !exists || item.ItemID != recipe.TargetItem {
		return fmt.Errorf("target equipment instance is missing")
	}
	definition, exists := itemDefinition(item.ItemID)
	if !exists || (definition.Kind != ItemKindWeapon && definition.Kind != ItemKindArmour) {
		return fmt.Errorf("only weapons and armour can be reinforced")
	}
	if item.Reinforcement >= maximumReinforcementLevel {
		return fmt.Errorf("equipment is already at +%d", maximumReinforcementLevel)
	}
	candidate := cloneInventory(player.inventory)
	if !payItemCosts(&candidate, recipe.Costs) {
		return fmt.Errorf("required reinforcement bars are missing")
	}
	return nil
}

func applyReinforcement(player *playerState, recipe reinforcementRecipeDefinition, instanceID string) error {
	if err := validateReinforcement(player, recipe, instanceID); err != nil {
		return err
	}
	if !payItemCosts(&player.inventory, recipe.Costs) {
		return fmt.Errorf("required reinforcement bars are missing")
	}
	for index := range player.inventory.items {
		if player.inventory.items[index].InstanceID == instanceID {
			player.inventory.items[index].Reinforcement++
			grantSkillExperience(player, SkillSmithing, recipe.RewardXP)
			return nil
		}
	}
	return fmt.Errorf("target equipment instance is missing")
}

func equippedReinforcement(player *playerState, instanceID string) int {
	if player == nil || instanceID == "" {
		return 0
	}
	if instanceID == player.equippedWeaponInstanceID {
		return player.equippedWeaponReinforcement
	}
	for slot, equippedInstanceID := range player.equippedArmourInstances {
		if instanceID == equippedInstanceID {
			return player.equippedArmourReinforcements[slot]
		}
	}
	return 0
}
