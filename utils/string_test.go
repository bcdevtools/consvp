package utils

import (
	"encoding/hex"
	"fmt"
	coreutils "github.com/bcdevtools/cvp-streaming-core/utils"
	"testing"
)

func TestTruncateStringUntilBufferLessThanXBytesOrFillWithSpaceSuffix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxBytes int
		want     string
	}{
		{
			name:     "less than 20 bytes",
			input:    "abc",
			maxBytes: 20,
			want:     "abc                 ",
		},
		{
			name:     "more than 20 bytes",
			input:    "123456789012345678901",
			maxBytes: 20,
			want:     "12345678901234567890",
		},
		{
			name:     "empty",
			input:    "",
			maxBytes: 20,
			want:     "                    ",
		},
		{
			name:     "UTF8 more than 20 bytes",
			input:    "✅✅✅✅✅✅✅✅✅✅✅✅✅✅✅✅✅✅✅✅",
			maxBytes: 20,
			want:     "✅✅✅✅✅✅  ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := coreutils.TruncateStringUntilBufferLessThanXBytesOrFillWithSpaceSuffix(tt.input, tt.maxBytes); string(got) != tt.want {
				fmt.Println("got buffer", hex.EncodeToString(got))
				t.Errorf("TruncateStringUntilBufferLessThanXBytesOrFillWithSpaceSuffix() = %s, want %s", string(got), tt.want)
			}
		})
	}
}
