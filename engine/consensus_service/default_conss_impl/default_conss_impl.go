package default_conss_impl

//goland:noinspection SpellCheckingInspection
import (
	"fmt"
	"github.com/bcdevtools/consvp/engine/consensus_service"
	"github.com/bcdevtools/consvp/engine/rpc_client"
	enginetypes "github.com/bcdevtools/consvp/engine/types"
	"github.com/pkg/errors"
	"regexp"
	"sort"
	"strings"
)

var _ consensus_service.ConsensusService = (*defaultConsensusServiceClientImpl)(nil) // ensure defaultConsensusServiceClientImpl implements ConsensusService interface

type defaultConsensusServiceClientImpl struct {
	rpcClient rpc_client.RpcClient
}

// NewDefaultConsensusServiceClientImpl returns the default implementation of ConsensusService,
// which uses the RPC client to query the consensus state.
func NewDefaultConsensusServiceClientImpl(rpcClient rpc_client.RpcClient) *defaultConsensusServiceClientImpl {
	return &defaultConsensusServiceClientImpl{
		rpcClient: rpcClient,
	}
}

// GetNextBlockVotingInformation returns the voting status of validators for the next block.
//
// Output voting information is sorted descending by voting power.
func (s *defaultConsensusServiceClientImpl) GetNextBlockVotingInformation(lightValidators enginetypes.LightValidators) (nextBlockVotingInfo *enginetypes.NextBlockVotingInformation, err error) {
	if len(lightValidators) < 1 {
		panic("light validator list is empty")
	}

	consensusState, err := s.rpcClient.ConsensusState()
	if err != nil {
		err = errors.Wrap(err, "failed to get consensus state")
		return
	}

	round, err := consensusState.GetRound()
	if err != nil {
		err = errors.Wrap(err, "failed to extract current round")
		return
	}

	const nilVoteStr = "nil-Vote"

	var validatorVoteStates []enginetypes.ValidatorVoteState

	for i, preVote := range consensusState.Votes[round].PreVotes {
		voted := false
		votedZeroes := false
		lightValidator := lightValidators.GetLightValidatorByIndex(i)

		if !strings.EqualFold(preVote, nilVoteStr) {
			voted = true
		}

		var fingerprintBlockHash string
		if voted {
			fingerprintBlockHash = extractFingerprintBlockHashVotedOn(preVote)

			if fingerprintBlockHash == "000000000000" {
				votedZeroes = true
			} else if strings.HasPrefix(fingerprintBlockHash, "?") { // extract failed
				if //goland:noinspection SpellCheckingInspection
				strings.Contains(preVote, "SIGNED_MSG_TYPE_PREVOTE(Prevote) 000000000000") {
					votedZeroes = true
				}
			}
		}

		validatorVoteStates = append(validatorVoteStates, enginetypes.ValidatorVoteState{
			Validator:       lightValidator,
			VotingBlockHash: fingerprintBlockHash,
			PreVoted:        voted,
			VotedZeroes:     votedZeroes,
		})

		// assert index is correct
		if voted && !strings.Contains(preVote, lightValidator.GetFingerPrintAddress()) {
			panic(fmt.Errorf("index mismatch for validator %s, finger print address %s could not be found in prevote %s", lightValidator.Moniker, lightValidator.GetFingerPrintAddress(), preVote))
		}
	}

	for i, preCommit := range consensusState.Votes[round].PreCommits {
		committed := false

		if !strings.EqualFold(preCommit, nilVoteStr) {
			committed = true
		}

		validatorVoteStates[i].PreCommitVoted = committed
	}

	preVotePercent, err := consensusState.GetPreVotePercent(round)
	if err != nil {
		err = errors.Wrap(err, "failed to extract pre-vote percent")
		return
	}

	preCommitPercent, err := consensusState.GetPreCommitPercent(round)
	if err != nil {
		err = errors.Wrap(err, "failed to extract pre-commit percent")
		return
	}

	startTimeUTC := consensusState.StartTime
	heightRoundStep := consensusState.HeightRoundStep

	sort.Slice(validatorVoteStates, func(i, j int) bool {
		return validatorVoteStates[i].Validator.VotingPower > validatorVoteStates[j].Validator.VotingPower
	})

	nextBlockVotingInfo = &enginetypes.NextBlockVotingInformation{
		SortedValidatorVoteStates: validatorVoteStates,
		PreVotePercent:            preVotePercent,
		PreCommitPercent:          preCommitPercent,
		HeightRoundStep:           heightRoundStep,
		StartTimeUTC:              startTimeUTC,
	}

	return
}

// Shutdown must be called when the service is no longer needed.
func (s *defaultConsensusServiceClientImpl) Shutdown() error {
	return s.rpcClient.Shutdown()
}

var regexpContainsFingerprintBlockHash = regexp.MustCompile(`\s+[a-fA-F\d]{12}\s+[a-fA-F\d]{12}\s+@\s+\d{4}`)

func extractFingerprintBlockHashVotedOn(voteString string) (blockHash string) {
	const defaultUnknownBlockHash = "????????????"
	blockHash = defaultUnknownBlockHash

	subString := regexpContainsFingerprintBlockHash.FindString(voteString)
	subString = strings.TrimSpace(subString)
	if subString != "" {
		blockHash = strings.Split(subString, " ")[0]
	}

	return
}
