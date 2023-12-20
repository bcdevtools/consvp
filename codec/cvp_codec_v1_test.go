package codec

import (
	"encoding/hex"
	"fmt"
	"github.com/bcdevtools/consvp/types"
	"reflect"
	"strings"
	"testing"
	"time"
)

var cvpV1CodecImpl = getCvpCodecV1()

func Test_cvpEncoderV1_EncodeDecodeStreamingLightValidators(t *testing.T) {
	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name                               string
		validators                         types.StreamingLightValidators
		wantPanicEncode                    bool
		wantEncodedData                    string
		wantErrDecode                      bool
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
			wantEncodedData: "1|00001010" + hex.EncodeToString([]byte("Val1")) + "|00100102" + hex.EncodeToString([]byte("Val2")),
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
			wantEncodedData: "1|00001010" + hex.EncodeToString([]byte("Val1")),
			wantErrDecode:   false,
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
			wantEncodedData: "1|00010000" + hex.EncodeToString([]byte("Val1")),
			wantErrDecode:   false,
		},
		{
			name:            "can not zero empty validator list",
			validators:      []types.StreamingLightValidator{},
			wantPanicEncode: false,
			wantEncodedData: "1|",
			wantErrDecode:   true,
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
			name: "keep only first 20 bytes of moniker",
			validators: []types.StreamingLightValidator{
				{
					Index:                     0,
					VotingPowerDisplayPercent: 99,
					Moniker:                   "123456789012345678901234567890",
				},
			},
			wantPanicEncode: false,
			wantEncodedData: "1|00009900" + hex.EncodeToString([]byte("12345678901234567890")),
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
			wantEncodedData: "1|00009900" + hex.EncodeToString([]byte(`<he'llo">`)),
			wantDecodedOrUseInputAsWantDecoded: []types.StreamingLightValidator{
				{
					Index:                     0,
					VotingPowerDisplayPercent: 99,
					Moniker:                   "(he`llo`)",
				},
			},
			wantErrDecode: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEncoded := func() (data string) {
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
				data = cvpV1CodecImpl.EncodeStreamingLightValidators(tt.validators)
				return
			}()

			if tt.wantPanicEncode {
				return
			}

			if len(tt.wantEncodedData) > 0 {
				if gotEncoded != tt.wantEncodedData {
					t.Errorf("EncodeStreamingLightValidators()\ngotEncoded = %v\nwant %v", gotEncoded, tt.wantEncodedData)
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
			}
		})
	}
}

