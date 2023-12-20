package codec

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/bcdevtools/consvp/types"
	"github.com/pkg/errors"
	"sort"
	"strconv"
	"strings"
	"time"
)

//goland:noinspection SpellCheckingInspection

var _ CvpCodec = (*cvpCodecV2)(nil)

const cvpCodecV2Separator = "|"

const prefixDataEncodedByCvpCodecV2 = "2" + cvpCodecV2Separator

type cvpCodecV2 struct {
}

func getCvpCodecV2() CvpCodec {
	return cvpCodecV2{}
}

func (c cvpCodecV2) EncodeStreamingLightValidators(validators types.StreamingLightValidators) []byte {
	var b strings.Builder
	b.WriteString(prefixDataEncodedByCvpCodecV2)

	for i, v := range validators {
		if i > 0 {
			b.WriteString(cvpCodecV2Separator)
		}

		if v.Index < 0 {
			panic(fmt.Errorf("invalid validator index: %d, must not be negative", v.Index))
		}
		if v.Index > 998 {
			panic(fmt.Errorf("invalid validator index: %d, must be less than 999", v.Index))
		}
		valIdxStr := strconv.Itoa(v.Index)
		for len(valIdxStr) < 3 {
			valIdxStr = "0" + valIdxStr
		}
		b.WriteString(valIdxStr)

		if v.VotingPowerDisplayPercent < 0 {
			panic(fmt.Errorf("invalid voting power display percent: %f, must not be negative", v.VotingPowerDisplayPercent))
		}
		if v.VotingPowerDisplayPercent > 100 {
			panic(fmt.Errorf("invalid voting power display percent: %f, must not be greater than 100", v.VotingPowerDisplayPercent))
		}
		valVpStr := strconv.Itoa(int(v.VotingPowerDisplayPercent * 100))
		for len(valVpStr) < 5 {
			valVpStr = "0" + valVpStr
		}
		b.WriteString(valVpStr)

		moniker := v.Moniker
		if len(moniker) > 0 {
			for len([]byte(moniker)) > 20 && len(moniker) > 1 {
				moniker = moniker[:len(moniker)-1]
			}
			b.WriteString(hex.EncodeToString([]byte(moniker)))
		}
	}

	return []byte(b.String())
}

func (c cvpCodecV2) DecodeStreamingLightValidators(bz []byte) (types.StreamingLightValidators, error) {
	if !bytes.HasPrefix(bz, []byte(prefixDataEncodedByCvpCodecV2)) {
		return nil, fmt.Errorf("bad encoding prefix")
	}

	var validators types.StreamingLightValidators

	spl := strings.Split(string(bz), cvpCodecV2Separator)

	for i := 1; i < len(spl); i++ {
		valRawData := spl[i]

		if len(valRawData) < 3 /*index*/ +5 /*percent x100*/ {
			return nil, fmt.Errorf("validator raw data too short: %s", valRawData)
		}
		if len(valRawData) > 3 /*index*/ +5 /*percent x100*/ +40 /*moniker*/ {
			return nil, fmt.Errorf("validator raw data too long: %s", valRawData)
		}

		var validator types.StreamingLightValidator

		validatorIndex, err := strconv.ParseInt(valRawData[:3], 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to parse validator index: %s", valRawData[:3]))
		}
		if validatorIndex < 0 || validatorIndex > 998 {
			return nil, fmt.Errorf("invalid validator index: %d", validatorIndex)
		}
		validator.Index = int(validatorIndex)

		votingPowerDisplayPercentX100, err := strconv.ParseInt(valRawData[3:8], 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to parse voting power display percent x100: %s", valRawData[3:8]))
		}
		validator.VotingPowerDisplayPercent = float64(votingPowerDisplayPercentX100) / 100
		if validator.VotingPowerDisplayPercent < 0 || validator.VotingPowerDisplayPercent > 100 {
			return nil, fmt.Errorf("invalid voting power display percent: %f", validator.VotingPowerDisplayPercent)
		}

		if len(valRawData) > 8 {
			monikerBytes, err := hex.DecodeString(valRawData[8:])
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("failed to decode moniker: %s", valRawData[8:]))
			}
			moniker := string(monikerBytes)
			moniker = strings.ReplaceAll(moniker, "<", "(")
			moniker = strings.ReplaceAll(moniker, ">", ")")
			moniker = strings.ReplaceAll(moniker, "'", "`")
			moniker = strings.ReplaceAll(moniker, "\"", "`")
			validator.Moniker = moniker
		}

		validators = append(validators, validator)
	}

	sort.Slice(validators, func(i, j int) bool {
		return validators[i].Index < validators[j].Index
	})
	for i, v := range validators {
		if v.Index != i {
			return nil, fmt.Errorf("invalid validator index sequence, %d at %d", v.Index, i)
		}
	}

	return validators, nil
}

