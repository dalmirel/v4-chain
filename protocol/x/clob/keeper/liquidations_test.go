package keeper_test

import (
	"math"
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dydxprotocol/v4/dtypes"
	"github.com/dydxprotocol/v4/indexer/indexer_manager"
	"github.com/dydxprotocol/v4/lib"
	"github.com/dydxprotocol/v4/mocks"
	big_testutil "github.com/dydxprotocol/v4/testutil/big"
	clobtest "github.com/dydxprotocol/v4/testutil/clob"
	"github.com/dydxprotocol/v4/testutil/constants"
	keepertest "github.com/dydxprotocol/v4/testutil/keeper"
	"github.com/dydxprotocol/v4/x/clob/memclob"
	"github.com/dydxprotocol/v4/x/clob/types"
	"github.com/dydxprotocol/v4/x/perpetuals"
	perptypes "github.com/dydxprotocol/v4/x/perpetuals/types"
	"github.com/dydxprotocol/v4/x/prices"
	satypes "github.com/dydxprotocol/v4/x/subaccounts/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPlacePerpetualLiquidation(t *testing.T) {
	tests := map[string]struct {
		// Perpetuals state.
		perpetuals []perptypes.Perpetual
		// Subaccount state.
		subaccounts []satypes.Subaccount
		// CLOB state.
		clobs          []types.ClobPair
		existingOrders []types.Order

		// Parameters.
		order types.LiquidationOrder

		// Expectations.
		expectedPlacedOrders  []*types.MsgPlaceOrder
		expectedMatchedOrders []*types.ClobMatch
	}{
		`Can place a liquidation that doesn't match any maker orders`: {
			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_SmallMarginRequirement,
			},
			subaccounts: []satypes.Subaccount{
				constants.Dave_Num0_1BTC_Long_46000USD_Short,
			},
			clobs: []types.ClobPair{constants.ClobPair_Btc},

			order: constants.LiquidationOrder_Dave_Num0_Clob0_Sell1BTC_Price50000,

			expectedPlacedOrders:  []*types.MsgPlaceOrder{},
			expectedMatchedOrders: []*types.ClobMatch{},
		},
		`Can place a liquidation that matches maker orders`: {
			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_SmallMarginRequirement,
			},
			subaccounts: []satypes.Subaccount{
				constants.Carl_Num0_1BTC_Short,
				constants.Dave_Num0_1BTC_Long_46000USD_Short,
			},
			clobs: []types.ClobPair{constants.ClobPair_Btc},
			existingOrders: []types.Order{
				constants.Order_Carl_Num0_Id0_Clob0_Buy1BTC_Price50000_GTB10,
			},

			order: constants.LiquidationOrder_Dave_Num0_Clob0_Sell1BTC_Price50000,

			expectedPlacedOrders: []*types.MsgPlaceOrder{
				{
					Order: constants.Order_Carl_Num0_Id0_Clob0_Buy1BTC_Price50000_GTB10,
				},
			},
			expectedMatchedOrders: []*types.ClobMatch{
				types.NewClobMatchFromMatchPerpetualLiquidation(
					&types.MatchPerpetualLiquidation{
						ClobPairId:  constants.ClobPair_Btc.Id,
						IsBuy:       false,
						TotalSize:   100_000_000,
						Liquidated:  constants.Dave_Num0,
						PerpetualId: constants.ClobPair_Btc.GetPerpetualClobMetadata().PerpetualId,
						Fills: []types.MakerFill{
							{
								MakerOrderId: types.OrderId{},
								FillAmount:   100_000_000,
							},
						},
					},
				),
			},
		},
		`Can place a liquidation that matches maker orders and removes undercollateralized ones`: {
			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_SmallMarginRequirement,
			},
			subaccounts: []satypes.Subaccount{
				constants.Carl_Num0_1BTC_Short,
				constants.Dave_Num0_1BTC_Long_46000USD_Short,
			},
			clobs: []types.ClobPair{constants.ClobPair_Btc},
			existingOrders: []types.Order{
				// Note this order will be removed when matching.
				constants.Order_Carl_Num1_Id0_Clob0_Buy1BTC_Price50000_GTB10,
				constants.Order_Carl_Num0_Id0_Clob0_Buy1BTC_Price50000_GTB10,
			},

			order: constants.LiquidationOrder_Dave_Num0_Clob0_Sell1BTC_Price50000,

			expectedPlacedOrders: []*types.MsgPlaceOrder{
				{
					Order: constants.Order_Carl_Num0_Id0_Clob0_Buy1BTC_Price50000_GTB10,
				},
			},
			expectedMatchedOrders: []*types.ClobMatch{
				types.NewClobMatchFromMatchPerpetualLiquidation(
					&types.MatchPerpetualLiquidation{
						ClobPairId:  constants.ClobPair_Btc.Id,
						IsBuy:       false,
						TotalSize:   100_000_000,
						Liquidated:  constants.Dave_Num0,
						PerpetualId: constants.ClobPair_Btc.GetPerpetualClobMetadata().PerpetualId,
						Fills: []types.MakerFill{
							{
								MakerOrderId: types.OrderId{},
								FillAmount:   100_000_000,
							},
						},
					},
				),
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup keeper state.
			memClob := memclob.NewMemClobPriceTimePriority(false)
			mockBankKeeper := &mocks.BankKeeper{}
			mockBankKeeper.On(
				"SendCoinsFromModuleToModule",
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(nil)

			ctx,
				clobKeeper,
				pricesKeeper,
				assetsKeeper,
				perpetualsKeeper,
				subaccountsKeeper,
				_,
				_ := keepertest.ClobKeepers(t, memClob, mockBankKeeper, indexer_manager.NewIndexerEventManagerNoop())

			ctx = ctx.WithIsCheckTx(true)
			// Create the default markets.
			keepertest.CreateTestMarketsAndExchangeFeeds(t, ctx, pricesKeeper)

			// Create liquidity tiers.
			keepertest.CreateTestLiquidityTiers(t, ctx, perpetualsKeeper)

			// Set up USDC asset in assets module.
			err := keepertest.CreateUsdcAsset(ctx, assetsKeeper)
			require.NoError(t, err)

			// Create all perpetuals.
			for _, p := range tc.perpetuals {
				_, err := perpetualsKeeper.CreatePerpetual(
					ctx,
					p.Ticker,
					p.MarketId,
					p.AtomicResolution,
					p.DefaultFundingPpm,
					p.LiquidityTier,
				)
				require.NoError(t, err)
			}

			// Create all subaccounts.
			for _, subaccount := range tc.subaccounts {
				subaccountsKeeper.SetSubaccount(ctx, subaccount)
			}

			// Create all CLOBs.
			for _, clobPair := range tc.clobs {
				_, err = clobKeeper.CreatePerpetualClobPair(
					ctx,
					clobtest.MustPerpetualId(clobPair),
					satypes.BaseQuantums(clobPair.StepBaseQuantums),
					satypes.BaseQuantums(clobPair.MinOrderBaseQuantums),
					clobPair.QuantumConversionExponent,
					clobPair.SubticksPerTick,
					clobPair.Status,
					clobPair.MakerFeePpm,
					clobPair.TakerFeePpm,
				)
				require.NoError(t, err)
			}

			// Initialize the liquidations config.
			require.NoError(
				t,
				clobKeeper.InitializeLiquidationsConfig(ctx, types.LiquidationsConfig_Default),
			)

			// Create all existing orders.
			for _, order := range tc.existingOrders {
				_, _, err := clobKeeper.CheckTxPlaceOrder(ctx, &types.MsgPlaceOrder{Order: order})
				require.NoError(t, err)
			}

			// Run the test.
			_, _, err = clobKeeper.CheckTxPlacePerpetualLiquidation(ctx, tc.order)
			require.NoError(t, err)

			// Verify test expectations.
			// TODO(DEC-1979): Refactor these tests to support the operations queue refactor.
			// placedOrders, matchedOrders := memClob.GetPendingFills(ctx)

			// require.Equal(t, tc.expectedPlacedOrders, placedOrders, "Placed orders lists are not equal")
			// require.Equal(t, tc.expectedMatchedOrders, matchedOrders, "Matched orders lists are not equal")
		})
	}
}

