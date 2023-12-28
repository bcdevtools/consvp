package utils

import (
	"fmt"
	"os"
)

func PrintlnStdErr(a ...any) {
	_, _ = fmt.Fprintln(os.Stderr, a...)
}

//goland:noinspection GoUnusedExportedFunction
func PrintfStdErr(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
}
