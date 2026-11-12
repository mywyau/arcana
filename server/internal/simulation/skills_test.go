package simulation

import "testing"

func TestAttackLevelCurve(t *testing.T) {
	tests := []struct {
		experience int64
		level      int
		next       int64
	}{
		{experience: 0, level: 1, next: 83},
		{experience: 82, level: 1, next: 83},
		{experience: 83, level: 2, next: 174},
		{experience: 173, level: 2, next: 174},
		{experience: 174, level: 3, next: 276},
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

func TestSkillCurveUsesLevelNinetyNineCap(t *testing.T) {
	maximumExperience := SkillExperienceForLevel(MaximumSkillLevel)
	if MaximumSkillLevel != 99 {
		t.Fatalf("maximum skill level = %d, want 99", MaximumSkillLevel)
	}
	if maximumExperience != 13_034_431 {
		t.Fatalf("level 99 XP = %d, want 13034431", maximumExperience)
	}
	if got := SkillLevelForExperience(maximumExperience + 1_000_000); got != MaximumSkillLevel {
		t.Fatalf("level above maximum XP = %d, want %d", got, MaximumSkillLevel)
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
		{WeaponWand, CombatStyleRapid, SkillMagic, true},
		{WeaponStaff, CombatStyleAccurate, SkillMagic, true},
		{WeaponEnchantedStaff, CombatStyleLongRange, SkillMagic, true},
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

func TestVitalityLevelAddsTenMaximumHealth(t *testing.T) {
	player := &playerState{
		health: playerMaximumHealth, maximumHealth: playerMaximumHealth,
		skills: newDefaultSkills(),
	}
	grantSkillExperience(player, SkillVitality, SkillExperienceForLevel(2))

	if got, want := player.maximumHealth, 110; got != want {
		t.Fatalf("maximum health at Vitality level 2 = %d, want %d", got, want)
	}
	if got, want := player.health, playerMaximumHealth; got != want {
		t.Fatalf("current health after Vitality level = %d, want unchanged %d", got, want)
	}

	player.skills[SkillVitality] = SkillExperienceForLevel(MaximumSkillLevel)
	refreshPlayerMaximumHealth(player)
	if got, want := player.maximumHealth, 1_080; got != want {
		t.Fatalf("maximum health at Vitality level 99 = %d, want %d", got, want)
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

func TestDebugSkillShrineSetsExactLevelExperience(t *testing.T) {
	world := NewWorld()
	shrine := world.gatheringSpots[debugSkillShrineID]
	mustAddPlayer(t, world, "player-one", shrine.Position, 4)
	if err := world.QueueDebugSetSkill(DebugSetSkillCommand{
		PlayerID: "player-one", Sequence: 1, Skill: SkillAttack, Level: 30,
	}); err != nil {
		t.Fatalf("QueueDebugSetSkill() returned an unexpected error: %v", err)
	}
	world.Step(50)
	attack := mustSkill(t, mustPlayer(t, world, "player-one"), SkillAttack)
	if attack.Level != 30 || attack.Experience != SkillExperienceForLevel(30) {
		t.Fatalf("Attack = %+v, want level 30 at its exact threshold", attack)
	}
}

func TestDebugSkillShrineRequiresProximity(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{X: 40, Z: 40}, 4)
	err := world.QueueDebugSetSkill(DebugSetSkillCommand{
		PlayerID: "player-one", Sequence: 1, Skill: SkillAttack, Level: 30,
	})
	if err == nil {
		t.Fatal("QueueDebugSetSkill() error = nil while away from the shrine")
	}
}

func TestDebugSkillShrineSetsExactExperience(t *testing.T) {
	world := NewWorld()
	shrine := world.gatheringSpots[debugSkillShrineID]
	mustAddPlayer(t, world, "player-one", shrine.Position, 4)
	if err := world.QueueDebugSetSkill(DebugSetSkillCommand{
		PlayerID: "player-one", Sequence: 1, Skill: SkillMagic,
		Experience: 173, SetExperience: true,
	}); err != nil {
		t.Fatalf("QueueDebugSetSkill() returned an unexpected error: %v", err)
	}
	world.Step(50)
	magic := mustSkill(t, mustPlayer(t, world, "player-one"), SkillMagic)
	if magic.Level != 2 || magic.Experience != 173 {
		t.Fatalf("Magic = %+v, want level 2 with 173 XP", magic)
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
