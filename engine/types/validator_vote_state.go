package types

//goland:noinspection SpellCheckingInspection
import (
	consstypes "github.com/bcdevtools/consvp/engine/consensus_service/types"
)

type ValidatorVoteState struct {
	Validator      consstypes.LightValidator
	PreVoted       bool
	VotedZeroes    bool
	PreCommitVoted bool
}
