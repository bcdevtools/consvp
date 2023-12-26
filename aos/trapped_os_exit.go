package aos

import (
	"github.com/bcdevtools/consvp/utils"
	"os"
)

// Exit is a wrapper for os.Exit() to execute exit functions to cleanups, etc... before exiting.
//
// Why? Because os.Exit() directly terminate application and does not execute defer functions.
func Exit(code int) {
	utils.AppExitHelper.ExecuteFunctionsUponAppExit()
	os.Exit(code)
}
