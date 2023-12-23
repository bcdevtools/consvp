package prevote_streaming_service

//goland:noinspection SpellCheckingInspection
import (
	enginetypes "github.com/bcdevtools/consvp/engine/types"
)

// PreVoteStreamingService is the interface for Pre-Vote & PreCommit-Vote streaming.
type PreVoteStreamingService interface {
	// OpenSession starts a new session for streaming pre-vote & pre-commit vote status.
	// It returns the URL that can be shared for others to join view.
	// Or returns error if failed on registering the session.
	// If a session had been started, it no-op and returns the URL for the existing session.
	OpenSession(lightValidators enginetypes.LightValidators) (shareViewUrl string, err error)

	// ExposeSessionIdAndKey returns the session ID and session key. Can be used to ResumeSession.
	ExposeSessionIdAndKey() (enginetypes.PreVoteStreamingSessionId, enginetypes.PreVoteStreamingSessionKey)

	// ResumeSession resumes the session with the given session ID and session key.
	// Usefully when mistakenly closed process and want to resume, without having to create a new one.
	ResumeSession(
		sessionId enginetypes.PreVoteStreamingSessionId, sessionKey enginetypes.PreVoteStreamingSessionKey,
	) error

	// BroadcastPreVote broadcasts the given pre-vote information to all viewers.
	// It returns error if failed on broadcasting.
	// It returns shouldStop=true if the broadcasting should be stopped.
	BroadcastPreVote(*enginetypes.NextBlockVotingInformation) (err error, shouldStop bool)

	Stop()

	IsStopped() bool
}
