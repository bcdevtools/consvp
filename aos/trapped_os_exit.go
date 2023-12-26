package aos

import (
	"github.com/bcdevtools/consvp/utils"
	"os"
)

// Exit is a wrapper for os.Exit() to execute exit functions to cleanups, etc... before exiting.
func Exit(code int) {
	utils.AppExitHelper.ExecuteFunctionsUponAppExit()
	os.Exit(code)
}
