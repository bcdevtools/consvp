package codec

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/bcdevtools/consvp/types"
	"github.com/bcdevtools/consvp/utils"
	"github.com/pkg/errors"
	"sort"
	"strconv"
	"strings"
	"time"
)

//goland:noinspection SpellCheckingInspection

var _ CvpCodec = (*cvpCodecV2)(nil)

const cvpCodecV2Separator byte = '|'

var prefixDataEncodedByCvpCodecV2 = []byte{0x2, cvpCodecV2Separator}

const cvpCodecV2MonikerBufferSize = 20
const cvpCodecV2Base64EncodedMonikerBufferSize = 28

type cvpCodecV2 struct {
}

func getCvpCodecV2() CvpCodec {
	return cvpCodecV2{}
}

func (c cvpCodecV2) EncodeStreamingLightValidators(validators types.StreamingLightValidators) []byte {
	var b bytes.Buffer
	b.Write(prefixDataEncodedByCvpCodecV2)

	for i, v := range validators {
		if i > 0 {
			b.WriteByte(cvpCodecV2Separator)
		}

		if v.Index < 0 {
			panic(fmt.Errorf("invalid validator index: %d, must not be negative", v.Index))
		}
		if v.Index > 998 {
			panic(fmt.Errorf("invalid validator index: %d, must be less than 999", v.Index))
		}
		b.Write(toUint16Buffer(v.Index))

		if v.VotingPowerDisplayPercent < 0 {
			panic(fmt.Errorf("invalid voting power display percent: %f, must not be negative", v.VotingPowerDisplayPercent))
		}
		if v.VotingPowerDisplayPercent > 100 {
			panic(fmt.Errorf("invalid voting power display percent: %f, must not be greater than 100", v.VotingPowerDisplayPercent))
		}
		b.Write(toPercentBuffer(v.VotingPowerDisplayPercent))

		moniker := v.Moniker
		if len(moniker) > 0 {
			bzMoniker := utils.TruncateStringUntilBufferLessThanXBytesOrFillWithSpaceSuffix(moniker, cvpCodecV2MonikerBufferSize)
			b.Write([]byte(base64.StdEncoding.EncodeToString(bzMoniker)))
		}
	}

	return b.Bytes()
}

func (c cvpCodecV2) DecodeStreamingLightValidators(bz []byte) (types.StreamingLightValidators, error) {
	if !bytes.HasPrefix(bz, prefixDataEncodedByCvpCodecV2) {
		return nil, fmt.Errorf("bad encoding prefix")
	}

	var validators types.StreamingLightValidators

	cursor := 1 // skipped first byte as version, starts with separator
	for cursor < len(bz) {
		cursor++

		bzValRawData := takeUntilSeparatorOrEnd(bz, cursor, cvpCodecV2Separator)

		const lengthOmittingMoniker = 2 /*index*/ + 2                                              /*percent*/
		const lengthWithMoniker = lengthOmittingMoniker + cvpCodecV2Base64EncodedMonikerBufferSize /*moniker*/

		if len(bzValRawData) == 0 {
			return nil, fmt.Errorf("invalid empty validator raw data")
		} else if len(bzValRawData) == lengthOmittingMoniker {
			// OK
		} else if len(bzValRawData) == lengthWithMoniker {
			// OK
		} else {
			return nil, fmt.Errorf("invalid validator raw data length %d", len(bzValRawData))
		}

		var validator types.StreamingLightValidator

		bzIndex := mustTakeNBytesFrom(bz, cursor, 2)
		validatorIndex := fromUint16Buffer(bzIndex)
		if validatorIndex < 0 || validatorIndex > 998 {
			return nil, fmt.Errorf("invalid validator index: %d", validatorIndex)
		}
		validator.Index = validatorIndex

		cursor += 2

		bzVotingPowerDisplayPercent := mustTakeNBytesFrom(bz, cursor, 2)
		validator.VotingPowerDisplayPercent = fromPercentBuffer(bzVotingPowerDisplayPercent)
		if validator.VotingPowerDisplayPercent < 0 || validator.VotingPowerDisplayPercent > 100 {
			return nil, fmt.Errorf("invalid voting power display percent: %f", validator.VotingPowerDisplayPercent)
		}

		cursor += 2

		if len(bzValRawData) == lengthWithMoniker {
			bzBase64EncodedOfBzMoniker := takeUntilSeparatorOrEnd(bz, cursor, cvpCodecV2Separator)
			bzMoniker, err := base64.StdEncoding.DecodeString(string(bzBase64EncodedOfBzMoniker))
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("failed to decode base64 encoded moniker: %s", string(bzBase64EncodedOfBzMoniker)))
			}
			validator.Moniker = strings.TrimSpace(sanitizeMoniker(string(bzMoniker)))

			cursor += len(bzBase64EncodedOfBzMoniker)
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
	var b bytes.Buffer
	b.Write(prefixDataEncodedByCvpCodecV2)

	b.Write([]byte(inf.HeightRoundStep))
	b.WriteByte(cvpCodecV2Separator)

	duration := inf.Duration
	if duration < 0 {
		duration = 0
	}
	b.Write([]byte(strconv.Itoa(int(duration.Seconds()))))
	b.WriteByte(cvpCodecV2Separator)

	b.Write(toPercentBuffer(inf.PreVotedPercent))
	b.Write(toPercentBuffer(inf.PreCommitVotedPercent))
	b.WriteByte(cvpCodecV2Separator)

	for _, v := range inf.ValidatorVoteStates {
		if v.ValidatorIndex < 0 {
			panic(fmt.Errorf("invalid validator index: %d, must not be negative", v.ValidatorIndex))
		}
		if v.ValidatorIndex > 998 {
			panic(fmt.Errorf("invalid validator index: %d, must be less than 999", v.ValidatorIndex))
		}
		b.Write(toUint16Buffer(v.ValidatorIndex))

		if len(v.PreVotedBlockHash) == 0 {
			b.Write([]byte("----"))
		} else if len(v.PreVotedBlockHash) != 4 {
			panic(fmt.Errorf("invalid pre-voted fingerprint block hash length: %s, must be 2 bytes", v.PreVotedBlockHash))
		} else {
			b.Write([]byte(v.PreVotedBlockHash))
		}

		if v.PreCommitVoted {
			b.WriteByte('C')
		} else if v.VotedZeroes {
			b.WriteByte('0')
		} else if v.PreVoted {
			b.WriteByte('V')
		} else {
			b.WriteByte('X')
		}
	}

	return b.Bytes()
}

