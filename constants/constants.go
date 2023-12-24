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
	STREAMING_BASE_URL       = "https://cvp.bcdev.tools"
	STREAMING_BASE_URL_LOCAL = "http://localhost:8080" // for development purpose only

	STREAMING_PATH_REGISTER_PRE_VOTE_PREFIX  = "register-session/pre-vote"
	STREAMING_PATH_RESUME_PRE_VOTE_PREFIX    = "resume-session/pre-vote"
	STREAMING_PATH_BROADCAST_PRE_VOTE_PREFIX = "broadcast/pre-vote"
	STREAMING_PATH_VIEW_PRE_VOTE_PREFIX      = "pvtop"

	STREAMING_CONTENT_TYPE       = "application/octet-stream"
	STREAMING_HEADER_SESSION_KEY = "X-Session-Key"
)

//goland:noinspection GoSnakeCaseUsage
const (
	MAX_VALIDATORS                             = 250
	MAX_ENCODED_LIGHT_VALIDATORS_BYTES         = 12251 // 12251 v1, 8251 v2
	MAX_ENCODED_NEXT_BLOCK_PRE_VOTE_INFO_BYTES = 2044  // 2044 v1, 1786 v2
)
