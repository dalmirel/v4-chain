package subaccounts

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dydxprotocol/v4/x/subaccounts/keeper"
	"github.com/dydxprotocol/v4/x/subaccounts/types"
)

// InitGenesis initializes the subaccounts module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Set all the subaccounts
	for _, elem := range genState.Subaccounts {
		k.SetSubaccount(ctx, elem)
	}
}

// ExportGenesis returns the subaccounts module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()

	genesis.Subaccounts = k.GetAllSubaccount(ctx)

	return genesis
}