func (c cvpCodecV2) DecodeStreamingNextBlockVotingInformation(bz []byte) (*types.StreamingNextBlockVotingInformation, error) {
	if !bytes.HasPrefix(bz, prefixDataEncodedByCvpCodecV2) {
		return nil, fmt.Errorf("bad encoding prefix")
	}

	var result types.StreamingNextBlockVotingInformation

	var countSeparator int
	for i := 1; i < len(bz); i++ {
		if bz[i] == cvpCodecV2Separator {
			countSeparator++
		}
	}

	if countSeparator != 4 {
		return nil, fmt.Errorf("wrong number of elements")
	}

	cursor := 2 // skipped first byte is version and second byte is separator

	bzHeightRoundStep := takeUntilSeparatorOrEnd(bz, cursor, cvpCodecV2Separator)
	result.HeightRoundStep = string(bzHeightRoundStep)
	if !regexpHeightRoundStep.MatchString(result.HeightRoundStep) {
		return nil, fmt.Errorf("invalid height round step: %s", result.HeightRoundStep)
	}

	cursor += len(bzHeightRoundStep) + 1 /*separator*/

	bzDurationSec := takeUntilSeparatorOrEnd(bz, cursor, cvpCodecV2Separator)
	durationSecStr := string(bzDurationSec)
	durationSec, err := strconv.ParseInt(durationSecStr, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to parse duration sec: %s", durationSecStr))
	}
	if durationSec < 0 {
		return nil, fmt.Errorf("negative duration sec: %d", durationSec)
	}
	result.Duration = time.Duration(durationSec) * time.Second

	cursor += len(bzDurationSec) + 1 /*separator*/

	bzPreVotedAndPreCommitVotedPercent := takeUntilSeparatorOrEnd(bz, cursor, cvpCodecV2Separator)
	if len(bzPreVotedAndPreCommitVotedPercent) != 4 {
		return nil, fmt.Errorf("invalid buffer of pre-voted and pre-commit voted percent length: %d", len(bzPreVotedAndPreCommitVotedPercent))
	}
	bzPreVotedPercent := bzPreVotedAndPreCommitVotedPercent[:2]
	result.PreVotedPercent = fromPercentBuffer(bzPreVotedPercent)
	if result.PreVotedPercent < 0 || result.PreVotedPercent > 100 {
		return nil, fmt.Errorf("invalid pre-voted percent: %f", result.PreVotedPercent)
	}
	bzPreCommitVotedPercent := bzPreVotedAndPreCommitVotedPercent[2:]
	result.PreCommitVotedPercent = fromPercentBuffer(bzPreCommitVotedPercent)
	if result.PreCommitVotedPercent < 0 || result.PreCommitVotedPercent > 100 {
		return nil, fmt.Errorf("invalid pre-commit voted percent: %f", result.PreCommitVotedPercent)
	}
	cursor += 4

	cursor += 1 // separator

	if cursor >= len(bz)-1 {
		return nil, fmt.Errorf("missing validator vote states")
	}

	bzValidatorVoteStates := bz[cursor:]
	if len(bzValidatorVoteStates)%7 != 0 {
		return nil, fmt.Errorf("invalid validator vote states length: %d", len(bzValidatorVoteStates))
	}

	validatorVoteStates := make([]types.StreamingValidatorVoteState, 0)

	cursor = 0 // reset cursor to work on new buffer

	for cursor < len(bzValidatorVoteStates) {
		bzValidatorVoteState := bzValidatorVoteStates[cursor : cursor+7]

		bzValidatorIndex := bzValidatorVoteState[:2]
		validatorIndex := fromUint16Buffer(bzValidatorIndex)
		if validatorIndex < 0 || validatorIndex > 998 {
			return nil, fmt.Errorf("invalid validator index: %d", validatorIndex)
		}

		bzPreVotedBlockHash := bzValidatorVoteState[2:6]
		preVotedBlockHash := string(bzPreVotedBlockHash)
		if preVotedBlockHash != "----" {
			if !regexpPreVotedFingerprintBlockHash.MatchString(preVotedBlockHash) {
				return nil, fmt.Errorf("invalid pre-voted fingerprint block hash: %s, must be 2 bytes", preVotedBlockHash)
			}
		}

		preCommitVoted := false
		votedZeroes := false
		preVoted := false
		voteFlag := bzValidatorVoteState[6]
		switch voteFlag {
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
			return nil, fmt.Errorf("invalid validator vote flag: %s", string(voteFlag))
		}

		validatorVoteStates = append(validatorVoteStates, types.StreamingValidatorVoteState{
			ValidatorIndex:    validatorIndex,
			PreVotedBlockHash: preVotedBlockHash,
			PreVoted:          preVoted,
			VotedZeroes:       votedZeroes,
			PreCommitVoted:    preCommitVoted,
		})

		cursor += 7
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
