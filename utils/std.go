package utils

import (
	"fmt"
	"os"
)

func PrintlnStdErr(a ...any) {
	fmt.Fprintln(os.Stderr, a...)
}

func PrintfStdErr(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format, a...)
}
