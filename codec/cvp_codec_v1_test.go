package codec

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/bcdevtools/consvp/constants"
	"github.com/bcdevtools/consvp/types"
	"reflect"
	"strings"
	"testing"
	"time"
)

var cvpV1CodecImpl = getCvpCodecV1()

func Test_cvpCodecV1_EncodeDecodeStreamingLightValidators(t *testing.T) {
	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name                               string
		validators                         types.StreamingLightValidators
		wantPanicEncode                    bool
		wantEncodedData                    []byte
		wantErrDecode                      bool
		wantErrDecodeContains              string
		wantDecodedOrUseInputAsWantDecoded types.StreamingLightValidators // if missing, use input as expect
	}{
		{
			name: "normal, 2 validators",
			validators: []types.StreamingLightValidator{
				{
					Index:                     0,
					VotingPowerDisplayPercent: 10.10,
					Moniker:                   "Val1",
				},
				{
					Index:                     1,
					VotingPowerDisplayPercent: 01.02,
					Moniker:                   "Val2",
				},
			},
			wantPanicEncode: false,
			wantEncodedData: []byte("1|00001010" + hex.EncodeToString(fssut("Val1", 20)) + "|00100102" + hex.EncodeToString(fssut("Val2", 20))),
			wantErrDecode:   false,
		},
		{
			name: "normal, 1 validator",
			validators: []types.StreamingLightValidator{
				{
					Index:                     0,
					VotingPowerDisplayPercent: 10.10,
					Moniker:                   "Val1",
				},
			},
			wantPanicEncode: false,
			wantEncodedData: []byte("1|00001010" + hex.EncodeToString(fssut("Val1", 20))),
			wantErrDecode:   false,
		},
		{
			name: "truncate before encode then decode correct moniker UTF-8",
			validators: []types.StreamingLightValidator{
				{
					Index:                     0,
					VotingPowerDisplayPercent: 10.10,
					Moniker:                   "✅✅✅✅✅✅✅✅✅✅✅✅✅✅✅✅✅✅✅✅",
				},
				{
					Index:                     1,
					VotingPowerDisplayPercent: 01.02,
					Moniker:                   "❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌❌",
				},
			},
			wantPanicEncode: false,
			wantErrDecode:   false,
			wantDecodedOrUseInputAsWantDecoded: []types.StreamingLightValidator{
				// moniker of validators are truncated to max 20 bytes of runes
				{
					Index:                     0,
					VotingPowerDisplayPercent: 10.10,
					Moniker:                   "✅✅✅✅✅✅",
				},
				{
					Index:                     1,
					VotingPowerDisplayPercent: 01.02,
					Moniker:                   "❌❌❌❌❌❌",
				},
			},
		},
		{
			name: "normal, validator with 100% VP",
			validators: []types.StreamingLightValidator{
				{
					Index:                     0,
					VotingPowerDisplayPercent: 100,
					Moniker:                   "Val1",
				},
			},
			wantPanicEncode: false,
			wantEncodedData: []byte("1|00010000" + hex.EncodeToString(fssut("Val1", 20))),
			wantErrDecode:   false,
		},
		{
			name:                  "not accept empty validator list",
			validators:            []types.StreamingLightValidator{},
			wantPanicEncode:       false,
			wantEncodedData:       []byte("1|"),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid empty validator raw data",
		},
		{
			name: "not accept validator negative index",
			validators: []types.StreamingLightValidator{
				{
					Index:                     -1,
					VotingPowerDisplayPercent: 99,
					Moniker:                   "Val1",
				},
			},
			wantPanicEncode: true,
		},
		{
			name: "not accept validator index greater than 998",
			validators: []types.StreamingLightValidator{
				{
					Index:                     999,
					VotingPowerDisplayPercent: 99,
					Moniker:                   "Val1",
				},
			},
			wantPanicEncode: true,
		},
		{
			name: "not accept validator negative voting power percent",
			validators: []types.StreamingLightValidator{
				{
					Index:                     0,
					VotingPowerDisplayPercent: -0.01,
					Moniker:                   "Val1",
				},
			},
			wantPanicEncode: true,
		},
		{
			name: "not accept validator voting power percent greater than 100%",
			validators: []types.StreamingLightValidator{
				{
					Index:                     0,
					VotingPowerDisplayPercent: 100.01,
					Moniker:                   "Val1",
				},
			},
			wantPanicEncode: true,
		},
		{
			name: "validator list size larger than cap",
			validators: func() types.StreamingLightValidators {
				var validators types.StreamingLightValidators
				for v := 1; v <= constants.MAX_VALIDATORS+1; v++ {
					validators = append(validators, types.StreamingLightValidator{
						Index:                     v - 1,
						VotingPowerDisplayPercent: 99,
						Moniker:                   fmt.Sprintf("Val%d", v),
					})
				}
				return validators
			}(),
			wantPanicEncode: true,
		},
		{
			name: "keep only first 20 bytes of moniker",
			validators: []types.StreamingLightValidator{
				{
					Index:                     0,
					VotingPowerDisplayPercent: 99,
					Moniker:                   "123456789012345678901234567890",
				},
			},
			wantPanicEncode: false,
			wantEncodedData: []byte("1|00009900" + hex.EncodeToString([]byte("12345678901234567890"))),
			wantDecodedOrUseInputAsWantDecoded: []types.StreamingLightValidator{
				{
					Index:                     0,
					VotingPowerDisplayPercent: 99,
					Moniker:                   "12345678901234567890", // truncated
				},
			},
			wantErrDecode: false,
		},
		{
			name: "sanitize moniker",
			validators: []types.StreamingLightValidator{
				{
					Index:                     0,
					VotingPowerDisplayPercent: 99,
					Moniker:                   `<he'llo">`,
				},
			},
			wantPanicEncode: false,
			wantEncodedData: []byte("1|00009900" + hex.EncodeToString(fssut(`<he'llo">`, 20))),
			wantDecodedOrUseInputAsWantDecoded: []types.StreamingLightValidator{
				{
					Index:                     0,
					VotingPowerDisplayPercent: 99,
					Moniker:                   "(he`llo`)",
				},
			},
			wantErrDecode: false,
		},
		{
			name: "collision of separator byte and bytes index",
			validators: func() types.StreamingLightValidators {
				var result types.StreamingLightValidators
				for i := 0; i < constants.MAX_VALIDATORS; i++ {
					result = append(result, types.StreamingLightValidator{
						Index:                     i,
						VotingPowerDisplayPercent: 99,
						Moniker:                   fmt.Sprintf("Val%d", i+1),
					})
				}
				return result
			}(),
			wantPanicEncode: false,
			wantErrDecode:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEncoded := func() (bz []byte) {
				defer func() {
					err := recover()
					if err != nil {
						if !tt.wantPanicEncode {
							t.Errorf("EncodeStreamingLightValidators() panic = %v but not wanted", err)
						}
					} else {
						if tt.wantPanicEncode {
							t.Errorf("EncodeStreamingLightValidators() panic = %v but wanted panic", err)
						}
					}
				}()
				bz = cvpV1CodecImpl.EncodeStreamingLightValidators(tt.validators)
				return
			}()

			if tt.wantPanicEncode {
				return
			}

			if len(tt.wantEncodedData) > 0 {
				if !bytes.Equal(gotEncoded, tt.wantEncodedData) {
					t.Errorf("EncodeStreamingLightValidators()\ngotEncoded = %v\nwant %v", string(gotEncoded), string(tt.wantEncodedData))
					return
				}
			}

			gotDecoded, err := cvpV1CodecImpl.DecodeStreamingLightValidators(gotEncoded)
			if (err != nil) != tt.wantErrDecode {
				t.Errorf("DecodeStreamingLightValidators() error = %v, wantErr %v", err, tt.wantErrDecode)
				return
			}
			if err == nil {
				if tt.wantDecodedOrUseInputAsWantDecoded == nil {
					tt.wantDecodedOrUseInputAsWantDecoded = tt.validators
				}
				if !reflect.DeepEqual(gotDecoded, tt.wantDecodedOrUseInputAsWantDecoded) {
					t.Errorf("DecodeStreamingLightValidators()\ngot = %v,\nwant %v", gotDecoded, tt.wantDecodedOrUseInputAsWantDecoded)
				}
			} else {
				if tt.wantErrDecodeContains == "" {
					t.Errorf("missing setup check error content, actual error: %v", err)
				} else {
					if !strings.Contains(err.Error(), tt.wantErrDecodeContains) {
						t.Errorf("DecodeStreamingLightValidators() error = %v, wantErr contains %v", err, tt.wantErrDecodeContains)
					}
				}
			}
		})
	}
}

