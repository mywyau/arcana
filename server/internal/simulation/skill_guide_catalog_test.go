package simulation

import "testing"

func TestSkillGuideCatalogIncludesGatheringAndEquipmentUnlocks(t *testing.T) {
	entries := SkillGuideCatalog()
	if len(entries) == 0 {
		t.Fatal("SkillGuideCatalog() returned no entries")
	}
	foundGathering := false
	foundEquipment := false
	for index, entry := range entries {
		if entry.ID == "" || entry.Skill == "" || entry.Name == "" || entry.RequiredLevel < 1 {
			t.Fatalf("skill guide entry is incomplete: %+v", entry)
		}
		foundGathering = foundGathering || entry.Category == "gathering"
		foundEquipment = foundEquipment || entry.Category == "equipment"
		if index > 0 {
			previous := entries[index-1]
			if previous.Skill > entry.Skill ||
				(previous.Skill == entry.Skill && previous.RequiredLevel > entry.RequiredLevel) {
				t.Fatalf("skill guide is not ordered: %+v before %+v", previous, entry)
			}
		}
	}
	if !foundGathering || !foundEquipment {
		t.Fatalf("skill guide categories: gathering=%t equipment=%t", foundGathering, foundEquipment)
	}
}
