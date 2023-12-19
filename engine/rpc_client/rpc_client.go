package rpc_client

//goland:noinspection SpellCheckingInspection
import (
	enginetypes "github.com/bcdevtools/consvp/engine/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

// RpcClient is the interface that abstract the interaction with the RPC server.
//
//goland:noinspection GoNameStartsWithPackageName
type RpcClient interface {
	// NodeInfo returns upstream RPC server chain id, consensus version and moniker if validator.
	NodeInfo() (chainId, consensusVersion, moniker string)

	// LightValidators returns the list of bonded validators with minimal information needed for application business logic.
	//
	// CONTRACT: must maintain the same order as the result from the RPC server.
	LightValidators() ([]enginetypes.LightValidator, error)

	// BondedValidators returns the list of bonded validators
	BondedValidators() ([]stakingtypes.Validator, error)

	// ConsensusState fetches the current consensus state from the RPC server ':26657/consensus_state'.
	ConsensusState() (*enginetypes.RoundState, error)

	// Status fetches the current status from the RPC server ':26657/status'.
	Status() (*coretypes.ResultStatus, error)

	// LatestValidators returns the most recent validator set from the RPC server ':26657/validators'.
	//
	// CONTRACT: must maintain the same order as the result from the RPC server.
	LatestValidators() ([]*tmtypes.Validator, error)

	// Shutdown must be called when the RPC client is no longer needed.
	// It does close up all the connections to the RPC server and free resources.
	Shutdown() error
}
