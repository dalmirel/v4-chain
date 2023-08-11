package types

import (
	"math/big"

	"github.com/dydxprotocol/v4/lib"
	pricestypes "github.com/dydxprotocol/v4/x/prices/types"
)

// PriceToSubticks converts price value from Prices module into subticks.
// By construction of the Clob module:
// `sizeQuoteQuantums = subticks * baseQuantums * 10^quantumConversionExponent`
// Substituting `baseQuantums` with a full coin of the base currency:
//
// `usdcPrice * 10^(-quoteAtomicResolution) = subticks * 10^(-baseAtomicResolution) *
// 10^quantumConversionExponent` (A)
//
// By construction of Prices module:
//
// `usdcPrice = market.Price * 10^market.Exponent` (B)
//
// Combining equations (A) & (B), we get:
//
// `subticks = market.Price * 10^(market.Exponent - quantumConversionExponent +
// baseAtomicResolution - quoteAtomicResolution)`
func PriceToSubticks(
	market pricestypes.Market,
	clobPair ClobPair,
	baseAtomicResolution int32,
	quoteAtomicResolution int32,
) (
	ratSubticks *big.Rat,
) {
	exponent := int32(market.Exponent - clobPair.QuantumConversionExponent + baseAtomicResolution - quoteAtomicResolution)
	return lib.BigMulPow10(
		// TODO(DEC-1256): Use index price from the price daemon, instead of oracle price.
		new(big.Int).SetUint64(market.Price),
		exponent,
	)
}
