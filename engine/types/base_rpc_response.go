package types

import "fmt"

type BaseRpcResponse[T any] struct {
	Error  *BaseRpcResponseError `json:"error"`
	Result *T                    `json:"result"`
}

type BaseRpcResponseError struct {
	Message string `json:"message"`
	Data    string `json:"data"`
}

func (re *BaseRpcResponseError) GetError() error {
	if re == nil {
		return nil
	}

	if len(re.Message) < 1 {
		return nil
	}

	return fmt.Errorf("%s: %s", re.Message, re.Data)
}
