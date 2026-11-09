package protocol

import "testing"

func TestDecodeAuthenticateMessage(t *testing.T) {
	message, err := DecodeClientMessage([]byte(`{"type":"authenticate","guest_token":"token"}`))
	if err != nil {
		t.Fatalf("DecodeClientMessage() returned an unexpected error: %v", err)
	}
	if message.Type != MessageAuthenticate || message.GuestToken != "token" || message.Sequence != 0 {
		t.Fatalf("DecodeClientMessage() = %+v", message)
	}
}

func TestDecodeAccountMessages(t *testing.T) {
	login, err := DecodeClientMessage([]byte(`{"type":"authenticate","email":"player@example.com","password":"strong-password"}`))
	if err != nil || login.Email != "player@example.com" || login.Password != "strong-password" {
		t.Fatalf("decoded login = %+v, error %v", login, err)
	}
	session, err := DecodeClientMessage([]byte(`{"type":"authenticate","session_token":"session-token"}`))
	if err != nil || session.SessionToken != "session-token" {
		t.Fatalf("decoded session authentication = %+v, error %v", session, err)
	}
	registration, err := DecodeClientMessage([]byte(`{"type":"register_account","sequence":1,"email":"player@example.com","password":"strong-password"}`))
	if err != nil || registration.Sequence != 1 || registration.Email != "player@example.com" {
		t.Fatalf("decoded registration = %+v, error %v", registration, err)
	}
}

func TestDecodeClientMessage(t *testing.T) {
	message, err := DecodeClientMessage([]byte(`{"type":"move_to","sequence":3,"destination":{"x":1.5,"y":4,"z":-2}}`))
	if err != nil {
		t.Fatalf("DecodeClientMessage() returned an unexpected error: %v", err)
	}
	if message.Type != MessageMoveTo || message.Sequence != 3 {
		t.Fatalf("DecodeClientMessage() = %+v", message)
	}
	if message.Destination == nil || message.Destination.X != 1.5 || message.Destination.Y != 4 || message.Destination.Z != -2 {
		t.Fatalf("DecodeClientMessage().Destination = %+v", message.Destination)
	}
}

func TestDecodeDebugSetSkillMessage(t *testing.T) {
	message, err := DecodeClientMessage([]byte(`{"type":"debug_set_skill","sequence":4,"skill_id":"attack","level":30}`))
	if err != nil {
		t.Fatalf("DecodeClientMessage() returned an unexpected error: %v", err)
	}
	if message.Type != MessageDebugSetSkill || message.SkillID != "attack" || message.Level != 30 {
		t.Fatalf("DecodeClientMessage() = %+v", message)
	}
}

func TestDecodeEngageTargetMessage(t *testing.T) {
	message, err := DecodeClientMessage([]byte(`{"type":"engage_target","sequence":2,"target_id":"dummy"}`))
	if err != nil {
		t.Fatalf("DecodeClientMessage() returned an unexpected error: %v", err)
	}
	if message.Type != MessageEngageTarget || message.Sequence != 2 || message.TargetID != "dummy" {
		t.Fatalf("DecodeClientMessage() = %+v", message)
	}
}

func TestDecodeDisengageMessage(t *testing.T) {
	message, err := DecodeClientMessage([]byte(`{"type":"disengage","sequence":3}`))
	if err != nil {
		t.Fatalf("DecodeClientMessage() returned an unexpected error: %v", err)
	}
	if message.Type != MessageDisengage || message.Sequence != 3 {
		t.Fatalf("DecodeClientMessage() = %+v", message)
	}
}

func TestDecodeEquipItemMessage(t *testing.T) {
	message, err := DecodeClientMessage([]byte(`{"type":"equip_item","sequence":4,"item_id":"dagger"}`))
	if err != nil {
		t.Fatalf("DecodeClientMessage() returned an unexpected error: %v", err)
	}
	if message.Type != MessageEquipItem || message.Sequence != 4 || message.ItemID != "dagger" {
		t.Fatalf("DecodeClientMessage() = %+v", message)
	}
}

func TestDecodeUnequipItemMessage(t *testing.T) {
	message, err := DecodeClientMessage([]byte(`{"type":"unequip_item","sequence":5,"item_id":"training_helm"}`))
	if err != nil {
		t.Fatalf("DecodeClientMessage() returned an unexpected error: %v", err)
	}
	if message.Type != MessageUnequipItem || message.Sequence != 5 || message.ItemID != "training_helm" {
		t.Fatalf("DecodeClientMessage() = %+v", message)
	}
}

func TestDecodeConsumeItemMessage(t *testing.T) {
	message, err := DecodeClientMessage([]byte(`{"type":"consume_item","sequence":6,"item_id":"cooked_minnow","inventory_slot":3}`))
	if err != nil {
		t.Fatalf("DecodeClientMessage() returned an unexpected error: %v", err)
	}
	if message.Type != MessageConsumeItem || message.Sequence != 6 || message.ItemID != "cooked_minnow" || message.InventorySlot != 3 {
		t.Fatalf("decoded consume message = %+v", message)
	}
}