func TestPlacePerpetualLiquidation_PreexistingLiquidation(t *testing.T) {
	tests := map[string]struct {
		// State.
		subaccounts         []satypes.Subaccount
		setupMockBankKeeper func(m *mocks.BankKeeper)

		// Parameters.
		liquidationConfig     types.LiquidationsConfig
		placedMatchableOrders []types.MatchableOrder
		order                 types.LiquidationOrder

		// Expectations.
		panics                            bool
		expectedError                     error
		expectedFilledSize                satypes.BaseQuantums
		expectedOrderStatus               types.OrderStatus
		expectedPlacedOrders              []*types.MsgPlaceOrder
		expectedMatchedOrders             []*types.ClobMatch
		expectedSubaccountLiquidationInfo map[satypes.SubaccountId]types.SubaccountLiquidationInfo
	}{
		`PlacePerpetualLiquidation succeeds with pre-existing liquidations in the block`: {
			subaccounts: []satypes.Subaccount{
				{
					Id: &constants.Carl_Num0,
					AssetPositions: []*satypes.AssetPosition{
						{
							AssetId:  0,
							Quantums: dtypes.NewInt(54_999_000_000), // $54,999
						},
					},
					PerpetualPositions: []*satypes.PerpetualPosition{
						{
							PerpetualId: 0,
							Quantums:    dtypes.NewInt(-100_000_000), // -1 BTC
						},
						{
							PerpetualId: 1,
							Quantums:    dtypes.NewInt(-1_000_000_000), // -1 ETH
						},
					},
				},
				constants.Dave_Num0_1BTC_Long,
			},

			liquidationConfig: constants.LiquidationsConfig_No_Limit,
			placedMatchableOrders: []types.MatchableOrder{
				&constants.Order_Dave_Num0_Id3_Clob1_Sell1ETH_Price3000,
				&constants.LiquidationOrder_Carl_Num0_Clob1_Buy1ETH_Price3000,
				&constants.Order_Dave_Num0_Id0_Clob0_Sell1BTC_Price50000_GTB10,
			},
			order: constants.LiquidationOrder_Carl_Num0_Clob0_Buy1BTC_Price50000,

			expectedOrderStatus: types.Success,
			expectedPlacedOrders: []*types.MsgPlaceOrder{
				{
					Order: constants.Order_Dave_Num0_Id3_Clob1_Sell1ETH_Price3000,
				},
				{
					Order: constants.Order_Dave_Num0_Id0_Clob0_Sell1BTC_Price50000_GTB10,
				},
			},
			expectedMatchedOrders: []*types.ClobMatch{
				types.NewClobMatchFromMatchPerpetualLiquidation(
					&types.MatchPerpetualLiquidation{
						ClobPairId:  constants.ClobPair_Eth.Id,
						IsBuy:       true,
						TotalSize:   1_000_000_000,
						Liquidated:  constants.Carl_Num0,
						PerpetualId: constants.ClobPair_Eth.GetPerpetualClobMetadata().PerpetualId,
						Fills: []types.MakerFill{
							{
								MakerOrderId: types.OrderId{},
								FillAmount:   1_000_000_000,
							},
						},
					},
				),
				types.NewClobMatchFromMatchPerpetualLiquidation(
					&types.MatchPerpetualLiquidation{
						ClobPairId:  constants.ClobPair_Btc.Id,
						IsBuy:       true,
						TotalSize:   100_000_000,
						Liquidated:  constants.Carl_Num0,
						PerpetualId: constants.ClobPair_Btc.GetPerpetualClobMetadata().PerpetualId,
						Fills: []types.MakerFill{
							{
								MakerOrderId: types.OrderId{},
								FillAmount:   100_000_000,
							},
						},
					},
				),
			},
			expectedSubaccountLiquidationInfo: map[satypes.SubaccountId]types.SubaccountLiquidationInfo{
				constants.Carl_Num0: {
					PerpetualsLiquidated:  []uint32{1, 0},
					NotionalLiquidated:    53_000_000_000, // $53,000
					QuantumsInsuranceLost: 0,
				},
				constants.Dave_Num0: {},
			},
		},
		`PlacePerpetualLiquidation considers pre-existing liquidations and stops before exceeding
		max notional liquidated per block`: {
			subaccounts: []satypes.Subaccount{
				{
					Id: &constants.Carl_Num0,
					AssetPositions: []*satypes.AssetPosition{
						{
							AssetId:  0,
							Quantums: dtypes.NewInt(54_999_000_000), // $54,999
						},
					},
					PerpetualPositions: []*satypes.PerpetualPosition{
						{
							PerpetualId: 0,
							Quantums:    dtypes.NewInt(-100_000_000), // -1 BTC
						},
						{
							PerpetualId: 1,
							Quantums:    dtypes.NewInt(-1_000_000_000), // -1 ETH
						},
					},
				},
				constants.Dave_Num0_1BTC_Long,
			},

			liquidationConfig: types.LiquidationsConfig{
				MaxLiquidationFeePpm: 5_000,
				FillablePriceConfig:  constants.FillablePriceConfig_Default,
				PositionBlockLimits:  constants.PositionBlockLimits_No_Limit,
				SubaccountBlockLimits: types.SubaccountBlockLimits{
					MaxNotionalLiquidated:    10_000_000_000, // $10,000
					MaxQuantumsInsuranceLost: math.MaxUint64,
				},
			},
			placedMatchableOrders: []types.MatchableOrder{
				&constants.Order_Dave_Num0_Id3_Clob1_Sell1ETH_Price3000,
				&constants.LiquidationOrder_Carl_Num0_Clob1_Buy1ETH_Price3000,
				&constants.Order_Dave_Num0_Id0_Clob0_Sell1BTC_Price50000_GTB10,
			},
			order: constants.LiquidationOrder_Carl_Num0_Clob0_Buy1BTC_Price50000,

			// Only matches one order since matching both orders would exceed `MaxNotionalLiquidated`.
			expectedOrderStatus: types.Success,
			expectedPlacedOrders: []*types.MsgPlaceOrder{
				{
					Order: constants.Order_Dave_Num0_Id3_Clob1_Sell1ETH_Price3000,
				},
			},
			expectedMatchedOrders: []*types.ClobMatch{
				types.NewClobMatchFromMatchPerpetualLiquidation(
					&types.MatchPerpetualLiquidation{
						ClobPairId:  constants.ClobPair_Eth.Id,
						IsBuy:       true,
						TotalSize:   1_000_000_000,
						Liquidated:  constants.Carl_Num0,
						PerpetualId: constants.ClobPair_Eth.GetPerpetualClobMetadata().PerpetualId,
						Fills: []types.MakerFill{
							{
								MakerOrderId: types.OrderId{},
								FillAmount:   1_000_000_000,
							},
						},
					},
				),
			},
			expectedSubaccountLiquidationInfo: map[satypes.SubaccountId]types.SubaccountLiquidationInfo{
				constants.Carl_Num0: {
					PerpetualsLiquidated:  []uint32{1, 0},
					NotionalLiquidated:    3_000_000_000, // $3,000
					QuantumsInsuranceLost: 0,
				},
				constants.Dave_Num0: {},
			},
		},
		`PlacePerpetualLiquidation matches some order and stops before exceeding max notional liquidated per block`: {
			subaccounts: []satypes.Subaccount{
				constants.Carl_Num0_1BTC_Short_54999USD,
				constants.Dave_Num0_1BTC_Long,
			},

			liquidationConfig: types.LiquidationsConfig{
				MaxLiquidationFeePpm: 5_000,
				FillablePriceConfig:  constants.FillablePriceConfig_Default,
				PositionBlockLimits:  constants.PositionBlockLimits_No_Limit,
				SubaccountBlockLimits: types.SubaccountBlockLimits{
					MaxNotionalLiquidated:    20_000_000_000, // $20,000
					MaxQuantumsInsuranceLost: math.MaxUint64,
				},
			},
			placedMatchableOrders: []types.MatchableOrder{
				&constants.Order_Dave_Num0_Id1_Clob0_Sell025BTC_Price50000_GTB11,
				&constants.Order_Dave_Num0_Id2_Clob0_Sell025BTC_Price50000_GTB12,
			},
			order: constants.LiquidationOrder_Carl_Num0_Clob0_Buy1BTC_Price50000,

			// Only matches one order since matching both orders would exceed `MaxNotionalLiquidated`.
			expectedOrderStatus: types.Success,
			expectedPlacedOrders: []*types.MsgPlaceOrder{
				{
					Order: constants.Order_Dave_Num0_Id1_Clob0_Sell025BTC_Price50000_GTB11,
				},
			},
			expectedMatchedOrders: []*types.ClobMatch{
				types.NewClobMatchFromMatchPerpetualLiquidation(
					&types.MatchPerpetualLiquidation{
						ClobPairId:  constants.ClobPair_Btc.Id,
						IsBuy:       true,
						TotalSize:   100_000_000,
						Liquidated:  constants.Carl_Num0,
						PerpetualId: constants.ClobPair_Btc.GetPerpetualClobMetadata().PerpetualId,
						Fills: []types.MakerFill{
							{
								MakerOrderId: types.OrderId{},
								FillAmount:   25_000_000,
							},
						},
					},
				),
			},
			expectedSubaccountLiquidationInfo: map[satypes.SubaccountId]types.SubaccountLiquidationInfo{
				constants.Carl_Num0: {
					PerpetualsLiquidated:  []uint32{0},
					NotionalLiquidated:    12_500_000_000, // $12,500
					QuantumsInsuranceLost: 0,
				},
				constants.Dave_Num0: {},
			},
		},
		`PlacePerpetualLiquidation considers pre-existing liquidations and stops before exceeding
		max insurance fund lost per block`: {
			subaccounts: []satypes.Subaccount{
				{
					Id: &constants.Carl_Num0,
					AssetPositions: []*satypes.AssetPosition{
						{
							AssetId:  0,
							Quantums: dtypes.NewInt(53_000_000_000), // $53,000
						},
					},
					PerpetualPositions: []*satypes.PerpetualPosition{
						{
							PerpetualId: 0,
							Quantums:    dtypes.NewInt(-100_000_000), // -1 BTC
						},
						{
							PerpetualId: 1,
							Quantums:    dtypes.NewInt(-1_000_000_000), // -1 ETH
						},
					},
				},
				constants.Dave_Num0_1BTC_Long,
			},

			liquidationConfig: types.LiquidationsConfig{
				MaxLiquidationFeePpm: 5_000,
				FillablePriceConfig:  constants.FillablePriceConfig_Default,
				PositionBlockLimits:  constants.PositionBlockLimits_No_Limit,
				SubaccountBlockLimits: types.SubaccountBlockLimits{
					MaxNotionalLiquidated:    math.MaxUint64,
					MaxQuantumsInsuranceLost: 50_000_000, // $50
				},
			},
			placedMatchableOrders: []types.MatchableOrder{
				&constants.Order_Dave_Num0_Id4_Clob1_Sell1ETH_Price3030,
				&constants.LiquidationOrder_Carl_Num0_Clob1_Buy1ETH_Price3030,
				&constants.Order_Dave_Num0_Id0_Clob0_Sell1BTC_Price50500_GTB10,
			},
			order: constants.LiquidationOrder_Carl_Num0_Clob0_Buy1BTC_Price50500,

			// Only matches one order since matching both orders would exceed `MaxQuantumsInsuranceLost`.
			expectedOrderStatus: types.Success,
			expectedPlacedOrders: []*types.MsgPlaceOrder{
				{
					Order: constants.Order_Dave_Num0_Id4_Clob1_Sell1ETH_Price3030,
				},
			},
			expectedMatchedOrders: []*types.ClobMatch{
				types.NewClobMatchFromMatchPerpetualLiquidation(
					&types.MatchPerpetualLiquidation{
						ClobPairId:  constants.ClobPair_Eth.Id,
						IsBuy:       true,
						TotalSize:   1_000_000_000,
						Liquidated:  constants.Carl_Num0,
						PerpetualId: constants.ClobPair_Eth.GetPerpetualClobMetadata().PerpetualId,
						Fills: []types.MakerFill{
							{
								MakerOrderId: types.OrderId{},
								FillAmount:   1_000_000_000,
							},
						},
					},
				),
			},
			expectedSubaccountLiquidationInfo: map[satypes.SubaccountId]types.SubaccountLiquidationInfo{
				constants.Carl_Num0: {
					PerpetualsLiquidated:  []uint32{1, 0},
					NotionalLiquidated:    3_030_000_000, // $3,030
					QuantumsInsuranceLost: 30_000_000,    // $30
				},
				constants.Dave_Num0: {},
			},
		},
		`PlacePerpetualLiquidation matches some order and stops before exceeding max insurance lost per block`: {
			subaccounts: []satypes.Subaccount{
				constants.Carl_Num0_1BTC_Short_50499USD,
				constants.Dave_Num0_1BTC_Long,
			},

			liquidationConfig: types.LiquidationsConfig{
				MaxLiquidationFeePpm: 5_000,
				FillablePriceConfig:  constants.FillablePriceConfig_Default,
				PositionBlockLimits:  constants.PositionBlockLimits_No_Limit,
				SubaccountBlockLimits: types.SubaccountBlockLimits{
					MaxNotionalLiquidated:    math.MaxUint64,
					MaxQuantumsInsuranceLost: 500_000, // $0.5
				},
			},
			placedMatchableOrders: []types.MatchableOrder{
				&constants.Order_Dave_Num0_Id2_Clob0_Sell025BTC_Price50500_GTB12,
				&constants.Order_Dave_Num0_Id0_Clob0_Sell1BTC_Price50500_GTB10,
			},
			// Overall insurance lost when liquidating at $50,500 is $1.
			order: constants.LiquidationOrder_Carl_Num0_Clob0_Buy1BTC_Price50500,

			// Only matches one order since matching both orders would exceed `MaxQuantumsInsuranceLost`.
			expectedOrderStatus: types.Success,
			expectedPlacedOrders: []*types.MsgPlaceOrder{
				{
					Order: constants.Order_Dave_Num0_Id2_Clob0_Sell025BTC_Price50500_GTB12,
				},
			},
			expectedMatchedOrders: []*types.ClobMatch{
				types.NewClobMatchFromMatchPerpetualLiquidation(
					&types.MatchPerpetualLiquidation{
						ClobPairId:  constants.ClobPair_Btc.Id,
						IsBuy:       true,
						TotalSize:   100_000_000,
						Liquidated:  constants.Carl_Num0,
						PerpetualId: constants.ClobPair_Btc.GetPerpetualClobMetadata().PerpetualId,
						Fills: []types.MakerFill{
							{
								MakerOrderId: types.OrderId{},
								FillAmount:   25_000_000,
							},
						},
					},
				),
			},
			expectedSubaccountLiquidationInfo: map[satypes.SubaccountId]types.SubaccountLiquidationInfo{
				constants.Carl_Num0: {
					PerpetualsLiquidated:  []uint32{0},
					NotionalLiquidated:    12_625_000_000, // $12,625
					QuantumsInsuranceLost: 250_000,
				},
				constants.Dave_Num0: {},
			},
		},
		`Liquidation buy order does not generate a match when deleveraging is required`: {
			subaccounts: []satypes.Subaccount{
				constants.Carl_Num0_1BTC_Short_50499USD,
				constants.Dave_Num0_1BTC_Long,
			},
			setupMockBankKeeper: func(bk *mocks.BankKeeper) {
				bk.On(
					"SendCoinsFromModuleToModule",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
				bk.On(
					"GetBalance",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(sdk.NewCoin("USDC", sdk.NewIntFromUint64(0))) // Insurance fund is empty.
			},

			liquidationConfig: constants.LiquidationsConfig_No_Limit, // `MaxInsuranceFundQuantumsForDeleveraging` is zero.
			placedMatchableOrders: []types.MatchableOrder{
				&constants.Order_Dave_Num0_Id0_Clob0_Sell1BTC_Price50500_GTB10,
			},
			order: constants.LiquidationOrder_Carl_Num0_Clob0_Buy1BTC_Price50500, // Expected insurance fund delta is $-1.

			// Does not generate a match since insurance fund does not have enough to cover the losses.
			expectedOrderStatus:   types.LiquidationRequiresDeleveraging,
			expectedPlacedOrders:  []*types.MsgPlaceOrder{},
			expectedMatchedOrders: []*types.ClobMatch{},
			expectedSubaccountLiquidationInfo: map[satypes.SubaccountId]types.SubaccountLiquidationInfo{
				constants.Carl_Num0: {
					PerpetualsLiquidated:  []uint32{0},
					NotionalLiquidated:    0,
					QuantumsInsuranceLost: 0,
				},
				constants.Dave_Num0: {},
			},
		},
		`Liquidation sell order does not generate a match when deleveraging is required`: {
			subaccounts: []satypes.Subaccount{
				constants.Carl_Num0_1BTC_Short,
				constants.Dave_Num0_1BTC_Long_49501USD_Short,
			},
			setupMockBankKeeper: func(bk *mocks.BankKeeper) {
				bk.On(
					"SendCoinsFromModuleToModule",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
				bk.On(
					"GetBalance",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(sdk.NewCoin("USDC", sdk.NewIntFromUint64(0))) // Insurance fund is empty.
			},

			liquidationConfig: constants.LiquidationsConfig_No_Limit, // `MaxInsuranceFundQuantumsForDeleveraging` is zero.
			placedMatchableOrders: []types.MatchableOrder{
				&constants.Order_Carl_Num0_Id0_Clob0_Buy1BTC_Price49500_GTB10,
			},
			order: constants.LiquidationOrder_Dave_Num0_Clob0_Sell1BTC_Price49500, // Expected insurance fund delta is $-1.

			// Does not generate a match since insurance fund does not have enough to cover the losses.
			expectedOrderStatus:   types.LiquidationRequiresDeleveraging,
			expectedPlacedOrders:  []*types.MsgPlaceOrder{},
			expectedMatchedOrders: []*types.ClobMatch{},
			expectedSubaccountLiquidationInfo: map[satypes.SubaccountId]types.SubaccountLiquidationInfo{
				constants.Carl_Num0: {},
				constants.Dave_Num0: {
					PerpetualsLiquidated:  []uint32{0},
					NotionalLiquidated:    0,
					QuantumsInsuranceLost: 0,
				},
			},
		},
		`Liquidation buy order matches with some orders and stops when deleveraging is required`: {
			subaccounts: []satypes.Subaccount{
				constants.Carl_Num0_1BTC_Short_50499USD,
				constants.Dave_Num0_1BTC_Long,
			},
			setupMockBankKeeper: func(bk *mocks.BankKeeper) {
				bk.On(
					"SendCoinsFromModuleToModule",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
				bk.On(
					"GetBalance",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					// Insurance fund has $1 initially.
					sdk.NewCoin("USDC", sdk.NewIntFromUint64(1_000_000)),
				).Once()
				bk.On(
					"GetBalance",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					// Insurance fund has $0.75 after covering the loss of the first match.
					sdk.NewCoin("USDC", sdk.NewIntFromUint64(750_000)),
				).Once()
			},

			liquidationConfig: types.LiquidationsConfig{
				MaxInsuranceFundQuantumsForDeleveraging: 750_001,
				MaxLiquidationFeePpm:                    5_000,
				FillablePriceConfig:                     constants.FillablePriceConfig_Default,
				PositionBlockLimits:                     constants.PositionBlockLimits_No_Limit,
				SubaccountBlockLimits:                   constants.SubaccountBlockLimits_No_Limit,
			},
			placedMatchableOrders: []types.MatchableOrder{
				&constants.Order_Dave_Num0_Id2_Clob0_Sell025BTC_Price50500_GTB12,
				&constants.Order_Dave_Num0_Id0_Clob0_Sell1BTC_Price50500_GTB10,
			},
			// Overall insurance fund delta when liquidating at $50,500 is -$1.
			order: constants.LiquidationOrder_Carl_Num0_Clob0_Buy1BTC_Price50500,

			// Matches the first order since insurance fund balance is above `MaxInsuranceFundQuantumsForDeleveraging`
			// and has enough to cover the losses (-$0.25).
			// Does not match the second order since insurance fund delta is -$0.75 and insurance fund balance
			// is $0.75 which is lower than `MaxInsuranceFundQuantumsForDeleveraging`,
			// and therefore, deleveraging is required.
			expectedOrderStatus: types.LiquidationRequiresDeleveraging,
			expectedPlacedOrders: []*types.MsgPlaceOrder{
				{
					Order: constants.Order_Dave_Num0_Id2_Clob0_Sell025BTC_Price50500_GTB12,
				},
			},
			expectedMatchedOrders: []*types.ClobMatch{
				types.NewClobMatchFromMatchPerpetualLiquidation(
					&types.MatchPerpetualLiquidation{
						ClobPairId:  constants.ClobPair_Btc.Id,
						IsBuy:       true,
						TotalSize:   100_000_000,
						Liquidated:  constants.Carl_Num0,
						PerpetualId: constants.ClobPair_Btc.GetPerpetualClobMetadata().PerpetualId,
						Fills: []types.MakerFill{
							{
								MakerOrderId: types.OrderId{},
								FillAmount:   25_000_000,
							},
						},
					},
				),
			},
			expectedSubaccountLiquidationInfo: map[satypes.SubaccountId]types.SubaccountLiquidationInfo{
				constants.Carl_Num0: {
					PerpetualsLiquidated:  []uint32{0},
					NotionalLiquidated:    12_625_000_000, // $12,625
					QuantumsInsuranceLost: 250_000,
				},
				constants.Dave_Num0: {},
			},
		},
		`Liquidation sell order matches with some orders and stops when deleveraging is required`: {
			subaccounts: []satypes.Subaccount{
				constants.Carl_Num0_1BTC_Short,
				constants.Dave_Num0_1BTC_Long_49501USD_Short,
			},
			setupMockBankKeeper: func(bk *mocks.BankKeeper) {
				bk.On(
					"SendCoinsFromModuleToModule",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
				bk.On(
					"GetBalance",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					// Insurance fund has $1 initially.
					sdk.NewCoin("USDC", sdk.NewIntFromUint64(1_000_000)),
				).Once()
				bk.On(
					"GetBalance",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					// Insurance fund has $0.75 after covering the loss of the first match.
					sdk.NewCoin("USDC", sdk.NewIntFromUint64(750_000)),
				).Once()
			},

			liquidationConfig: types.LiquidationsConfig{
				MaxInsuranceFundQuantumsForDeleveraging: 750_001,
				MaxLiquidationFeePpm:                    5_000,
				FillablePriceConfig:                     constants.FillablePriceConfig_Default,
				PositionBlockLimits:                     constants.PositionBlockLimits_No_Limit,
				SubaccountBlockLimits:                   constants.SubaccountBlockLimits_No_Limit,
			},
			placedMatchableOrders: []types.MatchableOrder{
				&constants.Order_Carl_Num0_Id3_Clob0_Buy025BTC_Price49500,
				&constants.Order_Carl_Num0_Id0_Clob0_Buy1BTC_Price49500_GTB10,
			},
			// Overall insurance fund delta when liquidating at $50,500 is -$1.
			order: constants.LiquidationOrder_Dave_Num0_Clob0_Sell1BTC_Price49500,

			// Matches the first order since insurance fund balance is above `MaxInsuranceFundQuantumsForDeleveraging`
			// and has enough to cover the losses (-$0.25).
			// Does not match the second order since insurance fund delta is -$0.75 and insurance fund balance
			// is $0.75 which is lower than `MaxInsuranceFundQuantumsForDeleveraging`,
			// and therefore, deleveraging is required.
			expectedOrderStatus: types.LiquidationRequiresDeleveraging,
			expectedPlacedOrders: []*types.MsgPlaceOrder{
				{
					Order: constants.Order_Carl_Num0_Id3_Clob0_Buy025BTC_Price49500,
				},
			},
			expectedMatchedOrders: []*types.ClobMatch{
				types.NewClobMatchFromMatchPerpetualLiquidation(
					&types.MatchPerpetualLiquidation{
						ClobPairId:  constants.ClobPair_Btc.Id,
						IsBuy:       false,
						TotalSize:   100_000_000,
						Liquidated:  constants.Dave_Num0,
						PerpetualId: constants.ClobPair_Btc.GetPerpetualClobMetadata().PerpetualId,
						Fills: []types.MakerFill{
							{
								MakerOrderId: types.OrderId{},
								FillAmount:   25_000_000,
							},
						},
					},
				),
			},
			expectedSubaccountLiquidationInfo: map[satypes.SubaccountId]types.SubaccountLiquidationInfo{
				constants.Carl_Num0: {},
				constants.Dave_Num0: {
					PerpetualsLiquidated:  []uint32{0},
					NotionalLiquidated:    12_375_000_000, // $12,375
					QuantumsInsuranceLost: 250_000,
				},
			},
		},
		`PlacePerpetualLiquidation panics when trying to liquidate the same perpeptual in a block`: {
			subaccounts: []satypes.Subaccount{
				{
					Id: &constants.Carl_Num0,
					AssetPositions: []*satypes.AssetPosition{
						{
							AssetId:  0,
							Quantums: dtypes.NewInt(54_999_000_000), // $54,999
						},
					},
					PerpetualPositions: []*satypes.PerpetualPosition{
						{
							PerpetualId: 0,
							Quantums:    dtypes.NewInt(-100_000_000), // -1 BTC
						},
						{
							PerpetualId: 1,
							Quantums:    dtypes.NewInt(-2_000_000_000), // -2 ETH
						},
					},
				},
				constants.Dave_Num0_1BTC_Long,
			},

			liquidationConfig: constants.LiquidationsConfig_No_Limit,
			placedMatchableOrders: []types.MatchableOrder{
				&constants.Order_Dave_Num0_Id3_Clob1_Sell1ETH_Price3000,
				&constants.LiquidationOrder_Carl_Num0_Clob1_Buy1ETH_Price3000,
				&constants.Order_Dave_Num0_Id4_Clob1_Sell1ETH_Price3000,
			},
			order: constants.LiquidationOrder_Carl_Num0_Clob1_Buy1ETH_Price3000,

			expectedError: sdkerrors.Wrapf(
				types.ErrSubaccountHasLiquidatedPerpetual,
				"Subaccount %v and perpetual %v have already been liquidated within the last block",
				constants.Carl_Num0,
				1,
			),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup memclob state and test expectations.
			memclob := memclob.NewMemClobPriceTimePriority(false)

			bankKeeper := &mocks.BankKeeper{}
			if tc.setupMockBankKeeper != nil {
				tc.setupMockBankKeeper(bankKeeper)
			} else {
				bankKeeper.On(
					"SendCoinsFromModuleToModule",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
				bankKeeper.On(
					"GetBalance",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(sdk.NewCoin("USDC", sdk.NewIntFromUint64(math.MaxUint64)))
			}

			mockIndexerEventManager := &mocks.IndexerEventManager{}
			mockIndexerEventManager.On("Enabled").Return(false)
			ctx,
				keeper,
				pricesKeeper,
				assetsKeeper,
				perpKeeper,
				saKeeper,
				_,
				_ := keepertest.ClobKeepers(t, memclob, bankKeeper, mockIndexerEventManager)

			ctx = ctx.WithIsCheckTx(true)

			keepertest.CreateTestMarketsAndExchangeFeeds(t, ctx, pricesKeeper)

			// Create liquidity tiers.
			keepertest.CreateTestLiquidityTiers(t, ctx, perpKeeper)

			// Set up USDC asset in assets module.
			err := keepertest.CreateUsdcAsset(ctx, assetsKeeper)
			require.NoError(t, err)

			for _, perpetual := range []perptypes.Perpetual{
				constants.BtcUsd_100PercentMarginRequirement,
				constants.EthUsd_100PercentMarginRequirement,
			} {
				_, err = perpKeeper.CreatePerpetual(
					ctx,
					perpetual.Ticker,
					perpetual.MarketId,
					perpetual.AtomicResolution,
					perpetual.DefaultFundingPpm,
					perpetual.LiquidityTier,
				)
				require.NoError(t, err)
			}

			for _, s := range tc.subaccounts {
				saKeeper.SetSubaccount(ctx, s)
			}

			_, err = keeper.CreatePerpetualClobPair(
				ctx,
				clobtest.MustPerpetualId(constants.ClobPair_Btc),
				satypes.BaseQuantums(constants.ClobPair_Btc.StepBaseQuantums),
				satypes.BaseQuantums(constants.ClobPair_Btc.MinOrderBaseQuantums),
				constants.ClobPair_Btc.QuantumConversionExponent,
				constants.ClobPair_Btc.SubticksPerTick,
				constants.ClobPair_Btc.Status,
				constants.ClobPair_Btc.MakerFeePpm,
				constants.ClobPair_Btc.TakerFeePpm,
			)
			require.NoError(t, err)
			_, err = keeper.CreatePerpetualClobPair(
				ctx,
				clobtest.MustPerpetualId(constants.ClobPair_Eth),
				satypes.BaseQuantums(constants.ClobPair_Eth.StepBaseQuantums),
				satypes.BaseQuantums(constants.ClobPair_Eth.MinOrderBaseQuantums),
				constants.ClobPair_Eth.QuantumConversionExponent,
				constants.ClobPair_Eth.SubticksPerTick,
				constants.ClobPair_Eth.Status,
				constants.ClobPair_Eth.MakerFeePpm,
				constants.ClobPair_Eth.TakerFeePpm,
			)
			require.NoError(t, err)

			require.NoError(
				t,
				keeper.InitializeLiquidationsConfig(ctx, tc.liquidationConfig),
			)

			// Place all existing orders on the orderbook.
			for _, matchableOrder := range tc.placedMatchableOrders {
				// If the order is a liquidation order, place the liquidation.
				// Else, assume it's a regular order and place it.
				if liquidationOrder, ok := matchableOrder.(*types.LiquidationOrder); ok {
					_, _, err := keeper.CheckTxPlacePerpetualLiquidation(
						ctx,
						*liquidationOrder,
					)
					require.NoError(t, err)
				} else {
					order := matchableOrder.MustGetOrder()
					_, _, err := keeper.CheckTxPlaceOrder(ctx, &types.MsgPlaceOrder{Order: order.MustGetOrder()})
					require.NoError(t, err)
				}
			}

			// Run the test case and verify expectations.
			if tc.expectedError != nil {
				require.PanicsWithError(
					t,
					tc.expectedError.Error(),
					func() {
						_, _, _ = keeper.CheckTxPlacePerpetualLiquidation(ctx, tc.order)
					},
				)
			} else {
				_, orderStatus, err := keeper.CheckTxPlacePerpetualLiquidation(ctx, tc.order)
				require.NoError(t, err)
				require.Equal(t, tc.expectedOrderStatus, orderStatus)

				for subaccountId, liquidationInfo := range tc.expectedSubaccountLiquidationInfo {
					require.Equal(
						t,
						liquidationInfo,
						keeper.GetSubaccountLiquidationInfo(ctx, subaccountId),
					)
				}

				// Verify test expectations.
				// TODO(DEC-1979): Refactor these tests to support the operations queue refactor.
				// placedOrders, matchedOrders := memclob.GetPendingFills(ctx)

				// require.Equal(t, tc.expectedPlacedOrders, placedOrders, "Placed orders lists are not equal")
				// require.Equal(t, tc.expectedMatchedOrders, matchedOrders, "Matched orders lists are not equal")
			}
		})
	}
}

func TestPlacePerpetualLiquidation_SendOffchainMessages(t *testing.T) {
	indexerEventManager := &mocks.IndexerEventManager{}
	for _, message := range constants.TestOffchainMessages {
		indexerEventManager.On("SendOffchainData", message).Once().Return()
	}

	memClob := &mocks.MemClob{}
	memClob.On("SetClobKeeper", mock.Anything).Return()

	ctx, keeper, pricesKeeper, _, perpetualsKeeper, _, _, _ :=
		keepertest.ClobKeepers(t, memClob, &mocks.BankKeeper{}, indexerEventManager)
	ctx = ctx.WithTxBytes(constants.TestTxBytes)
	// CheckTx mode set correctly
	ctx = ctx.WithIsCheckTx(true)
	prices.InitGenesis(ctx, *pricesKeeper, constants.Prices_DefaultGenesisState)
	perpetuals.InitGenesis(ctx, *perpetualsKeeper, constants.Perpetuals_DefaultGenesisState)

	memClob.On("CreateOrderbook", ctx, constants.ClobPair_Btc).Return()
	_, err := keeper.CreatePerpetualClobPair(
		ctx,
		clobtest.MustPerpetualId(constants.ClobPair_Btc),
		satypes.BaseQuantums(constants.ClobPair_Btc.StepBaseQuantums),
		satypes.BaseQuantums(constants.ClobPair_Btc.MinOrderBaseQuantums),
		constants.ClobPair_Btc.QuantumConversionExponent,
		constants.ClobPair_Btc.SubticksPerTick,
		constants.ClobPair_Btc.Status,
		constants.ClobPair_Btc.MakerFeePpm,
		constants.ClobPair_Btc.TakerFeePpm,
	)
	require.NoError(t, err)

	order := constants.LiquidationOrder_Dave_Num0_Clob0_Sell1BTC_Price50000
	memClob.On("PlacePerpetualLiquidation", ctx, order).Return(
		satypes.BaseQuantums(100_000_000),
		types.Success,
		constants.TestOffchainUpdates,
		nil,
	)

	_, _, err = keeper.CheckTxPlacePerpetualLiquidation(ctx, order)
	require.NoError(t, err)

	indexerEventManager.AssertNumberOfCalls(t, "SendOffchainData", len(constants.TestOffchainMessages))
	indexerEventManager.AssertExpectations(t)
	memClob.AssertExpectations(t)
}

func TestIsLiquidatable(t *testing.T) {
	tests := map[string]struct {
		// State.
		perpetuals []perptypes.Perpetual

		// Subaccount state.
		assetPositions     []*satypes.AssetPosition
		perpetualPositions []*satypes.PerpetualPosition

		// Expectations.
		expectedIsLiquidatable bool
	}{
		"Subaccount with no open positions but positive net collateral is not liquidatable": {
			expectedIsLiquidatable: false,
			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * 1),
			),
		},
		"Subaccount with no open positions but negative net collateral is not liquidatable": {
			expectedIsLiquidatable: false,
			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -1),
			),
		},
		"Subaccount at initial margin requirements is not liquidatable": {
			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			perpetualPositions: []*satypes.PerpetualPosition{
				{
					PerpetualId: uint32(0),
					Quantums:    dtypes.NewInt(10_000_000), // 0.1 BTC, $5,000 notional.
				},
			},
			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -4_000),
			),
			expectedIsLiquidatable: false,
		},
		"Subaccount below initial but at maintenance margin requirements is not liquidatable": {
			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			perpetualPositions: []*satypes.PerpetualPosition{
				{
					PerpetualId: uint32(0),
					Quantums:    dtypes.NewInt(10_000_000), // 0.1 BTC, $5,000 notional.
				},
			},
			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -4_500),
			),
			expectedIsLiquidatable: false,
		},
		"Subaccount below maintenance margin requirements is liquidatable": {
			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			perpetualPositions: []*satypes.PerpetualPosition{
				{
					PerpetualId: uint32(0),
					Quantums:    dtypes.NewInt(10_000_000), // 0.1 BTC, $5,000 notional.
				},
			},
			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -4_501),
			),
			expectedIsLiquidatable: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup keeper state.
			memClob := memclob.NewMemClobPriceTimePriority(false)
			ctx,
				clobKeeper,
				pricesKeeper,
				_,
				perpetualsKeeper,
				subaccountsKeeper,
				_,
				_ := keepertest.ClobKeepers(t, memClob, &mocks.BankKeeper{}, &mocks.IndexerEventManager{})

			// Create the default markets.
			keepertest.CreateTestMarketsAndExchangeFeeds(t, ctx, pricesKeeper)

			// Create liquidity tiers.
			keepertest.CreateTestLiquidityTiers(t, ctx, perpetualsKeeper)

			// Create all perpetuals.
			for _, p := range tc.perpetuals {
				_, err := perpetualsKeeper.CreatePerpetual(
					ctx,
					p.Ticker,
					p.MarketId,
					p.AtomicResolution,
					p.DefaultFundingPpm,
					p.LiquidityTier,
				)
				require.NoError(t, err)
			}

			// Create the subaccount.
			subaccount := satypes.Subaccount{
				Id: &satypes.SubaccountId{
					Owner:  "liquidations_test",
					Number: 0,
				},
				AssetPositions:     tc.assetPositions,
				PerpetualPositions: tc.perpetualPositions,
			}
			subaccountsKeeper.SetSubaccount(ctx, subaccount)
			isLiquidatable, err := clobKeeper.IsLiquidatable(ctx, *subaccount.Id)

			// Note that there should never be errors when passing the empty update.
			require.NoError(t, err)
			require.Equal(t, tc.expectedIsLiquidatable, isLiquidatable)
		})
	}
}