func (c cvpCodecV2) EncodeStreamingNextBlockVotingInformation(inf *types.StreamingNextBlockVotingInformation) []byte {
	var b strings.Builder
	b.WriteString(prefixDataEncodedByCvpCodecV2)

	b.WriteString(inf.HeightRoundStep)
	b.WriteString(cvpCodecV2Separator)

	duration := inf.Duration
	if duration < 0 {
		b.WriteString("0")
	} else {
		b.WriteString(strconv.Itoa(int(duration.Milliseconds())))
	}
	b.WriteString(cvpCodecV2Separator)

	b.WriteString(strconv.Itoa(int(inf.PreVotedPercent * 100)))
	b.WriteString(cvpCodecV2Separator)

	b.WriteString(strconv.Itoa(int(inf.PreCommitVotedPercent * 100)))
	b.WriteString(cvpCodecV2Separator)

	for _, v := range inf.ValidatorVoteStates {
		if v.ValidatorIndex < 0 {
			panic(fmt.Errorf("invalid validator index: %d, must not be negative", v.ValidatorIndex))
		}
		if v.ValidatorIndex > 998 {
			panic(fmt.Errorf("invalid validator index: %d, must be less than 999", v.ValidatorIndex))
		}
		valIdxStr := strconv.Itoa(v.ValidatorIndex)
		for len(valIdxStr) < 3 {
			valIdxStr = "0" + valIdxStr
		}
		b.WriteString(valIdxStr)

		if len(v.PreVotedBlockHash) == 0 {
			b.WriteString("----")
		} else if len(v.PreVotedBlockHash) != 4 {
			panic(fmt.Errorf("invalid pre-voted fingerprint block hash length: %s, must be 2 bytes", v.PreVotedBlockHash))
		} else {
			b.WriteString(v.PreVotedBlockHash)
		}

		if v.PreCommitVoted {
			b.WriteString("C")
		} else if v.VotedZeroes {
			b.WriteString("0")
		} else if v.PreVoted {
			b.WriteString("V")
		} else {
			b.WriteString("X")
		}
	}

	return []byte(b.String())
}

func (c cvpCodecV2) DecodeStreamingNextBlockVotingInformation(bz []byte) (*types.StreamingNextBlockVotingInformation, error) {
	if !bytes.HasPrefix(bz, []byte(prefixDataEncodedByCvpCodecV2)) {
		return nil, fmt.Errorf("bad encoding prefix")
	}

	data := strings.ToUpper(string(bz))

	var result types.StreamingNextBlockVotingInformation

	spl := strings.Split(data, cvpCodecV2Separator)
	if len(spl) != 6 {
		return nil, fmt.Errorf("wrong number of elements")
	}

	result.HeightRoundStep = spl[1]
	if !regexpHeightRoundStep.MatchString(result.HeightRoundStep) {
		return nil, fmt.Errorf("invalid height round step: %s", result.HeightRoundStep)
	}

	durationMs, err := strconv.ParseInt(spl[2], 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to parse duration ms: %s", spl[2]))
	}
	if durationMs < 0 {
		return nil, fmt.Errorf("negative duration ms: %d", durationMs)
	}
	result.Duration = time.Duration(durationMs) * time.Millisecond

	preVotedPercentX100, err := strconv.ParseInt(spl[3], 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to parse pre-voted percent x100: %s", spl[3]))
	}
	result.PreVotedPercent = float64(preVotedPercentX100) / 100
	if result.PreVotedPercent < 0 || result.PreVotedPercent > 100 {
		return nil, fmt.Errorf("invalid pre-voted percent: %f", result.PreVotedPercent)
	}

	preCommitVotedPercentX100, err := strconv.ParseInt(spl[4], 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to parse pre-commit voted percent x100: %s", spl[4]))
	}
	result.PreCommitVotedPercent = float64(preCommitVotedPercentX100) / 100
	if result.PreCommitVotedPercent < 0 || result.PreCommitVotedPercent > 100 {
		return nil, fmt.Errorf("invalid pre-commit voted percent: %f", result.PreCommitVotedPercent)
	}

	validatorVoteStates := make([]types.StreamingValidatorVoteState, 0)
	validatorVoteStatesStr := spl[5]
	if len(validatorVoteStatesStr) < 1 {
		return nil, fmt.Errorf("missing validator vote states")
	}
	if len(validatorVoteStatesStr)%8 != 0 {
		return nil, fmt.Errorf("invalid validator vote states length: %d", len(validatorVoteStatesStr))
	}
	var cursor int
	for cursor < len(validatorVoteStatesStr) {
		validatorIndex, err := strconv.ParseInt(validatorVoteStatesStr[cursor:cursor+3], 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to parse validator index: %s", validatorVoteStatesStr[cursor:cursor+3]))
		}
		if validatorIndex < 0 || validatorIndex > 998 {
			return nil, fmt.Errorf("invalid validator index: %d", validatorIndex)
		}
		cursor += 3

		preVotedBlockHash := validatorVoteStatesStr[cursor : cursor+4]
		if preVotedBlockHash != "----" {
			if !regexpPreVotedFingerprintBlockHash.MatchString(preVotedBlockHash) {
				return nil, fmt.Errorf("invalid pre-voted fingerprint block hash: %s, must be 2 bytes", preVotedBlockHash)
			}
		}
		cursor += 4

		preCommitVoted := false
		votedZeroes := false
		preVoted := false
		switch validatorVoteStatesStr[cursor] {
		case 'C':
			preCommitVoted = true
			preVoted = true
		case '0':
			votedZeroes = true
			preVoted = true
		case 'V':
			preVoted = true
		case 'X':
		default:
			return nil, fmt.Errorf("invalid validator vote state: %s", string(validatorVoteStatesStr[cursor]))
		}
		cursor++

		validatorVoteStates = append(validatorVoteStates, types.StreamingValidatorVoteState{
			ValidatorIndex:    int(validatorIndex),
			PreVotedBlockHash: preVotedBlockHash,
			PreVoted:          preVoted,
			VotedZeroes:       votedZeroes,
			PreCommitVoted:    preCommitVoted,
		})
	}
	sort.Slice(validatorVoteStates, func(i, j int) bool {
		return validatorVoteStates[i].ValidatorIndex < validatorVoteStates[j].ValidatorIndex
	})
	for i, state := range validatorVoteStates {
		if state.ValidatorIndex != i {
			return nil, fmt.Errorf("invalid validator index sequence, %d at %d", state.ValidatorIndex, i)
		}
	}
	result.ValidatorVoteStates = validatorVoteStates

	return &result, nil
}
