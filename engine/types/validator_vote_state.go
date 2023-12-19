package types

//goland:noinspection SpellCheckingInspection

type ValidatorVoteState struct {
	Validator      LightValidator
	PreVoted       bool
	VotedZeroes    bool
	PreCommitVoted bool
}