func TestGetBankruptcyPriceInQuoteQuantums(t *testing.T) {
	tests := map[string]struct {
		// Parameters.
		perpetualId   uint32
		deltaQuantums int64

		// Perpetual state.
		perpetuals []perptypes.Perpetual

		// Subaccount state.
		assetPositions     []*satypes.AssetPosition
		perpetualPositions []*satypes.PerpetualPosition

		// Expectations.
		expectedBankruptcyPriceQuoteQuantums *big.Int
		expectedError                        error
	}{
		`Can calculate bankruptcy price in quote quantums for a subaccount that is fully closing
		one long position that is slightly below maintenance margin requirements`: {
			perpetualId:   0,
			deltaQuantums: -10_000_000,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -4_501),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCLong,
			},

			// 4,501,000,000 quote quantums = $4,501. This means if 0.1 BTC can't be sold for at
			// least $4,501 then the subaccount will be bankrupt when this position is closed.
			expectedBankruptcyPriceQuoteQuantums: big.NewInt(4_501_000_000),
		},
		`Can calculate bankruptcy price in quote quantums for a subaccount that is fully closing
		one short position that is slightly below maintenance margin requirements`: {
			perpetualId:   0,
			deltaQuantums: 10_000_000,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * 5_499),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCShort,
			},

			// -5,499,000,000 quote quantums = -$5,499. This means if 0.1 BTC can't be bought for
			// at most $5,499 then the subaccount will be bankrupt when this position is closed.
			expectedBankruptcyPriceQuoteQuantums: big.NewInt(-5_499_000_000),
		},
		`Can calculate bankruptcy price in quote quantums for a subaccount that is fully closing
		one long position that is at the bankruptcy price`: {
			perpetualId:   0,
			deltaQuantums: -10_000_000,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -5_000),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCLong,
			},

			// 5,000,000,000 quote quantums = $5,000. This means if 0.1 BTC can't be sold for at
			// least $5,000 then the subaccount will be bankrupt when this position is closed.
			expectedBankruptcyPriceQuoteQuantums: big.NewInt(5_000_000_000),
		},
		`Can calculate bankruptcy price in quote quantums for a subaccount that is partially closing
		one long position that is at the bankruptcy price`: {
			perpetualId:   0,
			deltaQuantums: -5_000_000,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -5_000),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCLong,
			},

			// 2,500,000,000 quote quantums = $2,500. This means if 0.1 BTC can't be sold for at
			// least $2,500 then the subaccount will be bankrupt when this position is closed.
			expectedBankruptcyPriceQuoteQuantums: big.NewInt(2_500_000_000),
		},
		`Can calculate bankruptcy price in quote quantums for a subaccount that is partially closing
		one short position that is at the bankruptcy price`: {
			perpetualId:   0,
			deltaQuantums: 5_000_000,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * 5_000),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCShort,
			},

			// -2,500,000,000 quote quantums = -$2,500. This means if 0.1 BTC can't be bought for at
			// most $2,500 then the subaccount will be bankrupt when this position is closed.
			expectedBankruptcyPriceQuoteQuantums: big.NewInt(-2_500_000_000),
		},
		`Can calculate bankruptcy price in quote quantums for a subaccount that is fully closing
		one short position that is at the bankruptcy price`: {
			perpetualId:   0,
			deltaQuantums: 10_000_000,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * 5_000),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCShort,
			},

			// -5,000,000,000 quote quantums = -$5,000. This means if 0.1 BTC can't be bought for at
			// most $5,000 then the subaccount will be bankrupt when this position is closed.
			expectedBankruptcyPriceQuoteQuantums: big.NewInt(-5_000_000_000),
		},
		`Can calculate bankruptcy price in quote quantums for a subaccount that is fully closing
		one long position that is below the bankruptcy price`: {
			perpetualId:   0,
			deltaQuantums: -10_000_000,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -5_100),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCLong,
			},

			// 5,100,000,000 quote quantums = $5,100. This means if 0.1 BTC can't be sold for at
			// least $5,100 then the subaccount will be bankrupt when this position is closed.
			expectedBankruptcyPriceQuoteQuantums: big.NewInt(5_100_000_000),
		},
		`Can calculate bankruptcy price in quote quantums for a subaccount that is fully closing
		one short position that is below the bankruptcy price`: {
			perpetualId:   0,
			deltaQuantums: 10_000_000,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * 4_900),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCShort,
			},

			// -4,900,000,000 quote quantums = -$4,900. This means if 0.1 BTC can't be bought for at
			// most $4,900 then the subaccount will be bankrupt when this position is closed.
			expectedBankruptcyPriceQuoteQuantums: big.NewInt(-4_900_000_000),
		},
		`Can calculate bankruptcy price in quote quantums for a subaccount that is fully closing
		one long position and has multiple long positions`: {
			perpetualId:   1,
			deltaQuantums: -100_000_000,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
				constants.EthUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -490),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_FourThousandthsBTCLong,
				&constants.PerpetualPosition_OneTenthEthLong,
			},

			// 294,000,000 quote quantums = $294. This means if 0.1 ETH can't be sold for at
			// least $294 then the subaccount will be bankrupt when this position is closed.
			expectedBankruptcyPriceQuoteQuantums: big.NewInt(294_000_000),
		},
		`Can calculate bankruptcy price in quote quantums for a subaccount that is fully closing
		one short position and has multiple short positions`: {
			perpetualId:   1,
			deltaQuantums: 100_000_000,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
				constants.EthUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * 510),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_FourThousandthsBTCShort,
				&constants.PerpetualPosition_OneTenthEthShort,
			},

			// -306,000,000 quote quantums = -$306. This means if 0.1 ETH can't be bought for at
			// most $306 then the subaccount will be bankrupt when this position is closed.
			expectedBankruptcyPriceQuoteQuantums: big.NewInt(-306_000_000),
		},
		`Can calculate bankruptcy price in quote quantums for a subaccount that is fully closing
		one short position and has a long and short position`: {
			perpetualId:   1,
			deltaQuantums: 100_000_000,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
				constants.EthUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * 110),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_FourThousandthsBTCLong,
				&constants.PerpetualPosition_OneTenthEthShort,
			},

			// -306,000,000 quote quantums = -$306. This means if 0.1 ETH can't be bought for at
			// most $306 then the subaccount will be bankrupt when this position is closed.
			expectedBankruptcyPriceQuoteQuantums: big.NewInt(-306_000_000),
		},
		`Rounds up bankruptcy price in quote quantums for a subaccount that is partially closing
		one long position that is slightly below maintenance margin requirements`: {
			perpetualId:   0,
			deltaQuantums: -21_347,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -13),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneBTCLong,
			},

			// 2,776 quote quantums = $0.002776. This means if 0.00021347 BTC can't be sold for
			// at least $0.002776 then the subaccount will be bankrupt when this position is closed.
			// Note that the result is rounded up from 2,775.11 quote quantums.
			expectedBankruptcyPriceQuoteQuantums: big.NewInt(2_776),
		},
		`Rounds up bankruptcy price in quote quantums for a subaccount that is partially closing
		one short position that is below the bankruptcy price`: {
			perpetualId:   0,
			deltaQuantums: 21_347,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * 13),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneBTCShort,
			},

			// -2,775 quote quantums = $0.002775. This means if 0.00021347 BTC can't be bought for
			// at most $0.002775 then the subaccount will be bankrupt when this position is closed.
			// Note that the result is rounded down from 2,775.11 quote quantums.
			expectedBankruptcyPriceQuoteQuantums: big.NewInt(-2_775),
		},
		`Account with a long position that cannot be liquidated at a loss has a negative
		bankruptcy price in quote quantums`: {
			perpetualId:   0,
			deltaQuantums: -100_000_000,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			// Note that if quote balance is positive for longs, this indicates that the subaccount's
			// quote balance exceeds the notional value of their long position.
			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneBTCLong,
			},

			// -1,000,000 quote quantums = -$1,000,000. This means if 1 BTC can't be sold for
			// at least -$1,000,000 then the subaccount will be bankrupt when this position is closed.
			// Note this is not possible since it's impossible to sell a position for less than 0 dollars.
			expectedBankruptcyPriceQuoteQuantums: big.NewInt(-1_000_000),
		},
		`Returns error when deltaQuantums is zero`: {
			perpetualId:   0,
			deltaQuantums: 0,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},
			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -4_501),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneBTCLong,
			},

			expectedError: types.ErrInvalidPerpetualPositionSizeDelta,
		},
		`Returns error when subaccount does not have an open position for perpetual id`: {
			perpetualId:   0,
			deltaQuantums: -10_000_000,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},
			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -4_501),
			),
			perpetualPositions: []*satypes.PerpetualPosition{},

			expectedError: types.ErrInvalidPerpetualPositionSizeDelta,
		},
		`Returns error when delta quantums and perpetual position have the same sign`: {
			perpetualId:   0,
			deltaQuantums: 10_000_000,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},
			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -4_501),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneBTCLong,
			},

			expectedError: types.ErrInvalidPerpetualPositionSizeDelta,
		},
		`Returns error when abs delta quantums is greater than position size`: {
			perpetualId:   0,
			deltaQuantums: -100_000_001,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},
			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -4_501),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneBTCLong,
			},

			expectedError: types.ErrInvalidPerpetualPositionSizeDelta,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup keeper state.
			memClob := memclob.NewMemClobPriceTimePriority(false)
			ctx,
				clobKeeper,
				pricesKeeper,
				_,
				perpetualsKeeper,
				subaccountsKeeper,
				_,
				_ := keepertest.ClobKeepers(t, memClob, &mocks.BankKeeper{}, &mocks.IndexerEventManager{})

			// Create the default markets.
			keepertest.CreateTestMarketsAndExchangeFeeds(t, ctx, pricesKeeper)

			// Create liquidity tiers.
			keepertest.CreateTestLiquidityTiers(t, ctx, perpetualsKeeper)

			// Create all perpetuals.
			for _, p := range tc.perpetuals {
				_, err := perpetualsKeeper.CreatePerpetual(
					ctx,
					p.Ticker,
					p.MarketId,
					p.AtomicResolution,
					p.DefaultFundingPpm,
					p.LiquidityTier,
				)
				require.NoError(t, err)
			}

			// Create the subaccount.
			subaccountId := satypes.SubaccountId{
				Owner:  "liquidations_test",
				Number: 0,
			}
			subaccount := satypes.Subaccount{
				Id:                 &subaccountId,
				AssetPositions:     tc.assetPositions,
				PerpetualPositions: tc.perpetualPositions,
			}
			subaccountsKeeper.SetSubaccount(ctx, subaccount)

			bankruptcyPriceInQuoteQuantums, err := clobKeeper.GetBankruptcyPriceInQuoteQuantums(
				ctx,
				*subaccount.Id,
				tc.perpetualId,
				big.NewInt(tc.deltaQuantums),
			)

			if tc.expectedError != nil {
				require.ErrorIs(t, err, tc.expectedError)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedBankruptcyPriceQuoteQuantums, bankruptcyPriceInQuoteQuantums)

				// Verify that the returned delta quote quantums can pass `CanUpdateSubaccounts` function.
				success, _, err := subaccountsKeeper.CanUpdateSubaccounts(
					ctx,
					[]satypes.Update{
						{
							SubaccountId: subaccountId,
							AssetUpdates: keepertest.CreateUsdcAssetUpdate(bankruptcyPriceInQuoteQuantums),
							PerpetualUpdates: []satypes.PerpetualUpdate{
								{
									PerpetualId:      tc.perpetualId,
									BigQuantumsDelta: big.NewInt(tc.deltaQuantums),
								},
							},
						},
					},
				)

				require.True(t, success)
				require.NoError(t, err)
			}
		})
	}
}