func Test_cvpCodecV1_DecodeStreamingLightValidators(t *testing.T) {
	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name                  string
		inputEncodedData      []byte
		wantDecoded           types.StreamingLightValidators
		wantErrDecode         bool
		wantErrDecodeContains string
	}{
		{
			name:             "normal, 2 validators",
			inputEncodedData: []byte("1|00001010" + hex.EncodeToString(fssut("Val1", 20)) + "|00100102" + hex.EncodeToString(fssut("Val2", 20))),
			wantDecoded: types.StreamingLightValidators{
				{
					Index:                     0,
					VotingPowerDisplayPercent: 10.10,
					Moniker:                   "Val1",
				},
				{
					Index:                     1,
					VotingPowerDisplayPercent: 01.02,
					Moniker:                   "Val2",
				},
			},
			wantErrDecode: false,
		},
		{
			name:             "normal, 1 validator",
			inputEncodedData: []byte("1|00001010" + hex.EncodeToString(fssut("Val1", 20))),
			wantDecoded: types.StreamingLightValidators{
				{
					Index:                     0,
					VotingPowerDisplayPercent: 10.10,
					Moniker:                   "Val1",
				},
			},
			wantErrDecode: false,
		},
		{
			name:             "decode upper case input",
			inputEncodedData: []byte(strings.ToUpper("1|00001010" + hex.EncodeToString(fssut("Val1", 20)))),
			wantDecoded: types.StreamingLightValidators{
				{
					Index:                     0,
					VotingPowerDisplayPercent: 10.10,
					Moniker:                   "Val1",
				},
			},
			wantErrDecode: false,
		},
		{
			name:             "decode lower case input",
			inputEncodedData: []byte(strings.ToLower("1|00001010" + hex.EncodeToString(fssut("Val1", 20)))),
			wantDecoded: types.StreamingLightValidators{
				{
					Index:                     0,
					VotingPowerDisplayPercent: 10.10,
					Moniker:                   "Val1",
				},
			},
			wantErrDecode: false,
		},
		{
			name:                  "icorrect codec version",
			inputEncodedData:      []byte("2|00001010" + hex.EncodeToString(fssut("Val1", 20))),
			wantErrDecode:         true,
			wantErrDecodeContains: "bad encoding prefix",
		},
		{
			name:                  "bad format validator index",
			inputEncodedData:      []byte("1|aaa01010" + hex.EncodeToString(fssut("Val1", 20))),
			wantErrDecode:         true,
			wantErrDecodeContains: "failed to parse validator index",
		},
		{
			name:                  "validator index can not be negative",
			inputEncodedData:      []byte("1|-0101010" + hex.EncodeToString(fssut("Val1", 20))),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid validator index",
		},
		{
			name:                  "validator index can not be greater than 998",
			inputEncodedData:      []byte("1|99901010" + hex.EncodeToString(fssut("Val1", 20))),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid validator index",
		},
		{
			name:                  "bad format voting power percent x100",
			inputEncodedData:      []byte("1|000aaaaa" + hex.EncodeToString(fssut("Val1", 20))),
			wantErrDecode:         true,
			wantErrDecodeContains: "failed to parse voting power display percent x100",
		},
		{
			name:                  "voting power percent can not be negative",
			inputEncodedData:      []byte("1|000-0001" + hex.EncodeToString(fssut("Val1", 20))),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid voting power display percent",
		},
		{
			name:                  "voting power percent can not greater than 100",
			inputEncodedData:      []byte("1|00010001" + hex.EncodeToString(fssut("Val1", 20))),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid voting power display percent",
		},
		{
			name:                  "moniker longer than 20 bytes",
			inputEncodedData:      []byte("1|00001010" + hex.EncodeToString([]byte("123456789012345678901"))),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid validator raw data length 50",
		},
		{
			name:                  "bad moniker bytes",
			inputEncodedData:      []byte("1|00001010" + hex.EncodeToString([]byte("1234567890123456789")) + "ZZ"),
			wantErrDecode:         true,
			wantErrDecodeContains: "failed to decode moniker",
		},
		{
			name:                  "bad validators index",
			inputEncodedData:      []byte("1|00001010" + hex.EncodeToString(fssut("Val1", 20)) + "|00200102" + hex.EncodeToString(fssut("Val2", 20))), // index 0 jump to 2
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid validator index sequence",
		},
		{
			name:                  "bad validators index",
			inputEncodedData:      []byte("1|00101010" + hex.EncodeToString(fssut("Val1", 20)) + "|00200102" + hex.EncodeToString(fssut("Val2", 20))), // missing index 0
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid validator index sequence",
		},
		{
			name:             "santinize moniker",
			inputEncodedData: []byte(strings.ToLower("1|00001010" + hex.EncodeToString(fssut(`<he'llo">`, 20)))),
			wantDecoded: types.StreamingLightValidators{
				{
					Index:                     0,
					VotingPowerDisplayPercent: 10.10,
					Moniker:                   "(he`llo`)",
				},
			},
			wantErrDecode: false,
		},
		{
			name:             "no moniker",
			inputEncodedData: []byte("1|00001010"),
			wantDecoded: types.StreamingLightValidators{
				{
					Index:                     0,
					VotingPowerDisplayPercent: 10.10,
					Moniker:                   "",
				},
			},
			wantErrDecode: false,
		},
		{
			name:                  "wrong size moniker",
			inputEncodedData:      []byte("1|0000101030"),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid validator raw data length 10",
		},
		{
			name:                  "not accept empty validator list",
			inputEncodedData:      []byte("1|"),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid empty validator raw data",
		},
		{
			name:                  "not accept validator part with empty data",
			inputEncodedData:      []byte("1|00001010" + hex.EncodeToString(fssut("Val1", 20)) + "|"),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid empty validator raw data",
		},
		{
			name:             "moniker contains separator",
			inputEncodedData: []byte(strings.ToLower("1|00001010" + hex.EncodeToString(fssut(cvpCodecV1Separator+`Val1`, 20)))),
			wantDecoded: types.StreamingLightValidators{
				{
					Index:                     0,
					VotingPowerDisplayPercent: 10.10,
					Moniker:                   cvpCodecV1Separator + "Val1",
				},
			},
			wantErrDecode: false,
		},
		{
			name:             "moniker contains separator",
			inputEncodedData: []byte(strings.ToLower("1|00001010" + hex.EncodeToString(fssut(cvpCodecV1Separator+`Val1`, 20)) + cvpCodecV1Separator + "00101010" + hex.EncodeToString(fssut(cvpCodecV1Separator+`Val2`, 20)))),
			wantDecoded: types.StreamingLightValidators{
				{
					Index:                     0,
					VotingPowerDisplayPercent: 10.10,
					Moniker:                   cvpCodecV1Separator + "Val1",
				},
				{
					Index:                     1,
					VotingPowerDisplayPercent: 10.10,
					Moniker:                   cvpCodecV1Separator + "Val2",
				},
			},
			wantErrDecode: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDecoded, err := cvpV1CodecImpl.DecodeStreamingLightValidators(tt.inputEncodedData)

			if (err != nil) != tt.wantErrDecode {
				if err == nil {
					fmt.Println("Un-expected result:", gotDecoded)
				}
				t.Errorf("DecodeStreamingLightValidators() error = %v, wantErr %v", err, tt.wantErrDecode)
				return
			}
			if err == nil {
				if !reflect.DeepEqual(gotDecoded, tt.wantDecoded) {
					t.Errorf("DecodeStreamingLightValidators()\ngot = %v,\nwant %v", gotDecoded, tt.wantDecoded)
				}
			} else {
				if tt.wantErrDecodeContains == "" {
					t.Errorf("missing setup check error content, actual error: %v", err)
				} else {
					if !strings.Contains(err.Error(), tt.wantErrDecodeContains) {
						t.Errorf("DecodeStreamingLightValidators() error = %v, wantErr contains %v", err, tt.wantErrDecodeContains)
					}
				}
			}
		})
	}
}

