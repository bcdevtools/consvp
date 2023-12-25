package codec

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

func Test_fromToUint16Buffer(t *testing.T) {
	for n1 := 0; n1 <= math.MaxUint16; n1++ {
		bz := toUint16Buffer(n1)
		n2 := fromUint16Buffer(bz)

		if n1 != n2 {
			t.Errorf("n: %d, n2: %d", n1, n2)
		}
	}

	t.Run("overflow", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expect panic")
			}
		}()
		toUint16Buffer(math.MaxUint16 + 1)
	})

	t.Run("underflow", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expect panic")
			}
		}()
		toUint16Buffer(-1)
	})

	t.Run("invalid buffer length", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expect panic")
			}
		}()
		fromUint16Buffer([]byte{1, 2, 3})
	})

	t.Run("invalid buffer length", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expect panic")
			}
		}()
		fromUint16Buffer([]byte{1})
	})
}

func Test_fromToPercentBuffer(t *testing.T) {
	for p1 := 0.0; p1 <= 100; p1 += 0.01 {
		bz := toPercentBuffer(p1)
		p2 := fromPercentBuffer(bz)

		if p1 != p2 {
			if math.Abs(p1-p2) > 0.01 {
				t.Errorf("p1: %f, p2: %f", p1, p2)
			}
		}
	}

	t.Run("overflow", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expect panic")
			}
		}()
		toPercentBuffer(100.01)
	})

	t.Run("underflow", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expect panic")
			}
		}()
		toPercentBuffer(-0.01)
	})

	t.Run("invalid buffer length", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expect panic")
			}
		}()
		fromPercentBuffer([]byte{1, 2, 3})
	})

	t.Run("invalid buffer length", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expect panic")
			}
		}()
		fromPercentBuffer([]byte{1})
	})
}

func Test_tryTakeNBytesFrom(t *testing.T) {
	var tests = []struct {
		name        string
		bz          []byte
		fromIndex   int
		size        int
		want        []byte
		wantSuccess bool
	}{
		{
			name:        "normal",
			bz:          []byte{1, 2, 3, 4, 5},
			fromIndex:   0,
			size:        5,
			want:        []byte{1, 2, 3, 4, 5},
			wantSuccess: true,
		},
		{
			name:        "from start, take less than size",
			bz:          []byte{1, 2, 3, 4, 5},
			fromIndex:   0,
			size:        4,
			want:        []byte{1, 2, 3, 4},
			wantSuccess: true,
		},
		{
			name:        "from middle, take less than remaining",
			bz:          []byte{1, 2, 3, 4, 5},
			fromIndex:   1,
			size:        3,
			want:        []byte{2, 3, 4},
			wantSuccess: true,
		},
		{
			name:        "from middle, take to end",
			bz:          []byte{1, 2, 3, 4, 5},
			fromIndex:   1,
			size:        4,
			want:        []byte{2, 3, 4, 5},
			wantSuccess: true,
		},
		{
			name:        "from start, take more than size",
			bz:          []byte{1, 2, 3, 4, 5},
			fromIndex:   0,
			size:        6,
			wantSuccess: false,
		},
		{
			name:        "from middle, take more than remaining",
			bz:          []byte{1, 2, 3, 4, 5},
			fromIndex:   1,
			size:        5,
			wantSuccess: false,
		},
		{
			name:        "from out of bound",
			bz:          []byte{1, 2, 3, 4, 5},
			fromIndex:   5,
			size:        5,
			wantSuccess: false,
		},
		{
			name:        "from end, take one",
			bz:          []byte{1, 2, 3, 4, 5},
			fromIndex:   4,
			size:        1,
			want:        []byte{5},
			wantSuccess: true,
		},
		{
			name:        "from end, take more than remaining",
			bz:          []byte{1, 2, 3, 4, 5},
			fromIndex:   4,
			size:        2,
			wantSuccess: false,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_try", tt.name), func(t *testing.T) {
			got, gotSuccess := tryTakeNBytesFrom(tt.bz, tt.fromIndex, tt.size)
			if gotSuccess != tt.wantSuccess {
				t.Errorf("tryTakeNBytesFrom() success = %v, want succes %v", gotSuccess, tt.wantSuccess)
			} else if gotSuccess {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("tryTakeNBytesFrom() got = %v, want %v", got, tt.want)
				}
			}
		})
		t.Run(fmt.Sprintf("%s_must", tt.name), func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) == tt.wantSuccess {
					t.Errorf("mustTakeNBytesFrom() success = %t, want succes %t", r == nil, tt.wantSuccess)
				}
			}()

			got := mustTakeNBytesFrom(tt.bz, tt.fromIndex, tt.size)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mustTakeNBytesFrom() got = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("invalid size", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expect panic")
			}
		}()
		tryTakeNBytesFrom([]byte{1, 2, 3}, 0, 0)
	})

	t.Run("invalid size", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expect panic")
			}
		}()
		mustTakeNBytesFrom([]byte{1, 2, 3}, 0, 0)
	})

	t.Run("invalid beginning index", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expect panic")
			}
		}()
		tryTakeNBytesFrom([]byte{1, 2, 3}, -1, 1)
	})

	t.Run("invalid beginning index", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expect panic")
			}
		}()
		mustTakeNBytesFrom([]byte{1, 2, 3}, -1, 1)
	})
}

func Test_takeUntilSeparatorOrEnd(t *testing.T) {
	tests := []struct {
		name      string
		bz        []byte
		fromIndex int
		separator byte
		wantTaken []byte
	}{
		{
			name:      "normal",
			bz:        []byte{1, 2, 3, 4, 5},
			fromIndex: 0,
			separator: 6,
			wantTaken: []byte{1, 2, 3, 4, 5},
		},
		{
			name:      "normal",
			bz:        []byte{1, 2, 3, 4, 5},
			fromIndex: 0,
			separator: 4,
			wantTaken: []byte{1, 2, 3},
		},
		{
			name:      "from out of bound",
			bz:        []byte{1, 2, 3, 4, 5},
			fromIndex: 5,
			separator: 4,
			wantTaken: []byte{},
		},
		{
			name:      "normal",
			bz:        []byte{1, 2, 3, 4, 5},
			fromIndex: 3,
			separator: 2,
			wantTaken: []byte{4, 5},
		},
		{
			name:      "normal",
			bz:        []byte{1, 2, 3, 4, 5, 2, 3},
			fromIndex: 3,
			separator: 2,
			wantTaken: []byte{4, 5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTaken := takeUntilSeparatorOrEnd(tt.bz, tt.fromIndex, tt.separator)
			if len(gotTaken) != len(tt.wantTaken) {
				t.Errorf("takeUntilSeparatorOrEnd() = %v, want %v", gotTaken, tt.wantTaken)
			} else if len(gotTaken) == 0 {
				// ok
			} else if !reflect.DeepEqual(gotTaken, tt.wantTaken) {
				t.Errorf("takeUntilSeparatorOrEnd() = %v, want %v", gotTaken, tt.wantTaken)
			}
		})
	}
}

func Test_sanitizeMoniker(t *testing.T) {
	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name    string
		moniker string
		want    string
	}{
		{
			name:    "all replace able",
			moniker: `<he'llo">`,
			want:    "(he`llo`)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeMoniker(tt.moniker); got != tt.want {
				t.Errorf("sanitizeMoniker() = %v, want %v", got, tt.want)
			}
		})
	}
}