func TestGetFillablePrice(t *testing.T) {
	tests := map[string]struct {
		// Parameters.
		perpetualId   uint32
		deltaQuantums int64

		// Perpetual state.
		perpetuals []perptypes.Perpetual

		// Subaccount state.
		assetPositions     []*satypes.AssetPosition
		perpetualPositions []*satypes.PerpetualPosition

		// Liquidation config.
		liquidationConfig *types.LiquidationsConfig

		// Expectations.
		expectedFillablePrice *big.Rat
		expectedError         error
	}{
		`Can calculate fillable price for a subaccount with one long position that is slightly
		below maintenance margin requirements`: {
			perpetualId:   0,
			deltaQuantums: -10_000_000,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -4_501),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCLong,
			},

			// $49,999 = (49,999 / 100) subticks * 10^(QuoteCurrencyAtomicResolution - BaseCurrencyAtomicResolution).
			// This means we should close the 0.1 BTC long with a $4,999.9 notional sell order.
			expectedFillablePrice: big.NewRat(49_999, 100),
		},
		`Can calculate fillable price for a subaccount with one long position when bankruptcyAdjustmentPpm is 2_000_000`: {
			perpetualId:   0,
			deltaQuantums: -10_000_000,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -4_501),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCLong,
			},

			liquidationConfig: &types.LiquidationsConfig{
				MaxLiquidationFeePpm: 5_000,
				FillablePriceConfig: types.FillablePriceConfig{
					BankruptcyAdjustmentPpm:           2_000_000,
					SpreadToMaintenanceMarginRatioPpm: 100_000,
				},
				PositionBlockLimits:   constants.PositionBlockLimits_No_Limit,
				SubaccountBlockLimits: constants.SubaccountBlockLimits_No_Limit,
			},
			// $49,998 = (49,998 / 100) subticks * 10^(QuoteCurrencyAtomicResolution - BaseCurrencyAtomicResolution).
			// This means we should close the 0.1 BTC long with a $4,999.8 notional sell order.
			expectedFillablePrice: big.NewRat(49_998, 100),
		},
		`Can calculate fillable price for a subaccount with one long position when 
		spreadToMaintenanceMarginRatioPpm is 200_000`: {
			perpetualId:   0,
			deltaQuantums: -10_000_000,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -4_501),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCLong,
			},

			liquidationConfig: &types.LiquidationsConfig{
				MaxLiquidationFeePpm: 5_000,
				FillablePriceConfig: types.FillablePriceConfig{
					BankruptcyAdjustmentPpm:           lib.OneMillion,
					SpreadToMaintenanceMarginRatioPpm: 200_000,
				},
				PositionBlockLimits:   constants.PositionBlockLimits_No_Limit,
				SubaccountBlockLimits: constants.SubaccountBlockLimits_No_Limit,
			},
			// $49,998 = (49,998 / 100) subticks * 10^(QuoteCurrencyAtomicResolution - BaseCurrencyAtomicResolution).
			// This means we should close the 0.1 BTC long with a $4,999.8 notional sell order.
			expectedFillablePrice: big.NewRat(49_998, 100),
		},
		`Can calculate fillable price for a subaccount with one short position that is slightly
		below maintenance margin requirements`: {
			perpetualId:   0,
			deltaQuantums: 10_000_000,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * 5_499),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCShort,
			},

			// $50,001 = (50,001 / 100) subticks * 10^(QuoteCurrencyAtomicResolution - BaseCurrencyAtomicResolution).
			// This means we should close the 0.1 BTC short with a $5,000.1 notional buy order.
			expectedFillablePrice: big.NewRat(50_001, 100),
		},
		`Can calculate fillable price for a subaccount with one short position when bankruptcyAdjustmentPpm is 2_000_000`: {
			perpetualId:   0,
			deltaQuantums: 10_000_000,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * 5_499),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCShort,
			},

			liquidationConfig: &types.LiquidationsConfig{
				MaxLiquidationFeePpm: 5_000,
				FillablePriceConfig: types.FillablePriceConfig{
					BankruptcyAdjustmentPpm:           2_000_000,
					SpreadToMaintenanceMarginRatioPpm: 100_000,
				},
				PositionBlockLimits:   constants.PositionBlockLimits_No_Limit,
				SubaccountBlockLimits: constants.SubaccountBlockLimits_No_Limit,
			},

			// $50,002 = (50,002 / 100) subticks * 10^(QuoteCurrencyAtomicResolution - BaseCurrencyAtomicResolution).
			// This means we should close the 0.1 BTC short with a $5,000.2 notional buy order.
			expectedFillablePrice: big.NewRat(50_002, 100),
		},
		`Can calculate fillable price for a subaccount with one short position when 
		SpreadToMaintenanceMarginRatioPpm is 200_000`: {
			perpetualId:   0,
			deltaQuantums: 10_000_000,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * 5_499),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCShort,
			},

			liquidationConfig: &types.LiquidationsConfig{
				MaxLiquidationFeePpm: 5_000,
				FillablePriceConfig: types.FillablePriceConfig{
					BankruptcyAdjustmentPpm:           lib.OneMillion,
					SpreadToMaintenanceMarginRatioPpm: 200_000,
				},
				PositionBlockLimits:   constants.PositionBlockLimits_No_Limit,
				SubaccountBlockLimits: constants.SubaccountBlockLimits_No_Limit,
			},

			// $50,002 = (50,002 / 100) subticks * 10^(QuoteCurrencyAtomicResolution - BaseCurrencyAtomicResolution).
			// This means we should close the 0.1 BTC short with a $5,000.2 notional buy order.
			expectedFillablePrice: big.NewRat(50_002, 100),
		},
		"Can calculate fillable price for a subaccount with one long position at the bankruptcy price": {
			perpetualId:   0,
			deltaQuantums: -10_000_000,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -5_000),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCLong,
			},

			// $49,500 = (495 / 1) subticks * 10^(QuoteCurrencyAtomicResolution - BaseCurrencyAtomicResolution).
			// This means we should close the 0.1 BTC long with a $4,950 notional sell order.
			expectedFillablePrice: big.NewRat(495, 1),
		},
		`Can calculate fillable price for a subaccount with one long position at the bankruptcy price
		where we are liquidating half of the position`: {
			perpetualId:   0,
			deltaQuantums: -5_000_000,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -5_000),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCLong,
			},

			// $49,500 = (495 / 1) subticks * 10^(QuoteCurrencyAtomicResolution - BaseCurrencyAtomicResolution).
			// This means we should close the 0.1 BTC long with a $4,950 notional sell order.
			// Note that even though we are closing half of the position, the fillable price is the same as
			// if we were closing the full position because it's calculated based on the position size.
			expectedFillablePrice: big.NewRat(495, 1),
		},
		"Can calculate fillable price for a subaccount with one short position at the bankruptcy price": {
			perpetualId:   0,
			deltaQuantums: 10_000_000,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * 5_000),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCShort,
			},

			// $50,500 = (505 / 1) subticks * 10^(QuoteCurrencyAtomicResolution - BaseCurrencyAtomicResolution).
			// This means we should close the 0.1 BTC short with a $5,050 notional buy order.
			expectedFillablePrice: big.NewRat(505, 1),
		},
		"Can calculate fillable price for a subaccount with one long position below the bankruptcy price": {
			perpetualId:   0,
			deltaQuantums: -10_000_000,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -5_500),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCLong,
			},

			// $49,500 = (495 / 1) subticks * 10^(QuoteCurrencyAtomicResolution - BaseCurrencyAtomicResolution).
			// This means we should close the 0.1 BTC long with a $4,950 notional sell order.
			expectedFillablePrice: big.NewRat(495, 1),
		},
		"Can calculate fillable price for a subaccount with one short position below the bankruptcy price": {
			perpetualId:   0,
			deltaQuantums: 10_000_000,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * 4_500),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCShort,
			},

			// $50,500 = (505 / 1) subticks * 10^(QuoteCurrencyAtomicResolution - BaseCurrencyAtomicResolution).
			// This means we should close the 0.1 BTC short with a $5,050 notional buy order.
			expectedFillablePrice: big.NewRat(505, 1),
		},
		"Can calculate fillable price for a subaccount with multiple long positions": {
			perpetualId:   1,
			deltaQuantums: -100_000_000,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
				constants.EthUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -490),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_FourThousandthsBTCLong,
				&constants.PerpetualPosition_OneTenthEthLong,
			},

			// $2976 = (372 / 125) subticks * QuoteCurrencyAtomicResolution / BaseCurrencyAtomicResolution.
			// This means we should close our 0.1 ETH long for $2,976 dollars.
			expectedFillablePrice: big.NewRat(372, 125),
		},
		`Can calculate fillable price when bankruptcyAdjustmentPpm is max uint32`: {
			perpetualId:   0,
			deltaQuantums: -10_000_000,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -4_501),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCLong,
			},

			liquidationConfig: &types.LiquidationsConfig{
				MaxLiquidationFeePpm: 5_000,
				FillablePriceConfig: types.FillablePriceConfig{
					BankruptcyAdjustmentPpm:           math.MaxUint32,
					SpreadToMaintenanceMarginRatioPpm: 100_000,
				},
				PositionBlockLimits:   constants.PositionBlockLimits_No_Limit,
				SubaccountBlockLimits: constants.SubaccountBlockLimits_No_Limit,
			},

			// $49,500 = (495 / 1) subticks * 10^(QuoteCurrencyAtomicResolution - BaseCurrencyAtomicResolution).
			// This means we should close the 0.1 BTC long with a $4,950 notional sell order.
			expectedFillablePrice: big.NewRat(495, 1),
		},
		`Can calculate fillable price when SpreadTomaintenanceMarginRatioPpm is 1`: {
			perpetualId:   0,
			deltaQuantums: -10_000_000,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -4_501),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCLong,
			},

			liquidationConfig: &types.LiquidationsConfig{
				MaxLiquidationFeePpm: 5_000,
				FillablePriceConfig: types.FillablePriceConfig{
					BankruptcyAdjustmentPpm:           lib.OneMillion,
					SpreadToMaintenanceMarginRatioPpm: 1,
				},
				PositionBlockLimits:   constants.PositionBlockLimits_No_Limit,
				SubaccountBlockLimits: constants.SubaccountBlockLimits_No_Limit,
			},

			expectedFillablePrice: big.NewRat(4_999_999_999, 10_000_000),
		},
		`Can calculate fillable price when SpreadTomaintenanceMarginRatioPpm is one million`: {
			perpetualId:   0,
			deltaQuantums: -10_000_000,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -4_501),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCLong,
			},

			liquidationConfig: &types.LiquidationsConfig{
				MaxLiquidationFeePpm: 5_000,
				FillablePriceConfig: types.FillablePriceConfig{
					BankruptcyAdjustmentPpm:           lib.OneMillion,
					SpreadToMaintenanceMarginRatioPpm: lib.OneMillion,
				},
				PositionBlockLimits:   constants.PositionBlockLimits_No_Limit,
				SubaccountBlockLimits: constants.SubaccountBlockLimits_No_Limit,
			},

			// $49,990 = (49990 / 100) subticks * 10^(QuoteCurrencyAtomicResolution - BaseCurrencyAtomicResolution).
			// This means we should close the 0.1 BTC long with a $4,999 notional sell order.
			expectedFillablePrice: big.NewRat(49_990, 100),
		},
		`Returns error when deltaQuantums is zero`: {
			perpetualId:   0,
			deltaQuantums: 0,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},
			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -4_501),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneBTCLong,
			},

			expectedError: types.ErrInvalidPerpetualPositionSizeDelta,
		},
		`Returns error when subaccount does not have an open position for perpetual id`: {
			perpetualId:   0,
			deltaQuantums: -10_000_000,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},
			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -4_501),
			),
			perpetualPositions: []*satypes.PerpetualPosition{},

			expectedError: types.ErrInvalidPerpetualPositionSizeDelta,
		},
		`Returns error when delta quantums and perpetual position have the same sign`: {
			perpetualId:   0,
			deltaQuantums: 10_000_000,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},
			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -4_501),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneBTCLong,
			},

			expectedError: types.ErrInvalidPerpetualPositionSizeDelta,
		},
		`Returns error when abs delta quantums is greater than position size`: {
			perpetualId:   0,
			deltaQuantums: -100_000_001,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},
			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -4_501),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneBTCLong,
			},

			expectedError: types.ErrInvalidPerpetualPositionSizeDelta,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup keeper state.
			memClob := memclob.NewMemClobPriceTimePriority(false)
			ctx,
				clobKeeper,
				pricesKeeper,
				_,
				perpetualsKeeper,
				subaccountsKeeper,
				_,
				_ := keepertest.ClobKeepers(t, memClob, &mocks.BankKeeper{}, &mocks.IndexerEventManager{})

			// Initialize the liquidations config.
			if tc.liquidationConfig != nil {
				require.NoError(t,
					clobKeeper.InitializeLiquidationsConfig(ctx, *tc.liquidationConfig),
				)
			} else {
				require.NoError(t,
					clobKeeper.InitializeLiquidationsConfig(ctx, types.LiquidationsConfig_Default),
				)
			}

			// Create the default markets.
			keepertest.CreateTestMarketsAndExchangeFeeds(t, ctx, pricesKeeper)

			// Create liquidity tiers.
			keepertest.CreateTestLiquidityTiers(t, ctx, perpetualsKeeper)

			// Create all perpetuals.
			for _, p := range tc.perpetuals {
				_, err := perpetualsKeeper.CreatePerpetual(
					ctx,
					p.Ticker,
					p.MarketId,
					p.AtomicResolution,
					p.DefaultFundingPpm,
					p.LiquidityTier,
				)
				require.NoError(t, err)
			}

			// Create the subaccount.
			subaccount := satypes.Subaccount{
				Id: &satypes.SubaccountId{
					Owner:  "liquidations_test",
					Number: 0,
				},
				AssetPositions:     tc.assetPositions,
				PerpetualPositions: tc.perpetualPositions,
			}
			subaccountsKeeper.SetSubaccount(ctx, subaccount)

			fillablePrice, err := clobKeeper.GetFillablePrice(
				ctx,
				*subaccount.Id,
				tc.perpetualId,
				big.NewInt(tc.deltaQuantums),
			)

			if tc.expectedError != nil {
				require.ErrorIs(t, err, tc.expectedError)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedFillablePrice, fillablePrice)
			}
		})
	}
}

