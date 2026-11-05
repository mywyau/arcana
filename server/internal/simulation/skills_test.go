package simulation

import "testing"

func TestAttackLevelCurve(t *testing.T) {
	tests := []struct {
		experience int64
		level      int
		next       int64
	}{
		{experience: 0, level: 1, next: 100},
		{experience: 99, level: 1, next: 100},
		{experience: 100, level: 2, next: 400},
		{experience: 399, level: 2, next: 400},
		{experience: 400, level: 3, next: 900},
	}
	for _, test := range tests {
		if got := SkillLevelForExperience(test.experience); got != test.level {
			t.Errorf("SkillLevelForExperience(%d) = %d, want %d", test.experience, got, test.level)
		}
		if got := SkillExperienceForLevel(test.level + 1); got != test.next {
			t.Errorf("next XP after level %d = %d, want %d", test.level, got, test.next)
		}
	}
}

func TestWeaponsTrainExpectedSkills(t *testing.T) {
	tests := map[WeaponID]SkillID{
		WeaponSword: SkillAttack, WeaponDagger: SkillAttack,
		WeaponMace: SkillAttack, WeaponBattleAxe: SkillAttack,
		WeaponBow: SkillRange, WeaponStaff: SkillMagic,
	}
	for weapon, want := range tests {
		got, trains := skillForWeapon(weapon)
		if !trains || got != want {
			t.Errorf("skillForWeapon(%q) = %q, %t; want %q, true", weapon, got, trains, want)
		}
	}
}

func mustSkill(t *testing.T, player PlayerSnapshot, id SkillID) SkillSnapshot {
	t.Helper()
	for _, skill := range player.Skills {
		if skill.ID == id {
			return skill
		}
	}
	t.Fatalf("skill %q not found", id)
	return SkillSnapshot{}
}
