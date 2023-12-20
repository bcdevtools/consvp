package codec

import (
	"bytes"
	"github.com/bcdevtools/consvp/types"
	"regexp"
)

var _ CvpCodec = (*proxyCvpCodec)(nil)

var defaultCvpCodec = getCvpCodecV1()

// proxyCvpCodec is an implementation of CvpCodec.
//
// The proxy automatically detect version of encoded data and forward to the corresponding implementation for decoding.
//
// When invoking encode functions, it forward to default CvpCodec.
type proxyCvpCodec struct {
}

// NewProxyCvpCodec returns new instance of proxy CvpCodec.
//
// The proxy automatically detect version of encoded data and forward to the corresponding implementation for decoding.
//
// When invoking encode functions, it forward to default CvpCodec.
func NewProxyCvpCodec() CvpCodec {
	return proxyCvpCodec{}
}

func (p proxyCvpCodec) EncodeStreamingLightValidators(validators types.StreamingLightValidators) []byte {
	return defaultCvpCodec.EncodeStreamingLightValidators(validators)
}

func (p proxyCvpCodec) DecodeStreamingLightValidators(bz []byte) (types.StreamingLightValidators, error) {
	if bytes.HasPrefix(bz, []byte(prefixDataEncodedByCvpCodecV1)) {
		return getCvpCodecV1().DecodeStreamingLightValidators(bz)
	}

	panic("unable to detect encoder version")
}

func (p proxyCvpCodec) EncodeStreamingNextBlockVotingInformation(information *types.StreamingNextBlockVotingInformation) []byte {
	return defaultCvpCodec.EncodeStreamingNextBlockVotingInformation(information)
}

var regexpHeightRoundStep = regexp.MustCompile(`^\d+/\d+/\d+$`)
var regexpPreVotedFingerprintBlockHash = regexp.MustCompile(`^[a-fA-F\d]{4}$`)

func (p proxyCvpCodec) DecodeStreamingNextBlockVotingInformation(bz []byte) (*types.StreamingNextBlockVotingInformation, error) {
	if bytes.HasPrefix(bz, []byte(prefixDataEncodedByCvpCodecV1)) {
		return getCvpCodecV1().DecodeStreamingNextBlockVotingInformation(bz)
	}

	panic("unable to detect encoder version")
}
