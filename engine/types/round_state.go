package types

import (
	"fmt"
	"github.com/pkg/errors"
	"strconv"
	"strings"
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

func (rs RoundState) GetRound() (round int, err error) {
	spl := strings.Split(rs.HeightRoundStep, "/")
	if len(spl) >= 2 {
		round, err = strconv.Atoi(spl[1])
		if err != nil {
			round = 0
			err = errors.Wrap(err, fmt.Sprintf("failed to parse round %s", spl[1]))
		}
	} else {
		err = fmt.Errorf("no round for current height %s", rs.HeightRoundStep)
	}
	return
}

func (rs RoundState) GetPreVotePercent(round int) (percent float64, err error) {
	bitArray := strings.Split(rs.Votes[round].PreVotesBitArray, " ")
	if len(bitArray) >= 3 {
		finalBitArray := bitArray[len(bitArray)-1]
		percent, err = strconv.ParseFloat(finalBitArray, 64)
		if err != nil {
			percent = 0.0
			err = errors.Wrap(err, fmt.Sprintf("failed to parse pre-vote percent %s", finalBitArray))
		} else {
			percent = percent * 100
		}
	} else {
		err = fmt.Errorf("invalid pre-vote bit array [%s] for round %d", bitArray, round)
	}
	return
}

func (rs RoundState) GetPreCommitPercent(round int) (percent float64, err error) {
	bitArray := strings.Split(rs.Votes[round].PreCommitsBitArray, " ")
	if len(bitArray) >= 3 {
		finalBitArray := bitArray[len(bitArray)-1]
		percent, err = strconv.ParseFloat(finalBitArray, 64)
		if err != nil {
			percent = 0.0
			err = errors.Wrap(err, fmt.Sprintf("failed to parse pre-commit percent %s", finalBitArray))
		} else {
			percent = percent * 100
		}
	} else {
		err = fmt.Errorf("invalid pre-commit bit array [%s] for round %d", bitArray, round)
	}
	return
}

//goland:noinspection SpellCheckingInspection
type RoundVotes struct {
	PreVotes           []string `json:"prevotes"`
	PreVotesBitArray   string   `json:"prevotes_bit_array"`
	PreCommits         []string `json:"precommits"`
	PreCommitsBitArray string   `json:"precommits_bit_array"`
}
