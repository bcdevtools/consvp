package codec

//goland:noinspection SpellCheckingInspection
import (
	"github.com/bcdevtools/consvp/types"
)

// CvpCodec is the interface for encoding and decoding streaming data.
type CvpCodec interface {
	// EncodeStreamingLightValidators encodes the given light validators information into sorter string for streaming.
	// Data is assumed to be valid, otherwise panic.
	EncodeStreamingLightValidators(types.StreamingLightValidators) string

	// DecodeStreamingLightValidators decodes the given string into light validators.
	DecodeStreamingLightValidators(string) (types.StreamingLightValidators, error)

	// EncodeStreamingNextBlockVotingInformation encodes the given next block voting information into sorter string for streaming.
	// Data is assumed to be valid, otherwise panic.
	EncodeStreamingNextBlockVotingInformation(*types.StreamingNextBlockVotingInformation) string

	// DecodeStreamingNextBlockVotingInformation decodes the given string into next block voting information.
	DecodeStreamingNextBlockVotingInformation(string) (*types.StreamingNextBlockVotingInformation, error)
}
