package default_rpc_impl

//goland:noinspection SpellCheckingInspection
import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/bcdevtools/consvp/engine/rpc"
	enginetypes "github.com/bcdevtools/consvp/engine/types"
	"github.com/bcdevtools/consvp/types"
	"github.com/bcdevtools/consvp/utils"
	"github.com/cosmos/cosmos-sdk/types/query"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/libs/rand"
	tmservice "github.com/tendermint/tendermint/libs/service"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

var _ rpc.RpcClient = (*defaultRpcClientImpl)(nil) // ensure defaultRpcClientImpl implements RpcClient interface

// CONTRACT: must be a valid HTTP endpoint, not ends with '/'.
type normalizedRpcHttpEndpoint string

// defaultRpcClientImpl is the default implementation of RpcClient interface.
// It is expected to work with both Tendermint & CometBFT.
type defaultRpcClientImpl struct {
	mutex *sync.Mutex

	// endpoint is the HTTP endpoint of the RPC server.
	endpoint normalizedRpcHttpEndpoint

	// producerEndpoint is the HTTP endpoint of the RPC server.
	// It is used to query the validators.
	producerEndpoint normalizedRpcHttpEndpoint

	// rpcWebsocketClient is the Websocket client to the RPC server.
	// Default mode, but only available when the RPC server supports Websocket.
	rpcWebsocketClient *rpchttp.HTTP

	// producerRpcWebsocketClient is the Websocket client to the producer RPC server.
	// Default mode, but only available when the RPC server supports Websocket.
	producerRpcWebsocketClient *rpchttp.HTTP

	// cached-information from the RPC server
	statusNetwork string
	statusVersion string
	statusMoniker string
}

// NewDefaultRpcClient returns the default implementation of rpc.RPC interface.
// It does support an optional producer endpoint for compatible with Consumer-architecture chains.
func NewDefaultRpcClient(endpoint, optionalProducerEndpoint string, useWebsocket bool) *defaultRpcClientImpl {
	httpEndpoint := utils.ReplaceAnySchemeWithHttp(endpoint)
	httpEndpoint = strings.TrimSuffix(httpEndpoint, "/")
	producerHttpEndpoint := httpEndpoint
	if len(optionalProducerEndpoint) > 0 {
		producerHttpEndpoint = utils.ReplaceAnySchemeWithHttp(optionalProducerEndpoint)
		producerHttpEndpoint = strings.TrimSuffix(producerHttpEndpoint, "/")
	}

	result := &defaultRpcClientImpl{
		mutex:            &sync.Mutex{},
		endpoint:         normalizedRpcHttpEndpoint(httpEndpoint),
		producerEndpoint: normalizedRpcHttpEndpoint(producerHttpEndpoint),
	}
	var err error
	if useWebsocket {
		result.rpcWebsocketClient, err = createRpcWebsocketClientToRemoteServer(result.endpoint)
		if err != nil {
			fmt.Println("Failed to initialize Websocket connect to remote server, switching to use HTTP client")
			useWebsocket = false
		}
		if result.endpoint == result.producerEndpoint && result.rpcWebsocketClient != nil {
			result.producerRpcWebsocketClient = result.rpcWebsocketClient // reuse the same client
		} else {
			result.producerRpcWebsocketClient, err = createRpcWebsocketClientToRemoteServer(result.producerEndpoint)
			if err != nil {
				fmt.Println("Failed to initialize Websocket connect to remote server, switching to use HTTP client")
				useWebsocket = false
			}
		}
	}

	status, err := result.Status()
	if err != nil {
		panic(errors.Wrap(err, "Error getting status from RPC server, failed to initialize client"))
	}
	result.statusNetwork = status.NodeInfo.Network
	result.statusVersion = status.NodeInfo.Version
	result.statusMoniker = status.NodeInfo.Moniker

	return result
}

