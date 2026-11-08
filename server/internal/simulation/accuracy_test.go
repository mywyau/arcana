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

func TestAttackLevelReducesTargetDodgeChance(t *testing.T) {
	weapon, _ := weaponDefinition(WeaponSword)
	player := &playerState{skills: newDefaultSkills(), combatStyle: CombatStyleStab}
	target := &TargetSnapshot{Evasion: dummyEvasionScore, DodgeRate: dummyDodgeRate}
	levelOneChance := targetDodgeChance(player, target, weapon)
	if math.Abs(levelOneChance-0.096) > 0.000_001 {
		t.Fatalf("level-one target dodge chance = %.3f, want 0.096", levelOneChance)
	}

	player.skills[SkillAttack] = SkillExperienceForLevel(50)
	if got := targetDodgeChance(player, target, weapon); got != minimumDodgeChance {
		t.Fatalf("level-50 target dodge chance = %.3f, want floor %.3f", got, minimumDodgeChance)
	}
}

func TestAgilityAndPrayerIncreasePlayerDodgeChance(t *testing.T) {
	player := &playerState{
		skills:              newDefaultSkills(),
		dodgeRate:           playerDodgeRate,
		activePrayers:       make(map[PrayerID]bool),
		prayerDrainProgress: make(map[PrayerID]uint64),
	}
	if got := playerDodgeChance(dummyAccuracyScore, player); math.Abs(got-0.10) > 0.000_001 {
		t.Fatalf("level-one player dodge chance = %.3f, want 0.100", got)
	}

	player.skills[SkillAgility] = SkillExperienceForLevel(50)
	withoutPrayer := playerDodgeChance(dummyAccuracyScore, player)
	if math.Abs(withoutPrayer-0.198) > 0.000_001 {
		t.Fatalf("level-50 Agility dodge chance = %.3f, want 0.198", withoutPrayer)
	}
	player.activePrayers[PrayerEvasion] = true
	withPrayer := playerDodgeChance(dummyAccuracyScore, player)
	if math.Abs(withPrayer-(withoutPrayer+evasionDodgeBonus)) > 0.000_001 {
		t.Fatalf("Evasion prayer dodge chance = %.3f, want %.3f", withPrayer, withoutPrayer+evasionDodgeBonus)
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

func TestEnemyDefenceUsesWeaponStyleDamageType(t *testing.T) {
	weapon, _ := weaponDefinition(WeaponSword)
	player := &playerState{skills: newDefaultSkills(), combatStyle: CombatStyleStab}
	target := &TargetSnapshot{
		Evasion: 1, DodgeRate: defaultBaseDodgeBonus,
		Defence: DefenceBonuses{Stab: 20, Slash: 0},
	}
	stabChance := targetDodgeChance(player, target, weapon)
	player.combatStyle = CombatStyleSlash
	slashChance := targetDodgeChance(player, target, weapon)
	if stabChance <= slashChance {
		t.Fatalf("stab dodge chance %.3f <= slash dodge chance %.3f against stab defence", stabChance, slashChance)
	}
}
