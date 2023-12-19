package types

import "time"

type NextBlockVotingInformation struct {
	SortedValidatorVoteStates []ValidatorVoteState
	PreVotePercent            float64
	PreCommitPercent          float64
	HeightRoundStep           string
	StartTimeUTC              time.Time
}
