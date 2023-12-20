package codec

import (
	"fmt"
	"github.com/bcdevtools/consvp/types"
	"reflect"
	"strings"
	"testing"
	"time"
)

func Test_cvpCodecAllVersions_EncodeDecodeStreamingLightValidators(t *testing.T) {
	//goland:noinspection SpellCheckingInspection
	testsGeneral := []struct {
		name                               string
		validators                         types.StreamingLightValidators
		wantPanicEncode                    bool
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
			wantErrDecode:   false,
		},
		{
			name:                  "not accept empty validator list",
			validators:            []types.StreamingLightValidator{},
			wantPanicEncode:       false,
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
			name: "keep only first 20 bytes of moniker",
			validators: []types.StreamingLightValidator{
				{
					Index:                     0,
					VotingPowerDisplayPercent: 99,
					Moniker:                   "123456789012345678901234567890",
				},
			},
			wantPanicEncode: false,
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
	for _, tt := range testsGeneral {
		testHandler := func(codec CvpCodec, t *testing.T) {
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
				bz = codec.EncodeStreamingLightValidators(tt.validators)
				return
			}()

			if tt.wantPanicEncode {
				return
			}

			gotDecoded, err := codec.DecodeStreamingLightValidators(gotEncoded)
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
		}
		t.Run(fmt.Sprintf("%s_v1", tt.name), func(t *testing.T) {
			testHandler(cvpV1CodecImpl, t)
		})
		t.Run(fmt.Sprintf("%s_v2", tt.name), func(t *testing.T) {
			testHandler(cvpV2CodecImpl, t)
		})
	}

	//goland:noinspection SpellCheckingInspection
	testsMonikerNameContainsSeparator := []struct {
		name     string
		seedName string
	}{
		{
			name:     "empty one",
			seedName: "",
		},
		{
			name:     "single char",
			seedName: "a",
		},
		{
			name:     "multiple chars",
			seedName: "aa",
		},
		{
			name:     "multiple chars",
			seedName: "abcde",
		},
	}
	for _, tt := range testsMonikerNameContainsSeparator {
		assertEncodeDecodeKeepSame := func(validators types.StreamingLightValidators, codec CvpCodec, t *testing.T) {
			for _, validator := range validators {
				fmt.Println("Moniker:", validator.Moniker)
			}
			encoded := codec.EncodeStreamingLightValidators(validators)
			decoded, err := codec.DecodeStreamingLightValidators(encoded)
			if err != nil {
				t.Errorf("DecodeStreamingLightValidators() error = %v", err)
				return
			}
			if !reflect.DeepEqual(decoded, validators) {
				t.Errorf("DecodeStreamingLightValidators()\ngot = %v,\nwant %v", decoded, validators)
			}
		}
		monikerNameContainsSeparatorHandler := func(separator byte, codec CvpCodec, t *testing.T) {
			// separator in prefix
			assertEncodeDecodeKeepSame(
				[]types.StreamingLightValidator{
					{
						Index:                     0,
						VotingPowerDisplayPercent: 99,
						Moniker:                   string(append([]byte{separator}, []byte(tt.seedName)...)),
					},
					{
						Index:                     1,
						VotingPowerDisplayPercent: 98,
						Moniker:                   string(append([]byte{separator}, []byte(tt.seedName)...)),
					},
				},
				codec,
				t,
			)
			// separator in suffix
			assertEncodeDecodeKeepSame(
				[]types.StreamingLightValidator{
					{
						Index:                     0,
						VotingPowerDisplayPercent: 99,
						Moniker:                   string(append([]byte(tt.seedName), separator)),
					},
					{
						Index:                     1,
						VotingPowerDisplayPercent: 98,
						Moniker:                   string(append([]byte(tt.seedName), separator)),
					},
				},
				codec,
				t,
			)
			// separator in both prefix and suffix
			assertEncodeDecodeKeepSame(
				[]types.StreamingLightValidator{
					{
						Index:                     0,
						VotingPowerDisplayPercent: 99,
						Moniker:                   string(append(append([]byte{separator}, []byte(tt.seedName)...), separator)),
					},
					{
						Index:                     1,
						VotingPowerDisplayPercent: 98,
						Moniker:                   string(append(append([]byte{separator}, []byte(tt.seedName)...), separator)),
					},
				},
				codec,
				t,
			)
			if len(tt.seedName) >= 2 {
				// separator in middle
				assertEncodeDecodeKeepSame(
					[]types.StreamingLightValidator{
						{
							Index:                     0,
							VotingPowerDisplayPercent: 99,
							Moniker:                   string(tt.seedName[0]) + string(append([]byte{separator}, []byte(tt.seedName[1:])...)),
						},
						{
							Index:                     1,
							VotingPowerDisplayPercent: 98,
							Moniker:                   string(tt.seedName[0]) + string(append([]byte{separator}, []byte(tt.seedName[1:])...)),
						},
					},
					codec,
					t,
				)
				// separator in prefix, middle and suffix
				assertEncodeDecodeKeepSame(
					[]types.StreamingLightValidator{
						{
							Index:                     0,
							VotingPowerDisplayPercent: 99,
							Moniker: func() string {
								moniker := string(append([]byte{separator}, tt.seedName[0]))
								moniker += string(append([]byte{separator}, []byte(tt.seedName[1:])...))
								moniker += string(separator)
								return moniker
							}(),
						},
						{
							Index:                     1,
							VotingPowerDisplayPercent: 98,
							Moniker: func() string {
								moniker := string(append([]byte{separator}, tt.seedName[0]))
								moniker += string(append([]byte{separator}, []byte(tt.seedName[1:])...))
								moniker += string(separator)
								return moniker
							}(),
						},
					},
					codec,
					t,
				)
			}
		}
		t.Run(fmt.Sprintf("%s_v1", tt.name), func(t *testing.T) {
			monikerNameContainsSeparatorHandler(cvpCodecV1Separator[0], cvpV1CodecImpl, t)
		})
		t.Run(fmt.Sprintf("%s_v2", tt.name), func(t *testing.T) {
			monikerNameContainsSeparatorHandler(cvpCodecV2Separator, cvpV2CodecImpl, t)
		})
	}
}

func Test_cvpCodecAllVersions_EncodeAndDecodeStreamingNextBlockVotingInformation(t *testing.T) {
	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name                               string
		inf                                types.StreamingNextBlockVotingInformation
		wantPanicEncode                    bool
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
			wantErrDecode: false,
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
			wantErrDecode: false,
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
		testHandler := func(codec CvpCodec, t *testing.T) {
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
		}
		t.Run(fmt.Sprintf("%s_v1", tt.name), func(t *testing.T) {
			testHandler(cvpV1CodecImpl, t)
		})
		t.Run(fmt.Sprintf("%s_v2", tt.name), func(t *testing.T) {
			testHandler(cvpV2CodecImpl, t)
		})
	}
}

// fssut means fill suffix space chars up to X bytes.
//
// For testing purpose only.
//
//goland:noinspection SpellCheckingInspection
func fssut(str string, upto int) []byte {
	bz := []byte(str)
	for len(bz) < upto {
		bz = append(bz, ' ')
	}
	return bz
}
