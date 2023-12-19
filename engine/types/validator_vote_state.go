package types

//goland:noinspection SpellCheckingInspection

type ValidatorVoteState struct {
	Validator       LightValidator
	VotingBlockHash string // 6 bytes fingerprint of hash of the block that the validator voted for
	PreVoted        bool
	VotedZeroes     bool
	PreCommitVoted  bool
}
