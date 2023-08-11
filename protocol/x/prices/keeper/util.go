package keeper

import (
	"errors"
	"github.com/dydxprotocol/v4/testutil/constants"
	"math/big"

	"github.com/dydxprotocol/v4/lib"
	"github.com/dydxprotocol/v4/x/prices/types"
)

// isAboveRequiredMinPriceChange returns true if the new price meets the required min price change
// for the market. Otherwise, returns false.
func isAboveRequiredMinPriceChange(market types.Market, newPrice uint64) bool {
	minChangeAmt := getMinPriceChangeAmountForMarket(market)
	return lib.AbsDiffUint64(market.Price, newPrice) >= minChangeAmt
}

// getMinPriceChangeAmountForMarket returns the amount of price change that is needed to trigger
// a price update in accordance with the min price change parts-per-million value.
func getMinPriceChangeAmountForMarket(market types.Market) uint64 {
	bigPrice := new(big.Int).SetUint64(market.Price)
	// There's no need to multiply this by the market's exponent, because `Price` comparisons are
	// done without the market's exponent.
	bigMinChangeAmt := lib.BigIntMulPpm(bigPrice, market.MinPriceChangePpm)

	if !bigMinChangeAmt.IsUint64() {
		// This means that the min change amount is greater than the max uint64. This can only
		// happen if the `MinPriceChangePpm` > 1,000,000 and there's a validation when
		// creating/modifying the `Market`.
		panic(errors.New("getMinPriceChangeAmountForMarket: min price change amount is greater than max uint64 value"))
	}

	return bigMinChangeAmt.Uint64()
}

// isTowardsIndexPrice returns true if the new price is between the current price and the index
// price, inclusive. Otherwise, it returns false.
func isTowardsIndexPrice(
	oldPrice uint64,
	newPrice uint64,
	indexPrice uint64,
) bool {
	return newPrice <= lib.Max(oldPrice, indexPrice) && newPrice >= lib.Min(oldPrice, indexPrice)
}

// isCrossingIndexPrice returns true if index price is between the current and the new price,
// noninclusive. Otherwise, returns false.
func isCrossingIndexPrice(
	oldPrice uint64,
	newPrice uint64,
	indexPrice uint64,
) bool {
	return indexPrice < lib.Max(oldPrice, newPrice) && indexPrice > lib.Min(oldPrice, newPrice)
}

// computeTickSizePpm calculates the tick_size of the currency at the current price, in ppm.
// We keep the tick_size multiplied by 10^6 to reduce divisions in our calculations and avoid rounding errors.
func computeTickSizePpm(oldPrice uint64, minPriceChangePpm uint32) *big.Int {
	// tick_size = oldPrice * minPriceChangePpm / 1_000_000 ==>
	// tick_size_ppm = oldPrice * minPriceChangePpm = tick_size * 1_000_000
	return new(big.Int).Mul(
		new(big.Int).SetUint64(oldPrice),
		new(big.Int).SetUint64(uint64(minPriceChangePpm)))
}

// priceDeltaIsWithinOneTick returns true iff the price delta is within one tick, given the tick_size in ppm.
func priceDeltaIsWithinOneTick(priceDelta *big.Int, tickSizePpm *big.Int) bool {
	// To compare if a price_delta > tick_size, let's multiply by 1_000_000 and compare against the
	// tick size in ppm
	priceDeltaPpm := new(big.Int).Mul(priceDelta, new(big.Int).SetUint64(constants.OneMillion))
	return priceDeltaPpm.Cmp(tickSizePpm) <= 0
}

// newPriceMeetsSqrtCondition calculates the price acceptance condition when the new price crosses the index
// price and the price change from the current price to the index price, or old_ticks, is > 1 tick.
//
// Ticks are computed at the currency's current price.
//
// Under these conditions, price changes are valid when new_ticks <= sqrt(old_ticks)
func newPriceMeetsSqrtCondition(oldDelta *big.Int, newDelta *big.Int, tickSizePpm *big.Int) bool {
	// In order to avoid division / sqrt, which is potentially lossy, use big.Ints and refactor:
	// given that new_ticks = new_delta / tick_size, old_ticks = old_delta / tick_size
	// new_ticks < sqrt(old_ticks)                                  ==> sub in old_ticks, new_ticks
	// new_delta / tick_size <= sqrt(old_delta / tick_size)         ==>
	// new_delta * new_delta / tick_size <= old_delta               ==>
	// new_delta * new_delta <= old_delta * tick_size               ==>
	// new_delta * new_delta * 1_000_000 <= old_delta * tickSizePpm
	newDeltaSquaredPpm := new(big.Int).Mul(newDelta, newDelta)
	newDeltaSquaredPpm.Mul(newDeltaSquaredPpm, new(big.Int).SetUint64(constants.OneMillion))
	oldDeltaTimesTickSizePpm := new(big.Int).Mul(oldDelta, tickSizePpm)
	return newDeltaSquaredPpm.Cmp(oldDeltaTimesTickSizePpm) <= 0
}

// maximumAllowedPriceDelta computes the maximum allowable value of new_delta under the conditions
// that the proposed price is crossing in the index price, and old_ticks > 1. This method uses potentially
// lossy arithmetic and is only for logging purposes.
func maximumAllowedPriceDelta(oldDelta *big.Int, tickSizePpm *big.Int) *big.Int {
	// Compute maximum allowable new_delta, or price difference between the index price
	// and the proposed price:
	// max_allowed = sqrt(old_delta * tick_size_ppm / 1_000_000)
	maxAllowed := new(big.Int).Mul(oldDelta, tickSizePpm)
	maxAllowed.Div(maxAllowed, new(big.Int).SetUint64(constants.OneMillion))
	maxAllowed.Sqrt(maxAllowed)
	return maxAllowed
}
