package utils

import "testing"

func TestReplaceAnySchemeWithHttp(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		want     string
	}{
		{
			name:     "add HTTP to no scheme",
			endpoint: "localhost:26657",
			want:     "http://localhost:26657",
		},
		{
			name:     "add HTTP to relative-scheme",
			endpoint: "://localhost:26657",
			want:     "http://localhost:26657",
		},
		{
			name:     "add HTTP to no-scheme",
			endpoint: "//localhost:26657",
			want:     "http://localhost:26657",
		},
		{
			name:     "keep HTTP",
			endpoint: "http://localhost:26657",
			want:     "http://localhost:26657",
		},
		{
			name:     "keep HTTPS",
			endpoint: "https://localhost:26657",
			want:     "https://localhost:26657",
		},
		{
			name:     "tcp should be replaced with http",
			endpoint: "tcp://localhost:26657",
			want:     "http://localhost:26657",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReplaceAnySchemeWithHttp(tt.endpoint); got != tt.want {
				t.Errorf("ReplaceAnySchemeWithHttp() = %v, want %v", got, tt.want)
			}
		})
	}
}
