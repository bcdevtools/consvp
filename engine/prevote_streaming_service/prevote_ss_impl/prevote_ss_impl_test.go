package prevote_ss_impl

//goland:noinspection SpellCheckingInspection
import (
	"bytes"
	"fmt"
	"github.com/bcdevtools/consvp/constants"
	enginetypes "github.com/bcdevtools/consvp/engine/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/libs/json"
	"io"
	"net/http"
	"testing"
	"time"
)

type PreVoteStreamingServiceTestSuite struct {
	suite.Suite
	httpClient *mockPreVotedStreamingHttpClientImpl
	ss         *preVoteStreamingServiceImpl
}

func TestPreVoteStreamingServiceTestSuite(t *testing.T) {
	suite.Run(t, new(PreVoteStreamingServiceTestSuite))
}

func (suite *PreVoteStreamingServiceTestSuite) SetupSuite() {
	suite.Refresh()
}

func (suite *PreVoteStreamingServiceTestSuite) Refresh() {
	suite.ss = NewPreVoteStreamingService("cosmoshub-4").(*preVoteStreamingServiceImpl)

	// use mock HTTP client for mocking response
	suite.httpClient = &mockPreVotedStreamingHttpClientImpl{
		baseUrl: "http://localhost:8080",
	}
	suite.ss.httpClient = suite.httpClient
}

func (suite *PreVoteStreamingServiceTestSuite) RandomSession() {
	var err error
	suite.ss.sessionId, suite.ss.sessionKey, err = enginetypes.NewPreVoteStreamingSession(suite.ss.chainId)
	if err != nil {
		panic(err)
	}
}

