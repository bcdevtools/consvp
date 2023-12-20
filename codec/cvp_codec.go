package codec

//goland:noinspection SpellCheckingInspection
import (
	"github.com/bcdevtools/consvp/types"
)

// CvpCodec is the interface for encoding and decoding streaming data.
type CvpCodec interface {
	// EncodeStreamingLightValidators encodes the given light validators information into sorter string for streaming.
	// Input is assumed to be valid, otherwise panic.
	EncodeStreamingLightValidators(types.StreamingLightValidators) []byte

	// DecodeStreamingLightValidators decodes the given string into light validators.
	DecodeStreamingLightValidators([]byte) (types.StreamingLightValidators, error)

	// EncodeStreamingNextBlockVotingInformation encodes the given next block voting information into sorter string for streaming.
	// Input is assumed to be valid, otherwise panic.
	EncodeStreamingNextBlockVotingInformation(*types.StreamingNextBlockVotingInformation) []byte

	// DecodeStreamingNextBlockVotingInformation decodes the given string into next block voting information.
	DecodeStreamingNextBlockVotingInformation([]byte) (*types.StreamingNextBlockVotingInformation, error)
}
