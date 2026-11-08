package simulation

import (
	"testing"
	"time"
)

func TestMerchantBuysWholeStackAndDestroysTwentyPercentEveryTwoMinutes(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", starterMerchant.Position, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	player := world.players["player-one"]
	if !player.inventory.add(ItemNormalLogs, 10) {
		t.Fatal("failed to add logs")
	}
	if err := world.QueueMerchantSale(MerchantSellCommand{
		PlayerID: "player-one", Sequence: 1, InventorySlot: 0,
	}); err != nil {
		t.Fatalf("QueueMerchantSale() returned an unexpected error: %v", err)
	}
	world.Step(50 * time.Millisecond)
	snapshot := mustPlayer(t, world, "player-one")
	if snapshot.Inventory.itemsQuantity(ItemNormalLogs) != 0 || snapshot.Inventory.itemsQuantity(ItemCoins) != 10 {
		t.Fatalf("inventory after sale = %+v", snapshot.Inventory)
	}
	if snapshot.LastMerchantSequence != 1 || len(world.merchantStock) != 1 || world.merchantStock[0].Item.Quantity != 10 {
		t.Fatalf("merchant state after sale = %+v, player = %+v", world.merchantStock, snapshot)
	}

	world.Step(merchantStockLifetime)
	if got := world.merchantStock[0].Item.Quantity; got != 8 {
		t.Fatalf("merchant stock after cleanup = %d, want 8", got)
	}
}

func TestMerchantDestroysSmallStacksCompletely(t *testing.T) {
	world := NewWorld()
	world.merchantStock = []merchantStockEntry{{
		Item: InventoryItem{ItemID: ItemBlankRune, Quantity: 4}, Remaining: merchantStockLifetime,
	}}
	world.Step(merchantStockLifetime)
	if len(world.merchantStock) != 0 {
		t.Fatalf("merchant stock after small-stack cleanup = %+v, want empty", world.merchantStock)
	}
}

func TestMerchantRejectsCoinsAndOutOfRangeSales(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	world.players["player-one"].inventory.add(ItemCoins, 10)
	if err := world.QueueMerchantSale(MerchantSellCommand{
		PlayerID: "player-one", Sequence: 1, InventorySlot: 0,
	}); err == nil {
		t.Fatal("QueueMerchantSale() error = nil outside merchant range")
	}
	world.players["player-one"].position = starterMerchant.Position
	if err := world.QueueMerchantSale(MerchantSellCommand{
		PlayerID: "player-one", Sequence: 1, InventorySlot: 0,
	}); err == nil {
		t.Fatal("QueueMerchantSale() error = nil for coins")
	}
}
