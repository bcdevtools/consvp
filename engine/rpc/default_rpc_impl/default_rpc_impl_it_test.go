package default_rpc_impl

//goland:noinspection SpellCheckingInspection
import (
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"time"
)

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
