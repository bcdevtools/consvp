package codec

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/bcdevtools/consvp/types"
	"reflect"
	"strings"
	"testing"
	"time"
)

var cvpV2CodecImpl = getCvpCodecV2()

func mergeBuffers(bzs ...[]byte) []byte {
	var buf bytes.Buffer
	for _, bz := range bzs {
		buf.Write(bz)
	}
	return buf.Bytes()
}

func Test_cvpCodecV2_EncodeDecodeStreamingLightValidators(t *testing.T) {
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
					VotingPowerDisplayPercent: 10.11,
					Moniker:                   "Val1",
				},
				{
					Index:                     1,
					VotingPowerDisplayPercent: 01.02,
					Moniker:                   "Val2",
				},
			},
			wantPanicEncode: false,
			wantEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte{0x0, 0x0}, []byte{0x0a, 0x0b}, b64bz(fssut("Val1", 20)),
				[]byte{cvpCodecV2Separator},
				[]byte{0x0, 0x1}, []byte{0x01, 0x02}, b64bz(fssut("Val2", 20)),
			),
			wantErrDecode: false,
		},
		{
			name: "normal, 1 validator",
			validators: []types.StreamingLightValidator{
				{
					Index:                     0,
					VotingPowerDisplayPercent: 10.11,
					Moniker:                   "Val1",
				},
			},
			wantPanicEncode: false,
			wantEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte{0x0, 0x0}, []byte{0x0a, 0x0b}, b64bz(fssut("Val1", 20)),
			),
			wantErrDecode: false,
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
			wantEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte{0x0, 0x0}, []byte{0x64, 0x00}, b64bz(fssut("Val1", 20)),
			),
			wantErrDecode: false,
		},
		{
			name:                  "not accept empty validator list",
			validators:            []types.StreamingLightValidator{},
			wantPanicEncode:       false,
			wantEncodedData:       prefixDataEncodedByCvpCodecV2,
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
			wantEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte{0x0, 0x0}, []byte{0x63, 0x00}, b64bz([]byte("12345678901234567890")),
			),
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
			wantEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte{0x0, 0x0}, []byte{0x63, 0x00}, b64bz(fssut(`<he'llo">`, 20)),
			),
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
				bz = cvpV2CodecImpl.EncodeStreamingLightValidators(tt.validators)
				return
			}()

			if tt.wantPanicEncode {
				return
			}

			if len(tt.wantEncodedData) > 0 {
				if !bytes.Equal(gotEncoded, tt.wantEncodedData) {
					t.Errorf("EncodeStreamingLightValidators()\n%v (got)\n%v (want)", hex.EncodeToString(gotEncoded), hex.EncodeToString(tt.wantEncodedData))
					return
				}
			}

			gotDecoded, err := cvpV2CodecImpl.DecodeStreamingLightValidators(gotEncoded)
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