func (suite *PreVoteStreamingServiceTestSuite) Test_OpenSession() {
	pseudoSessionId, pseudoSessionKey, errGenPseudoSessionPair := enginetypes.NewPreVoteStreamingSession(suite.ss.chainId)
	if errGenPseudoSessionPair != nil {
		panic(errGenPseudoSessionPair)
	}

	tests := []struct {
		name                          string
		lightValidators               enginetypes.LightValidators
		streamingServerReturnResponse *http.Response
		streamingServerReturnError    error
		wantError                     bool
		wantErrorContains             string
		wantSession                   bool
		wantExactSessionId            string
		wantExactSessionKey           string
	}{
		{
			name: "register session success",
			lightValidators: enginetypes.LightValidators{
				{
					Index:                     0,
					Moniker:                   "A",
					VotingPowerDisplayPercent: 10.01,
				},
			},
			streamingServerReturnResponse: func() *http.Response {
				serverResponse := enginetypes.RegisterPreVotedStreamingSessionResponse{
					SessionId:  pseudoSessionId,
					SessionKey: pseudoSessionKey,
				}

				bz, err := json.Marshal(serverResponse)
				if err != nil {
					panic(err)
				}

				return &http.Response{
					StatusCode:    http.StatusCreated,
					Body:          io.NopCloser(bytes.NewBuffer(bz)),
					ContentLength: int64(len(bz)),
				}
			}(),
			streamingServerReturnError: nil,
			wantError:                  false,
			wantErrorContains:          "",
			wantSession:                true,
			wantExactSessionId:         string(pseudoSessionId),
			wantExactSessionKey:        string(pseudoSessionKey),
		},
		{
			name: "if HTTP status code is not 201, means error and ignore payload",
			lightValidators: enginetypes.LightValidators{
				{
					Index:                     0,
					Moniker:                   "A",
					VotingPowerDisplayPercent: 10.01,
				},
			},
			streamingServerReturnResponse: func() *http.Response {
				id, k, _ := enginetypes.NewPreVoteStreamingSession(suite.ss.chainId)
				serverResponse := enginetypes.RegisterPreVotedStreamingSessionResponse{
					SessionId:  id,
					SessionKey: k,
				}

				bz, err := json.Marshal(serverResponse)
				if err != nil {
					panic(err)
				}

				return &http.Response{
					StatusCode:    http.StatusOK, // 200, but we expect 201
					Body:          io.NopCloser(bytes.NewBuffer(bz)),
					ContentLength: int64(len(bz)),
				}
			}(),
			streamingServerReturnError: nil,
			wantError:                  true,
			wantErrorContains:          "failed to register pre-vote streaming session, server returned status code",
			wantSession:                false,
		},
		{
			name: "when server returns error",
			lightValidators: enginetypes.LightValidators{
				{
					Index:                     0,
					Moniker:                   "A",
					VotingPowerDisplayPercent: 10.01,
				},
			},
			streamingServerReturnResponse: nil,
			streamingServerReturnError:    fmt.Errorf("pseudo error"),
			wantError:                     true,
			wantErrorContains:             "pseudo error",
			wantSession:                   false,
		},
		{
			name: "when can not unmarshal response body",
			lightValidators: enginetypes.LightValidators{
				{
					Index:                     0,
					Moniker:                   "A",
					VotingPowerDisplayPercent: 10.01,
				},
			},
			streamingServerReturnResponse: func() *http.Response {
				return &http.Response{
					StatusCode:    http.StatusCreated,
					Body:          io.NopCloser(bytes.NewBuffer([]byte{0x01, 0x02})),
					ContentLength: int64(2),
				}
			}(),
			streamingServerReturnError: nil,
			wantError:                  true,
			wantErrorContains:          "failed to unmarshal response body",
			wantSession:                false,
		},
		{
			name: "when can not read response body",
			lightValidators: enginetypes.LightValidators{
				{
					Index:                     0,
					Moniker:                   "A",
					VotingPowerDisplayPercent: 10.01,
				},
			},
			streamingServerReturnResponse: func() *http.Response {
				return &http.Response{
					StatusCode:    http.StatusCreated,
					Body:          &mockClosedReadCloser{},
					ContentLength: int64(50),
				}
			}(),
			streamingServerReturnError: nil,
			wantError:                  true,
			wantErrorContains:          "failed to read response body",
			wantSession:                false,
		},
		{
			name: "when malformed session id",
			lightValidators: enginetypes.LightValidators{
				{
					Index:                     0,
					Moniker:                   "A",
					VotingPowerDisplayPercent: 10.01,
				},
			},
			streamingServerReturnResponse: func() *http.Response {
				serverResponse := enginetypes.RegisterPreVotedStreamingSessionResponse{
					SessionId:  "malformed",
					SessionKey: pseudoSessionKey,
				}

				bz, err := json.Marshal(serverResponse)
				if err != nil {
					panic(err)
				}

				return &http.Response{
					StatusCode:    http.StatusCreated,
					Body:          io.NopCloser(bytes.NewBuffer(bz)),
					ContentLength: int64(len(bz)),
				}
			}(),
			streamingServerReturnError: nil,
			wantError:                  true,
			wantErrorContains:          "invalid format",
		},
		{
			name: "when malformed session key",
			lightValidators: enginetypes.LightValidators{
				{
					Index:                     0,
					Moniker:                   "A",
					VotingPowerDisplayPercent: 10.01,
				},
			},
			streamingServerReturnResponse: func() *http.Response {
				serverResponse := enginetypes.RegisterPreVotedStreamingSessionResponse{
					SessionId:  pseudoSessionId,
					SessionKey: "malformed",
				}

				bz, err := json.Marshal(serverResponse)
				if err != nil {
					panic(err)
				}

				return &http.Response{
					StatusCode:    http.StatusCreated,
					Body:          io.NopCloser(bytes.NewBuffer(bz)),
					ContentLength: int64(len(bz)),
				}
			}(),
			streamingServerReturnError: nil,
			wantError:                  true,
			wantErrorContains:          "invalid format",
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			defer func() {
				suite.Refresh() // reset all state before coming to next test
			}()

			suite.httpClient.nextResponse = tt.streamingServerReturnResponse
			suite.httpClient.nextError = tt.streamingServerReturnError

			shareViewUrl, err := suite.ss.OpenSession(tt.lightValidators)

			suite.Equal(suite.ss.chainId, suite.httpClient.previousRegistrationChainId, "chain ID should be passed to HTTP client")
			bzPayload, errReadPayload := io.ReadAll(suite.httpClient.previousRegistrationPayload)
			if suite.NoError(errReadPayload) {
				if suite.NotEmpty(bzPayload) {
					gotStreamingLightValidators, errDecode := suite.ss.codec.DecodeStreamingLightValidators(bzPayload)
					if suite.NoError(errDecode) {
						suite.Equal(
							transformLightValidatorsToStreamingLightValidators(tt.lightValidators),
							gotStreamingLightValidators,
							"encoded light validators should be passed to HTTP client",
						)
					}
				}
			}

			if tt.wantError {
				suite.Require().Error(err)
				suite.Contains(err.Error(), tt.wantErrorContains)

				suite.Empty(suite.ss.sessionId)
				suite.Empty(suite.ss.sessionKey)
				return
			}

			suite.Require().NoError(err)
			suite.Equal(
				fmt.Sprintf(
					"%s/%s/%s",
					suite.httpClient.baseUrl,
					constants.STREAMING_PATH_VIEW_PRE_VOTE_PREFIX,
					suite.ss.sessionId,
				),
				shareViewUrl,
			)

			if tt.wantSession {
				suite.NotEqual("", suite.ss.sessionId)
				suite.NotEqual("", suite.ss.sessionKey)
				suite.NotEqual(string(suite.ss.sessionId), string(suite.ss.sessionKey))

				if tt.wantExactSessionId != "" {
					suite.Equal(tt.wantExactSessionId, string(suite.ss.sessionId))
				}
				if tt.wantExactSessionKey != "" {
					suite.Equal(tt.wantExactSessionKey, string(suite.ss.sessionKey))
				}

				suite.NoError(suite.ss.sessionId.ValidateBasic())
				suite.NoError(suite.ss.sessionKey.ValidateBasic())
			} else {
				suite.Equal("", suite.ss.sessionId)
				suite.Equal("", suite.ss.sessionKey)
			}
		})
	}

	suite.Run("light validators must be submitted correctly to streaming server", func() {
		defer func() {
			suite.Refresh() // reset all state before coming to next test
		}()

		lightValidators := enginetypes.LightValidators{
			{
				Index:                     0,
				Moniker:                   "A",
				VotingPowerDisplayPercent: 20.02,
			},
			{
				Index:                     1,
				Moniker:                   "B",
				VotingPowerDisplayPercent: 18.81,
			},
		}

		suite.httpClient.nextResponse = nil
		suite.httpClient.nextError = fmt.Errorf("ignored")

		_, _ = suite.ss.OpenSession(lightValidators)

		suite.Equal(suite.ss.chainId, suite.httpClient.previousRegistrationChainId, "chain ID should be passed to HTTP client")
		bzPayload, errReadPayload := io.ReadAll(suite.httpClient.previousRegistrationPayload)
		if suite.NoError(errReadPayload) {
			if suite.NotEmpty(bzPayload) {
				decoded, errDecode := suite.ss.codec.DecodeStreamingLightValidators(bzPayload)
				if suite.NoError(errDecode) {
					suite.Equal(
						transformLightValidatorsToStreamingLightValidators(lightValidators),
						decoded,
						"encoded light validators should be passed to HTTP client",
					)
				}
			}
		}
	})
}