func Test_cvpEncoderV1_DecodeStreamingLightValidators(t *testing.T) {
	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name             string
		inputEncodedData string
		wantDecoded      types.StreamingLightValidators
		wantErrDecode    bool
	}{
		{
			name:             "normal, 2 validators",
			inputEncodedData: "1|00001010" + hex.EncodeToString([]byte("Val1")) + "|00100102" + hex.EncodeToString([]byte("Val2")),
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
			inputEncodedData: "1|00001010" + hex.EncodeToString([]byte("Val1")),
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
			inputEncodedData: strings.ToUpper("1|00001010" + hex.EncodeToString([]byte("Val1"))),
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
			inputEncodedData: strings.ToLower("1|00001010" + hex.EncodeToString([]byte("Val1"))),
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
			name:             "icorrect codec version",
			inputEncodedData: "2|00001010" + hex.EncodeToString([]byte("Val1")),
			wantErrDecode:    true,
		},
		{
			name:             "bad format validator index",
			inputEncodedData: "1|aaa01010" + hex.EncodeToString([]byte("Val1")),
			wantErrDecode:    true,
		},
		{
			name:             "validator index can not be negative",
			inputEncodedData: "1|-0101010" + hex.EncodeToString([]byte("Val1")),
			wantErrDecode:    true,
		},
		{
			name:             "validator index can not be greater than 998",
			inputEncodedData: "1|99901010" + hex.EncodeToString([]byte("Val1")),
			wantErrDecode:    true,
		},
		{
			name:             "bad format voting power percent x100",
			inputEncodedData: "1|000aaaaa" + hex.EncodeToString([]byte("Val1")),
			wantErrDecode:    true,
		},
		{
			name:             "voting power percent x100 can not be negative",
			inputEncodedData: "1|000-0001" + hex.EncodeToString([]byte("Val1")),
			wantErrDecode:    true,
		},
		{
			name:             "voting power percent x100 can not greater than 100",
			inputEncodedData: "1|00010001" + hex.EncodeToString([]byte("Val1")),
			wantErrDecode:    true,
		},
		{
			name:             "moniker longer than 20 bytes",
			inputEncodedData: "1|00001010" + hex.EncodeToString([]byte("123456789012345678901")),
			wantErrDecode:    true,
		},
		{
			name:             "bad moniker bytes",
			inputEncodedData: "1|00001010" + "ZZ",
			wantErrDecode:    true,
		},
		{
			name:             "bad validators index",
			inputEncodedData: "1|00001010" + hex.EncodeToString([]byte("Val1")) + "|00200102" + hex.EncodeToString([]byte("Val2")), // index 0 jump to 2
			wantErrDecode:    true,
		},
		{
			name:             "bad validators index",
			inputEncodedData: "1|00101010" + hex.EncodeToString([]byte("Val1")) + "|00200102" + hex.EncodeToString([]byte("Val2")), // missing index 0
			wantErrDecode:    true,
		},
		{
			name:             "santinize moniker",
			inputEncodedData: strings.ToLower("1|00001010" + hex.EncodeToString([]byte(`<he'llo">`))),
			wantDecoded: types.StreamingLightValidators{
				{
					Index:                     0,
					VotingPowerDisplayPercent: 10.10,
					Moniker:                   "(he`llo`)",
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
			if err == nil && !reflect.DeepEqual(gotDecoded, tt.wantDecoded) {
				t.Errorf("DecodeStreamingLightValidators()\ngot = %v,\nwant %v", gotDecoded, tt.wantDecoded)
			}
		})
	}
}

func Test_cvpEncoderV1_EncodeAndDecodeStreamingNextBlockVotingInformation(t *testing.T) {
	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name                               string
		inf                                types.StreamingNextBlockVotingInformation
		wantPanicEncode                    bool
		wantEncodedData                    string
		wantDecodedOrUseInputAsWantDecoded *types.StreamingNextBlockVotingInformation // if missing, use input as expect
		wantErrDecode                      bool
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
			wantEncodedData: "1|1/2/3|1000|100|254|000ABCDC00100000002ABCDV003----X",
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
			wantEncodedData: "1|1/2/3|1000|100|254|000ABCDC",
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
			wantEncodedData: "1|1/2/3|1000|100|254|",
			wantErrDecode:   true,
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
			wantEncodedData: "1|1/2/3|0|100|254|000ABCDC",
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
			wantEncodedData: "1|1/2/3|2000|110|254|000ABCDC",
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
			wantEncodedData: "1|1/2/3|1000|100|200|000----X",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEncoded := func() (data string) {
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
				data = cvpV1CodecImpl.EncodeStreamingNextBlockVotingInformation(&tt.inf)
				return
			}()

			if tt.wantPanicEncode {
				return
			}

			if len(tt.wantEncodedData) > 0 {
				if gotEncoded != tt.wantEncodedData {
					t.Errorf("EncodeStreamingNextBlockVotingInformation()\ngotEncoded = %v\nwant %v", gotEncoded, tt.wantEncodedData)
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
			}
		})
	}
}

