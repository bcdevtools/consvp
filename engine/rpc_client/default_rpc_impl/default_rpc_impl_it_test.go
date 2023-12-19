package default_rpc_impl

//goland:noinspection SpellCheckingInspection
import (
	enginetypes "github.com/bcdevtools/consvp/engine/types"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"time"
)

func (suite *IntegrationTestSuite) Test_defaultRpcClientImpl_IT_LightValidators() {
	testHandler := func(client *defaultRpcClientImpl) {
		lightVals, err := client.LightValidators()
		suite.Require().NoError(err)
		suite.Require().NotEmpty(lightVals)

		filledIndexes := make([]bool, len(lightVals))

		for _, lightVal := range lightVals {
			suite.NotEmpty(lightVal.Moniker)
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

	suite.Run("Get light validators on Tendermint node", func() {
		testHandler(suite.TM)
	})

	suite.Run("Get light validators on CometBFT node", func() {
		testHandler(suite.COMETBFT)
	})
}

func (suite *IntegrationTestSuite) Test_defaultRpcClientImpl_IT_BondedValidators() {
	testHandler := func(client *defaultRpcClientImpl) {
		validatorsViaHTTP, err := client.bondedValidatorsViaHTTP()
		suite.Require().NoError(err)
		suite.NotEmpty(validatorsViaHTTP)
		suite.Greater(len(validatorsViaHTTP), 1)

		validatorsViaWs, err := client.bondedValidatorsViaWebsocket()
		if suite.NoError(err) {
			suite.Require().NotEmpty(validatorsViaWs)
			suite.Greater(len(validatorsViaWs), 1)

			suite.Equal(validatorsViaWs, validatorsViaHTTP, "mis-match result validators set between Websocket and HTTP response")
		}
	}

	suite.Run("Get bonded validators on Tendermint node", func() {
		testHandler(suite.TM)
	})

	suite.Run("Get bonded validators on CometBFT node", func() {
		testHandler(suite.COMETBFT)
	})
}

func (suite *IntegrationTestSuite) Test_defaultRpcClientImpl_IT_ConsensusState() {
	assertResult := func(roundState *enginetypes.RoundState) {
		suite.Require().NotNil(roundState)

		suite.NotEmpty(roundState.HeightRoundStep)
		suite.True(roundState.StartTime.After(time.Now().UTC().Add(-1*365*24*time.Hour)), "expect start time is not too old")
		suite.NotEmpty(roundState.Votes)
	}

	testHandler := func(client *defaultRpcClientImpl) {
		consensusStateViaHTTP, err := client.consensusStateViaHTTP()
		suite.Require().NoError(err)
		assertResult(consensusStateViaHTTP)

		consensusStateViaWs, err := client.consensusStateViaWebsocket()
		if suite.NoError(err) {
			assertResult(consensusStateViaWs)
		}
	}

	suite.Run("Get Consensus State on Tendermint node", func() {
		testHandler(suite.TM)
	})

	suite.Run("Get Consensus State on CometBFT node", func() {
		testHandler(suite.COMETBFT)
	})
}

func (suite *IntegrationTestSuite) Test_defaultRpcClientImpl_IT_Status() {
	// This case should test all information that to be fetched during application business logic

	assertResult := func(status *coretypes.ResultStatus) {
		suite.Require().NotNil(status)

		suite.NotEmpty(status.NodeInfo.Network)
		suite.NotEmpty(status.NodeInfo.Version)
		suite.NotEmpty(status.NodeInfo.Moniker)

		if suite.NotNil(status.ValidatorInfo, "expect validator info, even full node") {
			suite.NotEmpty(status.ValidatorInfo.Address)
			suite.NotEmpty(status.ValidatorInfo.PubKey.Bytes())
		}
	}

	testHandler := func(client *defaultRpcClientImpl) {
		statusViaHTTP, err := client.statusViaHTTP()
		suite.Require().NoError(err)
		assertResult(statusViaHTTP)

		statusViaWs, err := client.statusViaWebsocket()
		if suite.NoError(err) {
			assertResult(statusViaWs)
			suite.Equal(statusViaWs.NodeInfo, statusViaHTTP.NodeInfo)
		}
	}

	suite.Run("Get Status on Tendermint node", func() {
		testHandler(suite.TM)
	})

	suite.Run("Get Status on CometBFT node", func() {
		testHandler(suite.COMETBFT)
	})
}

func (suite *IntegrationTestSuite) Test_defaultRpcClientImpl_IT_LatestValidators() {
	testHandler := func(client *defaultRpcClientImpl) {
		status, err := client.Status()
		suite.Require().NoError(err)

		testHeight := status.SyncInfo.LatestBlockHeight // use same context for same validator set (even rarely changed)
		suite.Require().Greater(testHeight, int64(0))

		validatorsViaHTTP, err := client.latestValidatorsViaHttp(testHeight)
		suite.Require().NoError(err)
		suite.Require().NotEmpty(validatorsViaHTTP, "expect validator set via HTTP, but got none")

		validatorsViaWs, err := client.latestValidatorsViaWebsocket(testHeight)
		if suite.NoError(err) {
			suite.Require().NotEmpty(validatorsViaWs, "expect validator set via Websocket, but got none")

			suite.Equal(validatorsViaWs, validatorsViaHTTP, "mis-match result validators set between Websocket and HTTP response")
		}
	}

	suite.Run("Get latest validator set on Tendermint node", func() {
		testHandler(suite.TM)
	})

	suite.Run("Get latest validator set on CometBFT node", func() {
		testHandler(suite.COMETBFT)
	})
}

func (suite *IntegrationTestSuite) Test_defaultRpcClientImpl_IT_Shutdown() {
	testHandler := func(client *defaultRpcClientImpl) {
		suite.Require().True(client.rpcWebsocketClient.IsRunning(), "required status running at this point")

		err := client.Shutdown()
		suite.Require().NoError(err, "expect no error at first shutdown")

		for i := 0; i < 3; i++ { // try few more times
			time.Sleep(100 * time.Millisecond)
			err = client.Shutdown()
			suite.Require().NoError(err, "expect no error at later shutdown because error handled correctly")
		}
	}

	suite.Run("Shutdown client of Tendermint node", func() {
		testHandler(suite.TM)
	})

	suite.Run("Shutdown client of CometBFT node", func() {
		testHandler(suite.COMETBFT)
	})
}
