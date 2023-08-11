package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/dydxprotocol/v4/testutil/keeper"
	"github.com/dydxprotocol/v4/testutil/nullify"
	"github.com/dydxprotocol/v4/x/perpetuals/types"
)

func TestPerpetualQuerySingle(t *testing.T) {
	ctx, keeper, pricesKeeper, _, _ := keepertest.PerpetualsKeepers(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs, err := createLiquidityTiersAndNPerpetuals(t, ctx, keeper, pricesKeeper, 2)
	require.NoError(t, err)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryPerpetualRequest
		response *types.QueryPerpetualResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryPerpetualRequest{
				Id: msgs[0].Id,
			},
			response: &types.QueryPerpetualResponse{Perpetual: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryPerpetualRequest{
				Id: msgs[1].Id,
			},
			response: &types.QueryPerpetualResponse{Perpetual: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryPerpetualRequest{
				Id: uint32(100000),
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.Perpetual(wctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t,
					nullify.Fill(tc.response), //nolint:staticcheck
					nullify.Fill(response),    //nolint:staticcheck
				)
			}
		})
	}
}

func TestPerpetualQueryPaginated(t *testing.T) {
	ctx, keeper, pricesKeeper, _, _ := keepertest.PerpetualsKeepers(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs, err := createLiquidityTiersAndNPerpetuals(t, ctx, keeper, pricesKeeper, 5)
	require.NoError(t, err)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllPerpetualsRequest {
		return &types.QueryAllPerpetualsRequest{
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
			},
		}
	}
	t.Run("ByOffset", func(t *testing.T) {
		step := 2
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.AllPerpetuals(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Perpetual), step)
			require.Subset(t,
				nullify.Fill(msgs),           //nolint:staticcheck
				nullify.Fill(resp.Perpetual), //nolint:staticcheck
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.AllPerpetuals(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Perpetual), step)
			require.Subset(t,
				nullify.Fill(msgs),           //nolint:staticcheck
				nullify.Fill(resp.Perpetual), //nolint:staticcheck
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.AllPerpetuals(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),           //nolint:staticcheck
			nullify.Fill(resp.Perpetual), //nolint:staticcheck
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.AllPerpetuals(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
