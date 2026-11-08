package simulation

import (
	"math"
	"testing"
)

func TestWeaponFamiliesUseRelevantAccuracySkill(t *testing.T) {
	tests := []struct {
		name   string
		weapon WeaponID
		skill  SkillID
	}{
		{name: "melee", weapon: WeaponSword, skill: SkillAttack},
		{name: "ranged", weapon: WeaponBow, skill: SkillRange},
		{name: "magic", weapon: WeaponStaff, skill: SkillMagic},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			weapon, exists := weaponDefinition(test.weapon)
			if !exists {
				t.Fatalf("weaponDefinition(%q) does not exist", test.weapon)
			}
			if got := weaponAccuracySkill(weapon); got != test.skill {
				t.Fatalf("weaponAccuracySkill(%q) = %q, want %q", test.weapon, got, test.skill)
			}
		})
	}
}

func TestAttackLevelImprovesHitChanceFromModerateEarlyAccuracy(t *testing.T) {
	weapon, _ := weaponDefinition(WeaponSword)
	player := &playerState{skills: newDefaultSkills(), combatStyle: CombatStyleStab}
	target := &TargetSnapshot{Evasion: dummyEvasionScore, DodgeRate: dummyDodgeRate}
	levelOneHitChance := 1 - targetDodgeChance(player, target, weapon)
	if levelOneHitChance < 0.60 || levelOneHitChance > 0.70 {
		t.Fatalf("level-one hit chance = %.3f, want 0.60-0.70", levelOneHitChance)
	}

	player.skills[SkillAttack] = SkillExperienceForLevel(50)
	levelFiftyHitChance := 1 - targetDodgeChance(player, target, weapon)
	if levelFiftyHitChance < 0.80 || levelFiftyHitChance > 0.90 {
		t.Fatalf("level-50 hit chance = %.3f, want 0.80-0.90", levelFiftyHitChance)
	}
	if levelFiftyHitChance <= levelOneHitChance {
		t.Fatalf("level-50 hit chance %.3f <= level-one hit chance %.3f", levelFiftyHitChance, levelOneHitChance)
	}
}

func TestAgilityAndPrayerIncreasePlayerDodgeChance(t *testing.T) {
	player := &playerState{
		skills:              newDefaultSkills(),
		dodgeRate:           playerDodgeRate,
		activePrayers:       make(map[PrayerID]bool),
		prayerDrainProgress: make(map[PrayerID]uint64),
	}
	levelOneDodge := playerDodgeChance(dummyAccuracyScore, player)
	if levelOneDodge < 0.30 || levelOneDodge > 0.40 {
		t.Fatalf("level-one player dodge chance = %.3f, want 0.30-0.40", levelOneDodge)
	}

	player.skills[SkillAgility] = SkillExperienceForLevel(50)
	withoutPrayer := playerDodgeChance(dummyAccuracyScore, player)
	if withoutPrayer < 0.40 || withoutPrayer > 0.50 {
		t.Fatalf("level-50 Agility dodge chance = %.3f, want 0.40-0.50", withoutPrayer)
	}
	if withoutPrayer <= levelOneDodge {
		t.Fatalf("level-50 Agility dodge %.3f <= level-one dodge %.3f", withoutPrayer, levelOneDodge)
	}
	player.activePrayers[PrayerEvasion] = true
	withPrayer := playerDodgeChance(dummyAccuracyScore, player)
	if math.Abs(withPrayer-(withoutPrayer+evasionDodgeBonus)) > 0.000_001 {
		t.Fatalf("Evasion prayer dodge chance = %.3f, want %.3f", withPrayer, withoutPrayer+evasionDodgeBonus)
	}
}

func TestEarlyEnemiesHaveModerateHitChance(t *testing.T) {
	player := &playerState{skills: newDefaultSkills()}
	for _, test := range []struct {
		name     string
		accuracy int
	}{
		{name: "wild fowl", accuracy: 5},
		{name: "cow", accuracy: 7},
		{name: "training dummy", accuracy: dummyAccuracyScore},
	} {
		t.Run(test.name, func(t *testing.T) {
			hitChance := 1 - playerDodgeChance(test.accuracy, player)
			if hitChance < 0.60 || hitChance > 0.70 {
				t.Fatalf("early enemy hit chance = %.3f, want 0.60-0.70", hitChance)
			}
		})
	}
}

func TestEvasionEquipmentRaisesPlayerDodgeChance(t *testing.T) {
	player := &playerState{
		skills:         newDefaultSkills(),
		equippedArmour: make(map[ArmourSlot]string),
	}
	naked := playerDodgeChance(dummyAccuracyScore, player)
	player.equippedArmour = map[ArmourSlot]string{
		ArmourSlotHead: ItemSwiftHelm,
		ArmourSlotBody: ItemSwiftBody,
		ArmourSlotFeet: ItemSwiftBoots,
	}
	withSwiftSet := playerDodgeChance(dummyAccuracyScore, player)
	if got := equippedEvasionBonusForPlayer(player); got != 5 {
		t.Fatalf("Swift set evasion bonus = %d, want 5", got)
	}
	if withSwiftSet <= naked {
		t.Fatalf("Swift set dodge %.3f <= naked dodge %.3f", withSwiftSet, naked)
	}
}

func TestCombatStylesModifyAccuracyScore(t *testing.T) {
	player := &playerState{skills: newDefaultSkills(), combatStyle: CombatStyleAccurate}
	accurateWeapon, _ := weaponDefinitionForStyle(WeaponBow, CombatStyleAccurate)
	accurate := playerAccuracyScore(player, accurateWeapon)
	player.combatStyle = CombatStyleRapid
	rapidWeapon, _ := weaponDefinitionForStyle(WeaponBow, CombatStyleRapid)
	rapid := playerAccuracyScore(player, rapidWeapon)
	if accurate-rapid != 9 {
		t.Fatalf("accurate score %d minus rapid score %d = %d, want 9", accurate, rapid, accurate-rapid)
	}
}

func TestEnemyDefenceMitigatesDamageWithoutChangingEvasion(t *testing.T) {
	weapon, _ := weaponDefinition(WeaponSword)
	player := &playerState{skills: newDefaultSkills(), combatStyle: CombatStyleStab}
	target := &TargetSnapshot{
		Evasion: 5, DodgeRate: defaultBaseDodgeBonus,
		Defence: DefenceBonuses{Stab: 20, Slash: 0},
	}
	withDefence := targetDodgeChance(player, target, weapon)
	target.Defence = DefenceBonuses{}
	withoutDefence := targetDodgeChance(player, target, weapon)
	if withDefence != withoutDefence {
		t.Fatalf("target defence changed dodge chance from %.3f to %.3f", withoutDefence, withDefence)
	}
	target.Defence = DefenceBonuses{Stab: 20, Slash: 0}
	if stabReduction, slashReduction := targetDamageReduction(target, DamageTypeStab), targetDamageReduction(target, DamageTypeSlash); stabReduction <= slashReduction {
		t.Fatalf("stab reduction %.3f <= slash reduction %.3f", stabReduction, slashReduction)
	}
}
