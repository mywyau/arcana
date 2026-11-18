package simulation

import "testing"

func TestSkillGuideCatalogIncludesGatheringWeaponAndArmourUnlocks(t *testing.T) {
	entries := SkillGuideCatalog()
	if len(entries) == 0 {
		t.Fatal("SkillGuideCatalog() returned no entries")
	}
	foundGathering := false
	foundWeapons := false
	foundArmour := false
	for index, entry := range entries {
		if entry.ID == "" || entry.Skill == "" || entry.Name == "" || entry.RequiredLevel < 1 {
			t.Fatalf("skill guide entry is incomplete: %+v", entry)
		}
		foundGathering = foundGathering || entry.Category == "gathering"
		foundWeapons = foundWeapons || entry.Category == "weapons"
		foundArmour = foundArmour || entry.Category == "armour"
		if index > 0 {
			previous := entries[index-1]
			if previous.Skill > entry.Skill ||
				(previous.Skill == entry.Skill && previous.RequiredLevel > entry.RequiredLevel) {
				t.Fatalf("skill guide is not ordered: %+v before %+v", previous, entry)
			}
		}
	}
	if !foundGathering || !foundWeapons || !foundArmour {
		t.Fatalf("skill guide categories: gathering=%t weapons=%t armour=%t", foundGathering, foundWeapons, foundArmour)
	}
}

func TestSkillGuideCatalogCollapsesEquipmentTiers(t *testing.T) {
	entries := SkillGuideCatalog()
	foundBronzeWeapons := false
	foundBronzeArmour := false
	foundIronArmour := false
	for _, entry := range entries {
		foundBronzeWeapons = foundBronzeWeapons ||
			entry.Skill == SkillAttack && entry.Category == "weapons" &&
				entry.Name == "Bronze weapons" && entry.RequiredLevel == bronzeEquipmentLevel
		foundBronzeArmour = foundBronzeArmour ||
			entry.Skill == SkillDefence && entry.Category == "armour" &&
				entry.Name == "Bronze armour" && entry.RequiredLevel == bronzeEquipmentLevel
		foundIronArmour = foundIronArmour ||
			entry.Skill == SkillDefence && entry.Category == "armour" &&
				entry.Name == "Iron armour" && entry.RequiredLevel == ironEquipmentLevel
		if entry.Skill == SkillDefence && (entry.Name == "Bronze Helm" || entry.Name == "Iron Platebody") {
			t.Fatalf("tiered equipment remained as an individual guide entry: %+v", entry)
		}
	}
	if !foundBronzeWeapons || !foundBronzeArmour || !foundIronArmour {
		t.Fatalf(
			"tier summaries: bronze weapons=%t bronze armour=%t iron armour=%t",
			foundBronzeWeapons, foundBronzeArmour, foundIronArmour,
		)
	}
}
