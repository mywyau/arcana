package simulation

const maximumProductionBatchSize = 10

// productionBatch represents a bounded production order. A zero quantity is
// the client-facing "make available" request and resolves to at most ten lots.
type productionBatch struct {
	remaining int
}

func newProductionBatch(quantity int) productionBatch {
	if quantity == 0 || quantity > maximumProductionBatchSize {
		quantity = maximumProductionBatchSize
	}
	return productionBatch{remaining: quantity}
}

func (batch *productionBatch) completeStep() bool {
	if batch == nil || batch.remaining <= 0 {
		return true
	}
	if batch.remaining > 0 {
		batch.remaining--
	}
	return batch.remaining == 0
}

func usesProductionBatch(station craftingStationID) bool {
	return station == craftingStationCooking ||
		station == craftingStationSmelting ||
		station == craftingStationForging ||
		station == craftingStationRunecrafting
}
