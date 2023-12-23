package prevote_ss_impl

//goland:noinspection SpellCheckingInspection
import (
	"bytes"
	"fmt"
	"github.com/bcdevtools/consvp/codec"
	"github.com/bcdevtools/consvp/constants"
	ss "github.com/bcdevtools/consvp/engine/prevote_streaming_service"
	enginetypes "github.com/bcdevtools/consvp/engine/types"
	"github.com/bcdevtools/consvp/types"
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
	sessionId enginetypes.PreVoteStreamingSessionId

	// sessionKey is used to authorize broadcasting pre-vote information.
	sessionKey enginetypes.PreVoteStreamingSessionKey

	codec codec.CvpCodec

	httpClient preVotedStreamingHttpClient

	stopped bool
}

func (s *preVoteStreamingServiceImpl) Stop() {
	s.stopped = true
}

func (s *preVoteStreamingServiceImpl) IsStopped() bool {
	return s.stopped
}

// NewPreVoteStreamingService creates a new PreVoteStreamingService.
func NewPreVoteStreamingService(chainId string) ss.PreVoteStreamingService {
	return &preVoteStreamingServiceImpl{
		chainId: chainId,

		codec: codec.NewProxyCvpCodec(),

		httpClient: &preVotedStreamingHttpClientImpl{
			baseUrl: constants.STREAMING_BASE_URL,
		},
	}
}

func (s *preVoteStreamingServiceImpl) OpenSession(lightValidators enginetypes.LightValidators) (shareViewUrl string, err error) {
	if len(s.sessionKey) < 1 {
		var streamingLightValidators types.StreamingLightValidators
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

		var registrationResponse enginetypes.RegisterPreVotedStreamingSessionResponse
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

	return fmt.Sprintf("%s/%s/%s", s.httpClient.BaseUrl(), constants.STREAMING_PATH_VIEW_PRE_VOTE_PREFIX, s.sessionId), nil
}

func (s *preVoteStreamingServiceImpl) ExposeSessionIdAndKey() (enginetypes.PreVoteStreamingSessionId, enginetypes.PreVoteStreamingSessionKey) {
	if len(s.sessionId) < 1 || len(s.sessionKey) < 1 {
		panic("no active session found")
	}
	return s.sessionId, s.sessionKey
}

func (s *preVoteStreamingServiceImpl) ResumeSession(
	sessionId enginetypes.PreVoteStreamingSessionId,
	sessionKey enginetypes.PreVoteStreamingSessionKey,
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

func (s *preVoteStreamingServiceImpl) BroadcastPreVote(information *enginetypes.NextBlockVotingInformation) (err error, shouldStop bool) {
	if s.stopped {
		return fmt.Errorf("service is already marked as stopped"), true
	}

	var si *types.StreamingNextBlockVotingInformation
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

	err = genericHandleStatusCode(resp, http.StatusOK, "broadcast pre-vote")

	if err != nil {
		shouldStop = true
		switch resp.StatusCode {
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

func genericHandleStatusCode(resp *http.Response, acceptedStatusCode int, actionName string) error {
	if resp.StatusCode == acceptedStatusCode {
		return nil
	} else if resp.StatusCode == http.StatusBadRequest { // 400
		return fmt.Errorf("invalid request, probably due to server side deprecated this [%s] version, recommend to upgrade '%s'", actionName, constants.BINARY_NAME)
	} else if resp.StatusCode == http.StatusUnauthorized { // 401
		return fmt.Errorf("unauthorized, probably session timed out")
	} else if resp.StatusCode == http.StatusTooManyRequests { // 429
		return fmt.Errorf("slow down")
	} else if resp.StatusCode == http.StatusInternalServerError { // 500
		return fmt.Errorf("internal server issue, please try again later")
	} else if resp.StatusCode == http.StatusBadGateway || // 502
		resp.StatusCode == http.StatusServiceUnavailable { // 503
		return fmt.Errorf("upstream streaming server is currently unavailable, please try again later")
	} else if resp.StatusCode == http.StatusGatewayTimeout { // 504
		return fmt.Errorf("timed out connecting to upstream streaming server, please try again")
	} else if resp.StatusCode == http.StatusUpgradeRequired { // 426
		return fmt.Errorf("'%s' binary upgrade is required, probably due to server side changed [%s] behaviors and conditions", constants.BINARY_NAME, actionName)
	} else {
		return errors.Errorf("failed to [%s], server returned status code: %d", actionName, resp.StatusCode)
	}
}

func transformLightValidatorsToStreamingLightValidators(lightValidators enginetypes.LightValidators) types.StreamingLightValidators {
	var streamingLightValidators types.StreamingLightValidators
	for _, lightValidator := range lightValidators {
		streamingLightValidators = append(streamingLightValidators, types.StreamingLightValidator{
			Index:                     lightValidator.Index,
			VotingPowerDisplayPercent: lightValidator.VotingPowerDisplayPercent,
			Moniker:                   lightValidator.Moniker,
		})
	}
	return streamingLightValidators
}

func transformNextBlockVotingInformationToStreamingNextBlockVotingInformation(information *enginetypes.NextBlockVotingInformation) *types.StreamingNextBlockVotingInformation {
	si := &types.StreamingNextBlockVotingInformation{
		HeightRoundStep:       information.HeightRoundStep,
		Duration:              time.Now().UTC().Sub(information.StartTimeUTC),
		PreVotedPercent:       information.PreVotePercent,
		PreCommitVotedPercent: information.PreCommitPercent,
		ValidatorVoteStates:   nil,
	}

	for _, voteState := range information.SortedValidatorVoteStates {
		si.ValidatorVoteStates = append(si.ValidatorVoteStates, types.StreamingValidatorVoteState{
			ValidatorIndex:    voteState.Validator.Index,
			PreVotedBlockHash: voteState.VotingBlockHash,
			PreVoted:          voteState.PreVoted,
			VotedZeroes:       voteState.VotedZeroes,
			PreCommitVoted:    voteState.PreCommitVoted,
		})
	}

	return si
}
