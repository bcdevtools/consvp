package cmd

import "testing"

func Test_readPvTopArg(t *testing.T) {
	type args struct {
		args     []string
		index    int
		optional bool
	}
	tests := []struct {
		name    string
		args    args
		wantArg string
		wantErr bool
	}{
		{
			name: "empty input args",
			args: args{
				args:     []string{},
				index:    0,
				optional: false,
			},
			wantArg: "",
			wantErr: true,
		},
		{
			name: "empty input args",
			args: args{
				args:     []string{},
				index:    1,
				optional: false,
			},
			wantArg: "",
			wantErr: true,
		},
		{
			name: "mission optional arg",
			args: args{
				args:     []string{"arg1"},
				index:    1,
				optional: true,
			},
			wantArg: "",
			wantErr: false,
		},
		{
			name: "auto correct port",
			args: args{
				args:     []string{"26657"},
				index:    0,
				optional: false,
			},
			wantArg: "http://localhost:26657",
			wantErr: false,
		},
		{
			name: "auto correct :port",
			args: args{
				args:     []string{":26657"},
				index:    0,
				optional: false,
			},
			wantArg: "http://localhost:26657",
			wantErr: false,
		},
		{
			name: "do not auto correct invalid port pattern",
			args: args{
				args:     []string{"26657x"},
				index:    0,
				optional: false,
			},
			wantArg: "26657x",
			wantErr: false,
		},
		{
			name: "do not auto correct invalid port pattern",
			args: args{
				args:     []string{"x26657"},
				index:    0,
				optional: false,
			},
			wantArg: "x26657",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotArg, err := readPvTopArg(tt.args.args, tt.args.index, tt.args.optional)
			if (err != nil) != tt.wantErr {
				t.Errorf("readPvTopArg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotArg != tt.wantArg {
				t.Errorf("readPvTopArg() gotArg = %v, want %v", gotArg, tt.wantArg)
			}
		})
	}
}
