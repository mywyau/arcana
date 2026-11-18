package simulation

import (
	"sort"
	"strconv"
	"strings"
)

// SkillGuideEntrySnapshot exposes level unlock metadata without leaking the
// rest of the authoritative item or world definitions to the client.
type SkillGuideEntrySnapshot struct {
	ID            string
	Skill         SkillID
	Category      string
	Action        string
	Name          string
	RequiredLevel int
}

// SkillGuideCatalog returns gathering and equipment unlocks. Production,
// spell, and prayer unlocks already have dedicated client catalogues.
func SkillGuideCatalog() []SkillGuideEntrySnapshot {
	entries := make([]SkillGuideEntrySnapshot, 0)
	tierSummaries := make(map[string]SkillGuideEntrySnapshot)
	for _, spot := range newGatheringSpots() {
		if spot.RequiredLevel < 1 || spot.Skill == "" {
			continue
		}
		action := "Use"
		switch spot.Skill {
		case SkillFishing:
			action = "Catch"
		case SkillMining:
			action = "Mine"
		case SkillWoodcutting:
			action = "Cut"
		case SkillAgility:
			action = "Cross"
		case SkillRunecrafting:
			action = "Craft"
		default:
			continue
		}
		entries = append(entries, SkillGuideEntrySnapshot{
			ID: spot.ID, Skill: spot.Skill, Category: "gathering", Action: action,
			Name: spot.Name, RequiredLevel: spot.RequiredLevel,
		})
	}
	for _, item := range itemCatalog {
		if item.Virtual {
			continue
		}
		for _, requirement := range item.EquipRequirements {
			action := "Wear"
			if item.Kind == ItemKindWeapon {
				action = "Wield"
			} else if item.Ammunition {
				action = "Equip"
			}
			entry := SkillGuideEntrySnapshot{
				ID: item.ID, Skill: requirement.Skill, Category: skillGuideItemCategory(item), Action: action,
				Name: item.Name, RequiredLevel: requirement.Level,
			}
			if summary, ok := skillGuideTierSummary(item, entry); ok {
				key := strings.Join([]string{
					string(summary.Skill), summary.Category, summary.Name, strconv.Itoa(summary.RequiredLevel),
				}, "|")
				tierSummaries[key] = summary
				continue
			}
			entries = append(entries, entry)
		}
	}
	for _, summary := range tierSummaries {
		entries = append(entries, summary)
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Skill != entries[j].Skill {
			return entries[i].Skill < entries[j].Skill
		}
		if entries[i].RequiredLevel != entries[j].RequiredLevel {
			return entries[i].RequiredLevel < entries[j].RequiredLevel
		}
		if entries[i].Category != entries[j].Category {
			return entries[i].Category < entries[j].Category
		}
		return entries[i].Name < entries[j].Name
	})
	return entries
}

func skillGuideTierSummary(item ItemDefinition, entry SkillGuideEntrySnapshot) (SkillGuideEntrySnapshot, bool) {
	tier, ok := skillGuideTierName(item, entry.Category)
	if !ok {
		return SkillGuideEntrySnapshot{}, false
	}
	suffix := "equipment"
	switch entry.Category {
	case "weapons":
		suffix = "weapons"
	case "armour":
		suffix = "armour"
	case "ammunition":
		suffix = "ammunition"
	}
	entry.ID = "tier:" + strings.ToLower(strings.ReplaceAll(tier, "'", "")) + ":" + entry.Category
	entry.Name = tier + " " + suffix
	return entry, true
}

func skillGuideTierName(item ItemDefinition, category string) (string, bool) {
	for _, tier := range []string{"Bronze", "Iron", "Steel", "Mithril", "Adamantite", "Rune"} {
		if strings.HasPrefix(item.Name, tier+" ") {
			return tier, true
		}
	}
	if category != "armour" {
		return "", false
	}
	for _, tier := range []string{"Mage's", "Moonspun", "Enchanted", "Stardew", "Swift", "Scout", "Tracker", "Hunter", "Marksman"} {
		if strings.HasPrefix(item.Name, tier+" ") {
			return tier, true
		}
	}
	return "", false
}

func skillGuideItemCategory(item ItemDefinition) string {
	switch {
	case item.Kind == ItemKindWeapon:
		return "weapons"
	case item.Kind == ItemKindArmour:
		return "armour"
	case item.Ammunition:
		return "ammunition"
	case item.Kind == ItemKindSpellbook:
		return "spellbooks"
	case item.GatheringTool.Family != "":
		return "tools"
	default:
		return "equipment"
	}
}
