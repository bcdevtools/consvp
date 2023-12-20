package codec

import (
	"github.com/bcdevtools/consvp/types"
	"regexp"
	"strings"
)

var _ CvpCodec = (*proxyCvpCodec)(nil)

var defaultCvpCodec = getCvpCodecV1()

// proxyCvpCodec is an implementation of CvpCodec, automatically detect version and forward to the correct implementation.
type proxyCvpCodec struct {
}

// NewProxyCvpCodec returns a new instance of proxy CvpCodec.
// The proxy automatically detect version and forward to the correct implementation.
func NewProxyCvpCodec() CvpCodec {
	return proxyCvpCodec{}
}

func (p proxyCvpCodec) EncodeStreamingLightValidators(validators types.StreamingLightValidators) string {
	return defaultCvpCodec.EncodeStreamingLightValidators(validators)
}

func (p proxyCvpCodec) DecodeStreamingLightValidators(data string) (types.StreamingLightValidators, error) {
	if strings.HasPrefix(data, cvpCodecV1DataPrefix) {
		return getCvpCodecV1().DecodeStreamingLightValidators(data)
	}

	panic("unable to detect encoder version")
}

func (p proxyCvpCodec) EncodeStreamingNextBlockVotingInformation(information *types.StreamingNextBlockVotingInformation) string {
	return defaultCvpCodec.EncodeStreamingNextBlockVotingInformation(information)
}

var regexpHeightRoundStep = regexp.MustCompile(`^\d+/\d+/\d+$`)
var regexpPreVotedFingerprintBlockHash = regexp.MustCompile(`^[a-fA-F\d]{4}$`)

func (p proxyCvpCodec) DecodeStreamingNextBlockVotingInformation(data string) (*types.StreamingNextBlockVotingInformation, error) {
	if strings.HasPrefix(data, cvpCodecV1DataPrefix) {
		return getCvpCodecV1().DecodeStreamingNextBlockVotingInformation(data)
	}

	panic("unable to detect encoder version")
}
