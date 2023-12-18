package default_conss_impl

import (
	"fmt"
	"time"
)

func (suite *IntegrationTestSuite) Test_defaultConsensusServiceClientImpl_IT_GetNextBlockVotingInformation() {
	lightVals, err := suite.SVC.LightValidators()
	suite.Require().NoError(err)
	suite.Require().NotEmpty(lightVals)

	sortedValidatorVoteStates, preVotePercent, preCommitPercent, heightRoundStep, startTimeUTC, err := suite.SVC.GetNextBlockVotingInformation(lightVals)
	suite.Require().NoError(err)
	suite.NotEmpty(sortedValidatorVoteStates)
	suite.NotEmpty(heightRoundStep)
	suite.True(startTimeUTC.After(time.Now().UTC().Add(-24*time.Hour)), "expect start time is not too old")

	fmt.Println("Voting information for", heightRoundStep, ", starts at", startTimeUTC)
	fmt.Println("Pre-vote percent:", preVotePercent)
	fmt.Println("Pre-commit percent:", preCommitPercent)
	fmt.Println("Validators:")
	for i, val := range sortedValidatorVoteStates {
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

func (suite *IntegrationTestSuite) Test_defaultConsensusServiceClientImpl_IT_LightValidators() {
	lightVals, err := suite.SVC.LightValidators()
	suite.Require().NoError(err)
	suite.Require().NotEmpty(lightVals)

	filledIndexes := make([]bool, len(lightVals))

	for _, lightVal := range lightVals {
		suite.NotEmpty(lightVal.Address)
		suite.NotEmpty(lightVal.PubKey)
		suite.Greater(lightVal.VotingPower, int64(0))
		suite.Greater(lightVal.VotingPowerDisplayPercent, float64(0))

		filledIndexes[lightVal.Index] = true
	}

	for i, isFilled := range filledIndexes {
		suite.True(isFilled, "index %d not found", i)
	}
}

func (suite *IntegrationTestSuite) Test_defaultConsensusServiceClientImpl_IT_Shutdown() {
	for i := 0; i < 3; i++ {
		suite.Require().NoError(suite.SVC.Shutdown())
	}
}
