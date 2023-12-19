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
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"log"
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
	exitCallback := func() {
		if rpcClient != nil {
			_ = rpcClient.Shutdown()
		}
		if consensusService != nil {
			_ = consensusService.Shutdown()
		}
	}

	defer exitCallback()

	rpcClient = drpci.NewDefaultRpcClient(consumerUrl, providerUrl, !useHttp)
	consensusService = dconsi.NewDefaultConsensusServiceClientImpl(rpcClient)

	refreshTicker := time.NewTicker(func() time.Duration {
		if cmd.Flags().Changed(flagRapidRefresh) {
			return rapidRefreshInterval
		}
		return defaultRefreshInterval
	}())

	var chainId, consensusVersion, moniker string = rpcClient.NodeInfo()
	var lightValidators enginetypes.LightValidators

	votingInfoChan := make(chan interface{}) // accept both voting info and error

	go drawScreen(chainId, consensusVersion, moniker, votingInfoChan, exitCallback)

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

// drawScreen render pre-vote information into screen.
func drawScreen(chainId, consensusVersion, moniker string, votingInfoChan chan interface{}, exitCallback func()) {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	preVotePctGauge := widgets.NewGauge()
	preCommitVotePctGauge := widgets.NewGauge()

	p := widgets.NewParagraph()
	p.Title = fmt.Sprintf("%s, consensus v%s", chainId, consensusVersion)

	lists := make([]*widgets.List, 3)
	for i := range lists {
		lists[i] = widgets.NewList()
		lists[i].Border = false
	}
	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	grid.Set(
		ui.NewRow(0.1,
			ui.NewCol(1.0/3, p),
			ui.NewCol(1.0/3, preVotePctGauge),
			ui.NewCol(1.0/3, preCommitVotePctGauge),
		),
		ui.NewRow(0.9,
			ui.NewCol(.9/3, lists[0]),
			ui.NewCol(.9/3, lists[1]),
			ui.NewCol(1.2/3, lists[2]),
		),
	)
	ui.Render(grid)

	refresh := false
	tick := time.NewTicker(100 * time.Millisecond)
	uiEvents := ui.PollEvents()

	for {
		select {
		case <-tick.C:
			if !refresh {
				continue
			}
			refresh = false
			ui.Render(grid)

			break
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				ui.Clear()
				ui.Close()

				exitCallback()

				os.Exit(0)
			case "j", "<Down>":
				for _, list := range lists {
					if len(list.Rows) > 0 {
						list.ScrollBottom()
					}
				}

				break
			case "k", "<Up>":
				for _, list := range lists {
					if len(list.Rows) > 0 {
						list.ScrollTop()
					}
				}

				break
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				grid.SetRect(0, 0, payload.Width, payload.Height)
				ui.Clear()
				ui.Render(grid)

				break
			}

			break
		case votingInfoAny := <-votingInfoChan:
			refresh = true

			if err, ok := votingInfoAny.(error); ok {
				p.Text = err.Error()
				continue
			}

			votingInfo := votingInfoAny.(*enginetypes.NextBlockVotingInformation)

			duration := time.Now().UTC().Sub(votingInfo.StartTimeUTC)
			if duration < 0 {
				duration = 0
			}

			p.Text = fmt.Sprintf(
				"height/round/step: %s - v: %.0f%% c: %.0f%% (%v)\n\nProposer:\n(rank/%%/moniker) %s",
				votingInfo.HeightRoundStep,
				votingInfo.PreVotePercent*100,
				votingInfo.PreCommitPercent*100,
				duration,
				moniker,
			)

			split, columns, rows := splitVotes(votingInfo.SortedValidatorVoteStates)
			for i := 0; i < columns; i++ {
				lists[i].Rows = make([]string, rows)
				for j, voter := range split[i] {
					var preVote, preCommitVote string

					if voter.VotedZeroes {
						preVote = "🤷"
					} else if voter.PreVoted {
						preVote = "✅"
					} else {
						preVote = "❌"
					}
					if voter.PreCommitVoted {
						preCommitVote = "✅"
					} else {
						preCommitVote = "❌"
					}

					valMoniker := voter.Validator.Moniker
					if len([]byte(valMoniker)) > 20 {
						valMoniker = string(append([]byte(valMoniker[:14]), []byte("...")...))
					}
					if len([]byte(valMoniker)) > len(valMoniker) {
						valMoniker = valMoniker[:len([]byte(valMoniker))-len(valMoniker)]
					}

					validatorDescription := fmt.Sprintf("%-3d %-.2f%%   %-20s ", voter.Validator.Index+1, voter.Validator.VotingPowerDisplayPercent, valMoniker)

					lists[i].Rows[j] = fmt.Sprintf("%-3s %-3s %s", preVote, preCommitVote, validatorDescription)
				}
			}

			preVotePctGauge.Percent = int(votingInfo.PreVotePercent * 100)
			preCommitVotePctGauge.Percent = int(votingInfo.PreCommitPercent * 100)

			break
		}
	}
}

func splitVotes(votes []enginetypes.ValidatorVoteState) ([][]enginetypes.ValidatorVoteState, int, int) {
	// TODO review logic

	batches := make([][]enginetypes.ValidatorVoteState, 0)
	var columnsCount int
	var rows = 50

	switch {
	case len(votes) < 50:
		columnsCount = 1
		batches = append(batches, votes)
	case len(votes) < 100:
		columnsCount = 2
		batches = append(batches, votes[:50])
		batches = append(batches, votes[50:])
	default:
		columnsCount = 3
		rows = (len(votes) + columnsCount - 1) / 3
		batches = append(batches, votes[:rows])
		batches = append(batches, votes[rows:rows*2])
		batches = append(batches, votes[rows*2:])
	}
	return batches, columnsCount, rows
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