// BondedValidators returns the list of bonded validators
func (rpc *defaultRpcClientImpl) BondedValidators() ([]stakingtypes.Validator, error) {
	if rpc.rpcWebsocketClient != nil {
		return rpc.bondedValidatorsViaWebsocket()
	} else {
		return rpc.bondedValidatorsViaHTTP()
	}
}

func (rpc *defaultRpcClientImpl) bondedValidatorsViaWebsocket() ([]stakingtypes.Validator, error) {
	const limit uint64 = 200 // luckily, this endpoint support large page size. 500 is no problem.

	var validators []stakingtypes.Validator
	var nextKey []byte
	var stop = false
	page := 1

	for !stop {
		req := stakingtypes.QueryValidatorsRequest{
			Status: stakingtypes.BondStatusBonded,
			Pagination: &query.PageRequest{
				Limit: limit,
				Key:   nextKey,
			},
		}

		bz, err := req.Marshal()
		if err != nil {
			panic(errors.Wrap(err, "failed to marshal request, weird!"))
		}

		var resultABCIQuery *coretypes.ResultABCIQuery
		var queryValidatorsResponse stakingtypes.QueryValidatorsResponse

		retry := types.DefaultRetryCounterFetchingRpc()

		for retry.Continue() {
			resultABCIQuery, err = rpc.producerRpcWebsocketClient.ABCIQuery(context.Background(), "/cosmos.staking.v1beta1.Query/Validators", bz)
			if err == nil {
				break
			}

			sleepRetry()
		}

		if err != nil {
			return nil, err
		}

		if len(resultABCIQuery.Response.Value) == 0 {
			panic("empty response value, weird!")
		}

		err = queryValidatorsResponse.Unmarshal(resultABCIQuery.Response.Value)
		if err != nil {
			panic(errors.Wrap(err, "failed to unmarshal response, weird!"))
		}

		nextKey = queryValidatorsResponse.Pagination.NextKey
		stop = len(queryValidatorsResponse.Pagination.NextKey) == 0
		validators = append(validators, queryValidatorsResponse.Validators...)
		page++
	}

	return validators, nil
}

func (rpc *defaultRpcClientImpl) bondedValidatorsViaHTTP() ([]stakingtypes.Validator, error) {
	const limit uint64 = 200 // luckily, this endpoint support large page size. 500 is no problem.

	var validators []stakingtypes.Validator
	var nextKey []byte
	var stop = false
	page := 1

	fetchBondedValidators := func(nextKey []byte) (*stakingtypes.QueryValidatorsResponse, error) {
		req := stakingtypes.QueryValidatorsRequest{
			Status: stakingtypes.BondStatusBonded,
			Pagination: &query.PageRequest{
				Limit: limit,
				Key:   nextKey,
			},
		}

		bz, err := req.Marshal()
		if err != nil {
			panic(errors.Wrap(err, "failed to marshal request, weird!"))
		}

		payload := fmt.Sprintf(`{
			"jsonrpc": "2.0",
			"id":      "%s",
			"method":  "abci_query",
			"params": [
				"/cosmos.staking.v1beta1.Query/Validators",
				"%s",
				"0",
				false
			]
		}`, strconv.Itoa(int(rand.Uint16())+1), hex.EncodeToString(bz))

		resp, err := http.Post(string(rpc.producerEndpoint), "application/json", bytes.NewBuffer([]byte(payload)))
		if err != nil {
			return nil, errors.Wrap(err, "error request bonded validators via rpc '/abci_query' endpoint")
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		bz, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "error reading response bonded validators from rpc '/abci_query' endpoint")
		}

		var resContent enginetypes.BaseAbciQueryResponse
		err = json.Unmarshal(bz, &resContent)
		if err != nil {
			return nil, errors.Wrap(err, "error unmarshal response bonded validators from rpc '/abci_query' endpoint")
		}

		bz, err = resContent.GetBuffer()
		if err != nil {
			return nil, err
		}

		var queryValidatorsResponse stakingtypes.QueryValidatorsResponse
		err = queryValidatorsResponse.Unmarshal(bz)
		if err != nil {
			return nil, errors.Wrap(err, "error unmarshal response value bonded validators from rpc '/abci_query' endpoint")
		}

		return &queryValidatorsResponse, nil
	}

	for !stop {
		var queryValidatorsResponse *stakingtypes.QueryValidatorsResponse
		var err error

		retry := types.DefaultRetryCounterFetchingRpc()

		for retry.Continue() {
			queryValidatorsResponse, err = fetchBondedValidators(nextKey)
			if err == nil {
				break
			}

			sleepRetry()
		}

		if err != nil {
			return nil, err
		}

		nextKey = queryValidatorsResponse.Pagination.NextKey
		stop = len(queryValidatorsResponse.Pagination.NextKey) == 0
		validators = append(validators, queryValidatorsResponse.Validators...)
		page++
	}

	return validators, nil
}