func TestDecodeBrewMessage(t *testing.T) {
	message, err := DecodeClientMessage([]byte(`{"type":"brew","sequence":7,"item_ids":["vial_of_water","sunleaf","sacred_dust"]}`))
	if err != nil {
		t.Fatalf("DecodeClientMessage() returned an unexpected error: %v", err)
	}
	if message.Type != MessageBrew || message.Sequence != 7 || len(message.ItemIDs) != 3 {
		t.Fatalf("decoded brew message = %+v", message)
	}
}

func TestDecodeFletchMessage(t *testing.T) {
	message, err := DecodeClientMessage([]byte(`{"type":"fletch","sequence":8,"recipe_id":"iron_arrows","quantity":3}`))
	if err != nil {
		t.Fatalf("DecodeClientMessage() returned an unexpected error: %v", err)
	}
	if message.Type != MessageFletch || message.Sequence != 8 || message.RecipeID != "iron_arrows" || message.Quantity != 3 {
		t.Fatalf("decoded fletch message = %+v", message)
	}
}

func TestDecodeGatherMessage(t *testing.T) {
	message, err := DecodeClientMessage([]byte(`{"type":"gather","sequence":7,"spot_id":"starter_furnace","recipe_id":"bronze_bar","quantity":3}`))
	if err != nil {
		t.Fatalf("DecodeClientMessage() returned an unexpected error: %v", err)
	}
	if message.Type != MessageGather || message.Sequence != 7 || message.SpotID != "starter_furnace" || message.RecipeID != "bronze_bar" || message.Quantity != 3 {
		t.Fatalf("decoded gather message = %+v", message)
	}
}

func TestDecodeGatherMessageRejectsNegativeQuantity(t *testing.T) {
	_, err := DecodeClientMessage([]byte(`{"type":"gather","sequence":7,"spot_id":"starter_furnace","recipe_id":"bronze_bar","quantity":-1}`))
	if err == nil {
		t.Fatal("DecodeClientMessage() error = nil for negative gather quantity")
	}
}

func TestDecodeBankTransferMessages(t *testing.T) {
	deposit, err := DecodeClientMessage([]byte(`{"type":"bank_deposit","sequence":8,"item_id":"air_rune","inventory_slot":3}`))
	if err != nil || deposit.Type != MessageBankDeposit || deposit.ItemID != "air_rune" || deposit.InventorySlot != 3 {
		t.Fatalf("decoded deposit = %+v, error %v", deposit, err)
	}
	withdraw, err := DecodeClientMessage([]byte(`{"type":"bank_withdraw","sequence":9,"bank_slot":4,"quantity":12}`))
	if err != nil || withdraw.Type != MessageBankWithdraw || withdraw.BankSlot != 4 || withdraw.Quantity != 12 {
		t.Fatalf("decoded withdrawal = %+v, error %v", withdraw, err)
	}
}

func TestDecodeAmmunitionTransferMessages(t *testing.T) {
	deposit, err := DecodeClientMessage([]byte(`{"type":"ammunition_deposit","sequence":10,"inventory_slot":3}`))
	if err != nil || deposit.Type != MessageAmmunitionDeposit || deposit.InventorySlot != 3 {
		t.Fatalf("decoded ammunition deposit = %+v, error %v", deposit, err)
	}
	withdraw, err := DecodeClientMessage([]byte(`{"type":"ammunition_withdraw","sequence":11,"ammunition_slot":1}`))
	if err != nil || withdraw.Type != MessageAmmunitionWithdraw || withdraw.AmmunitionSlot != 1 {
		t.Fatalf("decoded ammunition withdrawal = %+v, error %v", withdraw, err)
	}
}

func TestDecodeRingAndRuneTransferMessages(t *testing.T) {
	ring, err := DecodeClientMessage([]byte(`{"type":"ring_deposit","sequence":12,"item_id":"ruby_ring","instance_id":"ring-1"}`))
	if err != nil || ring.Type != MessageRingDeposit || ring.ItemID != "ruby_ring" || ring.InstanceID != "ring-1" {
		t.Fatalf("decoded ring deposit = %+v, error %v", ring, err)
	}
	ring, err = DecodeClientMessage([]byte(`{"type":"ring_withdraw","sequence":13,"ring_slot":4}`))
	if err != nil || ring.Type != MessageRingWithdraw || ring.RingSlot != 4 {
		t.Fatalf("decoded ring withdrawal = %+v, error %v", ring, err)
	}
	runes, err := DecodeClientMessage([]byte(`{"type":"rune_deposit","sequence":14,"inventory_slot":2}`))
	if err != nil || runes.Type != MessageRuneDeposit || runes.InventorySlot != 2 {
		t.Fatalf("decoded rune deposit = %+v, error %v", runes, err)
	}
	runes, err = DecodeClientMessage([]byte(`{"type":"rune_withdraw","sequence":15,"rune_slot":5}`))
	if err != nil || runes.Type != MessageRuneWithdraw || runes.RuneSlot != 5 {
		t.Fatalf("decoded rune withdrawal = %+v, error %v", runes, err)
	}
}

