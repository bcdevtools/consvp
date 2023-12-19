package consensus_service

//goland:noinspection SpellCheckingInspection
import (
	enginetypes "github.com/bcdevtools/consvp/engine/types"
	"time"
)

type ConsensusService interface {
	// GetNextBlockVotingInformation returns the voting status of validators for the next block.
	//
	// Output voting information is sorted descending by voting power.
	GetNextBlockVotingInformation(lightValidators enginetypes.LightValidators) (sortedValidatorVoteStates []enginetypes.ValidatorVoteState, preVotePercent, preCommitPercent float64, heightRoundStep string, startTimeUTC time.Time, err error)
}
