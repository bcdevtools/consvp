package default_conss_impl

import (
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
}

func (suite *IntegrationTestSuite) Test_defaultConsensusServiceClientImpl_IT_Shutdown() {
	for i := 0; i < 3; i++ {
		suite.Require().NoError(suite.SVC.Shutdown())
	}
}
