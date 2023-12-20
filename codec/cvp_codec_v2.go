package codec

import (
	"bytes"
	"encoding/hex"
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
			monikerBz := utils.TruncateStringUntilBufferLessThanXBytesOrFillWithSpaceSuffix(moniker, 20)
			b.Write(monikerBz)
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
		if bz[cursor] != cvpCodecV2Separator {
			return nil, fmt.Errorf("expect separator at %d, got 0x%s", cursor, hex.EncodeToString([]byte{bz[cursor]}))
		}

		cursor++

		valRawDataBz := takeUntilSeparatorOrEnd(bz, cursor, cvpCodecV2Separator)

		const lengthOmittingMoniker = 2 /*index*/ + 2        /*percent*/
		const lengthWithMoniker = lengthOmittingMoniker + 20 /*moniker*/

		if len(valRawDataBz) < lengthOmittingMoniker {
			return nil, fmt.Errorf("validator raw data too short: %d", len(valRawDataBz))
		}
		if len(valRawDataBz) > lengthWithMoniker {
			return nil, fmt.Errorf("validator raw data too long: %d", len(valRawDataBz))
		}

		var validator types.StreamingLightValidator

		bzIndex, takeSuccess := tryTakeNBytesFrom(bz, cursor, 2)
		if !takeSuccess {
			return nil, fmt.Errorf("failed to take validator index from %d of buffer", cursor)
		}
		if bytes.ContainsRune(bzIndex, rune(cvpCodecV2Separator)) {
			return nil, fmt.Errorf("validator raw data too short, detected separator at buffer of validator index")
		}

		validatorIndex := fromUint16Buffer(bzIndex)
		if validatorIndex < 0 || validatorIndex > 998 {
			return nil, fmt.Errorf("invalid validator index: %d", validatorIndex)
		}
		validator.Index = validatorIndex

		cursor += 2

		bzVotingPowerDisplayPercent, takeSuccess := tryTakeNBytesFrom(bz, cursor, 2)
		if !takeSuccess {
			return nil, fmt.Errorf("failed to take voting power display percent from %d of buffer", cursor)
		}
		if bytes.ContainsRune(bzVotingPowerDisplayPercent, rune(cvpCodecV2Separator)) {
			return nil, fmt.Errorf("validator raw data too short, detected separator at buffer of voting power display percent")
		}
		validator.VotingPowerDisplayPercent = fromPercentBuffer(bzVotingPowerDisplayPercent)
		if validator.VotingPowerDisplayPercent < 0 || validator.VotingPowerDisplayPercent > 100 {
			return nil, fmt.Errorf("invalid voting power display percent: %f", validator.VotingPowerDisplayPercent)
		}

		cursor += 2

		if len(valRawDataBz) == lengthOmittingMoniker {
			// no moniker
		} else if len(valRawDataBz) == lengthWithMoniker {
			monikerBz := takeUntilSeparatorOrEnd(bz, cursor, cvpCodecV2Separator)
			validator.Moniker = strings.TrimSpace(sanitizeMoniker(string(monikerBz)))

			cursor += len(monikerBz)
		} else {
			return nil, fmt.Errorf("invalid validator raw data length %d: %s", len(valRawDataBz), valRawDataBz)
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

	cursor += len(bzHeightRoundStep)
	if bz[cursor] != cvpCodecV2Separator {
		return nil, fmt.Errorf("expect separator at %d, got 0x%s", cursor, hex.EncodeToString([]byte{bz[cursor]}))
	}
	cursor += 1

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

	cursor += len(bzDurationSec)
	if bz[cursor] != cvpCodecV2Separator {
		return nil, fmt.Errorf("expect separator at %d, got 0x%s", cursor, hex.EncodeToString([]byte{bz[cursor]}))
	}
	cursor += 1

	bzPreVotedPercent, takeSuccess := tryTakeNBytesFrom(bz, cursor, 2)
	if !takeSuccess {
		return nil, fmt.Errorf("failed to take pre-voted percent from %d of buffer", cursor)
	}
	if bytes.ContainsRune(bzPreVotedPercent, rune(cvpCodecV2Separator)) {
		return nil, fmt.Errorf("pre-voted percent raw too short, detected separator at buffer of pre-voted percent")
	}
	result.PreVotedPercent = fromPercentBuffer(bzPreVotedPercent)
	if result.PreVotedPercent < 0 || result.PreVotedPercent > 100 {
		return nil, fmt.Errorf("invalid pre-voted percent: %f", result.PreVotedPercent)
	}
	cursor += 2

	bzPreCommitVotedPercent, takeSuccess := tryTakeNBytesFrom(bz, cursor, 2)
	if !takeSuccess {
		return nil, fmt.Errorf("failed to take pre-commit voted percent from %d of buffer", cursor)
	}
	if bytes.ContainsRune(bzPreCommitVotedPercent, rune(cvpCodecV2Separator)) {
		return nil, fmt.Errorf("pre-commit voted percent raw too short, detected separator at buffer of pre-commit voted percent")
	}
	result.PreCommitVotedPercent = fromPercentBuffer(bzPreCommitVotedPercent)
	if result.PreCommitVotedPercent < 0 || result.PreCommitVotedPercent > 100 {
		return nil, fmt.Errorf("invalid pre-commit voted percent: %f", result.PreCommitVotedPercent)
	}
	cursor += 2

	if bz[cursor] != cvpCodecV2Separator {
		return nil, fmt.Errorf("expect separator at %d, got 0x%s", cursor, hex.EncodeToString([]byte{bz[cursor]}))
	}
	cursor += 1

	if cursor >= len(bz)-1 {
		return nil, fmt.Errorf("missing validator vote states")
	}

	validatorVoteStatesBz := bz[cursor:]
	if len(validatorVoteStatesBz)%7 != 0 {
		return nil, fmt.Errorf("invalid validator vote states length: %d", len(validatorVoteStatesBz))
	}

	validatorVoteStates := make([]types.StreamingValidatorVoteState, 0)

	cursor = 0 // reset cursor to work on new buffer

	for cursor < len(validatorVoteStatesBz) {
		bzValidatorIndex, takeSuccess := tryTakeNBytesFrom(validatorVoteStatesBz, cursor, 2)
		if !takeSuccess {
			return nil, fmt.Errorf("failed to take validator index from %d of buffer", cursor)
		}
		validatorIndex := fromUint16Buffer(bzValidatorIndex)
		if validatorIndex < 0 || validatorIndex > 998 {
			return nil, fmt.Errorf("invalid validator index: %d", validatorIndex)
		}
		cursor += 2

		bzPreVotedBlockHash, takeSuccess := tryTakeNBytesFrom(validatorVoteStatesBz, cursor, 4)
		if !takeSuccess {
			return nil, fmt.Errorf("failed to take pre-voted block hash from %d of buffer", cursor)
		}
		preVotedBlockHash := string(bzPreVotedBlockHash)
		if preVotedBlockHash != "----" {
			if !regexpPreVotedFingerprintBlockHash.MatchString(preVotedBlockHash) {
				return nil, fmt.Errorf("invalid pre-voted fingerprint block hash: %s, must be 2 bytes", preVotedBlockHash)
			}
		}
		cursor += 4

		preCommitVoted := false
		votedZeroes := false
		preVoted := false
		switch validatorVoteStatesBz[cursor] {
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
			return nil, fmt.Errorf("invalid validator vote state: %s", string(validatorVoteStatesBz[cursor]))
		}
		cursor++

		validatorVoteStates = append(validatorVoteStates, types.StreamingValidatorVoteState{
			ValidatorIndex:    validatorIndex,
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
