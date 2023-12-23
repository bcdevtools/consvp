package constants

//goland:noinspection GoSnakeCaseUsage,SpellCheckingInspection
const (
	APP_NAME       = "Consensus Voting Power Information Tool"
	GITHUB_ORG     = "https://github.com/bcdevtools"
	GITHUB_PROJECT = GITHUB_ORG + "/consvp"
	APP_INTRO      = "You are using " + APP_NAME + ", a product of bcdev.tools\nFollow us on GitHub for new tools and updates: " + GITHUB_ORG + " (don't forget to star our repo!)"

	BINARY_NAME = "cvp"
	VERSION     = "v1.0.0"
)

//goland:noinspection GoSnakeCaseUsage
const (
	STREAMING_BASE_URL = "https://cvp.bcdev.tools"

	STREAMING_PATH_REGISTER_PRE_VOTE_PREFIX = "register-session/pre-vote"
	STREAMING_PATH_BROADCAST_PRE_VOTE       = "broadcast/pre-vote"
	STREAMING_PATH_VIEW_PRE_VOTE_PREFIX     = "pvtop"

	STREAMING_CONTENT_TYPE       = "application/octet-stream"
	STREAMING_HEADER_SESSION_KEY = "X-Session-Key"
)
