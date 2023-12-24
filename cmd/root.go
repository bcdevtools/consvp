package cmd

import (
	"fmt"
	"github.com/bcdevtools/consvp/constants"
	"github.com/spf13/cobra"
	"os"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: fmt.Sprintf("%s [port/host/consumer] [?optionalProvider/optionalProviderPort]", constants.BINARY_NAME),
	Long: `Show pre-vote. Provider/consumer mode is typically for CosmosHub only.
If no arguments are provided, the default port is 26657 and the default host is localhost.
If only a port is provided, the default host is localhost.
`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if cmd.Flags().Changed(flagVersion) {
			versionHandler(cmd, args)
			os.Exit(0)
		}
		if cmd.Flags().Changed(flagUpdate) {
			fmt.Println("Updating binary version...")
			updateHandler(cmd, args)
			os.Exit(0)
		}
	},
	Run: pvtopHandler,
}

func Execute() {
	rootCmd.Flags().Bool(flagHttp, false, "use http call for rpc client instead of default is websocket")
	rootCmd.Flags().BoolP(flagRapidRefresh, "r", false, fmt.Sprintf("refresh rate quicker, default is %v will be changed to %v", defaultRefreshInterval, rapidRefreshInterval))
	rootCmd.Flags().BoolP(flagStreaming, "s", false, "open a live-streaming pre-vote session to be able to share the view with others.")
	rootCmd.Flags().Bool(flagResumeStreaming, false, "resume an opened live-streaming pre-vote session to keep the current shared URL.")
	rootCmd.Flags().StringP(flagMockStreamingServer, "t", "none", "for testing purpose only, mock a streaming server or connect to local streaming server to test the streaming client.")

	rootCmd.Flags().BoolP(flagVersion, "v", false, "print the binary version. WARN: This action will bypass the main command handler.")
	rootCmd.Flags().Bool(flagLongVersion, false, fmt.Sprintf("print extra version information, must be used with --%s", flagVersion))

	rootCmd.Flags().Bool(flagUpdate, false, "update binary version. WARN: This action will bypass the main command handler.")

	rootCmd.CompletionOptions.HiddenDefaultCmd = true    // hide the 'completion' subcommand
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true}) // hide the 'help' subcommand

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
