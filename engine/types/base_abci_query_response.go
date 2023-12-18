package types

//goland:noinspection SpellCheckingInspection
import (
	"fmt"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
)

type BaseAbciQueryResponse struct {
	Result *coretypes.ResultABCIQuery `json:"result"`
}

type BaseAbciQueryResponseResultResponse struct {
	Code  int    `json:"code"`
	Log   string `json:"log"`
	Value string `json:"value"`
	Index string `json:"index"`
}

func (re *BaseAbciQueryResponse) GetBuffer() ([]byte, error) {
	err := re.GetError()
	if err != nil {
		return nil, err
	}

	return re.Result.Response.Value, nil
}

func (re *BaseAbciQueryResponse) GetError() error {
	if re == nil {
		panic("struct has not been initialized")
	}

	if re.Result == nil {
		return fmt.Errorf("missing result")
	}

	if re.Result.Response.Code == 0 {
		return nil
	}

	return fmt.Errorf("code %d: %s", re.Result.Response.Code, re.Result.Response.Log)
}
