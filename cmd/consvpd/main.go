package main

import (
	"fmt"
	"github.com/bcdevtools/consvp/cmd"
	"github.com/bcdevtools/consvp/constants"
)

// main is entrypoint of the app
func main() {
	fmt.Println(constants.APP_INTRO)
	fmt.Println()

	cmd.Execute()
}
