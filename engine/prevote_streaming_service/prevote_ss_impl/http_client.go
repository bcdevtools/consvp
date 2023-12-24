package prevote_ss_impl

import (
	"fmt"
	"github.com/bcdevtools/consvp/constants"
	"io"
	"net/http"
)

type preVotedStreamingHttpClient interface {
	BaseUrl() string
	RegisterPreVotedStreamingSession(chainId string, body io.Reader) (*http.Response, error)
	ResumePreVotedStreamingSession(sessionId string, body io.Reader) (*http.Response, error)
	BroadcastPreVote(sessionId, sessionKey string, body io.Reader) (*http.Response, error)
}

var _ preVotedStreamingHttpClient = (*preVotedStreamingHttpClientImpl)(nil)

type preVotedStreamingHttpClientImpl struct {
	baseUrl string
}

func (c *preVotedStreamingHttpClientImpl) BaseUrl() string {
	return c.baseUrl
}

func (c *preVotedStreamingHttpClientImpl) RegisterPreVotedStreamingSession(chainId string, body io.Reader) (*http.Response, error) {
	return http.Post(
		fmt.Sprintf("%s/%s/%s", c.baseUrl, constants.STREAMING_PATH_REGISTER_PRE_VOTE_PREFIX, chainId),
		constants.STREAMING_CONTENT_TYPE,
		body,
	)
}

func (c *preVotedStreamingHttpClientImpl) ResumePreVotedStreamingSession(sessionId string, body io.Reader) (*http.Response, error) {
	return http.Post(
		fmt.Sprintf("%s/%s/%s", c.baseUrl, constants.STREAMING_PATH_RESUME_PRE_VOTE_PREFIX, sessionId),
		constants.STREAMING_CONTENT_TYPE,
		body,
	)
}

func (c *preVotedStreamingHttpClientImpl) BroadcastPreVote(sessionId, sessionKey string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s/%s", c.baseUrl, constants.STREAMING_PATH_BROADCAST_PRE_VOTE_PREFIX, sessionId), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", constants.STREAMING_CONTENT_TYPE)
	req.Header.Set(constants.STREAMING_HEADER_SESSION_KEY, sessionKey)
	return http.DefaultClient.Do(req)
}