func Test_cvpCodecV2_DecodeStreamingLightValidators(t *testing.T) {
	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name                  string
		inputEncodedData      []byte
		wantDecoded           types.StreamingLightValidators
		wantErrDecode         bool
		wantErrDecodeContains string
	}{
		{
			name: "normal, 2 validators",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte{0x0, 0x0}, []byte{0x0a, 0x0a}, b64bz(fssut("Val1", 20)),
				[]byte{cvpCodecV2Separator},
				[]byte{0x0, 0x1}, []byte{0x01, 0x02}, b64bz(fssut("Val2", 20)),
			),
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
			name: "normal, 1 validator",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte{0x0, 0x0}, []byte{0x0a, 0x0a}, b64bz(fssut("Val1", 20)),
			),
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
			name: "icorrect codec version",
			inputEncodedData: mergeBuffers(
				[]byte{'1', cvpCodecV2Separator},
				[]byte{0x0, 0x0}, []byte{0x0a, 0x0a}, b64bz(fssut("Val1", 20)),
			),
			wantErrDecode:         true,
			wantErrDecodeContains: "bad encoding prefix",
		},
		{
			name: "validator index can not be greater than 998",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte{0x03, 0xE7} /*999*/, []byte{0x0a, 0x0a}, b64bz(fssut("Val1", 20)),
			),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid validator index",
		},
		{
			name: "voting power percent can not greater than 100",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte{0x0, 0x0}, []byte{0x64, 0x01 /*100.01%*/}, b64bz(fssut("Val1", 20)),
			),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid voting power display percent",
		},
		{
			name: "voting power percent can not greater than 100",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte{0x0, 0x0}, []byte{0x65, 0x00 /*101%*/}, b64bz(fssut("Val1", 20)),
			),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid voting power display percent",
		},
		{
			name: "moniker longer than 20 bytes",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte{0x0, 0x0}, []byte{0x0a, 0x0a}, []byte("123456789012345678901"),
			),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid validator raw data length 25",
		},
		{
			name: "bad validators index",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte{0x0, 0x0}, []byte{0x0a, 0x0a}, b64bz(fssut("Val1", 20)),
				[]byte{cvpCodecV2Separator},
				[]byte{0x0, 0x2}, []byte{0x0a, 0x0a}, b64bz(fssut("Val2", 20)),
			), // index 0 jump to 2
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid validator index sequence",
		},
		{
			name: "bad validators index",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte{0x0, 0x1}, []byte{0x0a, 0x0a}, b64bz(fssut("Val1", 20)),
				[]byte{cvpCodecV2Separator},
				[]byte{0x0, 0x2}, []byte{0x0a, 0x0a}, b64bz(fssut("Val2", 20)),
			), // missing index 0
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid validator index sequence",
		},
		{
			name: "santinize moniker",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte{0x0, 0x0}, []byte{0x0a, 0x0a}, b64bz(fssut(`<he'llo">`, 20)),
			),
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
			name: "no moniker",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte{0x0, 0x0}, []byte{0x0a, 0x0a},
			),
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
			name: "wrong size moniker",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte{0x0, 0x0}, []byte{0x0a, 0x0a}, []byte{0x30},
			),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid validator raw data length 5",
		},
		{
			name: "not accept empty validator list",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
			),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid empty validator raw data",
		},
		{
			name: "not accept validator part with empty data",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte{0x0, 0x0}, []byte{0x0a, 0x0a}, b64bz(fssut("Val1", 20)),
				[]byte{cvpCodecV2Separator},
			),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid empty validator raw data",
		},
		{
			name: "buffer moniker is not base64",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte{0x0, 0x0}, []byte{0x0a, 0x0a}, make([]byte, cvpCodecV2Base64EncodedMonikerBufferSize),
			),
			wantErrDecode:         true,
			wantErrDecodeContains: "failed to decode base64 encoded moniker",
		},
		{
			name: "moniker contains separator",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte{0x0, 0x0}, []byte{0x0a, 0x0a}, b64bz(mergeBuffers([]byte{cvpCodecV2Separator}, fssut("Val1", 19))),
			),
			wantDecoded: types.StreamingLightValidators{
				{
					Index:                     0,
					VotingPowerDisplayPercent: 10.10,
					Moniker:                   string(cvpCodecV2Separator) + "Val1",
				},
			},
			wantErrDecode: false,
		},
		{
			name: "moniker contains separator",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte{0x0, 0x0}, []byte{0x0a, 0x0a}, b64bz(mergeBuffers([]byte{cvpCodecV2Separator}, fssut("Val1", 19))),
				[]byte{cvpCodecV2Separator},
				[]byte{0x0, 0x1}, []byte{0x0a, 0x0a}, b64bz(mergeBuffers([]byte{cvpCodecV2Separator}, fssut("Val2", 19))),
			),
			wantDecoded: types.StreamingLightValidators{
				{
					Index:                     0,
					VotingPowerDisplayPercent: 10.10,
					Moniker:                   string(cvpCodecV2Separator) + "Val1",
				},
				{
					Index:                     1,
					VotingPowerDisplayPercent: 10.10,
					Moniker:                   string(cvpCodecV2Separator) + "Val2",
				},
			},
			wantErrDecode: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDecoded, err := cvpV2CodecImpl.DecodeStreamingLightValidators(tt.inputEncodedData)

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