func TestGetLiquidationInsuranceFundDelta(t *testing.T) {
	tests := map[string]struct {
		// Parameters.
		perpetualId uint32
		isBuy       bool
		fillAmount  uint64
		subticks    types.Subticks

		liquidationConfig *types.LiquidationsConfig

		// Perpetual and subaccount state.
		perpetuals []perptypes.Perpetual

		// Subaccount state.
		assetPositions     []*satypes.AssetPosition
		perpetualPositions []*satypes.PerpetualPosition

		// Expectations.
		expectedLiquidationInsuranceFundDeltaBig *big.Int
		expectedError                            error
	}{
		`Fully closing one long position above the bankruptcy price and pays max liquidation fee`: {
			perpetualId: 0,
			isBuy:       false,
			fillAmount:  10_000_000,     // -0.1 BTC delta.
			subticks:    56_100_000_000, // 10% above bankruptcy price.

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -5_100),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCLong,
			},

			// Bankruptcy price in quote quantums is 5,100,000,000 quote quantums.
			// Liquidation price is 10% above bankruptcy price, 5,610,000,000 quote quantums.
			// abs(5,610,000,000) * 0.5% max liquidation fee < 5,610,000,000 - 5,100,000,000, so the max
			// liquidation fee is returned.
			expectedLiquidationInsuranceFundDeltaBig: big.NewInt(28_050_000),
		},
		`Fully closing one long position above the bankruptcy price pays max liquidation fee 
		when MaxLiquidationFeePpm is 25_000`: {
			perpetualId: 0,
			isBuy:       false,
			fillAmount:  10_000_000,     // -0.1 BTC delta.
			subticks:    56_100_000_000, // 10% above bankruptcy price.

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -5_100),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCLong,
			},
			liquidationConfig: &types.LiquidationsConfig{
				MaxLiquidationFeePpm:  25_000,
				FillablePriceConfig:   constants.FillablePriceConfig_Default,
				PositionBlockLimits:   constants.PositionBlockLimits_No_Limit,
				SubaccountBlockLimits: constants.SubaccountBlockLimits_No_Limit,
			},

			// Bankruptcy price in quote quantums is 5,100,000,000 quote quantums.
			// Liquidation price is 10% above bankruptcy price, 5,610,000,000 quote quantums.
			// abs(5,610,000,000) * 2.5% max liquidation fee < 5,610,000,000 - 5,100,000,000, so the max
			// liquidation fee is returned.
			expectedLiquidationInsuranceFundDeltaBig: big.NewInt(140_250_000),
		},
		`Fully closing one long position above the bankruptcy price pays less than max liquidation fee 
		when MaxLiquidationFeePpm is one million`: {
			perpetualId: 0,
			isBuy:       false,
			fillAmount:  10_000_000,     // -0.1 BTC delta.
			subticks:    56_100_000_000, // 10% above bankruptcy price.

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -5_100),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCLong,
			},
			liquidationConfig: &types.LiquidationsConfig{
				MaxLiquidationFeePpm:  1_000_000,
				FillablePriceConfig:   constants.FillablePriceConfig_Default,
				PositionBlockLimits:   constants.PositionBlockLimits_No_Limit,
				SubaccountBlockLimits: constants.SubaccountBlockLimits_No_Limit,
			},

			// Bankruptcy price in quote quantums is 5,100,000,000 quote quantums.
			// Liquidation price is 10% above bankruptcy price, 5,610,000,000 quote quantums.
			// abs(5,610,000,000) * 100% max liquidation fee > 5,610,000,000 - 5,100,000,000, so all
			// of the leftover collateral is transferred to the insurance fund.
			expectedLiquidationInsuranceFundDeltaBig: big.NewInt(510_000_000),
		},
		`Fully closing one short position above the bankruptcy price and pays max liquidation fee`: {
			perpetualId: 0,
			isBuy:       true,
			fillAmount:  10_000_000,     // 0.1 BTC delta.
			subticks:    44_100_000_000, // 10% above bankruptcy price.

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * 4_900),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCShort,
			},

			// Bankruptcy price in quote quantums is -4,900,000,000 quote quantums.
			// Liquidation price is 10% above bankruptcy price, -4,410,000,000 quote quantums.
			// abs(-4,410,000,000) * 0.5% max liquidation fee < -4,900,000,000 - -4,410,000,000, so
			// the max liquidation fee is returned.
			expectedLiquidationInsuranceFundDeltaBig: big.NewInt(22_050_000),
		},
		`Fully closing one short position above the bankruptcy price and pays max liquidation fee
		when MaxLiquidationFeePpm is 25_000`: {
			perpetualId: 0,
			isBuy:       true,
			fillAmount:  10_000_000,     // 0.1 BTC delta.
			subticks:    44_100_000_000, // 10% above bankruptcy price.

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * 4_900),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCShort,
			},
			liquidationConfig: &types.LiquidationsConfig{
				MaxLiquidationFeePpm:  25_000,
				FillablePriceConfig:   constants.FillablePriceConfig_Default,
				PositionBlockLimits:   constants.PositionBlockLimits_No_Limit,
				SubaccountBlockLimits: constants.SubaccountBlockLimits_No_Limit,
			},

			// Bankruptcy price in quote quantums is -4,900,000,000 quote quantums.
			// Liquidation price is 10% above bankruptcy price, -4,410,000,000 quote quantums.
			// abs(-4,410,000,000) * 2.5% max liquidation fee < -4,900,000,000 - -4,410,000,000, so
			// the max liquidation fee is returned.
			expectedLiquidationInsuranceFundDeltaBig: big.NewInt(110_250_000),
		},
		`Fully closing one short position above the bankruptcy price and pays less than max liquidation fee
		when MaxLiquidationFeePpm is one million`: {
			perpetualId: 0,
			isBuy:       true,
			fillAmount:  10_000_000,     // 0.1 BTC delta.
			subticks:    44_100_000_000, // 10% above bankruptcy price.

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * 4_900),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCShort,
			},
			liquidationConfig: &types.LiquidationsConfig{
				MaxLiquidationFeePpm:  1_000_000,
				FillablePriceConfig:   constants.FillablePriceConfig_Default,
				PositionBlockLimits:   constants.PositionBlockLimits_No_Limit,
				SubaccountBlockLimits: constants.SubaccountBlockLimits_No_Limit,
			},

			// Bankruptcy price in quote quantums is -4,900,000,000 quote quantums.
			// Liquidation price is 10% above bankruptcy price, -4,410,000,000 quote quantums.
			// abs(-4,410,000,000) * 100% max liquidation fee > -4,900,000,000 - -4,410,000,000, so all
			// of the leftover collateral is transferred to the insurance fund.
			expectedLiquidationInsuranceFundDeltaBig: big.NewInt(490_000_000),
		},
		`Fully closing one long position above the bankruptcy price and pays less than max
		liquidation fee`: {
			perpetualId: 0,
			isBuy:       false,
			fillAmount:  10_000_000,     // -0.1 BTC delta.
			subticks:    51_051_000_000, // 0.1% above bankruptcy price.

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -5_100),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCLong,
			},

			// Bankruptcy price in quote quantums is 5,100,000,000 quote quantums.
			// Liquidation price is 0.1% above bankruptcy price, 5,105,100,000 quote quantums.
			// 5,105,100,000 * 0.5% max liquidation fee > 5,105,100,000 - 5,100,000,000, so all
			// of the leftover collateral is transferred to the insurance fund.
			expectedLiquidationInsuranceFundDeltaBig: big.NewInt(5_100_000),
		},
		`Fully closing one short position above the bankruptcy price and pays less than max
		liquidation fee`: {
			perpetualId: 0,
			isBuy:       true,
			fillAmount:  10_000_000,     // 0.1 BTC delta.
			subticks:    48_951_000_000, // 0.1% above bankruptcy price.

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * 4_900),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCShort,
			},

			// Bankruptcy price in quote quantums is -4,900,000,000 quote quantums.
			// Liquidation price is 0.1% above bankruptcy price, -4,895,100,000 quote quantums.
			// -4,895,100,000 * 0.5% max liquidation fee < -4,895,100,000 - -4,900,000,000, so all
			// of the leftover collateral is transferred to the insurance fund.
			expectedLiquidationInsuranceFundDeltaBig: big.NewInt(4_900_000),
		},
		`Fully closing one long position at the bankruptcy price and the delta is 0`: {
			perpetualId: 0,
			isBuy:       false,
			fillAmount:  10_000_000,     // -0.1 BTC delta.
			subticks:    51_000_000_000, // 0% above bankruptcy price (equal).

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -5_100),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCLong,
			},

			// Bankruptcy price in quote quantums is 5,100,000,000 quote quantums.
			// Liquidation price is 0% above bankruptcy price, 5,100,000,000 quote quantums.
			// 5,100,000,000 * 0.5% max liquidation fee > 5,100,000,000 - 5,100,000,000, so all
			// of the leftover collateral (which is zero) is transferred to the insurance fund.
			expectedLiquidationInsuranceFundDeltaBig: big.NewInt(0),
		},
		`Fully closing one short position above the bankruptcy price and the delta is 0`: {
			perpetualId: 0,
			isBuy:       true,
			fillAmount:  10_000_000,     // 0.1 BTC delta.
			subticks:    49_000_000_000, // 0% above bankruptcy price.

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * 4_900),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCShort,
			},

			// Bankruptcy price in quote quantums is -4,900,000,000 quote quantums.
			// Liquidation price is 0.1% above bankruptcy price, -4,900,000,000 quote quantums.
			// -4,900,000,000 * 0.5% max liquidation fee < -4,900,000,000 - -4,900,000,000, so all
			// of the leftover collateral (which is zero) is transferred to the insurance fund.
			expectedLiquidationInsuranceFundDeltaBig: big.NewInt(0),
		},
		`Fully closing one long position below the bankruptcy price and the insurance fund must
		cover the loss`: {
			perpetualId: 0,
			isBuy:       false,
			fillAmount:  10_000_000,     // -0.1 BTC delta.
			subticks:    50_490_000_000, // 1% below bankruptcy price.

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * -5_100),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCLong,
			},

			// Bankruptcy price in quote quantums is 5,100,000,000 quote quantums.
			// Liquidation price is 1% below the bankruptcy price, 5,049,000,000 quote quantums.
			// 5,049,000,000 - 5,100,000,000 < 0, so the insurance fund must cover the losses.
			expectedLiquidationInsuranceFundDeltaBig: big.NewInt(-51_000_000),
		},
		`If fully closing one short position below the bankruptcy price the insurance fund must
		cover the loss`: {
			perpetualId: 0,
			isBuy:       true,
			fillAmount:  10_000_000,     // 0.1 BTC delta.
			subticks:    49_490_000_000, // 1% below bankruptcy price.

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * 4_900),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCShort,
			},

			// Bankruptcy price in quote quantums is -4,900,000,000 quote quantums.
			// Liquidation price is 1% below the bankruptcy price, -4,949,000,000 quote quantums.
			// -4,949,000,000 - -4,900,000,000 < 0, so the insurance fund msut cover the losses.
			expectedLiquidationInsuranceFundDeltaBig: big.NewInt(-49_000_000),
		},
		"Returns error when delta quantums is zero": {
			perpetualId: 0,
			isBuy:       true,
			fillAmount:  0,
			subticks:    50_000_000_000,

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * 4_900),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCShort,
			},
			expectedError: types.ErrInvalidQuantumsForInsuranceFundDeltaCalculation,
		},
		"Succeeds when delta quote quantums is zero": {
			perpetualId: 0,
			isBuy:       true,
			fillAmount:  10_000_000, // 0.1 BTC delta.
			subticks:    1,          // Quote quantums for 0.1 BTC is 1/10, rounded to zero.

			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},

			assetPositions: keepertest.CreateUsdcAssetPosition(
				big.NewInt(constants.QuoteBalance_OneDollar * 4_900),
			),
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCShort,
			},

			// Bankruptcy price in quote quantums is -4,900,000,000 quote quantums.
			// Insurance fund delta before applying position limit is 0 - -4,900,000,000 = 4,900,000,000.
			// abs(0) * 0.5% max liquidation fee < 4,900,000,000, so overall delta is zero.
			expectedLiquidationInsuranceFundDeltaBig: big.NewInt(0),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup keeper state.
			memClob := memclob.NewMemClobPriceTimePriority(false)
			ctx,
				clobKeeper,
				pricesKeeper,
				_,
				perpetualsKeeper,
				subaccountsKeeper,
				_,
				_ := keepertest.ClobKeepers(t, memClob, &mocks.BankKeeper{}, &mocks.IndexerEventManager{})

			// Create the default markets.
			keepertest.CreateTestMarketsAndExchangeFeeds(t, ctx, pricesKeeper)

			// Create liquidity tiers.
			keepertest.CreateTestLiquidityTiers(t, ctx, perpetualsKeeper)

			// Create all perpetuals.
			for _, p := range tc.perpetuals {
				_, err := perpetualsKeeper.CreatePerpetual(
					ctx,
					p.Ticker,
					p.MarketId,
					p.AtomicResolution,
					p.DefaultFundingPpm,
					p.LiquidityTier,
				)
				require.NoError(t, err)
			}

			// Create clob pair.
			_, err := clobKeeper.CreatePerpetualClobPair(
				ctx,
				clobtest.MustPerpetualId(constants.ClobPair_Btc),
				satypes.BaseQuantums(constants.ClobPair_Btc.StepBaseQuantums),
				satypes.BaseQuantums(constants.ClobPair_Btc.MinOrderBaseQuantums),
				constants.ClobPair_Btc.QuantumConversionExponent,
				constants.ClobPair_Btc.SubticksPerTick,
				constants.ClobPair_Btc.Status,
				constants.ClobPair_Btc.MakerFeePpm,
				constants.ClobPair_Btc.TakerFeePpm,
			)
			require.NoError(t, err)

			// Create the subaccount.
			subaccount := satypes.Subaccount{
				Id: &satypes.SubaccountId{
					Owner:  "liquidations_test",
					Number: 0,
				},
				AssetPositions:     tc.assetPositions,
				PerpetualPositions: tc.perpetualPositions,
			}
			subaccountsKeeper.SetSubaccount(ctx, subaccount)

			// Initialize the liquidations config.
			if tc.liquidationConfig != nil {
				require.NoError(
					t,
					clobKeeper.InitializeLiquidationsConfig(ctx, *tc.liquidationConfig),
				)
			} else {
				require.NoError(
					t,
					clobKeeper.InitializeLiquidationsConfig(ctx, types.LiquidationsConfig_Default),
				)
			}

			// Run the test and verify expectations.
			liquidationInsuranceFundDeltaBig, err := clobKeeper.GetLiquidationInsuranceFundDelta(
				ctx,
				*subaccount.Id,
				tc.perpetualId,
				tc.isBuy,
				tc.fillAmount,
				tc.subticks,
			)

			if tc.expectedError != nil {
				require.ErrorIs(t, err, tc.expectedError)
			} else {
				require.NoError(t, err)
				require.Equal(
					t,
					tc.expectedLiquidationInsuranceFundDeltaBig.Int64(),
					liquidationInsuranceFundDeltaBig.Int64(),
				)
			}
		})
	}
}