func Test_cvpCodecV1_EncodeAndDecodeStreamingNextBlockVotingInformation(t *testing.T) {
	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name                               string
		inf                                types.StreamingNextBlockVotingInformation
		wantPanicEncode                    bool
		wantEncodedData                    []byte
		wantDecodedOrUseInputAsWantDecoded *types.StreamingNextBlockVotingInformation // if missing, use input as expect
		wantErrDecode                      bool
		wantErrDecodeContains              string
	}{
		{
			name: "normal, 4 validators",
			inf: types.StreamingNextBlockVotingInformation{
				HeightRoundStep:       "1/2/3",
				Duration:              1 * time.Second,
				PreVotedPercent:       1,
				PreCommitVotedPercent: 2.54,
				ValidatorVoteStates: []types.StreamingValidatorVoteState{
					{
						ValidatorIndex:    0,
						PreVotedBlockHash: "ABCD",
						PreVoted:          true,
						VotedZeroes:       false,
						PreCommitVoted:    true,
					},
					{
						ValidatorIndex:    1,
						PreVotedBlockHash: "0000",
						PreVoted:          true,
						VotedZeroes:       true,
						PreCommitVoted:    false,
					},
					{
						ValidatorIndex:    2,
						PreVotedBlockHash: "ABCD",
						PreVoted:          true,
						VotedZeroes:       false,
						PreCommitVoted:    false,
					},
					{
						ValidatorIndex:    3,
						PreVotedBlockHash: "----",
						PreVoted:          false,
						VotedZeroes:       false,
						PreCommitVoted:    false,
					},
				},
			},
			wantEncodedData: []byte("1|1/2/3|1000|100|254|000ABCDC00100000002ABCDV003----X"),
			wantErrDecode:   false,
		},
		{
			name: "normal, 1 validators",
			inf: types.StreamingNextBlockVotingInformation{
				HeightRoundStep:       "1/2/3",
				Duration:              1 * time.Second,
				PreVotedPercent:       1,
				PreCommitVotedPercent: 2.54,
				ValidatorVoteStates: []types.StreamingValidatorVoteState{
					{
						ValidatorIndex:    0,
						PreVotedBlockHash: "ABCD",
						PreVoted:          true,
						VotedZeroes:       false,
						PreCommitVoted:    true,
					},
				},
			},
			wantEncodedData: []byte("1|1/2/3|1000|100|254|000ABCDC"),
			wantErrDecode:   false,
		},
		{
			name: "can not decode zero validators vote state",
			inf: types.StreamingNextBlockVotingInformation{
				HeightRoundStep:       "1/2/3",
				Duration:              1 * time.Second,
				PreVotedPercent:       1,
				PreCommitVotedPercent: 2.54,
				ValidatorVoteStates:   []types.StreamingValidatorVoteState{},
			},
			wantEncodedData:       []byte("1|1/2/3|1000|100|254|"),
			wantErrDecode:         true,
			wantErrDecodeContains: "missing validator vote states",
		},
		{
			name: "duration will be corrected to zero if negative",
			inf: types.StreamingNextBlockVotingInformation{
				HeightRoundStep:       "1/2/3",
				Duration:              -1 * time.Second,
				PreVotedPercent:       1,
				PreCommitVotedPercent: 2.54,
				ValidatorVoteStates: []types.StreamingValidatorVoteState{
					{
						ValidatorIndex:    0,
						PreVotedBlockHash: "ABCD",
						PreVoted:          true,
						VotedZeroes:       false,
						PreCommitVoted:    true,
					},
				},
			},
			wantEncodedData: []byte("1|1/2/3|0|100|254|000ABCDC"),
			wantDecodedOrUseInputAsWantDecoded: &types.StreamingNextBlockVotingInformation{
				HeightRoundStep:       "1/2/3",
				Duration:              0,
				PreVotedPercent:       1,
				PreCommitVotedPercent: 2.54,
				ValidatorVoteStates: []types.StreamingValidatorVoteState{
					{
						ValidatorIndex:    0,
						PreVotedBlockHash: "ABCD",
						PreVoted:          true,
						VotedZeroes:       false,
						PreCommitVoted:    true,
					},
				},
			},
		},
		{
			name: "percent will be x100 for saving space of dot",
			inf: types.StreamingNextBlockVotingInformation{
				HeightRoundStep:       "1/2/3",
				Duration:              2 * time.Second,
				PreVotedPercent:       1.1,
				PreCommitVotedPercent: 2.54,
				ValidatorVoteStates: []types.StreamingValidatorVoteState{
					{
						ValidatorIndex:    0,
						PreVotedBlockHash: "ABCD",
						PreVoted:          true,
						VotedZeroes:       false,
						PreCommitVoted:    true,
					},
				},
			},
			wantEncodedData: []byte("1|1/2/3|2000|110|254|000ABCDC"),
		},
		{
			name: "panic encode if negative validator index",
			inf: types.StreamingNextBlockVotingInformation{
				HeightRoundStep:       "1/2/3",
				Duration:              time.Second,
				PreVotedPercent:       1,
				PreCommitVotedPercent: 2,
				ValidatorVoteStates: []types.StreamingValidatorVoteState{
					{
						ValidatorIndex: -1,
					},
				},
			},
			wantPanicEncode: true,
		},
		{
			name: "panic encode if validator index greater than 998",
			inf: types.StreamingNextBlockVotingInformation{
				HeightRoundStep:       "1/2/3",
				Duration:              time.Second,
				PreVotedPercent:       1,
				PreCommitVotedPercent: 2,
				ValidatorVoteStates: []types.StreamingValidatorVoteState{
					{
						ValidatorIndex: 999,
					},
				},
			},
			wantPanicEncode: true,
		},
		{
			name: "panic encode if validator list size larger than cap",
			inf: func() types.StreamingNextBlockVotingInformation {
				inf := types.StreamingNextBlockVotingInformation{
					HeightRoundStep:       "1/2/3",
					Duration:              time.Second,
					PreVotedPercent:       1,
					PreCommitVotedPercent: 2,
				}

				for v := 1; v <= constants.MAX_VALIDATORS+1; v++ {
					inf.ValidatorVoteStates = append(inf.ValidatorVoteStates, types.StreamingValidatorVoteState{
						ValidatorIndex: v - 1,
					})
				}

				return inf
			}(),
			wantPanicEncode: true,
		},
		{
			name: "panic encode if block hash length is not 0 or 4",
			inf: types.StreamingNextBlockVotingInformation{
				HeightRoundStep:       "1/2/3",
				Duration:              time.Second,
				PreVotedPercent:       1,
				PreCommitVotedPercent: 2,
				ValidatorVoteStates: []types.StreamingValidatorVoteState{
					{
						ValidatorIndex:    0,
						PreVotedBlockHash: "123",
						PreVoted:          true,
					},
				},
			},
			wantPanicEncode: true,
		},
		{
			name: "panic encode if block hash length is not 0 or 4",
			inf: types.StreamingNextBlockVotingInformation{
				HeightRoundStep:       "1/2/3",
				Duration:              time.Second,
				PreVotedPercent:       1,
				PreCommitVotedPercent: 2,
				ValidatorVoteStates: []types.StreamingValidatorVoteState{
					{
						ValidatorIndex:    0,
						PreVotedBlockHash: "12345",
						PreVoted:          true,
					},
				},
			},
			wantPanicEncode: true,
		},
		{
			name: "automatically fill prevoted block hash if empty",
			inf: types.StreamingNextBlockVotingInformation{
				HeightRoundStep:       "1/2/3",
				Duration:              time.Second,
				PreVotedPercent:       1,
				PreCommitVotedPercent: 2,
				ValidatorVoteStates: []types.StreamingValidatorVoteState{
					{
						ValidatorIndex:    0,
						PreVotedBlockHash: "",
					},
				},
			},
			wantEncodedData: []byte("1|1/2/3|1000|100|200|000----X"),
			wantDecodedOrUseInputAsWantDecoded: &types.StreamingNextBlockVotingInformation{
				HeightRoundStep:       "1/2/3",
				Duration:              time.Second,
				PreVotedPercent:       1,
				PreCommitVotedPercent: 2,
				ValidatorVoteStates: []types.StreamingValidatorVoteState{
					{
						ValidatorIndex:    0,
						PreVotedBlockHash: "----",
					},
				},
			},
		},
		{
			name: "collision of separator byte with bytes index",
			inf: func() types.StreamingNextBlockVotingInformation {
				nextBlockVotingInfo := types.StreamingNextBlockVotingInformation{
					HeightRoundStep:       "1/2/3",
					Duration:              time.Second,
					PreVotedPercent:       1,
					PreCommitVotedPercent: 2,
				}

				for i := 0; i < constants.MAX_VALIDATORS; i++ {
					nextBlockVotingInfo.ValidatorVoteStates = append(nextBlockVotingInfo.ValidatorVoteStates, types.StreamingValidatorVoteState{
						ValidatorIndex:    i,
						PreVotedBlockHash: "C0FF",
						PreVoted:          true,
					})
				}

				return nextBlockVotingInfo
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEncoded := func() (bz []byte) {
				defer func() {
					err := recover()
					if err != nil {
						if !tt.wantPanicEncode {
							t.Errorf("EncodeStreamingNextBlockVotingInformation() panic = %v but not wanted", err)
						}
					} else {
						if tt.wantPanicEncode {
							t.Errorf("EncodeStreamingNextBlockVotingInformation() panic = %v but wanted panic", err)
						}
					}
				}()
				bz = cvpV1CodecImpl.EncodeStreamingNextBlockVotingInformation(&tt.inf)
				return
			}()

			if tt.wantPanicEncode {
				return
			}

			if len(tt.wantEncodedData) > 0 {
				if !bytes.Equal(gotEncoded, tt.wantEncodedData) {
					t.Errorf("EncodeStreamingNextBlockVotingInformation()\ngotEncoded = %v\nwant %v", string(gotEncoded), string(tt.wantEncodedData))
					return
				}
			}

			gotDecoded, err := cvpV1CodecImpl.DecodeStreamingNextBlockVotingInformation(gotEncoded)
			if (err != nil) != tt.wantErrDecode {
				t.Errorf("DecodeStreamingNextBlockVotingInformation() error = %v, wantErr %v", err, tt.wantErrDecode)
				return
			}
			if err == nil {
				if tt.wantDecodedOrUseInputAsWantDecoded == nil {
					tt.wantDecodedOrUseInputAsWantDecoded = &tt.inf
				}
				if !reflect.DeepEqual(gotDecoded, tt.wantDecodedOrUseInputAsWantDecoded) {
					t.Errorf("DecodeStreamingNextBlockVotingInformation()\ngot = %v,\nwant %v", gotDecoded, tt.wantDecodedOrUseInputAsWantDecoded)
				}
			} else {
				if tt.wantErrDecodeContains == "" {
					t.Errorf("missing setup check error content, actual error: %v", err)
				} else {
					if !strings.Contains(err.Error(), tt.wantErrDecodeContains) {
						t.Errorf("DecodeStreamingLightValidators() error = %v, wantErr contains %v", err, tt.wantErrDecodeContains)
					}
				}
			}
		})
	}
}

