package codec

//goland:noinspection SpellCheckingInspection
import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"strings"
)

func toUint16Buffer(num int) []byte {
	if num < 0 || num > math.MaxUint16 {
		panic(fmt.Errorf("overflow uint16: %d", num))
	}
	bz := make([]byte, 2)
	binary.BigEndian.PutUint16(bz, uint16(num))
	return bz
}

func fromUint16Buffer(bz []byte) int {
	if len(bz) != 2 {
		panic(fmt.Errorf("invalid uint16 buffer length: %d, require 2", len(bz)))
	}
	return int(binary.BigEndian.Uint16(bz))
}

func toPercentBuffer(percent float64) []byte {
	if percent < 0 || percent > 100 {
		panic(fmt.Errorf("overflow percent: %f", percent))
	}
	var pi, pf byte
	str := fmt.Sprintf("%.2f", percent)
	parts := strings.Split(str, ".")
	if len(parts) != 2 {
		pi = byte(percent)
	} else {
		ipi, _ := strconv.Atoi(parts[0])
		ipf, _ := strconv.Atoi(parts[1])
		pi = byte(ipi)
		pf = byte(ipf)
	}
	return []byte{pi, pf}
}

func fromPercentBuffer(bz []byte) float64 {
	if len(bz) != 2 {
		panic(fmt.Errorf("invalid percent buffer length: %d, require 2", len(bz)))
	}
	return float64(bz[0]) + float64(bz[1])/100
}

func tryTakeNBytesFrom(bz []byte, fromIndex, size int) ([]byte, bool) {
	if size < 1 {
		panic("invalid size")
	}
	if fromIndex+size > len(bz) {
		return nil, false
	}
	return bz[fromIndex : fromIndex+size], true
}

func takeUntilSeparatorOrEnd(bz []byte, fromIndex int, separator byte) (taken []byte) {
	for i := fromIndex; i < len(bz); i++ {
		if bz[i] == separator {
			break
		}
		taken = append(taken, bz[i])
	}
	return
}

func sanitizeMoniker(moniker string) string {
	moniker = strings.ReplaceAll(moniker, "<", "(")
	moniker = strings.ReplaceAll(moniker, ">", ")")
	moniker = strings.ReplaceAll(moniker, "'", "`")
	moniker = strings.ReplaceAll(moniker, "\"", "`")
	return moniker
}