func TestConvertFillablePriceToSubticks(t *testing.T) {
	tests := map[string]struct {
		// Parameters.
		fillablePrice     *big.Rat
		isLiquidatingLong bool
		clobPair          types.ClobPair

		// Expectations.
		expectedSubticks types.Subticks
	}{
		`Converts fillable price to subticks for liquidating a BTC long position`: {
			fillablePrice: big.NewRat(
				int64(constants.FiveBillion),
				1,
			),
			isLiquidatingLong: true,
			clobPair:          constants.ClobPair_Btc,

			expectedSubticks: 500_000_000_000_000_000,
		},
		`Converts fillable price to subticks for liquidating a BTC short position`: {
			fillablePrice: big.NewRat(
				int64(constants.FiveBillion),
				1,
			),
			isLiquidatingLong: false,
			clobPair:          constants.ClobPair_Btc,

			expectedSubticks: 500_000_000_000_000_000,
		},
		`Converts fillable price to subticks for liquidating a long position and rounds up`: {
			fillablePrice: big.NewRat(
				7,
				1,
			),
			isLiquidatingLong: true,
			clobPair: types.ClobPair{
				SubticksPerTick:           100,
				QuantumConversionExponent: 1,
			},

			expectedSubticks: 100,
		},
		`Converts fillable price to subticks for liquidating a short position and rounds down`: {
			fillablePrice: big.NewRat(
				197,
				1,
			),
			isLiquidatingLong: true,
			clobPair: types.ClobPair{
				SubticksPerTick:           100,
				QuantumConversionExponent: 1,
			},

			expectedSubticks: 100,
		},
		`Converts fillable price to subticks for liquidating a short position and rounds down, but
		the result is lower bounded at SubticksPerTick`: {
			fillablePrice: big.NewRat(
				7,
				1,
			),
			isLiquidatingLong: true,
			clobPair: types.ClobPair{
				SubticksPerTick:           100,
				QuantumConversionExponent: 1,
			},

			expectedSubticks: 100,
		},
		`Converts zero fillable price to subticks for liquidating a short position and rounds down,
		but the result is lower bounded at SubticksPerTick`: {
			fillablePrice: big.NewRat(
				0,
				1,
			),
			isLiquidatingLong: true,
			clobPair: types.ClobPair{
				SubticksPerTick:           100,
				QuantumConversionExponent: 1,
			},

			expectedSubticks: 100,
		},
		`Converts fillable price to subticks for liquidating a long position and rounds up, but
		the result is upper bounded at the max Uint64 that is most aligned with SubticksPerTick`: {
			fillablePrice: big_testutil.MustFirst(
				new(big.Rat).SetString("10000000000000000000000"),
			),
			isLiquidatingLong: true,
			clobPair: types.ClobPair{
				SubticksPerTick:           100,
				QuantumConversionExponent: 1,
			},

			expectedSubticks: 18_446_744_073_709_551_600,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup keeper state.
			memClob := memclob.NewMemClobPriceTimePriority(false)
			ctx,
				clobKeeper,
				_,
				_,
				_,
				_,
				_,
				_ := keepertest.ClobKeepers(t, memClob, &mocks.BankKeeper{}, &mocks.IndexerEventManager{})

			// Run the test.
			subticks := clobKeeper.ConvertFillablePriceToSubticks(
				ctx,
				tc.fillablePrice,
				tc.isLiquidatingLong,
				tc.clobPair,
			)
			require.Equal(
				t,
				tc.expectedSubticks.ToBigInt().String(),
				subticks.ToBigInt().String(),
			)
		})
	}
}

func TestConvertFillablePriceToSubticks_PanicsOnNegativeFillablePrice(t *testing.T) {
	// Setup keeper state.
	memClob := memclob.NewMemClobPriceTimePriority(false)
	ctx,
		clobKeeper,
		_,
		_,
		_,
		_,
		_,
		_ := keepertest.ClobKeepers(t, memClob, &mocks.BankKeeper{}, &mocks.IndexerEventManager{})

	// Run the test.
	require.Panics(t, func() {
		clobKeeper.ConvertFillablePriceToSubticks(
			ctx,
			big.NewRat(-1, 1),
			false,
			constants.ClobPair_Btc,
		)
	})
}

