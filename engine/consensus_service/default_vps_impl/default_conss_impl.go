package default_conss_impl

//goland:noinspection SpellCheckingInspection
import (
	"encoding/base64"
	"fmt"
	"github.com/bcdevtools/consvp/engine/consensus_service"
	consstypes "github.com/bcdevtools/consvp/engine/consensus_service/types"
	"github.com/bcdevtools/consvp/engine/rpc"
	enginetypes "github.com/bcdevtools/consvp/engine/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptoed25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	tmed25519 "github.com/tendermint/tendermint/crypto/ed25519"
	"sort"
	"strings"
	"time"
)

var _ consensus_service.ConsensusService = (*defaultConsensusServiceClientImpl)(nil) // ensure defaultConsensusServiceClientImpl implements ConsensusService interface

type defaultConsensusServiceClientImpl struct {
	rpcClient rpc.RpcClient
}

// NewDefaultConsensusServiceClientImpl returns the default implementation of ConsensusService,
// which uses the RPC client to query the consensus state.
func NewDefaultConsensusServiceClientImpl(rpcClient rpc.RpcClient) *defaultConsensusServiceClientImpl {
	return &defaultConsensusServiceClientImpl{
		rpcClient: rpcClient,
	}
}

// GetNextBlockVotingInformation returns the voting status of validators for the next block.
//
// Output voting information is sorted descending by voting power.
func (s *defaultConsensusServiceClientImpl) GetNextBlockVotingInformation(lightValidators consstypes.LightValidators) (sortedValidatorVoteStates []enginetypes.ValidatorVoteState, preVotePercent, preCommitPercent float64, heightRoundStep string, startTimeUTC time.Time, err error) {
	if len(lightValidators) < 1 {
		panic("light validator list is empty")
	}

	defer func() {
		if err != nil {
			// clean up
			sortedValidatorVoteStates = nil
			preVotePercent = 0.0
			preCommitPercent = 0.0
			heightRoundStep = ""
			startTimeUTC = time.Time{}
		}
	}()

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

		if //goland:noinspection SpellCheckingInspection
		strings.Contains(preVote, "SIGNED_MSG_TYPE_PREVOTE(Prevote) 000000000000") {
			votedZeroes = true
		}

		validatorVoteStates = append(validatorVoteStates, enginetypes.ValidatorVoteState{
			Validator:   lightValidator,
			PreVoted:    voted,
			VotedZeroes: votedZeroes,
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

	preVotePercent, err = consensusState.GetPreVotePercent(round)
	if err != nil {
		err = errors.Wrap(err, "failed to extract pre-vote percent")
		return
	}

	preCommitPercent, err = consensusState.GetPreCommitPercent(round)
	if err != nil {
		err = errors.Wrap(err, "failed to extract pre-commit percent")
		return
	}

	startTimeUTC = consensusState.StartTime
	heightRoundStep = consensusState.HeightRoundStep

	sort.Slice(validatorVoteStates, func(i, j int) bool {
		return validatorVoteStates[i].Validator.VotingPower > validatorVoteStates[j].Validator.VotingPower
	})

	sortedValidatorVoteStates = validatorVoteStates

	return
}

// LightValidators returns the light list of bonded validators.
//
// CONTRACT: must maintain the same order as the result from the RPC server.
func (s *defaultConsensusServiceClientImpl) LightValidators() ([]consstypes.LightValidator, error) {
	mapper := make(map[string]*consstypes.LightValidator)

	bondedVals, err := s.rpcClient.BondedValidators()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get bonded validators")
	}
	for _, bondedVal := range bondedVals {
		var pubKey cryptoed25519.PubKey
		err = proto.Unmarshal(bondedVal.ConsensusPubkey.Value, &pubKey)
		if err != nil {
			panic(errors.Wrap(err, "failed to unmarshal consensus public key"))
		}

		tmPublicKey, err := cryptocodec.ToTmProtoPublicKey(&pubKey)
		if err != nil {
			panic(errors.Wrap(err, "failed to cast to consensus public key"))
		}

		var tmPubKey tmcrypto.PubKey
		tmPubKey = tmed25519.PubKey(tmPublicKey.GetEd25519())

		val := consstypes.LightValidator{
			Moniker: bondedVal.Description.Moniker,
			Address: strings.ToUpper(tmPubKey.Address().String()),
			PubKey:  base64.StdEncoding.EncodeToString(tmPublicKey.GetEd25519()),
		}
		mapper[val.Address] = &val
	}

	latestVals, err := s.rpcClient.LatestValidators()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get latest validators")
	}

	for i, latestVal := range latestVals {
		address := strings.ToUpper(latestVal.PubKey.Address().String())
		if val, ok := mapper[address]; ok {
			val.Index = i
			val.VotingPower = latestVal.VotingPower
		}
	}

	var result consstypes.LightValidators
	var totalVotingPower uint64

	for _, val := range mapper {
		if val.VotingPower < 1 {
			continue
		}
		result = append(result, *val)
		totalVotingPower += uint64(val.VotingPower)
	}

	for i, val := range result {
		val.VotingPowerDisplayPercent = 100 * (float64(val.VotingPower) / float64(totalVotingPower))
		val.VotingPowerDisplayPercent = float64(int64(val.VotingPowerDisplayPercent*100)) / 100
		if val.VotingPower > 0 && val.VotingPowerDisplayPercent < 0.01 {
			// avoid 0.00% for small voting power because all validators at this point, has voting power
			val.VotingPowerDisplayPercent = 0.01
		}
		result[i] = val
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Index < result[j].Index
	})

	return result, nil
}

// Shutdown must be called when the service is no longer needed.
func (s *defaultConsensusServiceClientImpl) Shutdown() error {
	return s.rpcClient.Shutdown()
}
