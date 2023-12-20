package types

import "time"

type StreamingNextBlockVotingInformation struct {
	HeightRoundStep       string
	Duration              time.Duration
	PreVotedPercent       float64
	PreCommitVotedPercent float64
	ValidatorVoteStates   []StreamingValidatorVoteState
}

type StreamingValidatorVoteState struct {
	ValidatorIndex    int
	PreVotedBlockHash string
	PreVoted          bool
	VotedZeroes       bool
	PreCommitVoted    bool
}