func (rpc *defaultRpcClientImpl) ConsensusState() (*enginetypes.RoundState, error) {
	var resultRoundState *enginetypes.RoundState
	var err error

	retry := types.DefaultRetryCounterFetchingRpc()

	for retry.Continue() {
		if rpc.rpcWebsocketClient != nil {
			resultRoundState, err = rpc.consensusStateViaWebsocket()
		} else {
			resultRoundState, err = rpc.consensusStateViaHTTP()
		}
		if err == nil {
			break
		}

		sleepRetry()
	}

	return resultRoundState, err
}

func (rpc *defaultRpcClientImpl) consensusStateViaWebsocket() (*enginetypes.RoundState, error) {
	if rpc.rpcWebsocketClient == nil {
		return nil, errors.New("Websocket client is not available")
	}
	res, err := rpc.rpcWebsocketClient.ConsensusState(context.Background())
	if err != nil {
		return nil, err
	}
	var rs enginetypes.RoundState
	err = json.Unmarshal(res.RoundState, &rs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal RoundState")
	}
	return &rs, nil
}

func (rpc *defaultRpcClientImpl) consensusStateViaHTTP() (*enginetypes.RoundState, error) {
	resp, err := http.Get(fmt.Sprintf("%s/consensus_state", rpc.endpoint))
	if err != nil {
		return nil, errors.Wrap(err, "error request rpc '/consensus_state' endpoint")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	bz, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading response from rpc '/consensus_state' endpoint")
	}

	fmt.Println(string(bz))
	var resContent enginetypes.BaseRpcResponse[enginetypes.RoundStateResponse]
	err = json.Unmarshal(bz, &resContent)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshal response from rpc '/consensus_state' endpoint")
	}

	err = resContent.Error.GetError()
	if err != nil {
		return nil, err
	}

	if resContent.Result.RoundState == nil {
		return nil, errors.New("empty round state information")
	}

	return resContent.Result.RoundState, nil
}

// Status returns the status of the RPC server
func (rpc *defaultRpcClientImpl) Status() (*coretypes.ResultStatus, error) {
	var resultStatus *coretypes.ResultStatus
	var err error

	retry := types.DefaultRetryCounterFetchingRpc()

	for retry.Continue() {
		if rpc.rpcWebsocketClient != nil {
			resultStatus, err = rpc.statusViaWebsocket()
		} else {
			resultStatus, err = rpc.statusViaHTTP()
		}
		if err == nil {
			break
		}

		sleepRetry()
	}

	return resultStatus, err
}

func (rpc *defaultRpcClientImpl) statusViaWebsocket() (*coretypes.ResultStatus, error) {
	if rpc.rpcWebsocketClient == nil {
		return nil, errors.New("Websocket client is not available")
	}
	return rpc.rpcWebsocketClient.Status(context.Background())
}

func (rpc *defaultRpcClientImpl) statusViaHTTP() (*coretypes.ResultStatus, error) {
	resp, err := http.Get(fmt.Sprintf("%s/status", rpc.endpoint))
	if err != nil {
		return nil, errors.Wrap(err, "error request rpc '/status' endpoint")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	bz, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading response from rpc '/status' endpoint")
	}

	var resContent enginetypes.BaseRpcResponse[coretypes.ResultStatus]
	err = json.Unmarshal(bz, &resContent)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshal response from rpc '/status' endpoint")
	}

	err = resContent.Error.GetError()
	if err != nil {
		return nil, err
	}

	return resContent.Result, nil
}