func (suite *PreVoteStreamingServiceTestSuite) Test_ExposeSessionIdAndKey() {
	suite.Run("returns correct", func() {
		suite.Refresh()

		suite.RandomSession()

		id, key := suite.ss.ExposeSessionIdAndKey()
		suite.Equal(suite.ss.sessionId, id)
		suite.Equal(suite.ss.sessionKey, key)
	})

	suite.Run("panic if no active session", func() {
		suite.Refresh()

		suite.Require().Panics(func() {
			_, _ = suite.ss.ExposeSessionIdAndKey()
		})
	})

	suite.Run("panic if missing session id", func() {
		suite.Refresh()

		suite.RandomSession()
		suite.ss.sessionId = ""

		suite.Require().Panics(func() {
			_, _ = suite.ss.ExposeSessionIdAndKey()
		})
	})

	suite.Run("panic if missing session key", func() {
		suite.Refresh()

		suite.RandomSession()
		suite.ss.sessionKey = ""

		suite.Require().Panics(func() {
			_, _ = suite.ss.ExposeSessionIdAndKey()
		})
	})
}

func (suite *PreVoteStreamingServiceTestSuite) Test_ResumeSession() {
	pseudoSessionId, pseudoSessionKey, errGenPseudoSessionPair := enginetypes.NewPreVoteStreamingSession(suite.ss.chainId)
	if errGenPseudoSessionPair != nil {
		panic(errGenPseudoSessionPair)
	}

	newPseudoSessionIdAndKeyProvider := func() (enginetypes.PreVoteStreamingSessionId, enginetypes.PreVoteStreamingSessionKey) {
		pseudoSessionId, pseudoSessionKey, _ := enginetypes.NewPreVoteStreamingSession(suite.ss.chainId)
		return pseudoSessionId, pseudoSessionKey
	}

	tests := []struct {
		name                              string
		inputSessionProvider              func() (enginetypes.PreVoteStreamingSessionId, enginetypes.PreVoteStreamingSessionKey)
		streamingServerReturnResponse     *http.Response
		streamingServerReturnError        error
		wantError                         bool
		wantErrorContains                 string
		ignoreCheckDataPassedToHttpClient bool
	}{
		{
			name: "resume session success",
			inputSessionProvider: func() (enginetypes.PreVoteStreamingSessionId, enginetypes.PreVoteStreamingSessionKey) {
				return pseudoSessionId, pseudoSessionKey
			},
			streamingServerReturnResponse: func() *http.Response {
				return &http.Response{
					StatusCode: http.StatusAccepted,
				}
			}(),
			streamingServerReturnError: nil,
			wantError:                  false,
		},
		{
			name:                 "if HTTP status code is not 200, means error",
			inputSessionProvider: newPseudoSessionIdAndKeyProvider,
			streamingServerReturnResponse: func() *http.Response {
				return &http.Response{
					StatusCode: http.StatusOK, // 200, but we expect 202
				}
			}(),
			streamingServerReturnError: nil,
			wantError:                  true,
			wantErrorContains:          "failed to resume pre-vote streaming session, server returned status code",
		},
		{
			name: "ignore response body",
			inputSessionProvider: func() (enginetypes.PreVoteStreamingSessionId, enginetypes.PreVoteStreamingSessionKey) {
				return pseudoSessionId, pseudoSessionKey
			},
			streamingServerReturnResponse: func() *http.Response {
				return &http.Response{
					StatusCode:    http.StatusAccepted,
					Body:          io.NopCloser(bytes.NewBuffer([]byte{0x00})),
					ContentLength: 1,
				}
			}(),
			streamingServerReturnError: nil,
			wantError:                  false,
		},
		{
			name:                          "when server returns error",
			inputSessionProvider:          newPseudoSessionIdAndKeyProvider,
			streamingServerReturnResponse: nil,
			streamingServerReturnError:    fmt.Errorf("pseudo error"),
			wantError:                     true,
			wantErrorContains:             "pseudo error",
		},
		{
			name: "when malformed session id",
			inputSessionProvider: func() (enginetypes.PreVoteStreamingSessionId, enginetypes.PreVoteStreamingSessionKey) {
				return "malformed", pseudoSessionKey
			},
			wantError:                         true,
			wantErrorContains:                 "bad session ID",
			ignoreCheckDataPassedToHttpClient: true,
		},
		{
			name: "when malformed session key",
			inputSessionProvider: func() (enginetypes.PreVoteStreamingSessionId, enginetypes.PreVoteStreamingSessionKey) {
				return pseudoSessionId, "malformed"
			},
			wantError:                         true,
			wantErrorContains:                 "bad session key",
			ignoreCheckDataPassedToHttpClient: true,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			defer func() {
				suite.Refresh() // reset all state before coming to next test
			}()

			suite.httpClient.nextResponse = tt.streamingServerReturnResponse
			suite.httpClient.nextError = tt.streamingServerReturnError

			sessionId, sessionKey := tt.inputSessionProvider()

			err := suite.ss.ResumeSession(sessionId, sessionKey)

			if !tt.ignoreCheckDataPassedToHttpClient {
				suite.Equal(string(sessionId), suite.httpClient.previousResumeSessionId, "session ID should be passed to HTTP client")
				bzPayload, errReadPayload := io.ReadAll(suite.httpClient.previousResumePayload)
				if suite.NoError(errReadPayload) {
					if suite.NotEmpty(bzPayload) {
						suite.Equal(
							string(sessionKey),
							string(bzPayload),
							"session key should be passed to HTTP client",
						)
					}
				}
			}

			if tt.wantError {
				suite.Require().Error(err)
				suite.Contains(err.Error(), tt.wantErrorContains)

				suite.Empty(suite.ss.sessionId)
				suite.Empty(suite.ss.sessionKey)
				return
			}

			suite.Require().NoError(err)

			suite.Equal(sessionId, suite.ss.sessionId)
			suite.Equal(sessionKey, suite.ss.sessionKey)
		})
	}

	suite.Run("session ID & Key must be submitted correctly to streaming server", func() {
		defer func() {
			suite.Refresh() // reset all state before coming to next test
		}()

		sessionId, sessionKey := newPseudoSessionIdAndKeyProvider()

		suite.httpClient.nextResponse = nil
		suite.httpClient.nextError = fmt.Errorf("ignored")

		_ = suite.ss.ResumeSession(sessionId, sessionKey)

		suite.Equal(string(sessionId), suite.httpClient.previousResumeSessionId, "session ID should be passed to HTTP client")
		bzPayload, errReadPayload := io.ReadAll(suite.httpClient.previousResumePayload)
		if suite.NoError(errReadPayload) {
			if suite.NotEmpty(bzPayload) {
				suite.Equal(
					string(sessionKey),
					string(bzPayload),
					"session key should be passed to HTTP client",
				)
			}
		}
	})

	suite.Run("resume on same opened session", func() {
		defer func() {
			suite.Refresh() // reset all state before coming to next test
		}()

		suite.RandomSession()
		suite.Require().NotEmpty(suite.ss.sessionId)
		suite.Require().NotEmpty(suite.ss.sessionKey)

		err := suite.ss.ResumeSession(suite.ss.sessionId, suite.ss.sessionKey)
		suite.Require().NoError(err)
	})

	suite.Run("resume on while existing opened session, with new session", func() {
		defer func() {
			suite.Refresh() // reset all state before coming to next test
		}()

		suite.RandomSession()
		suite.Require().NotEmpty(suite.ss.sessionId)
		suite.Require().NotEmpty(suite.ss.sessionKey)

		sessionId, sessionKey := newPseudoSessionIdAndKeyProvider()

		err := suite.ss.ResumeSession(sessionId, sessionKey)
		suite.Require().Error(err)
		suite.Contains(err.Error(), "cannot resume session, because another session is currently active")
	})
}