func Test_cvpCodecV2_EncodeAndDecodeStreamingNextBlockVotingInformation(t *testing.T) {
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
			wantEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte("1/2/3"), []byte{cvpCodecV2Separator},
				[]byte("1"), []byte{cvpCodecV2Separator},
				[]byte{0x01, 0x00}, []byte{0x02, 0x36}, []byte{cvpCodecV2Separator},
				[]byte{0x00, 0x00}, []byte("ABCD"), []byte("C"),
				[]byte{0x00, 0x01}, []byte("0000"), []byte("0"),
				[]byte{0x00, 0x02}, []byte("ABCD"), []byte("V"),
				[]byte{0x00, 0x03}, []byte("----"), []byte("X"),
			),
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
			wantEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte("1/2/3"), []byte{cvpCodecV2Separator},
				[]byte("1"), []byte{cvpCodecV2Separator},
				[]byte{0x01, 0x00}, []byte{0x02, 0x36}, []byte{cvpCodecV2Separator},
				[]byte{0x00, 0x00}, []byte("ABCD"), []byte("C"),
			),
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
			wantEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte("1/2/3"), []byte{cvpCodecV2Separator},
				[]byte("1"), []byte{cvpCodecV2Separator},
				[]byte{0x01, 0x00}, []byte{0x02, 0x36}, []byte{cvpCodecV2Separator},
			),
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
			wantEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte("1/2/3"), []byte{cvpCodecV2Separator},
				[]byte("0"), []byte{cvpCodecV2Separator},
				[]byte{0x01, 0x00}, []byte{0x02, 0x36}, []byte{cvpCodecV2Separator},
				[]byte{0x00, 0x00}, []byte("ABCD"), []byte("C"),
			),
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
			name: "percent computed correctly",
			inf: types.StreamingNextBlockVotingInformation{
				HeightRoundStep:       "1/2/3",
				Duration:              2 * time.Second,
				PreVotedPercent:       99.98,
				PreCommitVotedPercent: 97.96,
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
			wantEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte("1/2/3"), []byte{cvpCodecV2Separator},
				[]byte("2"), []byte{cvpCodecV2Separator},
				[]byte{0x63, 0x62}, []byte{0x61, 0x60}, []byte{cvpCodecV2Separator},
				[]byte{0x00, 0x00}, []byte("ABCD"), []byte("C"),
			),
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
				Duration:              3 * time.Second,
				PreVotedPercent:       1,
				PreCommitVotedPercent: 2,
				ValidatorVoteStates: []types.StreamingValidatorVoteState{
					{
						ValidatorIndex:    0,
						PreVotedBlockHash: "",
					},
				},
			},
			wantEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte("1/2/3"), []byte{cvpCodecV2Separator},
				[]byte("3"), []byte{cvpCodecV2Separator},
				[]byte{0x01, 0x00}, []byte{0x02, 0x00}, []byte{cvpCodecV2Separator},
				[]byte{0x00, 0x00}, []byte("----"), []byte("X"),
			),
			wantDecodedOrUseInputAsWantDecoded: &types.StreamingNextBlockVotingInformation{
				HeightRoundStep:       "1/2/3",
				Duration:              3 * time.Second,
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
				bz = cvpV2CodecImpl.EncodeStreamingNextBlockVotingInformation(&tt.inf)
				return
			}()

			if tt.wantPanicEncode {
				return
			}

			if len(tt.wantEncodedData) > 0 {
				if !bytes.Equal(gotEncoded, tt.wantEncodedData) {
					t.Errorf("EncodeStreamingNextBlockVotingInformation()\n%v (got)\n%v (want)", hex.EncodeToString(gotEncoded), hex.EncodeToString(tt.wantEncodedData))
					return
				}
			}

			gotDecoded, err := cvpV2CodecImpl.DecodeStreamingNextBlockVotingInformation(gotEncoded)
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

