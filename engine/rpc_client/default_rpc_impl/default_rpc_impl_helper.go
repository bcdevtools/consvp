package default_rpc_impl

//goland:noinspection SpellCheckingInspection
import (
	"github.com/pkg/errors"
	tmservice "github.com/tendermint/tendermint/libs/service"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	jsonrpcclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
	"time"
)

func createRpcWebsocketClientToRemoteServer(endpoint normalizedRpcHttpEndpoint) (*rpchttp.HTTP, error) {
	client, err := jsonrpcclient.DefaultHTTPClient(string(endpoint))
	if err != nil {
		return nil, errors.Wrap(err, "Error creating HTTP client for RPC server")
	}
	websocketClient, err := rpchttp.NewWithClient(string(endpoint), "/websocket", client)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating WebSocket client for RPC server")
	}
	err = websocketClient.Start()
	if err != nil && err != tmservice.ErrAlreadyStarted {
		return nil, errors.Wrap(err, "Error starting WebSocket client for RPC server")
	}

	return websocketClient, nil
}

func sleepRetry() {
	time.Sleep(100 * time.Millisecond)
}
