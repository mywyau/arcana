package simulation

import "testing"

func TestPrayerCatalogIsCompleteAndDeterministic(t *testing.T) {
	first := PrayerCatalog()
	second := PrayerCatalog()
	if len(first) != len(prayerDefinitions) || len(first) != len(prayerCatalogOrder) || len(first) != 7 {
		t.Fatalf("PrayerCatalog() returned %d prayers, want %d", len(first), len(prayerDefinitions))
	}
	for index, prayer := range first {
		if prayer.ID == "" || prayer.Name == "" || prayer.RequiredLevel <= 0 || prayer.DrainAmount <= 0 || prayer.DrainIntervalPulses == 0 || prayer.DrainIntervalSeconds <= 0 {
			t.Fatalf("prayer %d is incomplete: %+v", index, prayer)
		}
		if prayer.ID != string(prayerCatalogOrder[index]) || prayer.ID != second[index].ID {
			t.Fatalf("catalogue order changed between calls: %q then %q", prayer.ID, second[index].ID)
		}
	}
}

func TestPrayerSetCatalogReturnsDetachedSortedMembership(t *testing.T) {
	sets := PrayerSetCatalog()
	if len(sets) != len(prayerSets) || len(sets) != 2 {
		t.Fatalf("PrayerSetCatalog() returned %d sets, want %d", len(sets), len(prayerSets))
	}
	for index, set := range sets {
		if set.ID == "" || set.Name == "" || len(set.PrayerIDs) == 0 {
			t.Fatalf("prayer set %d is incomplete: %+v", index, set)
		}
		if set.ID != string(prayerSetCatalogOrder[index]) {
			t.Fatalf("prayer set order changed: got %q at index %d", set.ID, index)
		}
	}

	original := sets[0].PrayerIDs[0]
	sets[0].PrayerIDs[0] = "changed"
	if got := PrayerSetCatalog()[0].PrayerIDs[0]; got != original {
		t.Fatalf("PrayerSetCatalog() exposed mutable membership: got %q, want %q", got, original)
	}
}
