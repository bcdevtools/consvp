package cmd

import (
	"fmt"
	"github.com/bcdevtools/consvp/aos"
	"github.com/bcdevtools/consvp/constants"
	corecodec "github.com/bcdevtools/cvp-streaming-core/codec"
	"github.com/spf13/cobra"
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
			aos.Exit(0)
		}
	},
	Run: pvtopHandler,
}

func Execute() {
	rootCmd.Flags().Bool(flagHttp, false, "use http call for rpc client instead of default is websocket")
	rootCmd.Flags().BoolP(flagRapidRefresh, "r", false, fmt.Sprintf("refresh rate quicker, default is %v will be changed to %v", defaultRefreshInterval, rapidRefreshInterval))
	rootCmd.Flags().BoolP(flagStreaming, "s", false, "open a live-streaming pre-vote session to be able to share the view with others.")
	rootCmd.Flags().String(flagCodec, string(corecodec.NewProxyCvpCodec().GetVersion()), "specify codec version to be used to encode the streaming data, mostly used for testing purpose or workaround when the default codec version has bug.")
	rootCmd.Flags().Bool(flagResumeStreaming, false, "resume an opened live-streaming pre-vote session to keep the current shared URL.")
	rootCmd.Flags().StringP(flagMockStreamingServer, "t", "none", "for testing purpose only, mock a streaming server or connect to local streaming server to test the streaming client.")

	rootCmd.Flags().BoolP(flagVersion, "v", false, "print the binary version. WARN: This action will bypass the main command handler.")
	rootCmd.Flags().Bool(flagLongVersion, false, fmt.Sprintf("print extra version information, must be used with --%s", flagVersion))

	rootCmd.CompletionOptions.HiddenDefaultCmd = true    // hide the 'completion' subcommand
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true}) // hide the 'help' subcommand

	err := rootCmd.Execute()
	if err != nil {
		aos.Exit(1)
	}
}
