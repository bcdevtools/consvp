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

//goland:noinspection SpellCheckingInspection
const updateCli = "go install -v github.com/bcdevtools/consvp/cmd/consvpd@latest"

// updateCmd represents the update command, it updates current binary to the latest version
//
//goland:noinspection SpellCheckingInspection
var updateCmd = &cobra.Command{
	Use: "update",
	Short: fmt.Sprintf(`Update binary %s to the latest version by running command:
> %s`, constants.BINARY_NAME, updateCli),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Current version: %s\n", constants.VERSION)
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
		bz, _ = exec.Command(constants.BINARY_NAME, "version").Output()
		fmt.Println(string(bz))
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
