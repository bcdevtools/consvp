package mock_local_prevote_ss_impl

import (
	"fmt"
	ss "github.com/bcdevtools/consvp/engine/prevote_streaming_service"
	enginetypes "github.com/bcdevtools/consvp/engine/types"
	"math/rand"
	"time"
)

var _ ss.PreVoteStreamingService = (*mockLocalPreVoteStreamingServiceImpl)(nil)

type mockLocalPreVoteStreamingServiceImpl struct {
	chainId       string
	sessionId     enginetypes.PreVoteStreamingSessionId
	sessionKey    enginetypes.PreVoteStreamingSessionKey
	sessionExpiry time.Time
	stopped       bool
}

func NewMockLocalPreVoteStreamingService(chainId string, sessionDuration time.Duration) ss.PreVoteStreamingService {
	return &mockLocalPreVoteStreamingServiceImpl{
		chainId:       chainId,
		sessionExpiry: time.Now().UTC().Add(sessionDuration),
	}
}

func (m *mockLocalPreVoteStreamingServiceImpl) OpenSession(enginetypes.LightValidators) (shareViewUrl string, err error) {
	id, key, err := enginetypes.NewPreVoteStreamingSession(m.chainId)
	if err != nil {
		panic("failed to generate pseudo pre-vote streaming session")
	}
	m.sessionId = id
	m.sessionKey = key
	return "http://localhost", nil
}

func (m *mockLocalPreVoteStreamingServiceImpl) ExposeSessionIdAndKey() (enginetypes.PreVoteStreamingSessionId, enginetypes.PreVoteStreamingSessionKey) {
	return m.sessionId, m.sessionKey
}

func (m *mockLocalPreVoteStreamingServiceImpl) ResumeSession(id enginetypes.PreVoteStreamingSessionId, key enginetypes.PreVoteStreamingSessionKey) error {
	m.sessionId = id
	m.sessionKey = key
	return nil
}

func (m *mockLocalPreVoteStreamingServiceImpl) BroadcastPreVote(*enginetypes.NextBlockVotingInformation) (err error, shouldStop bool) {
	if m.stopped {
		return fmt.Errorf("service is already marked as stopped"), true
	}

	if time.Now().UTC().After(m.sessionExpiry) {
		return fmt.Errorf("session timed out"), true
	}

	r4 := rand.Uint32() % 4
	if r4%2 != 0 { // 75% chance of no error
		return nil, false
	}

	return fmt.Errorf("mock server returns status code %d", 303+rand.Uint32()%202), false
}

func (m *mockLocalPreVoteStreamingServiceImpl) Stop() {
	m.stopped = true
}

func (m *mockLocalPreVoteStreamingServiceImpl) IsStopped() bool {
	return m.stopped
}
