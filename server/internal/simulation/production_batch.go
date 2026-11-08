package simulation

// productionBatch represents either a fixed number of crafts or an unlimited
// batch that continues until ingredients or output space run out.
type productionBatch struct {
	limited   bool
	remaining int
}

func newProductionBatch(quantity int) productionBatch {
	return productionBatch{
		limited:   quantity > 0,
		remaining: quantity,
	}
}

func (batch *productionBatch) completeStep() bool {
	if batch == nil || !batch.limited {
		return false
	}
	if batch.remaining > 0 {
		batch.remaining--
	}
	return batch.remaining == 0
}

func usesProductionBatch(station craftingStationID) bool {
	return station == craftingStationCooking || station == craftingStationRunecrafting
}
