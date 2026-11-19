package simulation

import "testing"

func TestPvPTargetRequiresBothPlayersInsideField(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.5))
	mustAddEmptyPlayer(t, world, "attacker", Position{Z: -100})
	mustAddEmptyPlayer(t, world, "defender", Position{})

	err := world.QueueTarget(TargetCommand{PlayerID: "attacker", Sequence: 1, TargetID: "defender"})
	if err == nil {
		t.Fatal("QueueTarget() accepted a defender outside the PvP field")
	}

	world.players["defender"].position = Position{X: 2, Z: -100}
	if err := world.QueueTarget(TargetCommand{PlayerID: "attacker", Sequence: 1, TargetID: "defender"}); err != nil {
		t.Fatalf("QueueTarget() inside PvP field error = %v", err)
	}
	world.applyTargetCommands()
	if got := world.players["attacker"].combatTarget; got != "defender" {
		t.Fatalf("combat target = %q, want defender", got)
	}
}

func TestPvPAttackDamagesPlayerWithoutGrantingExperience(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.5))
	mustAddEmptyPlayer(t, world, "attacker", Position{Z: -100})
	mustAddEmptyPlayer(t, world, "defender", Position{X: 1, Z: -100})
	attacker := world.players["attacker"]
	defender := world.players["defender"]
	attacker.equippedWeapon = WeaponSword
	attacker.combatStyle = CombatStyleSlash
	beforeAttackXP := attacker.skills[SkillAttack]
	beforeVitalityXP := attacker.skills[SkillVitality]

	if err := world.QueueTarget(TargetCommand{PlayerID: "attacker", Sequence: 1, TargetID: "defender"}); err != nil {
		t.Fatalf("QueueTarget() error = %v", err)
	}
	world.applyTargetCommands()
	world.resolveAutoAttacks()

	if defender.health >= defender.maximumHealth {
		t.Fatalf("defender health = %d, want damage below %d", defender.health, defender.maximumHealth)
	}
	if attacker.skills[SkillAttack] != beforeAttackXP || attacker.skills[SkillVitality] != beforeVitalityXP {
		t.Fatalf("PvP granted experience: attack=%d vitality=%d", attacker.skills[SkillAttack], attacker.skills[SkillVitality])
	}
}

func TestPvPCombatStopsWhenEitherPlayerLeavesField(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.5))
	mustAddEmptyPlayer(t, world, "attacker", Position{Z: -100})
	mustAddEmptyPlayer(t, world, "defender", Position{X: 1, Z: -100})
	if err := world.QueueTarget(TargetCommand{PlayerID: "attacker", Sequence: 1, TargetID: "defender"}); err != nil {
		t.Fatalf("QueueTarget() error = %v", err)
	}
	world.applyTargetCommands()
	world.players["defender"].position = Position{}
	world.resolveAutoAttacks()
	if got := world.players["attacker"].combatTarget; got != "" {
		t.Fatalf("combat target after leaving PvP field = %q, want empty", got)
	}
}

func TestPvPDefeatUsesSelectedDeathProtection(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.5))
	mustAddEmptyPlayer(t, world, "attacker", Position{Z: -100})
	mustAddEmptyPlayer(t, world, "defender", Position{X: 1, Z: -100})
	attacker := world.players["attacker"]
	defender := world.players["defender"]
	attacker.equippedWeapon = WeaponSword
	attacker.combatStyle = CombatStyleSlash
	defender.health = 1
	if !defender.inventory.add(ItemCoins, 10) || !defender.inventory.add(ItemAirRune, 5) {
		t.Fatal("could not prepare defender inventory")
	}
	if err := world.QueueSetDeathProtection(SetDeathProtectionCommand{
		PlayerID: "defender", Sequence: 1, Slots: []int{0},
	}); err != nil {
		t.Fatalf("QueueSetDeathProtection() error = %v", err)
	}
	world.applyPendingDeathProtections()

	if err := world.QueueTarget(TargetCommand{PlayerID: "attacker", Sequence: 1, TargetID: "defender"}); err != nil {
		t.Fatalf("QueueTarget() error = %v", err)
	}
	world.applyTargetCommands()
	world.resolveAutoAttacks()

	if !defender.defeated {
		t.Fatal("defender was not defeated")
	}
	if got := snapshotInventory(defender.inventory).itemsQuantity(ItemCoins); got != 10 {
		t.Fatalf("protected coins = %d, want 10", got)
	}
	if got := snapshotInventory(defender.inventory).itemsQuantity(ItemAirRune); got != 0 {
		t.Fatalf("unprotected runes = %d, want 0", got)
	}
	foundRunes := false
	for _, drop := range world.Snapshot().Drops {
		if drop.ItemID == ItemAirRune && drop.Quantity == 5 {
			foundRunes = true
		}
	}
	if !foundRunes {
		t.Fatal("unprotected runes were not dropped")
	}
}