func TestGetPerpetualPositionToLiquidate(t *testing.T) {
	tests := map[string]struct {
		// Subaccount state.
		perpetualPositions []*satypes.PerpetualPosition
		// Perpetual state.
		perpetuals []perptypes.Perpetual
		// Clob state.
		liquidationConfig types.LiquidationsConfig
		// CLOB pair state.
		clobPairs []types.ClobPair

		// Expectations.
		expectedClobPair types.ClobPair
		expectedQuantums *big.Int
		expectedError    error
	}{
		`Full position size is returned when subaccount has one perpetual long position`: {
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCLong,
			},
			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},
			liquidationConfig: constants.LiquidationsConfig_No_Limit,

			clobPairs: []types.ClobPair{
				constants.ClobPair_Btc,
			},

			expectedClobPair: constants.ClobPair_Btc,
			expectedQuantums: constants.PerpetualPosition_OneTenthBTCLong.GetBigQuantums(),
		},
		`Full position size is returned when MinPositionNotionalLiquidated is greater than position size`: {
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCLong,
			},
			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},
			liquidationConfig: types.LiquidationsConfig{
				MaxLiquidationFeePpm: 5_000,
				FillablePriceConfig:  constants.FillablePriceConfig_Default,
				PositionBlockLimits: types.PositionBlockLimits{
					MinPositionNotionalLiquidated:   10_000_000_000,
					MaxPositionPortionLiquidatedPpm: lib.OneMillion,
				},
				SubaccountBlockLimits: constants.SubaccountBlockLimits_No_Limit,
			},

			clobPairs: []types.ClobPair{
				constants.ClobPair_Btc,
			},

			expectedClobPair: constants.ClobPair_Btc,
			expectedQuantums: new(big.Int).SetUint64(10_000_000),
		},
		`Half position size is returned when MaxPositionPortionLiquidatedPpm is 500,000`: {
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCLong,
			},
			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},
			liquidationConfig: types.LiquidationsConfig{
				MaxLiquidationFeePpm: 5_000,
				FillablePriceConfig:  constants.FillablePriceConfig_Default,
				PositionBlockLimits: types.PositionBlockLimits{
					MinPositionNotionalLiquidated:   1_000,
					MaxPositionPortionLiquidatedPpm: 500_000,
				},
				SubaccountBlockLimits: constants.SubaccountBlockLimits_No_Limit,
			},

			clobPairs: []types.ClobPair{
				constants.ClobPair_Btc,
			},

			expectedClobPair: constants.ClobPair_Btc,
			expectedQuantums: new(big.Int).SetUint64(5_000_000),
		},
		`Full position is returned when position smaller than subaccount limit`: {
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCLong, // 0.1 BTC, $5,000 notional
			},
			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},
			liquidationConfig: types.LiquidationsConfig{
				MaxLiquidationFeePpm: 5_000,
				FillablePriceConfig:  constants.FillablePriceConfig_Default,
				PositionBlockLimits:  constants.PositionBlockLimits_No_Limit,
				SubaccountBlockLimits: types.SubaccountBlockLimits{
					MaxNotionalLiquidated:    10_000_000_000, // $10,000
					MaxQuantumsInsuranceLost: math.MaxUint64,
				},
			},

			clobPairs: []types.ClobPair{
				constants.ClobPair_Btc,
			},

			expectedClobPair: constants.ClobPair_Btc,
			expectedQuantums: new(big.Int).SetUint64(10_000_000), // 0.1 BTC
		},
		`Max subaccount limit is returned when position larger than subaccount limit`: {
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCLong, // 0.1 BTC, $5,000 notional
			},
			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},
			liquidationConfig: types.LiquidationsConfig{
				MaxLiquidationFeePpm: 5_000,
				FillablePriceConfig:  constants.FillablePriceConfig_Default,
				PositionBlockLimits:  constants.PositionBlockLimits_No_Limit,
				SubaccountBlockLimits: types.SubaccountBlockLimits{
					MaxNotionalLiquidated:    2_500_000_000, // $2,500
					MaxQuantumsInsuranceLost: math.MaxUint64,
				},
			},

			clobPairs: []types.ClobPair{
				constants.ClobPair_Btc,
			},

			expectedClobPair: constants.ClobPair_Btc,
			expectedQuantums: new(big.Int).SetUint64(5_000_000), // 0.05 BTC
		},
		`position size is capped by subaccount block limit when subaccount limit is lower than 
		position block limit`: {
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCLong,
			},
			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},
			liquidationConfig: types.LiquidationsConfig{
				MaxLiquidationFeePpm: 5_000,
				FillablePriceConfig:  constants.FillablePriceConfig_Default,
				PositionBlockLimits: types.PositionBlockLimits{
					MinPositionNotionalLiquidated:   1_000,
					MaxPositionPortionLiquidatedPpm: 500_000,
				},
				SubaccountBlockLimits: types.SubaccountBlockLimits{
					MaxNotionalLiquidated:    2_000_000_000, // $2,000
					MaxQuantumsInsuranceLost: math.MaxUint64,
				},
			},

			clobPairs: []types.ClobPair{
				constants.ClobPair_Btc,
			},

			expectedClobPair: constants.ClobPair_Btc,
			expectedQuantums: new(big.Int).SetUint64(4_000_000), // capped by subaccount block limit
		},
		`position size is capped by position block limit when position limit is lower than 
		subaccount block limit`: {
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCLong,
			},
			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},
			liquidationConfig: types.LiquidationsConfig{
				MaxLiquidationFeePpm: 5_000,
				FillablePriceConfig:  constants.FillablePriceConfig_Default,
				PositionBlockLimits: types.PositionBlockLimits{
					MinPositionNotionalLiquidated:   1_000,
					MaxPositionPortionLiquidatedPpm: 400_000, // 40%
				},
				SubaccountBlockLimits: types.SubaccountBlockLimits{
					MaxNotionalLiquidated:    2_500_000_000, // $2,500
					MaxQuantumsInsuranceLost: math.MaxUint64,
				},
			},

			clobPairs: []types.ClobPair{
				constants.ClobPair_Btc,
			},

			expectedClobPair: constants.ClobPair_Btc,
			expectedQuantums: new(big.Int).SetUint64(4_000_000), // capped by position block limit
		},
		`Result is rounded to nearest step size`: {
			perpetualPositions: []*satypes.PerpetualPosition{
				{
					PerpetualId: 0,
					Quantums:    dtypes.NewInt(21),
				},
			},
			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},
			liquidationConfig: types.LiquidationsConfig{
				MaxLiquidationFeePpm: 5_000,
				FillablePriceConfig:  constants.FillablePriceConfig_Default,
				PositionBlockLimits: types.PositionBlockLimits{
					MinPositionNotionalLiquidated:   1_000,
					MaxPositionPortionLiquidatedPpm: 500_000,
				},
				SubaccountBlockLimits: constants.SubaccountBlockLimits_No_Limit,
			},

			clobPairs: []types.ClobPair{
				{
					Metadata: &types.ClobPair_PerpetualClobMetadata{
						PerpetualClobMetadata: &types.PerpetualClobMetadata{
							PerpetualId: 0,
						},
					},
					Status:                    types.ClobPair_STATUS_ACTIVE,
					StepBaseQuantums:          3, // step size is 3
					SubticksPerTick:           100,
					MinOrderBaseQuantums:      12,
					QuantumConversionExponent: -8,
				},
			},

			expectedClobPair: types.ClobPair{
				Id: 0,
				Metadata: &types.ClobPair_PerpetualClobMetadata{
					PerpetualClobMetadata: &types.PerpetualClobMetadata{
						PerpetualId: 0,
					},
				},
				Status:                    types.ClobPair_STATUS_ACTIVE,
				StepBaseQuantums:          3, // step size is 3
				SubticksPerTick:           100,
				MinOrderBaseQuantums:      12,
				QuantumConversionExponent: -8,
			},
			expectedQuantums: new(big.Int).SetUint64(9), // result is rounded down
		},
		`Full position size is returned when subaccount has one perpetual short position`: {
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneBTCShort,
			},
			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},
			liquidationConfig: constants.LiquidationsConfig_No_Limit,

			clobPairs: []types.ClobPair{
				constants.ClobPair_Btc,
			},

			expectedClobPair: constants.ClobPair_Btc,
			expectedQuantums: constants.PerpetualPosition_OneBTCShort.GetBigQuantums(),
		},
		`Full position size (short) is returned when MinPositionNotionalLiquidated is 
		greater than position size`: {
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCShort,
			},
			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},
			liquidationConfig: types.LiquidationsConfig{
				MaxLiquidationFeePpm: 5_000,
				FillablePriceConfig:  constants.FillablePriceConfig_Default,
				PositionBlockLimits: types.PositionBlockLimits{
					MinPositionNotionalLiquidated:   10_000_000_000,
					MaxPositionPortionLiquidatedPpm: lib.OneMillion,
				},
				SubaccountBlockLimits: constants.SubaccountBlockLimits_No_Limit,
			},

			clobPairs: []types.ClobPair{
				constants.ClobPair_Btc,
			},

			expectedClobPair: constants.ClobPair_Btc,
			expectedQuantums: new(big.Int).SetInt64(-10_000_000),
		},
		`Half position size (short) is returned when MaxPositionPortionLiquidatedPpm is 500,000`: {
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCShort,
			},
			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},
			liquidationConfig: types.LiquidationsConfig{
				MaxLiquidationFeePpm: 5_000,
				FillablePriceConfig:  constants.FillablePriceConfig_Default,
				PositionBlockLimits: types.PositionBlockLimits{
					MinPositionNotionalLiquidated:   1_000,
					MaxPositionPortionLiquidatedPpm: 500_000,
				},
				SubaccountBlockLimits: constants.SubaccountBlockLimits_No_Limit,
			},

			clobPairs: []types.ClobPair{
				constants.ClobPair_Btc,
			},

			expectedClobPair: constants.ClobPair_Btc,
			expectedQuantums: new(big.Int).SetInt64(-5_000_000),
		},
		`Full position (short) is returned when position smaller than subaccount limit`: {
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCShort, // 0.1 BTC, $5,000 notional
			},
			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},
			liquidationConfig: types.LiquidationsConfig{
				MaxLiquidationFeePpm: 5_000,
				FillablePriceConfig:  constants.FillablePriceConfig_Default,
				PositionBlockLimits:  constants.PositionBlockLimits_No_Limit,
				SubaccountBlockLimits: types.SubaccountBlockLimits{
					MaxNotionalLiquidated:    10_000_000_000, // $10,000
					MaxQuantumsInsuranceLost: math.MaxUint64,
				},
			},

			clobPairs: []types.ClobPair{
				constants.ClobPair_Btc,
			},

			expectedClobPair: constants.ClobPair_Btc,
			expectedQuantums: new(big.Int).SetInt64(-10_000_000), // -0.1 BTC
		},
		`Max subaccount limit is returned when short position larger than subaccount limit`: {
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthBTCShort, // 0.1 BTC, $5,000 notional
			},
			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},
			liquidationConfig: types.LiquidationsConfig{
				MaxLiquidationFeePpm: 5_000,
				FillablePriceConfig:  constants.FillablePriceConfig_Default,
				PositionBlockLimits:  constants.PositionBlockLimits_No_Limit,
				SubaccountBlockLimits: types.SubaccountBlockLimits{
					MaxNotionalLiquidated:    2_500_000_000, // $2,500
					MaxQuantumsInsuranceLost: math.MaxUint64,
				},
			},

			clobPairs: []types.ClobPair{
				constants.ClobPair_Btc,
			},

			expectedClobPair: constants.ClobPair_Btc,
			expectedQuantums: new(big.Int).SetInt64(-5_000_000), // -0.05 BTC
		},
		`Full position size of first perpetual is returned when subaccount has multiple perpetual
		positions`: {
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthEthLong,
				&constants.PerpetualPosition_OneTenthBTCLong,
			},
			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
				constants.EthUsd_20PercentInitial_10PercentMaintenance,
			},
			liquidationConfig: constants.LiquidationsConfig_No_Limit,

			clobPairs: []types.ClobPair{
				constants.ClobPair_Btc,
				constants.ClobPair_Eth,
			},

			expectedClobPair: constants.ClobPair_Eth,
			expectedQuantums: constants.PerpetualPosition_OneTenthEthLong.GetBigQuantums(),
		},
		`Full position size and first CLOB pair are returned when subaccount has one long perpetual
		position and multiple CLOB pairs for a perpetual`: {
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneBTCLong,
			},
			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},
			liquidationConfig: constants.LiquidationsConfig_No_Limit,

			//	The definition order matters here
			clobPairs: []types.ClobPair{
				constants.ClobPair_Btc,
				{
					StepBaseQuantums:     5,
					Status:               types.ClobPair_STATUS_ACTIVE,
					SubticksPerTick:      7,
					MinOrderBaseQuantums: 10,
					Metadata: &types.ClobPair_PerpetualClobMetadata{
						PerpetualClobMetadata: &types.PerpetualClobMetadata{
							PerpetualId: constants.PerpetualPosition_OneBTCLong.PerpetualId,
						},
					},
				},
			},

			expectedClobPair: constants.ClobPair_Btc,
			expectedQuantums: constants.PerpetualPosition_OneTenthEthLong.GetBigQuantums(),
		},
		`Full position size of first perpetual and first CLOB pair are returned when subaccount has
		multiple perpetual positions and multiple CLOB pairs for a perpetual`: {
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_OneTenthEthLong,
				&constants.PerpetualPosition_OneBTCShort,
			},
			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
				constants.EthUsd_20PercentInitial_10PercentMaintenance,
			},
			liquidationConfig: constants.LiquidationsConfig_No_Limit,

			//	The definition order matters here
			clobPairs: []types.ClobPair{
				{
					StepBaseQuantums:     5,
					Status:               types.ClobPair_STATUS_ACTIVE,
					SubticksPerTick:      7,
					MinOrderBaseQuantums: 10,
					Metadata: &types.ClobPair_PerpetualClobMetadata{
						PerpetualClobMetadata: &types.PerpetualClobMetadata{
							PerpetualId: constants.PerpetualPosition_OneTenthEthLong.PerpetualId,
						},
					},
				},
				constants.ClobPair_Btc,
				constants.ClobPair_Eth,
			},

			expectedClobPair: types.ClobPair{
				Id:                   0,
				StepBaseQuantums:     5,
				Status:               types.ClobPair_STATUS_ACTIVE,
				SubticksPerTick:      7,
				MinOrderBaseQuantums: 10,
				Metadata: &types.ClobPair_PerpetualClobMetadata{
					PerpetualClobMetadata: &types.PerpetualClobMetadata{
						PerpetualId: constants.PerpetualPosition_OneTenthEthLong.PerpetualId,
					},
				},
			},
			expectedQuantums: constants.PerpetualPosition_OneTenthEthLong.GetBigQuantums(),
		},
		`Full position size of max uint64 of perpetual and CLOB pair are returned when subaccount
		has one long perpetual position at max position size`: {
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_MaxUint64EthLong,
			},
			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
				constants.EthUsd_20PercentInitial_10PercentMaintenance,
			},
			liquidationConfig: constants.LiquidationsConfig_No_Limit,

			clobPairs: []types.ClobPair{
				constants.ClobPair_Btc,
				constants.ClobPair_Eth,
			},

			expectedClobPair: constants.ClobPair_Eth,
			expectedQuantums: new(big.Int).SetUint64(6148914691236517000),
		},
		`Full position size of negated max uint64 of perpetual and CLOB pair are returned when
		subaccount has one short perpetual position at max position size`: {
			perpetualPositions: []*satypes.PerpetualPosition{
				&constants.PerpetualPosition_MaxUint64EthShort,
			},
			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
				constants.EthUsd_20PercentInitial_10PercentMaintenance,
			},
			liquidationConfig: constants.LiquidationsConfig_No_Limit,

			clobPairs: []types.ClobPair{
				constants.ClobPair_Btc,
				constants.ClobPair_Eth,
			},

			expectedClobPair: constants.ClobPair_Eth,
			expectedQuantums: big_testutil.MustFirst(
				new(big.Int).SetString("-6148914691236517000", 10),
			),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup keeper state.
			memClob := memclob.NewMemClobPriceTimePriority(false)
			ctx,
				clobKeeper,
				pricesKeeper,
				_,
				perpetualsKeeper,
				subaccountsKeeper,
				_,
				_ := keepertest.ClobKeepers(t, memClob, &mocks.BankKeeper{}, &mocks.IndexerEventManager{})

			// Create the default markets.
			keepertest.CreateTestMarketsAndExchangeFeeds(t, ctx, pricesKeeper)

			// Create liquidity tiers.
			keepertest.CreateTestLiquidityTiers(t, ctx, perpetualsKeeper)

			// Create all perpetuals.
			for _, p := range tc.perpetuals {
				_, err := perpetualsKeeper.CreatePerpetual(
					ctx,
					p.Ticker,
					p.MarketId,
					p.AtomicResolution,
					p.DefaultFundingPpm,
					p.LiquidityTier,
				)
				require.NoError(t, err)
			}

			// Create the subaccount.
			subaccount := satypes.Subaccount{
				Id: &satypes.SubaccountId{
					Owner:  "liquidations_test",
					Number: 0,
				},
				PerpetualPositions: tc.perpetualPositions,
			}
			subaccountsKeeper.SetSubaccount(ctx, subaccount)

			// Create the CLOB pairs and store the expected CLOB pair.
			for _, clobPair := range tc.clobPairs {
				_, err := clobKeeper.CreatePerpetualClobPair(
					ctx,
					clobtest.MustPerpetualId(clobPair),
					satypes.BaseQuantums(clobPair.StepBaseQuantums),
					satypes.BaseQuantums(clobPair.MinOrderBaseQuantums),
					clobPair.QuantumConversionExponent,
					clobPair.SubticksPerTick,
					clobPair.Status,
					clobPair.MakerFeePpm,
					clobPair.TakerFeePpm,
				)
				require.NoError(t, err)
			}
			// Initialize the liquidations config.
			err := clobKeeper.InitializeLiquidationsConfig(ctx, tc.liquidationConfig)
			require.NoError(t, err)

			clobPair, positionSize, err := clobKeeper.GetPerpetualPositionToLiquidate(
				ctx,
				*subaccount.Id,
			)
			require.ErrorIs(t, err, tc.expectedError)
			require.Equal(
				t,
				tc.expectedQuantums,
				positionSize,
			)
			require.Equal(
				t,
				tc.expectedClobPair,
				clobPair,
			)
		})
	}
}

func TestGetPerpetualPositionToLiquidate_PanicsClobDoesNotExist(t *testing.T) {
	// Setup keeper state.
	memClob := memclob.NewMemClobPriceTimePriority(false)
	ctx,
		clobKeeper,
		_,
		_,
		_,
		subaccountsKeeper,
		_,
		_ := keepertest.ClobKeepers(t, memClob, &mocks.BankKeeper{}, &mocks.IndexerEventManager{})

	// Create the subaccount.
	subaccount := satypes.Subaccount{
		Id: &satypes.SubaccountId{
			Owner:  "liquidations_test",
			Number: 0,
		},
		PerpetualPositions: []*satypes.PerpetualPosition{
			&constants.PerpetualPosition_OneTenthEthLong,
			&constants.PerpetualPosition_OneTenthBTCLong,
		},
	}
	subaccountsKeeper.SetSubaccount(ctx, subaccount)

	require.PanicsWithError(
		t,
		"Perpetual ID 1 has no associated CLOB pairs: The provided perpetual ID does not have "+
			"any associated CLOB pairs",
		func() {
			//nolint: errcheck
			clobKeeper.GetPerpetualPositionToLiquidate(
				ctx,
				*subaccount.Id,
			)
		},
	)
}

func TestGetPerpetualPositionToLiquidate_PanicsClobPairNotInState(t *testing.T) {
	// Setup keeper state.
	memClob := memclob.NewMemClobPriceTimePriority(false)
	ctx,
		clobKeeper,
		_,
		_,
		_,
		subaccountsKeeper,
		_,
		_ := keepertest.ClobKeepers(t, memClob, &mocks.BankKeeper{}, &mocks.IndexerEventManager{})

	// Create the subaccount.
	subaccount := satypes.Subaccount{
		Id: &satypes.SubaccountId{
			Owner:  "liquidations_test",
			Number: 0,
		},
		PerpetualPositions: []*satypes.PerpetualPosition{
			&constants.PerpetualPosition_OneTenthBTCLong,
			&constants.PerpetualPosition_OneTenthEthLong,
		},
	}
	subaccountsKeeper.SetSubaccount(ctx, subaccount)

	// Create the orderbook in the memclob.
	memClob.CreateOrderbook(ctx, constants.ClobPair_Btc)

	require.PanicsWithError(
		t,
		"CLOB pair ID 0 not found in state",
		func() {
			//nolint: errcheck
			clobKeeper.GetPerpetualPositionToLiquidate(
				ctx,
				*subaccount.Id,
			)
		},
	)
}

