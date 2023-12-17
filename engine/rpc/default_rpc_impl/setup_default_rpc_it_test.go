package default_rpc_impl

//goland:noinspection SpellCheckingInspection
import (
	"github.com/stretchr/testify/suite"
	"testing"
)

//goland:noinspection SpellCheckingInspection
type IntegrationTestSuite struct {
	suite.Suite
	// Node that running Tendermint, for testing compatible purpose
	TM *defaultRpcClientImpl
	// Node that running CometBFT, for testing compatible purpose
	COMETBFT *defaultRpcClientImpl
}

func TestDefaultRpcClientImplIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (suite *IntegrationTestSuite) SetupSuite() {
	// Setup clients with websocket enabled for testing both cases
	suite.TM = NewDefaultRpcClient(DEFAULT_TENDERMINT_RPC_URL_FOR_TEST, "", true)
	suite.COMETBFT = NewDefaultRpcClient(DEFAULT_COMET_BFT_RPC_URL_FOR_TEST, "", true)
}

func (suite *IntegrationTestSuite) SetupTest() {
}

func (suite *IntegrationTestSuite) TearDownTest() {
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	_ = suite.TM.Shutdown()
	_ = suite.COMETBFT.Shutdown()
}
