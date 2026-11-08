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
	tests := []struct {
		weapon WeaponID
		style  CombatStyleID
		skill  SkillID
		trains bool
	}{
		{WeaponSword, CombatStyleStab, SkillAttack, true},
		{WeaponSword, CombatStyleSlash, SkillStrength, true},
		{WeaponSword, CombatStyleDefensive, SkillDefence, true},
		{WeaponSteelSword, CombatStyleStab, SkillAttack, true},
		{WeaponSteelSword, CombatStyleSlash, SkillStrength, true},
		{WeaponRuneSword, CombatStyleStab, SkillAttack, true},
		{WeaponRuneSword, CombatStyleSlash, SkillStrength, true},
		{WeaponRuneSword, CombatStyleDefensive, SkillDefence, true},
		{WeaponMace, CombatStyleCrush, SkillAttack, true},
		{WeaponMace, CombatStyleHeavyCrush, SkillStrength, true},
		{WeaponBattleAxe, CombatStyleSlash, SkillAttack, true},
		{WeaponBow, CombatStyleRapid, SkillRange, true},
		{WeaponStaff, CombatStyleAccurate, SkillMagic, true},
		{WeaponUnarmed, CombatStyleCrush, SkillAttack, true},
		{WeaponUnarmed, CombatStyleHeavyCrush, SkillStrength, true},
		{WeaponUnarmed, CombatStyleDefensive, SkillDefence, true},
	}
	for _, test := range tests {
		got, trains := skillForCombatStyle(test.weapon, test.style)
		if trains != test.trains || got != test.skill {
			t.Errorf("skillForCombatStyle(%q, %q) = %q, %t; want %q, %t",
				test.weapon, test.style, got, trains, test.skill, test.trains)
		}
	}
}

func TestExperienceLockPreventsAllSkillExperience(t *testing.T) {
	player := &playerState{skills: newDefaultSkills(), experienceLocked: true}
	grantSkillExperience(player, SkillAttack, 10)
	grantSkillExperience(player, SkillCooking, 25)
	if player.skills[SkillAttack] != 0 || player.skills[SkillCooking] != 0 {
		t.Fatalf("locked skills gained XP: %+v", player.skills)
	}
	player.experienceLocked = false
	grantSkillExperience(player, SkillCooking, 25)
	if player.skills[SkillCooking] != 25 {
		t.Fatalf("unlocked Cooking XP = %d, want 25", player.skills[SkillCooking])
	}
}

func TestExperienceLockCommandAppliesOnSimulationStep(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	if err := world.QueueExperienceLock(ExperienceLockCommand{
		PlayerID: "player-one", Sequence: 1, Locked: true,
	}); err != nil {
		t.Fatalf("QueueExperienceLock() returned an unexpected error: %v", err)
	}
	world.Step(50)
	player := mustPlayer(t, world, "player-one")
	if !player.ExperienceLocked || player.LastExperienceLockSequence != 1 {
		t.Fatalf("experience lock snapshot = %+v", player)
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
