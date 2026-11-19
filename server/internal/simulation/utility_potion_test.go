package simulation

import "testing"

func utilityPotionWorld(t *testing.T, id string) (*World, *playerState) {
	t.Helper()
	w := NewWorldWithRollSource(fixedRollSource(0.62))
	if err := w.AddPlayer("tester", Position{}, 4); err != nil {
		t.Fatal(err)
	}
	p := w.players["tester"]
	if !p.inventory.add(id, 3) {
		t.Fatal("could not add potion")
	}
	return w, p
}

func drinkUtilityPotion(t *testing.T, w *World, id string, sequence uint64) {
	t.Helper()
	if err := w.QueueConsumeItem(ConsumeItemCommand{PlayerID: "tester", ItemID: id, Sequence: sequence}); err != nil {
		t.Fatal(err)
	}
	w.applyPendingConsumables()
}

func TestMendingDraughtCadenceAndRefresh(t *testing.T) {
	w, p := utilityPotionWorld(t, ItemMendingDraught)
	p.maximumHealth = 250
	p.health = 30
	drinkUtilityPotion(t, w, ItemMendingDraught, 1)
	for pulse := uint64(1); pulse <= 507; pulse++ {
		w.combat.pulse = pulse
		w.advanceTimedPotions()
		if pulse == 7 {
			drinkUtilityPotion(t, w, ItemMendingDraught, 2)
			timers := activeCombatBoostSnapshots(p)
			if len(timers) != 1 || timers[0].ExpiresAtPulse != 507 {
				t.Fatalf("refreshed timers = %+v", timers)
			}
		}
		if want := 30 + int(pulse/10)*3; p.health != want {
			t.Fatalf("pulse %d: hp = %d, want %d", pulse, p.health, want)
		}
	}
	if len(activeCombatBoostSnapshots(p)) != 0 || p.inventory.quantity(ItemMendingDraught) != 1 {
		t.Fatal("draught did not expire or consumed the wrong number of doses")
	}
	drinkUtilityPotion(t, w, ItemMendingDraught, 3)
	p.maximumHealth = 200
	p.health = 101
	for elapsed := uint64(1); elapsed <= 500; elapsed++ {
		w.combat.pulse++
		w.advanceTimedPotions()
		if want := min(200, 101+int(elapsed/10)*2); p.health != want {
			t.Fatalf("single dose at pulse %d: hp = %d, want %d", elapsed, p.health, want)
		}
	}
	if p.health != p.maximumHealth {
		t.Fatalf("healing exceeded maximum health: %d", p.health)
	}
}

func TestSurefootHalvesCurrentSlowPenalty(t *testing.T) {
	w, p := utilityPotionWorld(t, ItemSurefootTonic)
	if err := w.QueueConsumeItem(ConsumeItemCommand{PlayerID: p.id, ItemID: ItemSurefootTonic, Sequence: 1}); err == nil {
		t.Fatal("accepted a dose without Slow")
	}
	w.applyStatus(p, StatusRoot, "enemy", 10)
	w.applyStatus(p, StatusStun, "enemy", 10)
	w.applyStatus(p, StatusSlow, "enemy", 10)
	before := p.statusesRevision
	drinkUtilityPotion(t, w, ItemSurefootTonic, 1)
	if !statusActive(p, StatusSlow, 0) || !statusActive(p, StatusRoot, 0) || !statusActive(p, StatusStun, 0) ||
		p.statusesRevision != before+1 || p.nextPotionConsumePulse != 2 || p.inventory.quantity(ItemSurefootTonic) != 2 {
		t.Fatal("unexpected status, cooldown or inventory after reducing Slow")
	}
	if got := effectivePlayerMovementSpeed(p); got != 0.85*p.speed || p.activeStatuses[StatusSlow].ExpiresAtPulse != 10 {
		t.Fatalf("tonic speed = %v, expiry = %v", got, p.activeStatuses[StatusSlow].ExpiresAtPulse)
	}
	w.combat.pulse = 2
	if err := w.QueueConsumeItem(ConsumeItemCommand{PlayerID: p.id, ItemID: ItemSurefootTonic, Sequence: 2}); err == nil {
		t.Fatal("accepted a second tonic for the same slow")
	}
	w.applyStatus(p, StatusSlow, "enemy", 10)
	if got := effectivePlayerMovementSpeed(p); got != 0.7*p.speed {
		t.Fatalf("a fresh slow produced speed %v", got)
	}
	w.combat.pulse = 13
	w.expireStatuses()
	if got := effectivePlayerMovementSpeed(p); got != p.speed {
		t.Fatalf("speed after expiry = %v", got)
	}
}

func TestGatherersTonicSuccessCapAndExpiry(t *testing.T) {
	w, p := utilityPotionWorld(t, ItemGatherersTonic)
	p.inventory.add(ItemBronzePickaxe, 1)
	spot := newGatheringSpots()[starterCopperVeinID]
	if w.gatheringAttemptSucceeded(p, spot) {
		t.Fatal("baseline attempt should fail")
	}
	drinkUtilityPotion(t, w, ItemGatherersTonic, 1)
	if !w.gatheringAttemptSucceeded(p, spot) {
		t.Fatal("tonic should turn a 62% roll into a success")
	}
	w.rollSource = fixedRollSource(0.651)
	if w.gatheringAttemptSucceeded(p, spot) {
		t.Fatal("tonic added more than five percentage points")
	}
	w.rollSource = fixedRollSource(0.62)
	capped := spot
	capped.MaximumSuccessChance = 0.61
	if w.gatheringAttemptSucceeded(p, capped) {
		t.Fatal("tonic exceeded resource cap")
	}
	w.combat.pulse = 300
	w.advanceTimedPotions()
	if w.gatheringAttemptSucceeded(p, spot) || len(activeCombatBoostSnapshots(p)) != 0 {
		t.Fatal("gathering bonus or HUD timer remained after expiry")
	}
}
