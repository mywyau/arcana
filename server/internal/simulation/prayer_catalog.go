package simulation

type PrayerCatalogSnapshot struct {
	ID                       string
	Name                     string
	RequiredLevel            int
	MeleeDamageBonus         float64
	RangedDamageBonus        float64
	MagicDamageBonus         float64
	HealthRegeneration       float64
	MeleeHitChanceBonus      float64
	RangedHitChanceBonus     float64
	MagicHitChanceBonus      float64
	DamageReduction          float64
	MeleeProtection          float64
	RangedProtection         float64
	MagicProtection          float64
	EmergencyHealthThreshold float64
	EmergencyHealFraction    float64
	EmergencyPrayerCost      int
	DeathBurstDamageFraction float64
	DeathBurstRadius         float64
	DodgeBonus               float64
	DrainAmount              int
	DrainIntervalPulses      uint64
	DrainIntervalSeconds     float64
	ExclusiveGroup           string
}

type PrayerSetCatalogSnapshot struct {
	ID        string
	Name      string
	PrayerIDs []string
}

func PrayerCatalog() []PrayerCatalogSnapshot {
	prayers := make([]PrayerCatalogSnapshot, 0, len(prayerDefinitions))
	for _, prayerID := range prayerCatalogOrder {
		prayer, exists := prayerDefinitions[prayerID]
		if !exists {
			continue
		}
		prayers = append(prayers, PrayerCatalogSnapshot{
			ID: string(prayer.ID), Name: prayer.Name, RequiredLevel: prayer.RequiredLevel,
			MeleeDamageBonus: prayer.MeleeDamageBonus, RangedDamageBonus: prayer.RangedDamageBonus,
			MagicDamageBonus: prayer.MagicDamageBonus, HealthRegeneration: prayer.HealthRegeneration,
			MeleeHitChanceBonus:  prayer.MeleeHitChanceBonus,
			RangedHitChanceBonus: prayer.RangedHitChanceBonus, MagicHitChanceBonus: prayer.MagicHitChanceBonus,
			DamageReduction: prayer.DamageReduction,
			MeleeProtection: prayer.MeleeProtection, RangedProtection: prayer.RangedProtection,
			MagicProtection:          prayer.MagicProtection,
			EmergencyHealthThreshold: prayer.EmergencyHealthThreshold,
			EmergencyHealFraction:    prayer.EmergencyHealFraction, EmergencyPrayerCost: prayer.EmergencyPrayerCost,
			DeathBurstDamageFraction: prayer.DeathBurstDamageFraction, DeathBurstRadius: prayer.DeathBurstRadius,
			DodgeBonus: prayer.DodgeBonus, DrainAmount: prayer.DrainAmount,
			DrainIntervalPulses:  prayer.DrainIntervalPulses,
			DrainIntervalSeconds: float64(prayer.DrainIntervalPulses) * actionPulseInterval.Seconds(),
			ExclusiveGroup:       prayer.ExclusiveGroup,
		})
	}
	return prayers
}

func PrayerSetCatalog() []PrayerSetCatalogSnapshot {
	sets := make([]PrayerSetCatalogSnapshot, 0, len(prayerSets))
	for _, setID := range prayerSetCatalogOrder {
		set, exists := prayerSets[setID]
		if !exists {
			continue
		}
		prayerIDs := make([]string, 0, len(set.PrayerIDs))
		for _, prayerID := range prayerCatalogOrder {
			if set.PrayerIDs[prayerID] {
				prayerIDs = append(prayerIDs, string(prayerID))
			}
		}
		sets = append(sets, PrayerSetCatalogSnapshot{ID: string(setID), Name: set.Name, PrayerIDs: prayerIDs})
	}
	return sets
}
