package cmd

//goland:noinspection SpellCheckingInspection
import (
	"bufio"
	"fmt"
	"github.com/bcdevtools/consvp/constants"
	conss "github.com/bcdevtools/consvp/engine/consensus_service"
	dconsi "github.com/bcdevtools/consvp/engine/consensus_service/default_conss_impl"
	pvss "github.com/bcdevtools/consvp/engine/prevote_streaming_service"
	mpvssi "github.com/bcdevtools/consvp/engine/prevote_streaming_service/mock_local_prevote_ss_impl"
	pvssi "github.com/bcdevtools/consvp/engine/prevote_streaming_service/prevote_ss_impl"
	"github.com/bcdevtools/consvp/engine/rpc_client"
	drpci "github.com/bcdevtools/consvp/engine/rpc_client/default_rpc_impl"
	enginetypes "github.com/bcdevtools/consvp/engine/types"
	"github.com/bcdevtools/consvp/utils"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"math"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	flagHttp                = "http"
	flagRapidRefresh        = "rapid-refresh"
	flagStreaming           = "streaming"
	flagResumeStreaming     = "resume-streaming"
	flagMockStreamingServer = "mock-streaming-server"
)

const defaultRefreshInterval = 3 * time.Second
const rapidRefreshInterval = 1 * time.Second

func pvtopHandler(cmd *cobra.Command, args []string) {
	if len(args) > 2 {
		utils.PrintlnStdErr("ERR: Invalid number of arguments")
		os.Exit(1)
	}

	fmt.Println(constants.APP_INTRO)
	fmt.Println()

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

	useHttp := cmd.Flags().Changed(flagHttp)
	streamingMode := cmd.Flags().Changed(flagStreaming)
	resumeStreaming := cmd.Flags().Changed(flagResumeStreaming)
	if resumeStreaming {
		streamingMode = true
	}
	mockStreamingServer := cmd.Flags().Changed(flagMockStreamingServer)
	if mockStreamingServer && !streamingMode {
		panic(fmt.Errorf("cannot mock streaming server if not in streaming mode, requires --%s or --%s", flagStreaming, flagResumeStreaming))
	}

	var rpcClient rpc_client.RpcClient
	var consensusService conss.ConsensusService
	var preVoteStreamingService pvss.PreVoteStreamingService
	exitCallback1 := func() {
		if rpcClient != nil {
			_ = rpcClient.Shutdown()
		}
		if consensusService != nil {
			_ = consensusService.Shutdown()
		}
	}

	defer exitCallback1()

	rpcClient = drpci.NewDefaultRpcClient(consumerUrl, providerUrl, !useHttp)
	consensusService = dconsi.NewDefaultConsensusServiceClientImpl(rpcClient)

	mod5 := rand.Uint32() % 5
	if mod5 == 0 {
		fmt.Println("Tips: press 'Q' or 'Ctrl+C' to exit")
	} else if mod5 == 1 {
		fmt.Println("Tips: press 'K' / '↑' to scroll up and 'J' / '↓' to scroll down")
	}

	var chainId, consensusVersion, moniker string = rpcClient.NodeInfo()
	var lightValidators enginetypes.LightValidators
	var preVoteStreamingShareViewUrl string

	fmt.Println("Please wait, getting validators information...")
	lightValidators, _ = rpcClient.LightValidators()

	if streamingMode { // light validators is required to start a streaming session
		for len(lightValidators) < 1 {
			lightValidators, err = rpcClient.LightValidators()
			if err != nil {
				utils.PrintlnStdErr("ERR: failed to fetch light validators, waiting to retry...")
				time.Sleep(1 * time.Second)
			}
		}

		fmt.Println("Initializing pre-vote streaming service...")
		if mockStreamingServer {
			preVoteStreamingService = mpvssi.NewMockLocalPreVoteStreamingService(chainId, 2*time.Minute)
		} else {
			preVoteStreamingService = pvssi.NewPreVoteStreamingService(chainId)
		}

		if resumeStreaming {
			reader := bufio.NewReader(os.Stdin)

			sessionIdStr := readUntilValid(reader, "Enter session ID:", func(input string) error {
				if len(input) < 1 {
					return fmt.Errorf("must not be empty")
				}
				return enginetypes.PreVoteStreamingSessionId(input).ValidateBasic()
			}, "bad session ID, please check")
			sessionKeyStr := readUntilValid(reader, "Enter session key:", func(input string) error {
				if len(input) < 1 {
					return fmt.Errorf("must not be empty")
				}
				return enginetypes.PreVoteStreamingSessionKey(input).ValidateBasic()
			}, "bad session ID, please check")

			err = preVoteStreamingService.ResumeSession(enginetypes.PreVoteStreamingSessionId(sessionIdStr), enginetypes.PreVoteStreamingSessionKey(sessionKeyStr))
			if err != nil {
				utils.PrintlnStdErr("ERR: failed to resume streaming session id", sessionIdStr)
				utils.PrintlnStdErr(err)
				os.Exit(1)
			}
		} else {
			fmt.Println("Registering streaming session...")
			var errOpenSession error
			for {
				preVoteStreamingShareViewUrl, errOpenSession = preVoteStreamingService.OpenSession(lightValidators)
				if errOpenSession == nil {
					break
				}

				utils.PrintlnStdErr("ERR: failed to open streaming session, waiting to retry...")
				time.Sleep(1 * time.Second)
			}

			fmt.Println("Streaming session registered successfully")
			fmt.Println("use the following session ID and key to resume streaming the session if needed:")
			sessionId, sessionKey := preVoteStreamingService.ExposeSessionIdAndKey()
			fmt.Println("Session ID:", sessionId)
			fmt.Println("Session Key:", sessionKey)

			fmt.Println("*** Share the following URL to others to join:")
			fmt.Println(preVoteStreamingShareViewUrl)
			if !mockStreamingServer {
				time.Sleep(20 * time.Second)
			}
		}
	}

	renderVotingInfoChan := make(chan interface{}) // accept both voting info and error
	var broadcastingPreVoteInfoChan chan *enginetypes.NextBlockVotingInformation
	var broadcastingStatusChan chan string
	if streamingMode {
		broadcastingPreVoteInfoChan = make(chan *enginetypes.NextBlockVotingInformation)
		broadcastingStatusChan = make(chan string)
	}

	exitCallback2 := func() {
		exitCallback1()
		close(renderVotingInfoChan)
		if broadcastingPreVoteInfoChan != nil {
			close(broadcastingPreVoteInfoChan)
		}
		if broadcastingStatusChan != nil {
			close(broadcastingStatusChan)
		}
	}
	defer exitCallback2()

	go drawScreen(chainId, consensusVersion, moniker, renderVotingInfoChan, broadcastingStatusChan, exitCallback2)
	if streamingMode {
		go broadcastPreVoteInfo(preVoteStreamingService, broadcastingPreVoteInfoChan, broadcastingStatusChan)
	}

	refreshTicker := time.NewTicker(func() time.Duration {
		if cmd.Flags().Changed(flagRapidRefresh) {
			return rapidRefreshInterval
		}
		return defaultRefreshInterval
	}())

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
			renderVotingInfoChan <- errors.Wrap(err, "failed to get next block voting information")
			continue
		}

		renderVotingInfoChan <- nextBlockVotingInfo
		if streamingMode {
			if !preVoteStreamingService.IsStopped() { // prevent memory stacking due to no consumer
				broadcastingPreVoteInfoChan <- nextBlockVotingInfo
			}
		}
	}
}

