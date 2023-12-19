package default_conss_impl

import "testing"

func Test_extractFingerprintBlockHashVotedOn(t *testing.T) {
	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name       string
		voteString string
		want       string
	}{
		{
			name:       "Prevote",
			voteString: `Vote{56789:6AF1F4111082 12345/02/SIGNED_MSG_TYPE_PREVOTE(Prevote) 8B01023386C3 000000000000 @ 2017-12-25T03:00:01.234Z}`,
			want:       "8B01023386C3",
		},
		{
			name:       "Precommit",
			voteString: `Vote{56789:6AF1F4111082 12345/02/SIGNED_MSG_TYPE_PRECOMMIT(Precommit) 8B01023386C3 000000000000 @ 2017-12-25T03:00:01.234Z}`,
			want:       "8B01023386C3",
		},
		{
			name:       "Voting zero",
			voteString: `Vote{56789:6AF1F4111082 12345/02/SIGNED_MSG_TYPE_PRECOMMIT(Precommit) 000000000000 000000000000 @ 2017-12-25T03:00:01.234Z}`,
			want:       "000000000000",
		},
		{
			name:       "Abnormal, not enough 6 bytes",
			voteString: `Vote{56789:6AF1F4111082 12345/02/SIGNED_MSG_TYPE_PRECOMMIT(Precommit) 8B01023386C 000000000000 @ 2017-12-25T03:00:01.234Z}`,
			want:       "????????????",
		},
		{
			name:       "Abnormal, missing @",
			voteString: `Vote{56789:6AF1F4111082 12345/02/SIGNED_MSG_TYPE_PRECOMMIT(Precommit) 8B01023386C3 000000000000 # 2017-12-25T03:00:01.234Z}`,
			want:       "????????????",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractFingerprintBlockHashVotedOn(tt.voteString); got != tt.want {
				t.Errorf("extractFingerprintBlockHashVotedOn() = %v, want %v", got, tt.want)
			}
		})
	}
}
