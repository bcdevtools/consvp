package prevote_ss_impl

//goland:noinspection SpellCheckingInspection
import (
	"bytes"
	"fmt"
	"github.com/bcdevtools/consvp/constants"
	ss "github.com/bcdevtools/consvp/engine/prevote_streaming_service"
	enginetypes "github.com/bcdevtools/consvp/engine/types"
	corecodec "github.com/bcdevtools/cvp-streaming-core/codec"
	coretypes "github.com/bcdevtools/cvp-streaming-core/types"
	coreutils "github.com/bcdevtools/cvp-streaming-core/utils"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/libs/json"
	"io"
	"net/http"
	"time"
)

var _ ss.PreVoteStreamingService = (*preVoteStreamingServiceImpl)(nil)

type preVoteStreamingServiceImpl struct {
	// chainId is the chain ID that the upstream RPC server belong to.
	chainId string

	// sessionId is the unique identifier for this session. Can be used to generate URL for sharing or to broadcast.
	sessionId coretypes.PreVoteStreamingSessionId

	// sessionKey is used to authorize broadcasting pre-vote information.
	sessionKey coretypes.PreVoteStreamingSessionKey

	codec corecodec.CvpCodec

	httpClient preVotedStreamingHttpClient

	stopped bool
}

// NewPreVoteStreamingService creates a new PreVoteStreamingService.
func NewPreVoteStreamingService(chainId string, upstreamServerUrl string, optionalCodec corecodec.CvpCodec) ss.PreVoteStreamingService {
	var codec corecodec.CvpCodec
	if optionalCodec != nil {
		codec = optionalCodec
	} else {
		codec = corecodec.NewProxyCvpCodec()
	}

	return &preVoteStreamingServiceImpl{
		chainId: chainId,

		codec: codec,

		httpClient: &preVotedStreamingHttpClientImpl{
			baseUrl: upstreamServerUrl,
		},
	}
}

// OpenSession starts a new session for streaming pre-vote & pre-commit vote status.
// It returns the URL that can be shared for others to join view.
// Or returns error if failed on registering the session.
// If a session had been started, it no-op and returns the URL for the existing session.
func (s *preVoteStreamingServiceImpl) OpenSession(lightValidators enginetypes.LightValidators) (shareViewUrl string, err error) {
	if len(s.sessionKey) < 1 {
		var streamingLightValidators coretypes.StreamingLightValidators
		streamingLightValidators = transformLightValidatorsToStreamingLightValidators(lightValidators)

		encoded := s.codec.EncodeStreamingLightValidators(streamingLightValidators)

		resp, errRegister := s.httpClient.RegisterPreVotedStreamingSession(s.chainId, bytes.NewBuffer(encoded))
		if errRegister != nil {
			return "", errors.Wrap(errRegister, "failed to register pre-vote streaming session")
		}
		defer func() {
			if resp.Body != nil {
				_ = resp.Body.Close()
			}
		}()

		err = genericHandleStatusCode(resp, http.StatusCreated, "register pre-vote streaming session")
		if err != nil {
			return "", err
		}

		bz, errRead := io.ReadAll(resp.Body)
		if errRead != nil {
			return "", errors.Wrap(errRead, "failed to read response body")
		}

		var registrationResponse coretypes.PreVoteStreamingSessionRegistrationResponse
		errUnmarshal := json.Unmarshal(bz, &registrationResponse)
		if errUnmarshal != nil {
			return "", errors.Wrap(errUnmarshal, "failed to unmarshal response body")
		}

		if err := registrationResponse.SessionId.ValidateBasic(); err != nil {
			return "", errors.Wrap(err, "invalid session ID")
		}
		if err := registrationResponse.SessionKey.ValidateBasic(); err != nil {
			return "", errors.Wrap(err, "invalid session key")
		}

		s.sessionId = registrationResponse.SessionId
		s.sessionKey = registrationResponse.SessionKey
	}

	return coreutils.GetPublicUrlViewPreVoteStreamingSession(s.httpClient.BaseUrl(), string(s.sessionId)), nil
}

// ExposeSessionIdAndKey returns the session ID and session key. Can be used to ResumeSession.
func (s *preVoteStreamingServiceImpl) ExposeSessionIdAndKey() (coretypes.PreVoteStreamingSessionId, coretypes.PreVoteStreamingSessionKey) {
	if len(s.sessionId) < 1 || len(s.sessionKey) < 1 {
		panic("no active session found")
	}
	return s.sessionId, s.sessionKey
}

// ResumeSession resumes the session with the given session ID and session key.
// Usefully when mistakenly closed process and want to resume, without having to create a new one.
func (s *preVoteStreamingServiceImpl) ResumeSession(
	sessionId coretypes.PreVoteStreamingSessionId,
	sessionKey coretypes.PreVoteStreamingSessionKey,
) error {
	if err := sessionId.ValidateBasic(); err != nil {
		return errors.Wrap(err, "bad session ID")
	}
	if err := sessionKey.ValidateBasic(); err != nil {
		return errors.Wrap(err, "bad session key")
	}

	if s.sessionId == sessionId && s.sessionKey == sessionKey {
		return nil
	} else if len(s.sessionId) > 0 || len(s.sessionKey) > 0 {
		return errors.New("cannot resume session, because another session is currently active")
	}

	resp, err := s.httpClient.ResumePreVotedStreamingSession(string(sessionId), bytes.NewBuffer([]byte(sessionKey)))
	if err != nil {
		return errors.Wrapf(err, "failed to resume pre-vote streaming session %s", sessionId)
	}
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	err = genericHandleStatusCode(resp, http.StatusAccepted, "resume pre-vote streaming session")
	if err != nil {
		return err
	}

	s.sessionId = sessionId
	s.sessionKey = sessionKey

	return nil
}