func Test_cvpCodecV1_DecodeStreamingNextBlockVotingInformation(t *testing.T) {
	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name                  string
		inputEncodedData      []byte
		wantDecoded           *types.StreamingNextBlockVotingInformation
		wantErrDecode         bool
		wantErrDecodeContains string
	}{
		{
			name:             "normal, 4 validators",
			inputEncodedData: []byte("1|1/2/3|1000|100|254|000ABCDC00100000002ABCDV003----X"),
			wantDecoded: &types.StreamingNextBlockVotingInformation{
				HeightRoundStep:       "1/2/3",
				Duration:              1 * time.Second,
				PreVotedPercent:       1,
				PreCommitVotedPercent: 2.54,
				ValidatorVoteStates: []types.StreamingValidatorVoteState{
					{
						ValidatorIndex:    0,
						PreVotedBlockHash: "ABCD",
						PreVoted:          true,
						VotedZeroes:       false,
						PreCommitVoted:    true,
					},
					{
						ValidatorIndex:    1,
						PreVotedBlockHash: "0000",
						PreVoted:          true,
						VotedZeroes:       true,
						PreCommitVoted:    false,
					},
					{
						ValidatorIndex:    2,
						PreVotedBlockHash: "ABCD",
						PreVoted:          true,
						VotedZeroes:       false,
						PreCommitVoted:    false,
					},
					{
						ValidatorIndex:    3,
						PreVotedBlockHash: "----",
						PreVoted:          false,
						VotedZeroes:       false,
						PreCommitVoted:    false,
					},
				},
			},
			wantErrDecode: false,
		},
		{
			name:             "normal, 1 validator",
			inputEncodedData: []byte("1|1/2/3|1000|100|254|000ABCDC"),
			wantDecoded: &types.StreamingNextBlockVotingInformation{
				HeightRoundStep:       "1/2/3",
				Duration:              1 * time.Second,
				PreVotedPercent:       1,
				PreCommitVotedPercent: 2.54,
				ValidatorVoteStates: []types.StreamingValidatorVoteState{
					{
						ValidatorIndex:    0,
						PreVotedBlockHash: "ABCD",
						PreVoted:          true,
						VotedZeroes:       false,
						PreCommitVoted:    true,
					},
				},
			},
			wantErrDecode: false,
		},
		{
			name:             "decode upper case input",
			inputEncodedData: []byte(strings.ToUpper("1|1/2/3|1000|100|254|000ABCDC")),
			wantDecoded: &types.StreamingNextBlockVotingInformation{
				HeightRoundStep:       "1/2/3",
				Duration:              1 * time.Second,
				PreVotedPercent:       1,
				PreCommitVotedPercent: 2.54,
				ValidatorVoteStates: []types.StreamingValidatorVoteState{
					{
						ValidatorIndex:    0,
						PreVotedBlockHash: "ABCD",
						PreVoted:          true,
						VotedZeroes:       false,
						PreCommitVoted:    true,
					},
				},
			},
			wantErrDecode: false,
		},
		{
			name:             "decode lower case input",
			inputEncodedData: []byte(strings.ToLower("1|1/2/3|1000|100|254|000ABCDC")),
			wantDecoded: &types.StreamingNextBlockVotingInformation{
				HeightRoundStep:       "1/2/3",
				Duration:              1 * time.Second,
				PreVotedPercent:       1,
				PreCommitVotedPercent: 2.54,
				ValidatorVoteStates: []types.StreamingValidatorVoteState{
					{
						ValidatorIndex:    0,
						PreVotedBlockHash: "ABCD",
						PreVoted:          true,
						VotedZeroes:       false,
						PreCommitVoted:    true,
					},
				},
			},
			wantErrDecode: false,
		},
		{
			name:                  "icorrect codec version",
			inputEncodedData:      []byte("2|1/2/3|1000|100|254|000ABCDC"),
			wantErrDecode:         true,
			wantErrDecodeContains: "bad encoding prefix",
		},
		{
			name:                  "wrong number of elements",
			inputEncodedData:      []byte("1|1/2/3|1000|100|254"),
			wantErrDecode:         true,
			wantErrDecodeContains: "wrong number of elements",
		},
		{
			name:                  "wrong number of elements",
			inputEncodedData:      []byte("1|1/2/3|1000|100|254|000ABCDC|BAD"),
			wantErrDecode:         true,
			wantErrDecodeContains: "wrong number of elements",
		},
		{
			name:                  "bad format Height/Round/Step",
			inputEncodedData:      []byte("1|1/2/3/4|1000|100|254|000ABCDC"),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid height round step",
		},
		{
			name:                  "bad duration ms, negative",
			inputEncodedData:      []byte("1|1/2/3|-1000|100|254|000ABCDC"),
			wantErrDecode:         true,
			wantErrDecodeContains: "negative duration ms",
		},
		{
			name:                  "bad duration ms, invalid format",
			inputEncodedData:      []byte("1|1/2/3|aaa|100|254|000ABCDC"),
			wantErrDecode:         true,
			wantErrDecodeContains: "failed to parse duration ms",
		},
		{
			name:                  "bad pre-voted percent x100, negative",
			inputEncodedData:      []byte("1|1/2/3|1000|-1|1|000ABCDC"),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid pre-voted percent: -",
		},
		{
			name:                  "bad pre-voted percent x100, greater than 100x100",
			inputEncodedData:      []byte("1|1/2/3|1000|10100|1|000ABCDC"),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid pre-voted percent: 101",
		},
		{
			name:                  "bad format pre-voted percent x100, must be integer",
			inputEncodedData:      []byte("1|1/2/3|1000|aaa|1|000ABCDC"),
			wantErrDecode:         true,
			wantErrDecodeContains: "failed to parse pre-voted percent x100",
		},
		{
			name:                  "bad pre-commit voted percent x100, negative",
			inputEncodedData:      []byte("1|1/2/3|1000|1|-1|000ABCDC"),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid pre-commit voted percent: -",
		},
		{
			name:                  "bad pre-commit voted percent x100, greater than 100x100",
			inputEncodedData:      []byte("1|1/2/3|1000|1|10100|000ABCDC"),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid pre-commit voted percent: 101",
		},
		{
			name:                  "bad pre-commit voted percent x100, must be integer",
			inputEncodedData:      []byte("1|1/2/3|1000|1|aaa|000ABCDC"),
			wantErrDecode:         true,
			wantErrDecodeContains: "failed to parse pre-commit voted percent x100",
		},
		{
			name:                  "require validator vote states",
			inputEncodedData:      []byte("1|1/2/3|1000|1|1|"),
			wantErrDecode:         true,
			wantErrDecodeContains: "missing validator vote states",
		},
		{
			name:                  "bad size of validator vote states",
			inputEncodedData:      []byte("1|1/2/3|1000|1|1|000ABCD"), // missing 1 char
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid validator vote states length",
		},
		{
			name:                  "bad size of validator vote states",
			inputEncodedData:      []byte("1|1/2/3|1000|1|1|000ABCDC001"), // missing 1=5 chars
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid validator vote states length",
		},
		{
			name:                  "validator index can not greater than 998",
			inputEncodedData:      []byte("1|1/2/3|1000|1|1|999ABCDC"),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid validator index",
		},
		{
			name:                  "validator index can not be negative",
			inputEncodedData:      []byte("1|1/2/3|1000|1|1|-12ABCDC"),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid validator index: -",
		},
		{
			name:                  "validator index must be numeric",
			inputEncodedData:      []byte("1|1/2/3|1000|1|1|idxABCDC"),
			wantErrDecode:         true,
			wantErrDecodeContains: "failed to parse validator index",
		},
		{
			name:                  "bad pre-voted block hash",
			inputEncodedData:      []byte("1|1/2/3|1000|1|1|000GGGGC"),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid pre-voted fingerprint block hash: GGGG",
		},
		{
			name:                  "bad pre-voted flag",
			inputEncodedData:      []byte("1|1/2/3|1000|1|1|000ABCDZ"),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid validator vote flag: Z",
		},
		{
			name:                  "duplicated validator index",
			inputEncodedData:      []byte("1|1/2/3|1000|1|1|000ABCDX000ABCDC"),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid validator index sequence, 0 at 1",
		},
		{
			name:                  "invalid sequence of validator index",
			inputEncodedData:      []byte("1|1/2/3|1000|1|1|000ABCDX002ABCDC"), // index 0 jumped to 2
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid validator index sequence, 2 at 1",
		},
		{
			name:                  "invalid sequence of validator index",
			inputEncodedData:      []byte("1|1/2/3|1000|1|1|001ABCDX002ABCDC"), // missing index 0
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid validator index sequence, 1 at 0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDecoded, err := cvpV1CodecImpl.DecodeStreamingNextBlockVotingInformation(tt.inputEncodedData)

			if (err != nil) != tt.wantErrDecode {
				if err == nil {
					fmt.Println("Un-expected result:", gotDecoded)
				}
				t.Errorf("DecodeStreamingNextBlockVotingInformation() error = %v, wantErr %v", err, tt.wantErrDecode)
				return
			}
			if err == nil {
				if !reflect.DeepEqual(gotDecoded, tt.wantDecoded) {
					t.Errorf("DecodeStreamingNextBlockVotingInformation()\ngot = %v,\nwant %v", gotDecoded, tt.wantDecoded)
				}
			} else {
				if tt.wantErrDecodeContains == "" {
					t.Errorf("missing setup check error content, actual error: %v", err)
				} else {
					if !strings.Contains(err.Error(), tt.wantErrDecodeContains) {
						t.Errorf("DecodeStreamingLightValidators() error = %v, wantErr contains %v", err, tt.wantErrDecodeContains)
					}
				}
			}
		})
	}
}
