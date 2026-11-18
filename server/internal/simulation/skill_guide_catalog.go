package simulation

import "sort"

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
			entries = append(entries, SkillGuideEntrySnapshot{
				ID: item.ID, Skill: requirement.Skill, Category: "equipment", Action: action,
				Name: item.Name, RequiredLevel: requirement.Level,
			})
		}
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
