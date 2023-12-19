package default_conss_impl

import (
	"fmt"
	"time"
)

func (suite *IntegrationTestSuite) Test_defaultConsensusServiceClientImpl_IT_GetNextBlockVotingInformation() {
	lightVals, err := suite.SVC.rpcClient.LightValidators()
	suite.Require().NoError(err)
	suite.Require().NotEmpty(lightVals)

	nextBlockVotingInfo, err := suite.SVC.GetNextBlockVotingInformation(lightVals)
	suite.Require().NoError(err)
	suite.Require().NotNil(nextBlockVotingInfo)
	suite.NotEmpty(nextBlockVotingInfo.SortedValidatorVoteStates)
	suite.NotEmpty(nextBlockVotingInfo.HeightRoundStep)
	suite.True(nextBlockVotingInfo.StartTimeUTC.After(time.Now().UTC().Add(-24*time.Hour)), "expect start time is not too old")

	fmt.Println("Voting information for", nextBlockVotingInfo.HeightRoundStep, ", starts at", nextBlockVotingInfo.StartTimeUTC)
	fmt.Println("Pre-vote percent:", nextBlockVotingInfo.PreVotePercent)
	fmt.Println("Pre-commit percent:", nextBlockVotingInfo.PreCommitPercent)
	fmt.Println("Validators:")
	for i, val := range nextBlockVotingInfo.SortedValidatorVoteStates {
		if i > 0 {
			fmt.Println("__________________")
		}
		fmt.Println(val.Validator.Moniker)
		fmt.Println("Pre-vote:", val.PreVoted)
		fmt.Println("Pre-commit:", val.PreCommitVoted)
		fmt.Println("Voted zeroes:", val.VotedZeroes)
		fmt.Println("Voting power:", val.Validator.VotingPower, "(", val.Validator.VotingPowerDisplayPercent, "%)")
	}
}

func (suite *IntegrationTestSuite) Test_defaultConsensusServiceClientImpl_IT_Shutdown() {
	for i := 0; i < 3; i++ {
		suite.Require().NoError(suite.SVC.Shutdown())
	}
}
