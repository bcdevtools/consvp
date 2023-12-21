package prevote_ss_impl

import (
	"io"
	"net/http"
)

var _ preVotedStreamingHttpClient = (*mockPreVotedStreamingHttpClientImpl)(nil)

type mockPreVotedStreamingHttpClientImpl struct {
	baseUrl      string
	nextResponse *http.Response
	nextError    error

	previousRegistrationChainId string
	previousRegistrationPayload io.Reader

	previousResumeSessionId string
	previousResumePayload   io.Reader

	previousBroadcastSessionId  string
	previousBroadcastSessionKey string
	previousBroadcastPayload    io.Reader
}

func (c *mockPreVotedStreamingHttpClientImpl) BaseUrl() string {
	return c.baseUrl
}

func (c *mockPreVotedStreamingHttpClientImpl) RegisterPreVotedStreamingSession(chainId string, payload io.Reader) (*http.Response, error) {
	c.previousRegistrationChainId = chainId
	c.previousRegistrationPayload = payload

	return c.nextResponse, c.nextError
}

func (c *mockPreVotedStreamingHttpClientImpl) ResumePreVotedStreamingSession(sessionId string, payload io.Reader) (*http.Response, error) {
	c.previousResumeSessionId = sessionId
	c.previousResumePayload = payload

	return c.nextResponse, c.nextError
}

func (c *mockPreVotedStreamingHttpClientImpl) BroadcastPreVote(sessionId, sessionKey string, payload io.Reader) (*http.Response, error) {
	c.previousBroadcastSessionId = sessionId
	c.previousBroadcastSessionKey = sessionKey
	c.previousBroadcastPayload = payload

	return c.nextResponse, c.nextError
}