func TestMaybeLiquidateSubaccount(t *testing.T) {
	tests := map[string]struct {
		// Perpetuals state.
		perpetuals []perptypes.Perpetual
		// Subaccount state.
		subaccounts []satypes.Subaccount
		// CLOB state.
		clobs          []types.ClobPair
		existingOrders []types.Order

		// Parameters.
		liquidatableSubaccount satypes.SubaccountId

		// Expectations.
		expectedPlacedOrders  []*types.MsgPlaceOrder
		expectedMatchedOrders []*types.ClobMatch
	}{
		`Does not place a liquidation order for a non-liquidatable subaccount`: {
			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},
			subaccounts: []satypes.Subaccount{
				constants.Carl_Num0_1BTC_Short,
			},
			clobs: []types.ClobPair{constants.ClobPair_Btc},
			existingOrders: []types.Order{
				constants.Order_Carl_Num0_Id2_Clob0_Buy05BTC_Price50000,
			},

			liquidatableSubaccount: constants.Carl_Num0,

			expectedPlacedOrders:  []*types.MsgPlaceOrder{},
			expectedMatchedOrders: []*types.ClobMatch{},
		},
		`Subaccount liquidation matches no maker orders`: {
			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},
			subaccounts: []satypes.Subaccount{
				constants.Dave_Num0_1BTC_Long_46000USD_Short,
			},
			clobs: []types.ClobPair{constants.ClobPair_Btc},
			existingOrders: []types.Order{
				constants.Order_Carl_Num0_Id2_Clob0_Buy05BTC_Price50000,
			},

			liquidatableSubaccount: constants.Dave_Num0,

			expectedPlacedOrders:  []*types.MsgPlaceOrder{},
			expectedMatchedOrders: []*types.ClobMatch{},
		},
		`Subaccount liquidation matches maker orders`: {
			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},
			subaccounts: []satypes.Subaccount{
				constants.Carl_Num0_1BTC_Short,
				constants.Dave_Num0_1BTC_Long_46000USD_Short,
			},
			clobs: []types.ClobPair{constants.ClobPair_Btc},
			existingOrders: []types.Order{
				constants.Order_Carl_Num0_Id2_Clob0_Buy05BTC_Price50000,
				constants.Order_Carl_Num0_Id3_Clob0_Buy025BTC_Price50000,
				constants.Order_Carl_Num0_Id4_Clob0_Buy05BTC_Price40000,
			},

			liquidatableSubaccount: constants.Dave_Num0,

			expectedPlacedOrders: []*types.MsgPlaceOrder{
				{
					Order: constants.Order_Carl_Num0_Id2_Clob0_Buy05BTC_Price50000,
				},
				{
					Order: constants.Order_Carl_Num0_Id3_Clob0_Buy025BTC_Price50000,
				},
			},
			expectedMatchedOrders: []*types.ClobMatch{
				types.NewClobMatchFromMatchPerpetualLiquidation(
					&types.MatchPerpetualLiquidation{
						ClobPairId:  constants.ClobPair_Btc.Id,
						IsBuy:       false,
						TotalSize:   100_000_000,
						Liquidated:  constants.Dave_Num0,
						PerpetualId: constants.ClobPair_Btc.GetPerpetualClobMetadata().PerpetualId,
						Fills: []types.MakerFill{
							{
								MakerOrderId: types.OrderId{},
								FillAmount:   50_000_000,
							},
							{
								MakerOrderId: types.OrderId{},
								FillAmount:   25_000_000,
							},
						},
					},
				),
			},
		},
		`Does not place liquidation order if subaccount has no perpetual positions to liquidate`: {
			perpetuals: []perptypes.Perpetual{
				constants.BtcUsd_20PercentInitial_10PercentMaintenance,
			},
			subaccounts: []satypes.Subaccount{
				constants.Carl_Num1_Short_500USD,
				constants.Dave_Num0_1BTC_Long_46000USD_Short,
			},
			clobs:          []types.ClobPair{constants.ClobPair_Btc},
			existingOrders: []types.Order{},

			liquidatableSubaccount: constants.Carl_Num0,

			expectedPlacedOrders:  []*types.MsgPlaceOrder{},
			expectedMatchedOrders: []*types.ClobMatch{},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup keeper state.
			memClob := memclob.NewMemClobPriceTimePriority(false)
			mockBankKeeper := &mocks.BankKeeper{}
			mockBankKeeper.On(
				"SendCoinsFromModuleToModule",
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(nil)
			ctx,
				clobKeeper,
				pricesKeeper,
				assetsKeeper,
				perpetualsKeeper,
				subaccountsKeeper,
				_,
				_ := keepertest.ClobKeepers(t, memClob, mockBankKeeper, indexer_manager.NewIndexerEventManagerNoop())
			ctx = ctx.WithIsCheckTx(true)

			// Create the default markets.
			keepertest.CreateTestMarketsAndExchangeFeeds(t, ctx, pricesKeeper)

			// Create liquidity tiers.
			keepertest.CreateTestLiquidityTiers(t, ctx, perpetualsKeeper)

			err := keepertest.CreateUsdcAsset(ctx, assetsKeeper)
			require.NoError(t, err)

			// Create all perpetuals.
			for _, p := range tc.perpetuals {
				_, err := perpetualsKeeper.CreatePerpetual(
					ctx,
					p.Ticker,
					p.MarketId,
					p.AtomicResolution,
					p.DefaultFundingPpm,
					p.LiquidityTier,
				)
				require.NoError(t, err)
			}

			// Create all subaccounts.
			for _, subaccount := range tc.subaccounts {
				subaccountsKeeper.SetSubaccount(ctx, subaccount)
			}

			// Create all CLOBs.
			for _, clobPair := range tc.clobs {
				_, err = clobKeeper.CreatePerpetualClobPair(
					ctx,
					clobtest.MustPerpetualId(clobPair),
					satypes.BaseQuantums(clobPair.StepBaseQuantums),
					satypes.BaseQuantums(clobPair.MinOrderBaseQuantums),
					clobPair.QuantumConversionExponent,
					clobPair.SubticksPerTick,
					clobPair.Status,
					clobPair.MakerFeePpm,
					clobPair.TakerFeePpm,
				)
				require.NoError(t, err)
			}

			// Initialize the liquidations config.
			err = clobKeeper.InitializeLiquidationsConfig(ctx, types.LiquidationsConfig_Default)
			require.NoError(t, err)

			// Create all existing orders.
			for _, order := range tc.existingOrders {
				_, _, err := clobKeeper.CheckTxPlaceOrder(ctx, &types.MsgPlaceOrder{Order: order})
				require.NoError(t, err)
			}

			// Run the test.
			err = clobKeeper.MaybeLiquidateSubaccount(ctx, tc.liquidatableSubaccount)

			// Verify test expectations.
			require.NoError(t, err)
			// TODO(DEC-1979): Refactor these tests to support the operations queue refactor.
			// placedOrders, matchedOrders := memClob.GetPendingFills(ctx)
			// require.Equal(t, tc.expectedPlacedOrders, placedOrders, "Placed orders lists are not equal")
			// require.Equal(t, tc.expectedMatchedOrders, matchedOrders, "Matched orders lists are not equal")
		})
	}
}

func TestGetMaxLiquidatableNotionalAndInsuranceLost(t *testing.T) {
	tests := map[string]struct {
		// Setup
		liquidationConfig               types.LiquidationsConfig
		previouslyLiquidatedPerpetualId uint32
		previousNotionalLiquidated      *big.Int
		previousInsuranceFundLost       *big.Int

		// Expectations.
		panics                          bool
		expectedErr                     error
		expectedMaxNotionalLiquidatable *big.Int
		expectedMaxInsuranceLost        *big.Int
	}{
		"Can get max notional liquidatable and insurance lost": {
			liquidationConfig: types.LiquidationsConfig{
				MaxLiquidationFeePpm: 5_000,
				FillablePriceConfig:  constants.FillablePriceConfig_Default,
				PositionBlockLimits:  constants.PositionBlockLimits_No_Limit,
				SubaccountBlockLimits: types.SubaccountBlockLimits{
					MaxNotionalLiquidated:    150,
					MaxQuantumsInsuranceLost: 150,
				},
			},
			previouslyLiquidatedPerpetualId: uint32(1),
			previousNotionalLiquidated:      big.NewInt(100),
			previousInsuranceFundLost:       big.NewInt(-100),

			expectedMaxNotionalLiquidatable: big.NewInt(50),
			expectedMaxInsuranceLost:        big.NewInt(50),
		},
		"Same perpetual id": {
			liquidationConfig:          constants.LiquidationsConfig_No_Limit,
			previousNotionalLiquidated: big.NewInt(100),
			previousInsuranceFundLost:  big.NewInt(-100),

			expectedErr: types.ErrSubaccountHasLiquidatedPerpetual,
		},
		"Panics on invalid notional liquidated": {
			liquidationConfig: types.LiquidationsConfig{
				MaxLiquidationFeePpm: 5_000,
				FillablePriceConfig:  constants.FillablePriceConfig_Default,
				PositionBlockLimits:  constants.PositionBlockLimits_No_Limit,
				SubaccountBlockLimits: types.SubaccountBlockLimits{
					MaxNotionalLiquidated:    50,
					MaxQuantumsInsuranceLost: math.MaxUint64,
				},
			},
			previouslyLiquidatedPerpetualId: uint32(1),
			previousNotionalLiquidated:      big.NewInt(100),
			previousInsuranceFundLost:       big.NewInt(-100),

			panics:      true,
			expectedErr: types.ErrLiquidationExceedsSubaccountMaxNotionalLiquidated,
		},
		"Panics on invalid insurance lost": {
			liquidationConfig: types.LiquidationsConfig{
				MaxLiquidationFeePpm: 5_000,
				FillablePriceConfig:  constants.FillablePriceConfig_Default,
				PositionBlockLimits:  constants.PositionBlockLimits_No_Limit,
				SubaccountBlockLimits: types.SubaccountBlockLimits{
					MaxNotionalLiquidated:    math.MaxUint64,
					MaxQuantumsInsuranceLost: 50,
				},
			},
			previouslyLiquidatedPerpetualId: uint32(1),
			previousNotionalLiquidated:      big.NewInt(100),
			previousInsuranceFundLost:       big.NewInt(-100),

			panics:      true,
			expectedErr: types.ErrLiquidationExceedsSubaccountMaxInsuranceLost,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			memClob := memclob.NewMemClobPriceTimePriority(false)
			bankMock := &mocks.BankKeeper{}
			ctx,
				clobKeeper,
				_,
				_,
				_,
				_,
				_,
				_ := keepertest.ClobKeepers(t, memClob, bankMock, &mocks.IndexerEventManager{})

			err := clobKeeper.InitializeLiquidationsConfig(ctx, tc.liquidationConfig)
			require.NoError(t, err)

			subaccountId := constants.Alice_Num0
			perpetualId := uint32(0)
			clobKeeper.MustUpdateSubaccountPerpetualLiquidated(
				ctx,
				subaccountId,
				tc.previouslyLiquidatedPerpetualId,
			)
			clobKeeper.UpdateSubaccountLiquidationInfo(
				ctx,
				subaccountId,
				tc.previousNotionalLiquidated,
				tc.previousInsuranceFundLost,
			)

			if tc.panics {
				require.PanicsWithError(
					t,
					tc.expectedErr.Error(),
					func() {
						//nolint: errcheck
						clobKeeper.GetMaxLiquidatableNotionalAndInsuranceLost(
							ctx,
							subaccountId,
							perpetualId,
						)
					},
				)
			} else {
				actualMaxNotionalLiquidatable,
					actualMaxInsuranceLost,
					err := clobKeeper.GetMaxLiquidatableNotionalAndInsuranceLost(
					ctx,
					subaccountId,
					perpetualId,
				)
				if tc.expectedErr != nil {
					require.ErrorContains(t, err, tc.expectedErr.Error())
				} else {
					require.NoError(t, err)
					require.Equal(t, tc.expectedMaxNotionalLiquidatable, actualMaxNotionalLiquidatable)
					require.Equal(t, tc.expectedMaxInsuranceLost, actualMaxInsuranceLost)
				}
			}
		})
	}
}

func TestGetMaxAndMinPositionNotionalLiquidatable(t *testing.T) {
	tests := map[string]struct {
		// Setup
		liquidationConfig   types.LiquidationsConfig
		positionToLiquidate *satypes.PerpetualPosition

		// Expectations.
		expectedErr                        error
		expectedMinPosNotionalLiquidatable *big.Int
		expectedMaxPosNotionalLiquidatable *big.Int
	}{
		"Can get min notional liquidatable": {
			liquidationConfig: types.LiquidationsConfig{
				MaxLiquidationFeePpm: 5_000,
				FillablePriceConfig:  constants.FillablePriceConfig_Default,
				PositionBlockLimits: types.PositionBlockLimits{
					MinPositionNotionalLiquidated:   100,
					MaxPositionPortionLiquidatedPpm: lib.OneMillion,
				},
				SubaccountBlockLimits: constants.SubaccountBlockLimits_No_Limit,
			},
			positionToLiquidate: &satypes.PerpetualPosition{
				PerpetualId: uint32(0),
				Quantums:    dtypes.NewInt(100_000_000), // 1 BTC
			},
			expectedMinPosNotionalLiquidatable: big.NewInt(100),
			expectedMaxPosNotionalLiquidatable: big.NewInt(50_000_000_000), // $50,000
		},
		"Can get max notional liquidatable": {
			liquidationConfig: types.LiquidationsConfig{
				MaxLiquidationFeePpm: 5_000,
				FillablePriceConfig:  constants.FillablePriceConfig_Default,
				PositionBlockLimits: types.PositionBlockLimits{
					MinPositionNotionalLiquidated:   100,
					MaxPositionPortionLiquidatedPpm: 500_000,
				},
				SubaccountBlockLimits: constants.SubaccountBlockLimits_No_Limit,
			},
			positionToLiquidate: &satypes.PerpetualPosition{
				PerpetualId: uint32(0),
				Quantums:    dtypes.NewInt(100_000_000), // 1 BTC
			},
			expectedMinPosNotionalLiquidatable: big.NewInt(100),
			expectedMaxPosNotionalLiquidatable: big.NewInt(25_000_000_000), // $25,000
		},
		"min and max notional liquidatable can be overridden": {
			liquidationConfig: types.LiquidationsConfig{
				MaxLiquidationFeePpm: 5_000,
				FillablePriceConfig:  constants.FillablePriceConfig_Default,
				PositionBlockLimits: types.PositionBlockLimits{
					MinPositionNotionalLiquidated:   10_000_000, // $10
					MaxPositionPortionLiquidatedPpm: lib.OneMillion,
				},
				SubaccountBlockLimits: constants.SubaccountBlockLimits_No_Limit,
			},
			positionToLiquidate: &satypes.PerpetualPosition{
				PerpetualId: uint32(0),
				Quantums:    dtypes.NewInt(10_000), // $5 notional
			},
			expectedMinPosNotionalLiquidatable: big.NewInt(5_000_000), // $5
			expectedMaxPosNotionalLiquidatable: big.NewInt(5_000_000), // $5
		},
		"errors are propagated": {
			liquidationConfig: constants.LiquidationsConfig_No_Limit,
			positionToLiquidate: &satypes.PerpetualPosition{
				PerpetualId: uint32(999), // non-existent
				Quantums:    dtypes.NewInt(1),
			},
			expectedErr: perptypes.ErrPerpetualDoesNotExist,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup keeper state.
			memClob := memclob.NewMemClobPriceTimePriority(false)
			ctx,
				clobKeeper,
				pricesKeeper,
				_,
				perpetualsKeeper,
				_,
				_,
				_ := keepertest.ClobKeepers(t, memClob, &mocks.BankKeeper{}, &mocks.IndexerEventManager{})

			// Create the default markets.
			keepertest.CreateTestMarketsAndExchangeFeeds(t, ctx, pricesKeeper)

			// Create liquidity tiers.
			keepertest.CreateTestLiquidityTiers(t, ctx, perpetualsKeeper)

			// Create perpetual.
			_, err := perpetualsKeeper.CreatePerpetual(
				ctx,
				constants.BtcUsd_100PercentMarginRequirement.Ticker,
				constants.BtcUsd_100PercentMarginRequirement.MarketId,
				constants.BtcUsd_100PercentMarginRequirement.AtomicResolution,
				constants.BtcUsd_100PercentMarginRequirement.DefaultFundingPpm,
				constants.BtcUsd_100PercentMarginRequirement.LiquidityTier,
			)
			require.NoError(t, err)

			// Create all CLOBs.
			_, err = clobKeeper.CreatePerpetualClobPair(
				ctx,
				clobtest.MustPerpetualId(constants.ClobPair_Btc),
				satypes.BaseQuantums(constants.ClobPair_Btc.StepBaseQuantums),
				satypes.BaseQuantums(constants.ClobPair_Btc.MinOrderBaseQuantums),
				constants.ClobPair_Btc.QuantumConversionExponent,
				constants.ClobPair_Btc.SubticksPerTick,
				constants.ClobPair_Btc.Status,
				constants.ClobPair_Btc.MakerFeePpm,
				constants.ClobPair_Btc.TakerFeePpm,
			)
			require.NoError(t, err)

			err = clobKeeper.InitializeLiquidationsConfig(ctx, tc.liquidationConfig)
			require.NoError(t, err)

			actualMinPosNotionalLiquidatable,
				actualMaxPosNotionalLiquidatable,
				err := clobKeeper.GetMaxAndMinPositionNotionalLiquidatable(
				ctx,
				tc.positionToLiquidate,
			)
			if tc.expectedErr != nil {
				require.ErrorContains(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedMinPosNotionalLiquidatable, actualMinPosNotionalLiquidatable)
				require.Equal(t, tc.expectedMaxPosNotionalLiquidatable, actualMaxPosNotionalLiquidatable)
			}
		})
	}
}