// BroadcastPreVote broadcasts the given pre-vote information to all viewers.
// It returns error if failed on broadcasting.
// It returns shouldStop=true if the broadcasting should be stopped.
func (s *preVoteStreamingServiceImpl) BroadcastPreVote(information *enginetypes.NextBlockVotingInformation) (err error, shouldStop bool) {
	if s.stopped {
		return fmt.Errorf("service is already marked as stopped"), true
	}

	var si *coretypes.StreamingNextBlockVotingInformation
	si = transformNextBlockVotingInformationToStreamingNextBlockVotingInformation(information)

	encoded := s.codec.EncodeStreamingNextBlockVotingInformation(si)

	resp, err := s.httpClient.BroadcastPreVote(string(s.sessionId), string(s.sessionKey), bytes.NewBuffer(encoded))
	if err != nil {
		err = errors.Wrap(err, "failed to broadcast pre-vote")
		shouldStop = false
		return
	}
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	if resp.StatusCode == http.StatusNotModified {
		err = fmt.Errorf("upstream status has not changed, probably due to duplicated or outdated content")
	} else if resp.StatusCode == http.StatusNotFound {
		err = fmt.Errorf("session not found, please start a new streaming session")
	} else {
		err = genericHandleStatusCode(resp, http.StatusOK, "broadcast pre-vote")
	}

	if err != nil {
		shouldStop = true
		switch resp.StatusCode {
		case http.StatusNotModified:
			shouldStop = false
			break
		case http.StatusNotFound:
			shouldStop = false
			break
		case http.StatusTooManyRequests: // rate limit
			shouldStop = false
			break
		case http.StatusInternalServerError:
			shouldStop = false
			break
		case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout: // temporary unavailable
			shouldStop = false
			break
		default:
			break
		}
	}

	return
}

// Stop tells the service to stop.
func (s *preVoteStreamingServiceImpl) Stop() {
	s.stopped = true
}

// IsStopped returns true if the service is stopped.
func (s *preVoteStreamingServiceImpl) IsStopped() bool {
	return s.stopped
}

func genericHandleStatusCode(resp *http.Response, acceptedStatusCode int, actionName string) error {
	if resp.StatusCode == acceptedStatusCode {
		return nil
	} else if resp.StatusCode == http.StatusBadRequest { // 400
		return fmt.Errorf("bad request")
	} else if resp.StatusCode == http.StatusUnauthorized { // 401
		return fmt.Errorf("session timed out")
	} else if resp.StatusCode == http.StatusForbidden { // 403
		return fmt.Errorf("mis-match session key")
	} else if resp.StatusCode == http.StatusUnsupportedMediaType { // 415
		return fmt.Errorf("deprecated codec version or unsupported content type")
	} else if resp.StatusCode == http.StatusTooManyRequests { // 429
		return fmt.Errorf("slow down")
	} else if resp.StatusCode == http.StatusInternalServerError { // 500
		return fmt.Errorf("internal server issue")
	} else if resp.StatusCode == http.StatusBadGateway || // 502
		resp.StatusCode == http.StatusServiceUnavailable { // 503
		return fmt.Errorf("upstream server unavailable")
	} else if resp.StatusCode == http.StatusGatewayTimeout { // 504
		return fmt.Errorf("timed out connecting to upstream server")
	} else if resp.StatusCode == http.StatusUpgradeRequired { // 426
		return fmt.Errorf("'%s' binary upgrade is required", constants.BINARY_NAME)
	} else {
		return errors.Errorf("failed to [%s], server returned status code: %d", actionName, resp.StatusCode)
	}
}

func transformLightValidatorsToStreamingLightValidators(lightValidators enginetypes.LightValidators) coretypes.StreamingLightValidators {
	var streamingLightValidators coretypes.StreamingLightValidators
	for _, lightValidator := range lightValidators {
		streamingLightValidators = append(streamingLightValidators, coretypes.StreamingLightValidator{
			Index:                     lightValidator.Index,
			VotingPowerDisplayPercent: lightValidator.VotingPowerDisplayPercent,
			Moniker:                   lightValidator.Moniker,
		})
	}
	return streamingLightValidators
}

func transformNextBlockVotingInformationToStreamingNextBlockVotingInformation(information *enginetypes.NextBlockVotingInformation) *coretypes.StreamingNextBlockVotingInformation {
	si := &coretypes.StreamingNextBlockVotingInformation{
		HeightRoundStep:       information.HeightRoundStep,
		Duration:              time.Now().UTC().Sub(information.StartTimeUTC),
		PreVotedPercent:       information.PreVotePercent,
		PreCommitVotedPercent: information.PreCommitPercent,
		ValidatorVoteStates:   nil,
	}

	for _, voteState := range information.SortedValidatorVoteStates {
		var blockHash string
		if len(voteState.VotingBlockHash) > 0 {
			blockHash = voteState.VotingBlockHash[:4]
		}
		si.ValidatorVoteStates = append(si.ValidatorVoteStates, coretypes.StreamingValidatorVoteState{
			ValidatorIndex:    voteState.Validator.Index,
			PreVotedBlockHash: blockHash,
			PreVoted:          voteState.PreVoted,
			VotedZeroes:       voteState.VotedZeroes,
			PreCommitVoted:    voteState.PreCommitVoted,
		})
	}

	return si
}
