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
	"math"
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
	defer close(votingInfoChan)

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

const terminalColumnsCount = 3

// drawScreen render pre-vote information into screen.
func drawScreen(chainId, consensusVersion, moniker string, votingInfoChan chan interface{}, exitCallback func()) {
	if err := ui.Init(); err != nil {
		//goland:noinspection SpellCheckingInspection
		utils.PrintfStdErr("failed to initialize termui: %v\n", err)
	}

	pSummary := widgets.NewParagraph()
	summaryTitle := fmt.Sprintf(" %s, tm v%s", chainId, consensusVersion)
	if len(moniker) > 0 {
		summaryTitle += fmt.Sprintf(", %s", moniker)
	}
	summaryTitle += " "
	pSummary.Title = summaryTitle

	preVotePctGauge := widgets.NewGauge()
	preCommitVotePctGauge := widgets.NewGauge()

	lists := make([]*widgets.List, terminalColumnsCount)
	for i := range lists {
		lists[i] = widgets.NewList()
		lists[i].Border = false
	}
	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	grid.Set(
		ui.NewRow(0.1,
			ui.NewCol(1.0/terminalColumnsCount, pSummary),
			ui.NewCol(1.0/terminalColumnsCount, preVotePctGauge),
			ui.NewCol(1.0/terminalColumnsCount, preCommitVotePctGauge),
		),
		ui.NewRow(0.9,
			ui.NewCol(.96/terminalColumnsCount, lists[0]),
			ui.NewCol(.96/terminalColumnsCount, lists[1]),
			ui.NewCol(1.08/terminalColumnsCount, lists[2]),
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
				pSummary.Text = err.Error()
				continue
			}

			votingInfo := votingInfoAny.(*enginetypes.NextBlockVotingInformation)

			duration := time.Now().UTC().Sub(votingInfo.StartTimeUTC)
			if duration < 0 {
				duration = 0
			}

			pSummary.Text = fmt.Sprintf(
				"%v\nheight/round/step: %s\nv: %.0f%% c: %.0f%%",
				duration,
				votingInfo.HeightRoundStep,
				votingInfo.PreVotePercent*100,
				votingInfo.PreCommitPercent*100,
			)

			batches, rowsCount := splitVotesIntoColumnsForRendering(votingInfo.SortedValidatorVoteStates)
			totalVoteCount := len(votingInfo.SortedValidatorVoteStates)
			preVotedCount := totalVoteCount
			preCommitVotedCount := totalVoteCount
			for i := 0; i < terminalColumnsCount; i++ {
				lists[i].Rows = make([]string, rowsCount+1)

				lists[i].Rows[0] = fmt.Sprintf("%-3s %-3s %-4s %-3s %-6s %-15s ", "PV", "PC", "Hash", "Ord", "VPwr", "Moniker")

				for j, voter := range batches[i] {
					rowIndex := j + 1

					var preVote, preCommitVote string

					if voter.VotedZeroes {
						preVote = "🤷"
					} else if voter.PreVoted {
						preVote = "✅"
					} else {
						preVote = "❌"
						preVotedCount--
					}
					if voter.PreCommitVoted {

						preCommitVote = "✅"
					} else {
						preCommitVote = "❌"
						preCommitVotedCount--
					}

					valMoniker := voter.Validator.Moniker
					if len([]byte(valMoniker)) > 15 {
						valMoniker = string(append([]byte(valMoniker[:9]), []byte("...")...))
					}
					if len([]byte(valMoniker)) > len(valMoniker) {
						valMoniker = valMoniker[:len([]byte(valMoniker))-len(valMoniker)]
					}

					lists[i].Rows[rowIndex] = fmt.Sprintf(
						"%-2s %-2s %s %-3d %s%% %-15s ",
						preVote,
						preCommitVote,
						func() string {
							if len(voter.VotingBlockHash) >= 4 {
								return voter.VotingBlockHash[:4]
							} else {
								return "----"
							}
						}(),
						voter.Validator.Index+1,
						func() string {
							str := fmt.Sprintf("%-.2f", voter.Validator.VotingPowerDisplayPercent)
							if strings.Index(str, ".") == 1 { // VP percent < 10
								str = "0" + str
							}
							return str
						}(),
						valMoniker,
					)
				}
			}

			preVotePctGauge.Title = fmt.Sprintf(" Pre-vote: %d/%d ", preVotedCount, totalVoteCount)
			preVotePctGauge.Percent = int(votingInfo.PreVotePercent * 100)
			preCommitVotePctGauge.Title = fmt.Sprintf(" Pre-commit: %d/%d ", preCommitVotedCount, totalVoteCount)
			preCommitVotePctGauge.Percent = int(votingInfo.PreCommitPercent * 100)

			break
		}
	}
}

func splitVotesIntoColumnsForRendering(votes []enginetypes.ValidatorVoteState) (batches [][]enginetypes.ValidatorVoteState, rowsCount int) {
	rowsCount = int(math.Ceil(float64(len(votes)) / float64(terminalColumnsCount)))

	batches = make([][]enginetypes.ValidatorVoteState, terminalColumnsCount)

	colIndex := 0

	for i := 0; i < len(votes); i++ {
		batches[colIndex] = append(batches[colIndex], votes[i])

		colIndex++
		if colIndex >= terminalColumnsCount {
			colIndex = 0
		}
	}

	return
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
