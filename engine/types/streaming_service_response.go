package types

type RegisterPreVotedStreamingSessionResponse struct {
	SessionId  PreVoteStreamingSessionId  `json:"session-id"`
	SessionKey PreVoteStreamingSessionKey `json:"session-key"`
}
