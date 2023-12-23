package types

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
	"regexp"
)

type PreVoteStreamingSessionId string
type PreVoteStreamingSessionKey string

func NewPreVoteStreamingSession(chainId string) (PreVoteStreamingSessionId, PreVoteStreamingSessionKey, error) {
	bufferId := make([]byte, 32)
	_, err := rand.Read(bufferId)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to generate random bytes")
	}

	sid := PreVoteStreamingSessionId(fmt.Sprintf("%s_%X", chainId, bufferId))

	bufferKey := make([]byte, 32)
	_, err = rand.Read(bufferKey)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to generate random bytes")
	}

	sk := PreVoteStreamingSessionKey(hex.EncodeToString(bufferKey))

	return sid, sk, nil
}

var regexpPreVoteStreamingSessionId = regexp.MustCompile(`^[a-zA-Z\d_\-]+_[A-F\d]{64}$`)

func (sid PreVoteStreamingSessionId) ValidateBasic() error {
	if len(sid) == 0 {
		return fmt.Errorf("empty")
	}

	if !regexpPreVoteStreamingSessionId.MatchString(string(sid)) {
		return fmt.Errorf("invalid format %s", sid)
	}

	return nil
}

var regexpPreVoteStreamingSessionKey = regexp.MustCompile(`^[a-f\d]{64}$`)

func (sk PreVoteStreamingSessionKey) ValidateBasic() error {
	if len(sk) == 0 {
		return fmt.Errorf("empty")
	}

	if !regexpPreVoteStreamingSessionKey.MatchString(string(sk)) {
		return fmt.Errorf("invalid format %s", sk)
	}

	return nil
}
