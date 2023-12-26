package prevote_ss_impl

//goland:noinspection SpellCheckingInspection
import (
	coreconstants "github.com/bcdevtools/cvp-streaming-core/constants"
	coreutils "github.com/bcdevtools/cvp-streaming-core/utils"
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
		coreutils.GetRemoteUrlRegisterPreVoteStreamingSession(c.baseUrl, chainId),
		coreconstants.STREAMING_CONTENT_TYPE,
		body,
	)
}

func (c *preVotedStreamingHttpClientImpl) ResumePreVotedStreamingSession(sessionId string, body io.Reader) (*http.Response, error) {
	return http.Post(
		coreutils.GetRemoteUrlResumePreVoteStreamingSession(c.baseUrl, sessionId),
		coreconstants.STREAMING_CONTENT_TYPE,
		body,
	)
}

func (c *preVotedStreamingHttpClientImpl) BroadcastPreVote(sessionId, sessionKey string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(
		"POST",
		coreutils.GetRemoteUrlBroadcastPreVoteDuringStreamingSession(c.baseUrl, sessionId),
		body,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", coreconstants.STREAMING_CONTENT_TYPE)
	req.Header.Set(coreconstants.STREAMING_HEADER_SESSION_KEY, sessionKey)
	return http.DefaultClient.Do(req)
}
