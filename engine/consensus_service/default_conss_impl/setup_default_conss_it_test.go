package default_conss_impl

//goland:noinspection SpellCheckingInspection
import (
	"github.com/bcdevtools/consvp/engine/rpc_client/default_rpc_impl"
	"github.com/stretchr/testify/suite"
	"testing"
)

//goland:noinspection SpellCheckingInspection
type IntegrationTestSuite struct {
	suite.Suite
	SVC *defaultConsensusServiceClientImpl
}

func TestDefaultConsensusServiceImplIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (suite *IntegrationTestSuite) SetupSuite() {
	suite.SVC = NewDefaultConsensusServiceClientImpl(default_rpc_impl.NewDefaultRpcClient(DEFAULT_RPC_URL_FOR_TEST, "", true))
}

func (suite *IntegrationTestSuite) SetupTest() {
}

func (suite *IntegrationTestSuite) TearDownTest() {
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	_ = suite.SVC.Shutdown()
}
