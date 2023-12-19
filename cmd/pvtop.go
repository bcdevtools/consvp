package cmd

//goland:noinspection SpellCheckingInspection
import (
	"fmt"
	conss "github.com/bcdevtools/consvp/engine/consensus_service"
	dconsi "github.com/bcdevtools/consvp/engine/consensus_service/default_conss_impl"
	"github.com/bcdevtools/consvp/engine/rpc_client"
	drpci "github.com/bcdevtools/consvp/engine/rpc_client/default_rpc_impl"
	enginetypes "github.com/bcdevtools/consvp/engine/types"
	"github.com/bcdevtools/consvp/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	flagHttp         = "http"
	flagRapidRefresh = "rapid-refresh"
)

const defaultRefreshInterval = 3 * time.Second
const rapidRefreshInterval = 1 * time.Second

// pvtopCmd represents the version command, it prints the current version of the binary
var pvtopCmd = &cobra.Command{
	Use:     "pvtop [port/host/consumer] [?optionalProvider/optionalProviderPort]",
	Aliases: []string{"pv"},
	Long: `Show pre-vote. Provider/consumer mode is typically for CosmosHub only.
If no arguments are provided, the default port is 26657 and the default host is localhost.
If only a port is provided, the default host is localhost.
`,
	Run: pvtopHandler,
}

func pvtopHandler(cmd *cobra.Command, args []string) {
	if len(args) < 1 || len(args) > 2 {
		utils.PrintlnStdErr("ERR: Invalid number of arguments")
		os.Exit(1)
	}

	consumerUrl, err := readPvTopArg(args, 0, true)
	if err != nil {
		utils.PrintlnStdErr(err)
		os.Exit(1)
	}

	if consumerUrl == "" {
		consumerUrl = "http://localhost:26657"
		fmt.Println("No port/host/consumer provided, using default:", consumerUrl)
	}

	providerUrl, err := readPvTopArg(args, 1, true)
	if err != nil {
		utils.PrintlnStdErr(err)
		os.Exit(1)
	}

	useHttp, _ := cmd.Flags().GetBool(flagHttp)

	var rpcClient rpc_client.RpcClient
	var consensusService conss.ConsensusService
	defer func() {
		if rpcClient != nil {
			_ = rpcClient.Shutdown()
		}
		if consensusService != nil {
			_ = consensusService.Shutdown()
		}
	}()

	rpcClient = drpci.NewDefaultRpcClient(consumerUrl, providerUrl, !useHttp)
	consensusService = dconsi.NewDefaultConsensusServiceClientImpl(rpcClient)

	refreshTicker := time.NewTicker(func() time.Duration {
		if cmd.Flags().Changed(flagRapidRefresh) {
			return rapidRefreshInterval
		}
		return defaultRefreshInterval
	}())

	var lightValidators enginetypes.LightValidators

	votingInfoChan := make(chan interface{}) // accept both voting info and error

	for range refreshTicker.C {
		if len(lightValidators) < 1 {
			lightValidators, err = rpcClient.LightValidators()
			if err != nil {
				utils.PrintlnStdErr("ERR: failed to fetch light validators")
				utils.PrintlnStdErr(err)
				continue
			}
		}

		var nextBlockVotingInfo *enginetypes.NextBlockVotingInformation

		nextBlockVotingInfo, err = consensusService.GetNextBlockVotingInformation(lightValidators)
		if err != nil {
			votingInfoChan <- errors.Wrap(err, "failed to get next block voting information")
			continue
		}

		votingInfoChan <- nextBlockVotingInfo
	}
}

// readPvTopArg reads the argument at the given index, and returns an error if it is missing but required.
// If the argument is a number, it is assumed to be a port and the default host is localhost will be used.
func readPvTopArg(args []string, index int, optional bool) (arg string, err error) {
	if len(args) > index {
		arg = strings.TrimSpace(args[index])
	}

	if arg == "" && !optional {
		err = fmt.Errorf("missing required argument")
	} else if regexp.MustCompile("^\\d+$").MatchString(arg) {
		// automatically correct if only port
		arg = fmt.Sprintf("http://localhost:%s", arg)
	} else if regexp.MustCompile("^:\\d+$").MatchString(arg) {
		// automatically correct if only :port
		arg = fmt.Sprintf("http://localhost%s", arg)
	}

	return
}

func init() {
	pvtopCmd.Flags().Bool(flagHttp, false, "use http call for rpc client instead of default is websocket")
	pvtopCmd.Flags().BoolP(flagRapidRefresh, "r", false, fmt.Sprintf("refresh rate quicker, default is %v will be changed to %v", defaultRefreshInterval, rapidRefreshInterval))

	rootCmd.AddCommand(pvtopCmd)
}