const terminalColumnsCount = 3

// drawScreen render pre-vote information into screen.
func drawScreen(chainId, consensusVersion, moniker string, votingInfoChan <-chan interface{}, broadcastingStatusChan <-chan string, exitCallback func()) {
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

	pBroadcastStatus := widgets.NewParagraph()
	pBroadcastStatus.Title = " Broadcast Status "

	lists := make([]*widgets.List, terminalColumnsCount)
	for i := range lists {
		lists[i] = widgets.NewList()
		lists[i].Border = false
	}
	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	var gridHeader ui.GridItem
	if broadcastingStatusChan != nil {
		gridHeader = ui.NewRow(0.1,
			ui.NewCol(1.0/4, pSummary),
			ui.NewCol(1.0/4, pBroadcastStatus),
			ui.NewCol(1.0/4, preVotePctGauge),
			ui.NewCol(1.0/4, preCommitVotePctGauge),
		)
	} else {
		gridHeader = ui.NewRow(0.1,
			ui.NewCol(1.0/3, pSummary),
			ui.NewCol(1.0/3, preVotePctGauge),
			ui.NewCol(1.0/3, preCommitVotePctGauge),
		)
	}

	grid.Set(
		gridHeader,
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

					valMoniker := string(utils.TruncateStringUntilBufferLessThanXBytesOrFillWithSpaceSuffix(voter.Validator.Moniker, 15))
					valMoniker = strings.TrimSpace(valMoniker)

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
		case broadcastStatus := <-broadcastingStatusChan:
			refresh = true

			pBroadcastStatus.Text = broadcastStatus

			break
		}
	}
}

func broadcastPreVoteInfo(pvs pvss.PreVoteStreamingService, votingInfoChan <-chan *enginetypes.NextBlockVotingInformation, broadcastingStatusChan chan<- string) {
	for {
		select {
		case vi := <-votingInfoChan:
			if vi == nil {
				panic("voting info is nil")
			}

			err, shouldStop := pvs.BroadcastPreVote(vi)
			if shouldStop {
				if err == nil {
					broadcastingStatusChan <- "🔴 Broadcasting stopped"
				} else {
					broadcastingStatusChan <- fmt.Sprintf("🔴 Broadcasting stopped: %s", err)
				}
				pvs.Stop()
				return
			}

			if err != nil {
				broadcastingStatusChan <- fmt.Sprintf("❗️Last broadcasting failed with reason: %s", err)
				continue
			}

			broadcastingStatusChan <- "🟢 Pre-Vote streaming in progress"

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

func readUntilValid(reader *bufio.Reader, question string, validateFn func(t string) error, malformedErrMsg string) string {
	for {
		fmt.Println(question)
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)

		err := validateFn(line)
		if err != nil {
			utils.PrintfStdErr("ERR: %s\n", malformedErrMsg)
			utils.PrintlnStdErr(err)
			fmt.Println("----")
			continue
		}

		return line
	}
}
