package simulation

import "fmt"

const (
	starterAirRuneAltarID           = "starter_air_rune_altar"
	waterRuneAltarID                = "water_rune_altar"
	earthRuneAltarID                = "earth_rune_altar"
	fireRuneAltarID                 = "fire_rune_altar"
	focusRuneAltarID                = "focus_rune_altar"
	natureRuneAltarID               = "nature_rune_altar"
	spiritRuneAltarID               = "spirit_rune_altar"
	vitalRuneAltarID                = "vital_rune_altar"
	hexRuneAltarID                  = "hex_rune_altar"
	holyRuneAltarID                 = "holy_rune_altar"
	shadowRuneAltarID               = "shadow_rune_altar"
	aetherRuneAltarID               = "aether_rune_altar"
	wrathRuneAltarID                = "wrath_rune_altar"
	runecraftingInterval            = 3
	runecraftingLevelsPerMultiplier = 10
	maximumRunecraftingMultiplier   = 5
)

type runecraftingAltarDefinition struct {
	ID             string
	Name           string
	RuneItem       string
	RequiredLevel  int
	RewardQuantity int
	RewardXP       int64
	Position       Position
}

var runecraftingAltars = mustBuildRunecraftingAltars([]runecraftingAltarDefinition{
	{ID: starterAirRuneAltarID, Name: "Air Rune Altar", RuneItem: ItemAirRune, RequiredLevel: 1, RewardQuantity: 5, RewardXP: 12, Position: Position{X: 20, Z: -48}},
	{ID: waterRuneAltarID, Name: "Water Rune Altar", RuneItem: ItemWaterRune, RequiredLevel: 2, RewardQuantity: 5, RewardXP: 16, Position: Position{X: 32, Z: -48}},
	{ID: earthRuneAltarID, Name: "Earth Rune Altar", RuneItem: ItemEarthRune, RequiredLevel: 3, RewardQuantity: 5, RewardXP: 20, Position: Position{X: 44, Z: -48}},
	{ID: fireRuneAltarID, Name: "Fire Rune Altar", RuneItem: ItemFireRune, RequiredLevel: 4, RewardQuantity: 5, RewardXP: 24, Position: Position{X: 56, Z: -48}},
	{ID: focusRuneAltarID, Name: "Focus Rune Altar", RuneItem: ItemFocusRune, RequiredLevel: 5, RewardQuantity: 4, RewardXP: 30, Position: Position{X: 68, Z: -48}},
	{ID: natureRuneAltarID, Name: "Nature Rune Altar", RuneItem: ItemNatureRune, RequiredLevel: 6, RewardQuantity: 4, RewardXP: 36, Position: Position{X: 20, Z: -59}},
	{ID: spiritRuneAltarID, Name: "Spirit Rune Altar", RuneItem: ItemSpiritRune, RequiredLevel: 8, RewardQuantity: 4, RewardXP: 46, Position: Position{X: 32, Z: -59}},
	{ID: vitalRuneAltarID, Name: "Vital Rune Altar", RuneItem: ItemVitalRune, RequiredLevel: 10, RewardQuantity: 3, RewardXP: 58, Position: Position{X: 44, Z: -59}},
	{ID: hexRuneAltarID, Name: "Hex Rune Altar", RuneItem: ItemHexRune, RequiredLevel: 12, RewardQuantity: 3, RewardXP: 70, Position: Position{X: 56, Z: -59}},
	{ID: holyRuneAltarID, Name: "Holy Rune Altar", RuneItem: ItemHolyRune, RequiredLevel: 14, RewardQuantity: 3, RewardXP: 82, Position: Position{X: 68, Z: -59}},
	{ID: shadowRuneAltarID, Name: "Shadow Rune Altar", RuneItem: ItemShadowRune, RequiredLevel: 14, RewardQuantity: 3, RewardXP: 82, Position: Position{X: 20, Z: -70}},
	{ID: aetherRuneAltarID, Name: "Aether Rune Altar", RuneItem: ItemAetherRune, RequiredLevel: 16, RewardQuantity: 2, RewardXP: 96, Position: Position{X: 32, Z: -70}},
	{ID: wrathRuneAltarID, Name: "Wrath Rune Altar", RuneItem: ItemWrathRune, RequiredLevel: 18, RewardQuantity: 2, RewardXP: 110, Position: Position{X: 44, Z: -70}},
})

func buildRunecraftingAltars(definitions []runecraftingAltarDefinition) (map[string]runecraftingAltarDefinition, error) {
	altars := make(map[string]runecraftingAltarDefinition, len(definitions))
	runes := make(map[string]struct{}, len(definitions))
	for index, definition := range definitions {
		if definition.ID == "" || definition.Name == "" || definition.RequiredLevel < 1 ||
			definition.RequiredLevel > MaximumSkillLevel || definition.RewardQuantity < 1 || definition.RewardXP < 1 {
			return nil, fmt.Errorf("runecrafting altar %d is incomplete", index)
		}
		if !isRune(definition.RuneItem) {
			return nil, fmt.Errorf("runecrafting altar %q has invalid rune %q", definition.ID, definition.RuneItem)
		}
		if _, exists := altars[definition.ID]; exists {
			return nil, fmt.Errorf("runecrafting altar %q is duplicated", definition.ID)
		}
		if _, exists := runes[definition.RuneItem]; exists {
			return nil, fmt.Errorf("rune %q has more than one altar", definition.RuneItem)
		}
		altars[definition.ID] = definition
		runes[definition.RuneItem] = struct{}{}
	}
	return altars, nil
}

func mustBuildRunecraftingAltars(definitions []runecraftingAltarDefinition) map[string]runecraftingAltarDefinition {
	altars, err := buildRunecraftingAltars(definitions)
	if err != nil {
		panic(err)
	}
	return altars
}

func runecraftingAltar(definition runecraftingAltarDefinition) gatheringSpotDefinition {
	resultRune := definition.RuneItem
	if resultRune == ItemShadowRune {
		resultRune = "shadow_rune"
	}
	return gatheringSpotDefinition{
		GatheringSpotSnapshot: GatheringSpotSnapshot{
			ID: definition.ID, Name: definition.Name, Skill: SkillRunecrafting,
			Position: definition.Position, Range: gatheringRange, RequiredLevel: definition.RequiredLevel,
			Station: craftingStationRunecrafting, RecipeID: definition.RuneItem,
		},
		ActionKind: timedActionProduction,
		RewardItem: definition.RuneItem, RewardQuantity: definition.RewardQuantity, RewardXP: definition.RewardXP,
		Interval: runecraftingInterval, Costs: []ItemCost{{ItemID: ItemBlankRune, Quantity: 1}},
		ActiveResult: "runecrafting", RewardResult: "crafted_" + resultRune + "s",
		MissingResult: "missing_runestones",
	}
}

func runecraftingMultiplier(requiredLevel, currentLevel int) int {
	if currentLevel <= requiredLevel {
		return 1
	}
	multiplier := 1 + (currentLevel-requiredLevel)/runecraftingLevelsPerMultiplier
	return min(multiplier, maximumRunecraftingMultiplier)
}