func TestDecodeCombatStyleMessage(t *testing.T) {
	message, err := DecodeClientMessage([]byte(`{"type":"set_combat_style","sequence":6,"combat_style":"rapid"}`))
	if err != nil {
		t.Fatalf("DecodeClientMessage() returned an unexpected error: %v", err)
	}
	if message.Type != MessageCombatStyle || message.Sequence != 6 || message.CombatStyle != "rapid" {
		t.Fatalf("DecodeClientMessage() = %+v", message)
	}
}

func TestDecodeExperienceLockMessage(t *testing.T) {
	message, err := DecodeClientMessage([]byte(`{"type":"set_experience_lock","sequence":7,"active":true}`))
	if err != nil {
		t.Fatalf("DecodeClientMessage() returned an unexpected error: %v", err)
	}
	if message.Type != MessageExperienceLock || message.Sequence != 7 || !message.Active {
		t.Fatalf("DecodeClientMessage() = %+v", message)
	}
}

func TestDecodePrayerMessage(t *testing.T) {
	message, err := DecodeClientMessage([]byte(`{"type":"set_prayer","sequence":7,"prayer_id":"evasion","active":true}`))
	if err != nil {
		t.Fatalf("DecodeClientMessage() returned an unexpected error: %v", err)
	}
	if message.Type != MessageSetPrayer || message.Sequence != 7 || message.PrayerID != "evasion" || !message.Active {
		t.Fatalf("DecodeClientMessage() = %+v", message)
	}
}

func TestDecodeSpellMessage(t *testing.T) {
	message, err := DecodeClientMessage([]byte(`{"type":"set_spell","sequence":8,"spell_id":"wind_bolt"}`))
	if err != nil {
		t.Fatalf("DecodeClientMessage() returned an unexpected error: %v", err)
	}
	if message.Type != MessageSetSpell || message.Sequence != 8 || message.SpellID != "wind_bolt" {
		t.Fatalf("DecodeClientMessage() = %+v", message)
	}
}

func TestDecodeClearSpellMessage(t *testing.T) {
	message, err := DecodeClientMessage([]byte(`{"type":"set_spell","sequence":9,"spell_id":""}`))
	if err != nil {
		t.Fatalf("DecodeClientMessage() returned an unexpected error: %v", err)
	}
	if message.Type != MessageSetSpell || message.Sequence != 9 || message.SpellID != "" {
		t.Fatalf("DecodeClientMessage() = %+v", message)
	}
}

func TestDecodePickupItemMessage(t *testing.T) {
	message, err := DecodeClientMessage([]byte(`{"type":"pickup_item","sequence":5,"drop_id":"drop-1"}`))
	if err != nil {
		t.Fatalf("DecodeClientMessage() returned an unexpected error: %v", err)
	}
	if message.Type != MessagePickupItem || message.Sequence != 5 || message.DropID != "drop-1" {
		t.Fatalf("DecodeClientMessage() = %+v", message)
	}
}

func TestDecodeClientMessageRejectsInvalidMessages(t *testing.T) {
	tests := map[string]string{
		"malformed JSON":      `{"type":`,
		"unknown field":       `{"type":"move_to","sequence":1,"destination":{"x":1,"z":2},"extra":true}`,
		"unknown type":        `{"type":"dance","sequence":1,"destination":{"x":1,"z":2}}`,
		"missing sequence":    `{"type":"move_to","destination":{"x":1,"z":2}}`,
		"missing target":      `{"type":"move_to","sequence":1}`,
		"missing target ID":   `{"type":"engage_target","sequence":1}`,
		"missing item ID":     `{"type":"equip_item","sequence":1}`,
		"missing unequip ID":  `{"type":"unequip_item","sequence":1}`,
		"missing drop ID":     `{"type":"pickup_item","sequence":1}`,
		"missing style":       `{"type":"set_combat_style","sequence":1}`,
		"missing prayer ID":   `{"type":"set_prayer","sequence":1,"active":true}`,
		"missing debug skill": `{"type":"debug_set_skill","sequence":1,"level":10}`,
		"invalid debug level": `{"type":"debug_set_skill","sequence":1,"skill_id":"attack","level":101}`,
		"negative debug XP":   `{"type":"debug_set_skill","sequence":1,"skill_id":"attack","set_experience":true,"experience":-1}`,
		"multiple values":     `{"type":"move_to","sequence":1,"destination":{"x":1,"z":2}} {}`,
	}

	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeClientMessage([]byte(input)); err == nil {
				t.Fatal("DecodeClientMessage() error = nil")
			}
		})
	}
}
