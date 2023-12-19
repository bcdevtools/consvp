package consensus_service

//goland:noinspection SpellCheckingInspection
import (
	enginetypes "github.com/bcdevtools/consvp/engine/types"
)

type ConsensusService interface {
	// GetNextBlockVotingInformation returns the voting status of validators for the next block.
	//
	// Output voting information is sorted descending by voting power.
	GetNextBlockVotingInformation(lightValidators enginetypes.LightValidators) (nextBlockVotingInfo *enginetypes.NextBlockVotingInformation, err error)

	// Shutdown must be called when the service is no longer needed.
	Shutdown() error
}
