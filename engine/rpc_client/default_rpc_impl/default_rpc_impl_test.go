package default_rpc_impl

import (
	"github.com/stretchr/testify/require"
	"regexp"
	"strings"
	"testing"
)

func TestNewDefaultRpcClient(t *testing.T) {
	tests := []struct {
		name      string
		endpoint  string
		wantError bool
	}{
		{
			name:      "init default (Tendermint)",
			endpoint:  DEFAULT_TENDERMINT_RPC_URL_FOR_TEST,
			wantError: false,
		},
		{
			name:      "init default (CometBFT)",
			endpoint:  DEFAULT_COMET_BFT_RPC_URL_FOR_TEST,
			wantError: false,
		},
		{
			name:      "bad endpoint",
			endpoint:  "https://google.com:55555",
			wantError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				err := recover()
				if err != nil {
					if tt.wantError {
						// pass
					} else {
						require.Fail(t, "unexpected error", err)
					}
				} else {
					if tt.wantError {
						require.Fail(t, "expected error but got none")
					} else {
						// pass
					}
				}
			}()

			client := NewDefaultRpcClient(tt.endpoint, "", false)

			defer func() {
				_ = client.Shutdown()
			}()

			require.NotEmpty(t, string(client.endpoint))
			require.True(t, strings.HasPrefix(string(client.endpoint), "http"))
			require.NotNil(t, client.rpcWebsocketClient)
			require.NotEmpty(t, client.statusNetwork)
			require.True(t, regexp.MustCompile("^[a-zA-Z\\d]+-\\d+$").MatchString(client.statusNetwork))
			require.NotEmpty(t, client.statusMoniker)
			require.NotEmpty(t, client.statusVersion)
			require.True(t, regexp.MustCompile("^v?\\d+(\\.\\d+){2}$").MatchString(client.statusVersion))
		})
	}
}
