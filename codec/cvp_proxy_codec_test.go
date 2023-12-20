package codec

import (
	"github.com/bcdevtools/consvp/types"
	"reflect"
	"testing"
	"time"
)

//goland:noinspection GoVarAndConstTypeMayBeOmitted
var cvpProxyCodecImpl CvpCodec = NewProxyCvpCodec()

func Test_proxyCvpCodec_EncodeDecodeStreamingLightValidators(t *testing.T) {
	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name  string
		input types.StreamingLightValidators
		want  types.StreamingLightValidators
	}{
		{
			name: "can codec",
			input: []types.StreamingLightValidator{
				{
					Index:                     0,
					VotingPowerDisplayPercent: 0.1,
					Moniker:                   "Val1",
				},
				{
					Index:                     1,
					VotingPowerDisplayPercent: 2.5,
					Moniker:                   "Val2",
				},
			},
			want: []types.StreamingLightValidator{
				{
					Index:                     0,
					VotingPowerDisplayPercent: 0.1,
					Moniker:                   "Val1",
				},
				{
					Index:                     1,
					VotingPowerDisplayPercent: 2.5,
					Moniker:                   "Val2",
				},
			},
		},
		{
			name: "sanitize moniker",
			input: []types.StreamingLightValidator{
				{
					VotingPowerDisplayPercent: 0.1,
					Moniker:                   `<he'llo">`,
				},
			},
			want: []types.StreamingLightValidator{
				{
					VotingPowerDisplayPercent: 0.1,
					Moniker:                   "(he`llo`)",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEncoded := cvpProxyCodecImpl.EncodeStreamingLightValidators(tt.input)
			gotDecoded, err := cvpProxyCodecImpl.DecodeStreamingLightValidators(gotEncoded)
			if err != nil {
				t.Errorf("DecodeStreamingLightValidators() error = %v", err)
				return
			}
			if reflect.DeepEqual(gotDecoded, tt.want) {
				// ok, test detect v1
				gotEncodedByV1 := cvpV1CodecImpl.EncodeStreamingLightValidators(tt.input)
				_, errDecodeV1 := cvpProxyCodecImpl.DecodeStreamingLightValidators(gotEncodedByV1)
				if errDecodeV1 != nil {
					t.Errorf("proxy not forward v1 encoded data correctly, error = %v", errDecodeV1)
				}
			} else {
				t.Errorf("DecodeStreamingLightValidators()\ngotDecoded = %v\nwant %v", gotDecoded, tt.want)
			}
		})
	}

	t.Run("want panic if decoder not able to detect encoder version", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("DecodeStreamingNextBlockVotingInformation() did not panic")
			}
		}()
		_, _ = cvpProxyCodecImpl.DecodeStreamingLightValidators([]byte("invalid data"))
	})
}

func Test_proxyCvpCodec_EncodeDecodeStreamingNextBlockVotingInformation(t *testing.T) {
	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name  string
		input *types.StreamingNextBlockVotingInformation
	}{
		{
			name: "can codec",
			input: &types.StreamingNextBlockVotingInformation{
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEncoded := cvpProxyCodecImpl.EncodeStreamingNextBlockVotingInformation(tt.input)
			gotDecoded, err := cvpProxyCodecImpl.DecodeStreamingNextBlockVotingInformation(gotEncoded)
			if err != nil {
				t.Errorf("DecodeStreamingNextBlockVotingInformation() error = %v", err)
				return
			}
			if reflect.DeepEqual(gotDecoded, tt.input) {
				// ok, test detect v1
				gotEncodedByV1 := cvpV1CodecImpl.EncodeStreamingNextBlockVotingInformation(tt.input)
				_, errDecodeV1 := cvpProxyCodecImpl.DecodeStreamingNextBlockVotingInformation(gotEncodedByV1)
				if errDecodeV1 != nil {
					t.Errorf("proxy not forward v1 encoded data correctly, error = %v", errDecodeV1)
				}
			} else {
				t.Errorf("DecodeStreamingNextBlockVotingInformation()\ngotDecoded = %v\nwant %v", gotDecoded, tt.input)
			}
		})
	}

	t.Run("want panic if decoder not able to detect encoder version", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("DecodeStreamingNextBlockVotingInformation() did not panic")
			}
		}()
		_, _ = cvpProxyCodecImpl.DecodeStreamingNextBlockVotingInformation([]byte("invalid data"))
	})
}