// LatestValidators returns the most recent validator set
func (rpc *defaultRpcClientImpl) LatestValidators() ([]*tmtypes.Validator, error) {
	if rpc.rpcWebsocketClient != nil {
		return rpc.latestValidatorsViaWebsocket(0)
	} else {
		return rpc.latestValidatorsViaHttp(0)
	}
}

func (rpc *defaultRpcClientImpl) latestValidatorsViaWebsocket(height int64) ([]*tmtypes.Validator, error) {
	if rpc.rpcWebsocketClient == nil {
		return nil, errors.New("Websocket client is not available")
	}

	var latestHeight *int64
	var page int
	var perPage int

	if height > 0 {
		latestHeight = &height
	}

	page = 1
	perPage = 100
	var validators []*tmtypes.Validator

	var stop bool

	for !stop {
		var resVals *coretypes.ResultValidators
		var err error

		retry := types.DefaultRetryCounterFetchingRpc()

		for retry.Continue() {
			resVals, err = rpc.producerRpcWebsocketClient.Validators(context.Background(), latestHeight, &page, &perPage)

			if err == nil {
				break
			}

			sleepRetry()
		}

		if err != nil {
			return nil, err
		}

		page++

		for _, validator := range resVals.Validators {
			validators = append(validators, validator) // assume validator set not changed
		}

		stop = len(validators) >= resVals.Total
	}

	return validators, nil
}

func (rpc *defaultRpcClientImpl) latestValidatorsViaHttp(height int64) ([]*tmtypes.Validator, error) {
	var page int

	page = 1
	const perPage = 100
	var validators []*tmtypes.Validator

	var stop bool

	fetchValidators := func(page int) (*coretypes.ResultValidators, error) {
		url := fmt.Sprintf("%s/validators?per_page=%d&page=%d", rpc.producerEndpoint, perPage, page)
		if height > 0 {
			url += fmt.Sprintf("&height=%d", height)
		}
		resp, err := http.Get(url)
		if err != nil {
			return nil, errors.Wrap(err, "error request rpc '/validators' endpoint")
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		bz, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "error reading response from rpc '/validators' endpoint")
		}

		var resContent enginetypes.BaseRpcResponse[coretypes.ResultValidators]
		err = json.Unmarshal(bz, &resContent)
		if err != nil {
			return nil, errors.Wrap(err, "error unmarshal response from rpc '/validators' endpoint")
		}

		err = resContent.Error.GetError()
		if err != nil {
			return nil, err
		}

		return resContent.Result, nil
	}

	for !stop {
		var resVals *coretypes.ResultValidators
		var err error

		retry := types.DefaultRetryCounterFetchingRpc()

		for retry.Continue() {
			resVals, err = fetchValidators(page)

			if err == nil {
				break
			}

			sleepRetry()
		}

		if err != nil {
			return nil, err
		}

		page++

		for _, validator := range resVals.Validators {
			validators = append(validators, validator) // assume validator set not changed
		}

		stop = len(validators) >= resVals.Total
	}

	return validators, nil
}

// Shutdown must be called when the RPC client is no longer needed.
// It does close up all the connections to the RPC server and free resources.
func (rpc *defaultRpcClientImpl) Shutdown() error {
	shutdownWebsocketClient := func(client *rpchttp.HTTP) {
		if client == nil {
			return
		}

		err := client.Stop()
		if err == nil {
			return
		}

		if err == tmservice.ErrNotStarted {
			// ignore
		} else if err == tmservice.ErrAlreadyStopped {
			// ignore
		} else {
			panic(err)
		}
	}

	shutdownWebsocketClient(rpc.rpcWebsocketClient)
	shutdownWebsocketClient(rpc.producerRpcWebsocketClient)

	return nil
}