func Test_cvpEncoderV1_DecodeStreamingNextBlockVotingInformation(t *testing.T) {
	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name             string
		inputEncodedData string
		wantDecoded      *types.StreamingNextBlockVotingInformation
		wantErrDecode    bool
	}{
		{
			name:             "normal, 4 validators",
			inputEncodedData: "1|1/2/3|1000|100|254|000ABCDC00100000002ABCDV003----X",
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
			inputEncodedData: "1|1/2/3|1000|100|254|000ABCDC",
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
			inputEncodedData: strings.ToUpper("1|1/2/3|1000|100|254|000ABCDC"),
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
			inputEncodedData: strings.ToLower("1|1/2/3|1000|100|254|000ABCDC"),
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
			name:             "icorrect codec version",
			inputEncodedData: "2|1/2/3|1000|100|254|000ABCDC",
			wantErrDecode:    true,
		},
		{
			name:             "invalid number of elements",
			inputEncodedData: "1|1/2/3|1000|100|254",
			wantErrDecode:    true,
		},
		{
			name:             "invalid number of elements",
			inputEncodedData: "1|1/2/3|1000|100|254|000ABCDC|BAD",
			wantErrDecode:    true,
		},
		{
			name:             "bad format Height/Round/Step",
			inputEncodedData: "1|1/2/3/4|1000|100|254|000ABCDC",
			wantErrDecode:    true,
		},
		{
			name:             "bad duration ms, negative",
			inputEncodedData: "1|1/2/3|-1000|100|254|000ABCDC",
			wantErrDecode:    true,
		},
		{
			name:             "bad duration ms, invalid format",
			inputEncodedData: "1|1/2/3|aaa|100|254|000ABCDC",
			wantErrDecode:    true,
		},
		{
			name:             "bad pre-voted percent x100, negative",
			inputEncodedData: "1|1/2/3|1000|-1|1|000ABCDC",
			wantErrDecode:    true,
		},
		{
			name:             "bad pre-voted percent x100, greater than 100x100",
			inputEncodedData: "1|1/2/3|1000|10100|1|000ABCDC",
			wantErrDecode:    true,
		},
		{
			name:             "bad format pre-voted percent x100, must be integer",
			inputEncodedData: "1|1/2/3|1000|aaa|1|000ABCDC",
			wantErrDecode:    true,
		},
		{
			name:             "bad pre-commit voted percent x100, negative",
			inputEncodedData: "1|1/2/3|1000|1|-1|000ABCDC",
			wantErrDecode:    true,
		},
		{
			name:             "bad pre-commit voted percent x100, greater than 100x100",
			inputEncodedData: "1|1/2/3|1000|1|10100|000ABCDC",
			wantErrDecode:    true,
		},
		{
			name:             "bad pre-commit voted percent x100, must be integer",
			inputEncodedData: "1|1/2/3|1000|1|aaa|000ABCDC",
			wantErrDecode:    true,
		},
		{
			name:             "require validator vote states",
			inputEncodedData: "1|1/2/3|1000|1|1|",
			wantErrDecode:    true,
		},
		{
			name:             "bad size of validator vote states",
			inputEncodedData: "1|1/2/3|1000|1|1|000ABCD", // missing 1 char
			wantErrDecode:    true,
		},
		{
			name:             "bad size of validator vote states",
			inputEncodedData: "1|1/2/3|1000|1|1|000ABCDC001", // missing 1=5 chars
			wantErrDecode:    true,
		},
		{
			name:             "validator index can not greater than 998",
			inputEncodedData: "1|1/2/3|1000|1|1|999ABCDC",
			wantErrDecode:    true,
		},
		{
			name:             "validator index can not be negative",
			inputEncodedData: "1|1/2/3|1000|1|1|-12ABCDC",
			wantErrDecode:    true,
		},
		{
			name:             "validator index must be numeric",
			inputEncodedData: "1|1/2/3|1000|1|1|idxABCDC",
			wantErrDecode:    true,
		},
		{
			name:             "bad pre-voted block hash",
			inputEncodedData: "1|1/2/3|1000|1|1|000GGGGC",
			wantErrDecode:    true,
		},
		{
			name:             "bad pre-voted flag",
			inputEncodedData: "1|1/2/3|1000|1|1|000ABCDZ",
			wantErrDecode:    true,
		},
		{
			name:             "duplicated validator index",
			inputEncodedData: "1|1/2/3|1000|1|1|000ABCDX000ABCDC",
			wantErrDecode:    true,
		},
		{
			name:             "invalid sequence of validator index",
			inputEncodedData: "1|1/2/3|1000|1|1|000ABCDX002ABCDC", // index 0 jumped to 2
			wantErrDecode:    true,
		},
		{
			name:             "invalid sequence of validator index",
			inputEncodedData: "1|1/2/3|1000|1|1|001ABCDX002ABCDC", // missing index 0
			wantErrDecode:    true,
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
			if err == nil && !reflect.DeepEqual(gotDecoded, tt.wantDecoded) {
				t.Errorf("DecodeStreamingNextBlockVotingInformation()\ngot = %v,\nwant %v", gotDecoded, tt.wantDecoded)
			}
		})
	}
}