func Test_cvpCodecV2_DecodeStreamingNextBlockVotingInformation(t *testing.T) {
	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name                  string
		inputEncodedData      []byte
		wantDecoded           *types.StreamingNextBlockVotingInformation
		wantErrDecode         bool
		wantErrDecodeContains string
	}{
		{
			name: "normal, 4 validators",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte("1/2/3"), []byte{cvpCodecV2Separator},
				[]byte("1"), []byte{cvpCodecV2Separator},
				[]byte{0x01, 0x00}, []byte{0x02, 0x36}, []byte{cvpCodecV2Separator},
				[]byte{0x00, 0x00}, []byte("ABCD"), []byte("C"),
				[]byte{0x00, 0x01}, []byte("0000"), []byte("0"),
				[]byte{0x00, 0x02}, []byte("ABCD"), []byte("V"),
				[]byte{0x00, 0x03}, []byte("----"), []byte("X"),
			),
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
			name: "normal, 1 validator",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte("1/2/3"), []byte{cvpCodecV2Separator},
				[]byte("1"), []byte{cvpCodecV2Separator},
				[]byte{0x01, 0x00}, []byte{0x02, 0x36}, []byte{cvpCodecV2Separator},
				[]byte{0x00, 0x00}, []byte("ABCD"), []byte("C"),
			),
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
			name: "can not take pre-voted percent due to invalid size",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte("1/2/3"), []byte{cvpCodecV2Separator},
				[]byte("1"), []byte{cvpCodecV2Separator},
				[]byte{0x01} /* 1/4 */, []byte{cvpCodecV2Separator},
				[]byte{0x00, 0x00}, []byte("ABCD"), []byte("C"),
			),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid buffer of pre-voted and pre-commit voted percent length",
		},
		{
			name: "can not take pre-commit-voted percent due to invalid size",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte("1/2/3"), []byte{cvpCodecV2Separator},
				[]byte("1"), []byte{cvpCodecV2Separator},
				[]byte{0x01, 0x02} /* 2/4 */, []byte{cvpCodecV2Separator},
				[]byte{0x00, 0x00}, []byte("ABCD"), []byte("C"),
			),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid buffer of pre-voted and pre-commit voted percent length",
		},
		{
			name: "can not take pre-commit-voted percent due to invalid size",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte("1/2/3"), []byte{cvpCodecV2Separator},
				[]byte("1"), []byte{cvpCodecV2Separator},
				[]byte{0x01, 0x00, 0x02} /* 3/4 */, []byte{cvpCodecV2Separator},
				[]byte{0x00, 0x00}, []byte("ABCD"), []byte("C"),
			),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid buffer of pre-voted and pre-commit voted percent length",
		},
		{
			name: "icorrect codec version",
			inputEncodedData: mergeBuffers(
				[]byte{'1', cvpCodecV2Separator},
				[]byte("1/2/3"), []byte{cvpCodecV2Separator},
				[]byte("1"), []byte{cvpCodecV2Separator},
				[]byte{0x01, 0x00}, []byte{0x02, 0x36}, []byte{cvpCodecV2Separator},
				[]byte{0x00, 0x00}, []byte("ABCD"), []byte("C"),
			),
			wantErrDecode:         true,
			wantErrDecodeContains: "bad encoding prefix",
		},
		{
			name: "wrong number of elements",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte("1/2/3"), []byte{cvpCodecV2Separator},
				[]byte("1"), []byte{cvpCodecV2Separator},
				[]byte{0x01, 0x00}, []byte{0x02, 0x36},
			),
			wantErrDecode:         true,
			wantErrDecodeContains: "wrong number of elements",
		},
		{
			name: "wrong number of elements",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte("1/2/3"), []byte{cvpCodecV2Separator},
				[]byte("1"), []byte{cvpCodecV2Separator},
				[]byte{0x01, 0x00}, []byte{0x02, 0x36}, []byte{cvpCodecV2Separator},
				[]byte{0x00, 0x00}, []byte("ABCD"), []byte("C"), []byte{cvpCodecV2Separator}, []byte("BAD"),
			),
			wantErrDecode:         true,
			wantErrDecodeContains: "wrong number of elements",
		},
		{
			name: "bad format Height/Round/Step",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte("1/2/3/4"), []byte{cvpCodecV2Separator},
				[]byte("1"), []byte{cvpCodecV2Separator},
				[]byte{0x01, 0x00}, []byte{0x02, 0x36}, []byte{cvpCodecV2Separator},
				[]byte{0x00, 0x00}, []byte("ABCD"), []byte("C"),
			),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid height round step",
		},
		{
			name: "bad duration sec, negative",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte("1/2/3"), []byte{cvpCodecV2Separator},
				[]byte("-1"), []byte{cvpCodecV2Separator},
				[]byte{0x01, 0x00}, []byte{0x02, 0x36}, []byte{cvpCodecV2Separator},
				[]byte{0x00, 0x00}, []byte("ABCD"), []byte("C"),
			),
			wantErrDecode:         true,
			wantErrDecodeContains: "negative duration sec",
		},
		{
			name: "bad duration ms, invalid format",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte("1/2/3"), []byte{cvpCodecV2Separator},
				[]byte("aaa"), []byte{cvpCodecV2Separator},
				[]byte{0x01, 0x00}, []byte{0x02, 0x36}, []byte{cvpCodecV2Separator},
				[]byte{0x00, 0x00}, []byte("ABCD"), []byte("C"),
			),
			wantErrDecode:         true,
			wantErrDecodeContains: "failed to parse duration sec",
		},
		{
			name: "bad pre-voted percent, greater than 100",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte("1/2/3"), []byte{cvpCodecV2Separator},
				[]byte("1"), []byte{cvpCodecV2Separator},
				[]byte{0x65, 0x00} /*101%*/, []byte{0x00, 0x01}, []byte{cvpCodecV2Separator},
				[]byte{0x00, 0x00}, []byte("ABCD"), []byte("C"),
			),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid pre-voted percent: 101",
		},
		{
			name: "bad pre-voted percent, greater than 100",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte("1/2/3"), []byte{cvpCodecV2Separator},
				[]byte("1"), []byte{cvpCodecV2Separator},
				[]byte{0x64, 0x01} /*100.01%*/, []byte{0x00, 0x01}, []byte{cvpCodecV2Separator},
				[]byte{0x00, 0x00}, []byte("ABCD"), []byte("C"),
			),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid pre-voted percent: 100.01",
		},
		{
			name: "bad pre-commit voted percent, greater than 100",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte("1/2/3"), []byte{cvpCodecV2Separator},
				[]byte("1"), []byte{cvpCodecV2Separator},
				[]byte{0x00, 0x01}, []byte{0x65, 0x00} /*101%*/, []byte{cvpCodecV2Separator},
				[]byte{0x00, 0x00}, []byte("ABCD"), []byte("C"),
			),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid pre-commit voted percent: 101",
		},
		{
			name: "bad pre-commit voted percent, greater than 100",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte("1/2/3"), []byte{cvpCodecV2Separator},
				[]byte("1"), []byte{cvpCodecV2Separator},
				[]byte{0x00, 0x01}, []byte{0x64, 0x01} /*100.01%*/, []byte{cvpCodecV2Separator},
				[]byte{0x00, 0x00}, []byte("ABCD"), []byte("C"),
			),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid pre-commit voted percent: 100.01",
		},
		{
			name: "require validator vote states",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte("1/2/3"), []byte{cvpCodecV2Separator},
				[]byte("1"), []byte{cvpCodecV2Separator},
				[]byte{0x01, 0x00}, []byte{0x00, 0x00}, []byte{cvpCodecV2Separator},
			),
			wantErrDecode:         true,
			wantErrDecodeContains: "missing validator vote states",
		},
		{
			name: "bad size of validator vote states",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte("1/2/3"), []byte{cvpCodecV2Separator},
				[]byte("1"), []byte{cvpCodecV2Separator},
				[]byte{0x01, 0x00}, []byte{0x00, 0x00}, []byte{cvpCodecV2Separator},
				[]byte{0x00, 0x00}, []byte("ABCD"), // missing one char
			),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid validator vote states length",
		},
		{
			name: "bad size of validator vote states",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte("1/2/3"), []byte{cvpCodecV2Separator},
				[]byte("1"), []byte{cvpCodecV2Separator},
				[]byte{0x01, 0x00}, []byte{0x00, 0x00}, []byte{cvpCodecV2Separator},
				[]byte{0x00, 0x00}, []byte("ABCD"), []byte("C"),
				[]byte{0x00, 0x01},
			), // missing 1=5 chars
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid validator vote states length",
		},
		{
			name: "validator index can not greater than 998",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte("1/2/3"), []byte{cvpCodecV2Separator},
				[]byte("1"), []byte{cvpCodecV2Separator},
				[]byte{0x01, 0x00}, []byte{0x00, 0x00}, []byte{cvpCodecV2Separator},
				[]byte{0x03, 0xE7}, []byte("ABCD"), []byte("C"),
			),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid validator index",
		},
		{
			name: "bad pre-voted block hash",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte("1/2/3"), []byte{cvpCodecV2Separator},
				[]byte("1"), []byte{cvpCodecV2Separator},
				[]byte{0x01, 0x00}, []byte{0x00, 0x00}, []byte{cvpCodecV2Separator},
				[]byte{0x00, 0x00}, []byte("GGGG"), []byte("C"),
			),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid pre-voted fingerprint block hash: GGGG",
		},
		{
			name: "bad pre-voted flag",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte("1/2/3"), []byte{cvpCodecV2Separator},
				[]byte("1"), []byte{cvpCodecV2Separator},
				[]byte{0x01, 0x00}, []byte{0x00, 0x00}, []byte{cvpCodecV2Separator},
				[]byte{0x00, 0x00}, []byte("ABCD"), []byte("Z"),
			),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid validator vote flag: Z",
		},
		{
			name: "duplicated validator index",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte("1/2/3"), []byte{cvpCodecV2Separator},
				[]byte("1"), []byte{cvpCodecV2Separator},
				[]byte{0x01, 0x00}, []byte{0x00, 0x00}, []byte{cvpCodecV2Separator},
				[]byte{0x00, 0x00}, []byte("ABCD"), []byte("C"),
				[]byte{0x00, 0x00}, []byte("ABCD"), []byte("C"),
			),
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid validator index sequence, 0 at 1",
		},
		{
			name: "invalid sequence of validator index",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte("1/2/3"), []byte{cvpCodecV2Separator},
				[]byte("1"), []byte{cvpCodecV2Separator},
				[]byte{0x01, 0x00}, []byte{0x00, 0x00}, []byte{cvpCodecV2Separator},
				[]byte{0x00, 0x00}, []byte("ABCD"), []byte("C"),
				[]byte{0x00, 0x02}, []byte("ABCD"), []byte("C"),
			), // index 0 jumped to 2
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid validator index sequence, 2 at 1",
		},
		{
			name: "invalid sequence of validator index",
			inputEncodedData: mergeBuffers(
				prefixDataEncodedByCvpCodecV2,
				[]byte("1/2/3"), []byte{cvpCodecV2Separator},
				[]byte("1"), []byte{cvpCodecV2Separator},
				[]byte{0x01, 0x00}, []byte{0x00, 0x00}, []byte{cvpCodecV2Separator},
				[]byte{0x00, 0x01}, []byte("ABCD"), []byte("C"),
				[]byte{0x00, 0x02}, []byte("ABCD"), []byte("C"),
			), // missing index 0
			wantErrDecode:         true,
			wantErrDecodeContains: "invalid validator index sequence, 1 at 0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDecoded, err := cvpV2CodecImpl.DecodeStreamingNextBlockVotingInformation(tt.inputEncodedData)

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

func Test_cvpCodecV2_Base64EncodedMonikerBufferSize(t *testing.T) {
	bz := make([]byte, cvpCodecV2MonikerBufferSize)
	for r := 1; r <= 100; r++ {
		size, err := rand.Read(bz)
		if err != nil {
			t.Fatal(err)
			return
		}
		if size != cvpCodecV2MonikerBufferSize {
			t.Fatalf("bad rand read size: %v", size)
			return
		}
		base64EncodedMonikerBuffer := []byte(base64.StdEncoding.EncodeToString(bz))
		if len(base64EncodedMonikerBuffer) != cvpCodecV2Base64EncodedMonikerBufferSize {
			t.Fatalf("bad base64 encoded moniker buffer size: %v", len(base64EncodedMonikerBuffer))
			return
		}
	}
}
