package cmd

import (
	"fmt"
	"github.com/bcdevtools/consvp/constants"
	"github.com/bcdevtools/consvp/utils"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"strings"
)

const (
	flagUpdate = "update"
)

//goland:noinspection SpellCheckingInspection
const updateCli = "go install -v github.com/bcdevtools/consvp/cmd/cvp@latest"

func updateHandler(_ *cobra.Command, _ []string) {
	fmt.Printf("Current version: %s\n", constants.VERSION)

	fmt.Println("Executing command:")
	fmt.Println(">", updateCli)
	spl := strings.Split(updateCli, " ")
	bz, err := exec.Command(spl[0], spl[1:]...).Output()
	if err != nil {
		utils.PrintfStdErr("Failed to update binary %s: %s\n", constants.BINARY_NAME, err)
		if len(bz) > 0 {
			utils.PrintlnStdErr(string(bz))
		}
		os.Exit(1)
	}
	fmt.Println("New version:")
	bz, _ = exec.Command(constants.BINARY_NAME, fmt.Sprintf("--%s", flagVersion)).Output()
	fmt.Println(string(bz))
}
