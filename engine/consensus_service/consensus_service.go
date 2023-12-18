package consensus_service

//goland:noinspection SpellCheckingInspection
import (
	consstypes "github.com/bcdevtools/consvp/engine/consensus_service/types"
	enginetypes "github.com/bcdevtools/consvp/engine/types"
	"time"
)

type ConsensusService interface {
	// GetNextBlockVotingInformation returns the voting status of validators for the next block.
	//
	// Output voting information is sorted descending by voting power.
	GetNextBlockVotingInformation(lightValidators consstypes.LightValidators) (sortedValidatorVoteStates []enginetypes.ValidatorVoteState, preVotePercent, preCommitPercent float64, heightRoundStep string, startTimeUTC time.Time, err error)

	// LightValidators returns the light list of bonded validators.
	//
	// CONTRACT: must maintain the same order as the result from the RPC server.
	LightValidators() ([]consstypes.LightValidator, error)
}