func (suite *PreVoteStreamingServiceTestSuite) Test_BroadcastPreVote() {
	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name                          string
		information                   *enginetypes.NextBlockVotingInformation
		streamingServerReturnResponse *http.Response
		streamingServerReturnError    error
		wantError                     bool
		wantErrorContains             string
	}{
		{
			name: "broadcast success",
			information: &enginetypes.NextBlockVotingInformation{
				SortedValidatorVoteStates: []enginetypes.ValidatorVoteState{
					{
						Validator: enginetypes.LightValidator{
							Index:   0,
							Moniker: "moniker",
						},
						VotingBlockHash: "ABCD",
						PreVoted:        true,
						PreCommitVoted:  true,
					},
				},
				PreVotePercent:   100,
				PreCommitPercent: 1.01,
				HeightRoundStep:  "1/2/3",
				StartTimeUTC:     time.Now().UTC().Add(-1 * time.Second),
			},
			streamingServerReturnResponse: func() *http.Response {
				return &http.Response{
					StatusCode: http.StatusOK,
				}
			}(),
			streamingServerReturnError: nil,
			wantError:                  false,
			wantErrorContains:          "",
		},
		{
			name: "if HTTP status code is not 200, means error",
			information: &enginetypes.NextBlockVotingInformation{
				SortedValidatorVoteStates: []enginetypes.ValidatorVoteState{
					{
						Validator: enginetypes.LightValidator{
							Index:   0,
							Moniker: "moniker",
						},
						VotingBlockHash: "ABCD",
						PreVoted:        true,
						PreCommitVoted:  true,
					},
				},
				PreVotePercent:   100,
				PreCommitPercent: 1.01,
				HeightRoundStep:  "1/2/3",
				StartTimeUTC:     time.Now().UTC().Add(-1 * time.Second),
			},
			streamingServerReturnResponse: func() *http.Response {
				return &http.Response{
					StatusCode: http.StatusAccepted, // 202, but we expect 200
				}
			}(),
			streamingServerReturnError: nil,
			wantError:                  true,
			wantErrorContains:          "failed to broadcast pre-vote, server returned status code",
		},
		{
			name: "when server returns error",
			information: &enginetypes.NextBlockVotingInformation{
				SortedValidatorVoteStates: []enginetypes.ValidatorVoteState{
					{
						Validator: enginetypes.LightValidator{
							Index:   0,
							Moniker: "moniker",
						},
						VotingBlockHash: "ABCD",
						PreVoted:        true,
						PreCommitVoted:  true,
					},
				},
				PreVotePercent:   100,
				PreCommitPercent: 1.01,
				HeightRoundStep:  "1/2/3",
				StartTimeUTC:     time.Now().UTC().Add(-1 * time.Second),
			},
			streamingServerReturnResponse: nil,
			streamingServerReturnError:    fmt.Errorf("pseudo error"),
			wantError:                     true,
			wantErrorContains:             "pseudo error",
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			defer func() {
				suite.Refresh() // reset all state before coming to next test
			}()

			suite.RandomSession()

			suite.httpClient.nextResponse = tt.streamingServerReturnResponse
			suite.httpClient.nextError = tt.streamingServerReturnError

			err := suite.ss.BroadcastPreVote(tt.information)

			suite.NotEmpty(suite.httpClient.previousBroadcastSessionId)
			suite.Equal(string(suite.ss.sessionId), suite.httpClient.previousBroadcastSessionId, "session ID should be passed to HTTP client")
			suite.NotEmpty(suite.httpClient.previousBroadcastSessionKey)
			suite.Equal(string(suite.ss.sessionKey), suite.httpClient.previousBroadcastSessionKey, "session key should be passed to HTTP client")
			bzPayload, errReadPayload := io.ReadAll(suite.httpClient.previousBroadcastPayload)
			if suite.NoError(errReadPayload) {
				if suite.NotEmpty(bzPayload) {
					decoded, errDecode := suite.ss.codec.DecodeStreamingNextBlockVotingInformation(bzPayload)
					decoded.Duration = 1 * time.Second // ignore duration because it is computed from current time within method
					transformed := transformNextBlockVotingInformationToStreamingNextBlockVotingInformation(tt.information)
					transformed.Duration = 1 * time.Second // ignore duration because it is computed from current time within method
					if suite.NoError(errDecode) {
						suite.Equal(
							transformed,
							decoded,
							"encoded pre-vote information should be passed to HTTP client",
						)
					}
				}
			}

			if tt.wantError {
				suite.Require().Error(err)
				suite.Contains(err.Error(), tt.wantErrorContains)
				return
			}

			suite.Require().NoError(err)
		})
	}

	suite.Run("pre-vote information must be submitted correctly to streaming server", func() {
		defer func() {
			suite.Refresh() // reset all state before coming to next test
		}()

		suite.RandomSession()

		information := &enginetypes.NextBlockVotingInformation{
			SortedValidatorVoteStates: []enginetypes.ValidatorVoteState{
				{
					Validator: enginetypes.LightValidator{
						Index:   0,
						Moniker: "moniker",
					},
					VotingBlockHash: "ABCD",
					PreVoted:        true,
					PreCommitVoted:  true,
				},
			},
			PreVotePercent:   100,
			PreCommitPercent: 1.01,
			HeightRoundStep:  "1/2/3",
			StartTimeUTC:     time.Now().UTC().Add(-1 * time.Second),
		}

		suite.httpClient.nextResponse = &http.Response{
			StatusCode:    http.StatusOK,
			Body:          io.NopCloser(bytes.NewBuffer([]byte{})),
			ContentLength: 0,
		}
		suite.httpClient.nextError = nil

		_ = suite.ss.BroadcastPreVote(information)

		suite.NotEmpty(suite.httpClient.previousBroadcastSessionId)
		suite.Equal(string(suite.ss.sessionId), suite.httpClient.previousBroadcastSessionId, "session ID should be passed to HTTP client")
		suite.NotEmpty(suite.httpClient.previousBroadcastSessionKey)
		suite.Equal(string(suite.ss.sessionKey), suite.httpClient.previousBroadcastSessionKey, "session key should be passed to HTTP client")
		bzPayload, errReadPayload := io.ReadAll(suite.httpClient.previousBroadcastPayload)
		if suite.NoError(errReadPayload) {
			if suite.NotEmpty(bzPayload) {
				decoded, errDecode := suite.ss.codec.DecodeStreamingNextBlockVotingInformation(bzPayload)
				decoded.Duration = 1 * time.Second // ignore duration because it is computed from current time within method
				transformed := transformNextBlockVotingInformationToStreamingNextBlockVotingInformation(information)
				transformed.Duration = 1 * time.Second // ignore duration because it is computed from current time within method
				if suite.NoError(errDecode) {
					suite.Equal(
						transformed,
						decoded,
						"encoded pre-vote information should be passed to HTTP client",
					)
				}
			}
		}
	})

	testsPrettyErrMsg := []struct {
		statusCode      int
		wantErrContains string
	}{
		{
			statusCode:      http.StatusBadRequest,
			wantErrContains: "invalid request, probably due to server side deprecated this broadcasting version",
		},
		{
			statusCode:      http.StatusUnauthorized,
			wantErrContains: "unauthorized, probably session timed out",
		},
		{
			statusCode:      http.StatusTooManyRequests,
			wantErrContains: "slow down",
		},
		{
			statusCode:      http.StatusInternalServerError,
			wantErrContains: "internal server issue",
		},
		{
			statusCode:      http.StatusBadGateway,
			wantErrContains: "upstream streaming server is currently unavailable",
		},
		{
			statusCode:      http.StatusServiceUnavailable,
			wantErrContains: "upstream streaming server is currently unavailable",
		},
		{
			statusCode:      http.StatusGatewayTimeout,
			wantErrContains: "upstream streaming server is currently unavailable",
		},
		{
			statusCode:      http.StatusUpgradeRequired,
			wantErrContains: "binary upgrade is required, probably due to server side changed broadcast behaviors and conditions",
		},
	}
	for _, tt := range testsPrettyErrMsg {
		suite.Run(fmt.Sprintf("when server returns %d", tt.statusCode), func() {
			defer func() {
				suite.Refresh() // reset all state before coming to next test
			}()

			suite.RandomSession()

			suite.httpClient.nextResponse = &http.Response{
				StatusCode: tt.statusCode,
			}
			suite.httpClient.nextError = nil

			err := suite.ss.BroadcastPreVote(&enginetypes.NextBlockVotingInformation{})

			suite.Require().Error(err)
			suite.Contains(err.Error(), tt.wantErrContains)
		})
	}
}

var _ io.ReadCloser = (*mockClosedReadCloser)(nil)

type mockClosedReadCloser struct {
}

func (m mockClosedReadCloser) Read([]byte) (n int, err error) {
	return 0, fmt.Errorf("closed")
}

func (m mockClosedReadCloser) Close() error {
	return fmt.Errorf("already closed")
}
