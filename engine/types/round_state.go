package types

import (
	"time"
)

type RoundStateResponse struct {
	RoundState *RoundState `json:"round_state"`
}

type RoundState struct {
	HeightRoundStep string       `json:"height/round/step"`
	StartTime       time.Time    `json:"start_time"`
	Votes           []RoundVotes `json:"height_vote_set"`
}

//goland:noinspection SpellCheckingInspection
type RoundVotes struct {
	PreVotes           []string `json:"prevotes"`
	PreVotesBitArray   string   `json:"prevotes_bit_array"`
	PreCommits         []string `json:"precommits"`
	PreCommitsBitArray string   `json:"precommits_bit_array"`
}
