package rpc

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
	// BondedValidators returns the list of bonded validators
	BondedValidators() ([]stakingtypes.Validator, error)

	ConsensusState() (*enginetypes.RoundState, error)

	// Status returns the status of the RPC server
	Status() (*coretypes.ResultStatus, error)

	// LatestValidators returns the most recent validator set.
	//
	// CONTRACT: must maintain the same order as the result from the RPC server.
	LatestValidators() ([]*tmtypes.Validator, error)

	// Shutdown must be called when the RPC client is no longer needed.
	// It does close up all the connections to the RPC server and free resources.
	Shutdown() error
}
