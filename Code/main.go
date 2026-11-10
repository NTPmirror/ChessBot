package bot

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"chessbot/Code/board"
	"chessbot/Code/config"
	"chessbot/Code/duolingo"
	"chessbot/Code/engine"
	"chessbot/Code/history"
	"chessbot/Code/recognizer"
	"chessbot/Code/terminal"
	"chessbot/Code/web"
)

func Run() {
	helpFlag := flag.Bool("help", false, "Show help information")
	hFlag := flag.Bool("h", false, "Show help information")
	webFlag := flag.Bool("WEB", false, "Start live web board")
	historyFlag := flag.Bool("history", false, "Record match history to JSON")
	viewHistoryFlag := flag.Bool("view-history", false, "Start web server for viewing match history without chess engine")
	cfgPath := flag.String("config", "config.jsonc", "Path to config file")

	flag.Usage = func() {
		fmt.Printf("ChessBot (v%s) by %s\n\n", config.Version, config.Author)
		fmt.Println("Usage:")
		fmt.Println("  ChessBot [options]")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("  -config <path>   Path to configuration file (default: config.jsonc)")
		fmt.Println("  -WEB             Enable live web dashboard")
		fmt.Println("  -history         Enable recording match history to JSON")
		fmt.Println("  -view-history    Launch web server for viewing match history without chess engine")
		fmt.Println("  -h, --help       Show this help message")
		fmt.Println()
	}

	flag.Parse()

	if *helpFlag || *hFlag {
		flag.Usage()
		return
	}

	for _, arg := range os.Args[1:] {
		if arg == "help" || arg == "h" || arg == "-h" || arg == "--help" {
			flag.Usage()
			return
		}
	}

	enableViewHistory := *viewHistoryFlag
	for _, arg := range os.Args[1:] {
		if strings.EqualFold(arg, "-view-history") || strings.EqualFold(arg, "--view-history") {
			enableViewHistory = true
			break
		}
	}

	cfg := config.LoadConfig(*cfgPath)

	if enableViewHistory {
		historyDir := filepath.Join("DATA", "History")
		_ = os.MkdirAll(historyDir, 0755)
		history.StartHistoryViewerServer(cfg.WebPort, cfg.Duolingo.FiguresDir, historyDir)
		return
	}

	enableWeb := cfg.EnableWeb || *webFlag
	for _, arg := range os.Args[1:] {
		if strings.EqualFold(arg, "-WEB") || strings.EqualFold(arg, "--WEB") {
			enableWeb = true
			break
		}
	}

	enableHistory := *historyFlag
	for _, arg := range os.Args[1:] {
		if strings.EqualFold(arg, "-history") || strings.EqualFold(arg, "--history") {
			enableHistory = true
			break
		}
	}

	activeYRatio := cfg.Duolingo.YRatio
	isPeopleMode := strings.EqualFold(cfg.Duolingo.GameType, "People")
	isOscarMode := strings.EqualFold(cfg.Duolingo.GameType, "Oscar")
	if isPeopleMode {
		if cfg.Duolingo.YRatioPeople > 0 {
			activeYRatio = cfg.Duolingo.YRatioPeople
		} else {
			activeYRatio = 0.336
		}
	} else {
		if cfg.Duolingo.YRatioOscar > 0 {
			activeYRatio = cfg.Duolingo.YRatioOscar
		} else {
			activeYRatio = 0.355
		}
	}

	adbPath := recognizer.FindADBPath()
	log.Printf("ADB binary ready: %s", adbPath)

	templates, err := recognizer.LoadTemplates(cfg.Duolingo.FiguresDir)
	if err != nil {
		log.Fatalf("Template load error: %v", err)
	}

	stabilizer := board.NewBoardStabilizer(cfg.StabilityThreshold)

	var hub *web.WSHub
	if enableWeb {
		hub = web.NewHub()
		web.StartWebServer(hub, cfg.WebPort, cfg.Duolingo.FiguresDir)
	}

	recorder := history.NewRecorder(enableHistory)
	if enableHistory {
		log.Printf("[History] Match recording enabled (saving to DATA/History/<date>/)")
	}

	log.Printf("ChessBot online (v%s by %s | Mode: %s | AutoPlay: %v | AutoXP: %v | AutoStart: %v)...",
		config.Version, config.Author, cfg.Duolingo.GameType, cfg.AutoPlay, cfg.Duolingo.AutoClaimXP, cfg.Duolingo.AutoStartGame)

	sleepDuration := time.Duration(cfg.IntervalMs) * time.Millisecond

	var lastStableBoard [8][8]string
	var lastCalculatedBoard [8][8]string
	var lastPlayedMove engine.Move
	var lastBoardBeforeMove [8][8]string

	blunderBlacklist := make(map[string]map[string]bool)

	isFirstStable := true
	currentMove := "White"
	myTeam := ""
	bestMoveStr := "-"
	bestFrom := [2]int{-1, -1}
	bestTo := [2]int{-1, -1}
	xpClaimed := false
	var lastOpponentWaitStart time.Time

	resetMatchState := func() {
		recorder.FinishMatch("")
		lastStableBoard = [8][8]string{}
		lastCalculatedBoard = [8][8]string{}
		lastPlayedMove = engine.Move{}
		blunderBlacklist = make(map[string]map[string]bool)
		stabilizer.Reset()
		isFirstStable = true
		myTeam = ""
		currentMove = "White"
		lastOpponentWaitStart = time.Time{}
	}

	for {
		start := time.Now()

		frame, err := recognizer.CaptureScreen(adbPath)
		if err != nil {
			log.Printf("Screen capture failed: %v", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		w := frame.Bounds().Dx()
		h := frame.Bounds().Dy()

		if isPeopleMode && duolingo.IsMatchmakingScreen(frame) {
			terminal.RenderTerminal([8][8]string{}, 0, "-", "-", "Searching for opponent...", cfg.AutoPlay, false, cfg.Duolingo.AutoClaimXP, cfg.Duolingo.AutoStartGame, cfg.Duolingo.GameType)
			if enableWeb && hub != nil {
				hub.Broadcast([8][8]string{}, 0, "-", "-", "Searching for opponent...", [2]int{-1, -1}, [2]int{-1, -1}, false, false, false)
			}
			resetMatchState()
			time.Sleep(700 * time.Millisecond)
			continue
		}

		rawBoard := recognizer.Process(frame, templates, activeYRatio, cfg.MaxErrorThreshold)
		pieceCount := board.CountBoardPieces(rawBoard)

		if duolingo.IsLobbyScreen(frame) {
			if cfg.Duolingo.AutoStartGame {
				if duolingo.IsNewLobbyScreen(frame) {
					// Card 0: Oscar, Card 1: Lily, Card 2: People (Play)
					targetCard := 2
					if isOscarMode {
						targetCard = 0
					}

					currentSelected := duolingo.GetNewLobbyCardSelection(frame)
					if currentSelected != targetCard {
						var cardXFrac float64
						switch targetCard {
						case 0:
							cardXFrac = 0.20
						case 1:
							cardXFrac = 0.50
						case 2:
							cardXFrac = 0.80
						}
						cardX := int(float64(w) * cardXFrac)
						cardY := int(float64(h) * 0.25)
						terminal.RenderTerminal([8][8]string{}, 0, "-", "-", "Lobby: Selecting game mode card...", cfg.AutoPlay, false, cfg.Duolingo.AutoClaimXP, cfg.Duolingo.AutoStartGame, cfg.Duolingo.GameType)
						_ = recognizer.Tap(adbPath, cardX, cardY)
						time.Sleep(600 * time.Millisecond)
					}

					playX := int(float64(w) * 0.50)
					playY := int(float64(h) * 0.391)
					if isPeopleMode {
						terminal.RenderTerminal([8][8]string{}, 0, "-", "-", "Lobby: Clicking Play (Online)...", cfg.AutoPlay, false, cfg.Duolingo.AutoClaimXP, cfg.Duolingo.AutoStartGame, cfg.Duolingo.GameType)
					} else {
						terminal.RenderTerminal([8][8]string{}, 0, "-", "-", "Lobby: Clicking Play (Oscar)...", cfg.AutoPlay, false, cfg.Duolingo.AutoClaimXP, cfg.Duolingo.AutoStartGame, cfg.Duolingo.GameType)
					}
					_ = recognizer.Tap(adbPath, playX, playY)
				} else {
					var targetX, targetY int
					if isPeopleMode {
						targetX = int(float64(w) * 0.73)
						targetY = int(float64(h) * 0.31)
						terminal.RenderTerminal([8][8]string{}, 0, "-", "-", "Lobby: Joining online game...", cfg.AutoPlay, false, cfg.Duolingo.AutoClaimXP, cfg.Duolingo.AutoStartGame, cfg.Duolingo.GameType)
					} else {
						targetX = int(float64(w) * 0.27)
						targetY = int(float64(h) * 0.31)
						terminal.RenderTerminal([8][8]string{}, 0, "-", "-", "Lobby: Starting game with Oscar...", cfg.AutoPlay, false, cfg.Duolingo.AutoClaimXP, cfg.Duolingo.AutoStartGame, cfg.Duolingo.GameType)
					}
					_ = recognizer.Tap(adbPath, targetX, targetY)
				}

				time.Sleep(1500 * time.Millisecond)
				resetMatchState()
				xpClaimed = false
				continue
			} else {
				terminal.RenderTerminal([8][8]string{}, 0, "-", "-", "Lobby: AutoStart is off", cfg.AutoPlay, false, cfg.Duolingo.AutoClaimXP, cfg.Duolingo.AutoStartGame, cfg.Duolingo.GameType)
				time.Sleep(500 * time.Millisecond)
				continue
			}
		}

		btnType, btnX, btnY := duolingo.DetectActionButton(frame)

		if pieceCount < 4 {
			if cfg.Duolingo.AutoClaimXP && duolingo.IsChestScreen(frame) {
				chestX, chestY := duolingo.GetChestTapCoordinates(w, h)
				terminal.RenderTerminal([8][8]string{}, 0, "-", "-", "Opening reward chest...", cfg.AutoPlay, false, cfg.Duolingo.AutoClaimXP, cfg.Duolingo.AutoStartGame, cfg.Duolingo.GameType)
				_ = recognizer.Tap(adbPath, chestX, chestY)
				time.Sleep(1500 * time.Millisecond)
				continue
			}

			if duolingo.IsSuperAdScreen(frame) {
				terminal.RenderTerminal([8][8]string{}, 0, "-", "-", "Skipping Super Duolingo ad (No thanks)...", cfg.AutoPlay, false, cfg.Duolingo.AutoClaimXP, cfg.Duolingo.AutoStartGame, cfg.Duolingo.GameType)
				_ = recognizer.Tap(adbPath, w/2, int(float64(h)*0.928))
				time.Sleep(1500 * time.Millisecond)
				resetMatchState()
				continue
			}

			if btnType == "continue" {
				if cfg.Duolingo.AutoStartGame || cfg.AutoPlay || cfg.Duolingo.AutoClaimXP || xpClaimed {
					terminal.RenderTerminal([8][8]string{}, 0, "-", "-", "Clicking Play / Continue...", cfg.AutoPlay, false, cfg.Duolingo.AutoClaimXP, cfg.Duolingo.AutoStartGame, cfg.Duolingo.GameType)
					_ = recognizer.Tap(adbPath, btnX, btnY)
					time.Sleep(1500 * time.Millisecond)
					resetMatchState()
					xpClaimed = false
					continue
				}
			} else if btnType == "claim_xp" {
				if cfg.Duolingo.AutoClaimXP {
					terminal.RenderTerminal([8][8]string{}, 0, "-", "-", "Claiming XP reward...", cfg.AutoPlay, false, cfg.Duolingo.AutoClaimXP, cfg.Duolingo.AutoStartGame, cfg.Duolingo.GameType)
					_ = recognizer.Tap(adbPath, btnX, btnY)
					time.Sleep(1800 * time.Millisecond)
					resetMatchState()
					xpClaimed = true
					continue
				}
			}

			terminal.RenderTerminal([8][8]string{}, 0, "-", "-", "Waiting for board to load...", cfg.AutoPlay, false, cfg.Duolingo.AutoClaimXP, cfg.Duolingo.AutoStartGame, cfg.Duolingo.GameType)
			time.Sleep(300 * time.Millisecond)
			continue
		}

		rawBoard = board.EnsureKingPresence(rawBoard, lastStableBoard, myTeam)
		stableBoard := stabilizer.Filter(rawBoard)
		stableBoard = board.EnsureKingPresence(stableBoard, lastStableBoard, myTeam)

		state := 0
		if board.IsBoardEqual(rawBoard, stableBoard) {
			state = 1
		}

		highlightWho, _, _ := recognizer.DetectLastMoveHighlight(frame, activeYRatio, stableBoard)
		clockTurn := duolingo.DetectActiveClockTurn(frame)
		if clockTurn != "" {
			currentMove = clockTurn
		}

		if !board.IsBoardEqual(lastStableBoard, stableBoard) {
			if detectedTeam := board.DetermineMyTeam(stableBoard); detectedTeam != "" {
				myTeam = detectedTeam
			}

			if isFirstStable {
				if myTeam != "" {
					currentMove = board.InferInitialTurn(stableBoard, myTeam, highlightWho)
					isFirstStable = false
				}
			} else {
				whoMoved := board.DetectWhoMoved(lastStableBoard, stableBoard)
				if whoMoved == "" && highlightWho != "" {
					whoMoved = highlightWho
				}

				if whoMoved == "White" {
					currentMove = "Black"
				} else if whoMoved == "Black" {
					currentMove = "White"
				}
			}
			lastStableBoard = stableBoard
		}

		if state == 1 && myTeam != "" && currentMove != myTeam {
			if (myTeam == "Black" && highlightWho == "White") || (myTeam == "White" && highlightWho == "Black") {
				currentMove = myTeam
			}
		}

		if state == 1 && currentMove != myTeam && myTeam != "" {
			waitLimit := 2500 * time.Millisecond
			if isPeopleMode {
				waitLimit = 4000 * time.Millisecond
			}
			if lastOpponentWaitStart.IsZero() {
				lastOpponentWaitStart = time.Now()
			} else if time.Since(lastOpponentWaitStart) > waitLimit {
				currentMove = myTeam
				lastOpponentWaitStart = time.Time{}
			}
		} else {
			lastOpponentWaitStart = time.Time{}
		}

		promR, promC, hasPendingProm := board.FindPendingPromotion(stableBoard, myTeam)
		if !hasPendingProm {
			promR, promC, hasPendingProm = board.FindPendingPromotion(rawBoard, myTeam)
		}

		isGameOver := (btnType == "undo" || btnType == "continue" || btnType == "claim_xp")

		if btnType == "continue" {
			state = 0
			bestFrom = [2]int{-1, -1}
			bestTo = [2]int{-1, -1}
			bestMoveStr = "Checkmate! Victory! Continuing..."
			if cfg.Duolingo.AutoClaimXP || cfg.AutoPlay || cfg.Duolingo.AutoStartGame {
				_ = recognizer.Tap(adbPath, btnX, btnY)
				time.Sleep(1500 * time.Millisecond)
				resetMatchState()
				xpClaimed = false
				continue
			}

		} else if btnType == "claim_xp" {
			state = 0
			bestMoveStr = "Game complete. Claiming XP..."
			bestFrom = [2]int{-1, -1}
			bestTo = [2]int{-1, -1}

			if cfg.Duolingo.AutoClaimXP {
				_ = recognizer.Tap(adbPath, btnX, btnY)
				time.Sleep(1800 * time.Millisecond)
				resetMatchState()
				xpClaimed = true
			}

		} else if btnType == "undo" {
			state = 0
			bestMoveStr = "Stalemate detected. Undoing and searching for mate..."
			bestFrom = [2]int{-1, -1}
			bestTo = [2]int{-1, -1}

			if cfg.AutoPlay {
				if lastPlayedMove.Piece != "" {
					posKey := board.BoardToKey(lastBoardBeforeMove)
					if blunderBlacklist[posKey] == nil {
						blunderBlacklist[posKey] = make(map[string]bool)
					}
					blunderBlacklist[posKey][lastPlayedMove.String()] = true
					log.Printf("[Anti-Stalemate] Blacklisted move: %s", lastPlayedMove.String())
				}

				_ = recognizer.Tap(adbPath, btnX, btnY)
				time.Sleep(800 * time.Millisecond)

				lastCalculatedBoard = [8][8]string{}
				currentMove = myTeam
				lastPlayedMove = engine.Move{}
			}

		} else if hasPendingProm {
			state = 0
			bestMoveStr = "Promoting pawn to Queen..."
			bestFrom = [2]int{promR, promC}
			bestTo = [2]int{promR, promC}

			if cfg.AutoPlay {
				qX, qY := duolingo.GetPromotionQueenTap(w, h, activeYRatio, promR, promC)
				_ = recognizer.Tap(adbPath, qX, qY)
				time.Sleep(600 * time.Millisecond)
			}

		} else if state == 1 && currentMove == myTeam && myTeam != "" {
			if !board.IsBoardEqual(lastCalculatedBoard, stableBoard) {
				posKey := board.BoardToKey(stableBoard)
				m, ok := engine.FindBestMove(stableBoard, myTeam, 4, blunderBlacklist[posKey])
				if ok {
					bestMoveStr = m.String()
					bestFrom = [2]int{m.FromR, m.FromC}
					bestTo = [2]int{m.ToR, m.ToC}

					if cfg.AutoPlay {
						lastBoardBeforeMove = stableBoard
						lastPlayedMove = m

						tapDelay := 180
						if cfg.Humanize {
							minMove := cfg.MoveDelayMinMs
							maxMove := cfg.MoveDelayMaxMs
							if minMove <= 0 { minMove = 300 }
							if maxMove < minMove { maxMove = minMove + 500 }
							thinkingDelay := minMove + rand.Intn(maxMove-minMove+1)
							time.Sleep(time.Duration(thinkingDelay) * time.Millisecond)

							minTap := cfg.TapDelayMinMs
							maxTap := cfg.TapDelayMaxMs
							if minTap <= 0 { minTap = 120 }
							if maxTap < minTap { maxTap = minTap + 100 }
							tapDelay = minTap + rand.Intn(maxTap-minTap+1)
						}

						_ = recognizer.ExecuteMove(adbPath, w, h, activeYRatio, m, tapDelay)
						time.Sleep(450 * time.Millisecond)
					}
				} else {
					bestMoveStr = "Checkmate / Stalemate"
					bestFrom = [2]int{-1, -1}
					bestTo = [2]int{-1, -1}
				}
				lastCalculatedBoard = stableBoard
			}
		} else if state == 0 {
			bestMoveStr = "Pieces in motion..."
			bestFrom = [2]int{-1, -1}
			bestTo = [2]int{-1, -1}
		} else {
			bestMoveStr = "Opponent's turn (" + currentMove + ")..."
			bestFrom = [2]int{-1, -1}
			bestTo = [2]int{-1, -1}
		}

		terminal.RenderTerminal(stableBoard, state, currentMove, myTeam, bestMoveStr, cfg.AutoPlay, isGameOver, cfg.Duolingo.AutoClaimXP, cfg.Duolingo.AutoStartGame, cfg.Duolingo.GameType)

		isMyKingCheck := false
		isOppKingCheck := false
		if myTeam != "" {
			isMyKingCheck = engine.IsKingInCheck(stableBoard, myTeam, myTeam)
			oppTeam := "Black"
			if myTeam == "Black" {
				oppTeam = "White"
			}
			isOppKingCheck = engine.IsKingInCheck(stableBoard, oppTeam, myTeam)
		}

		if enableWeb && hub != nil {
			hub.Broadcast(stableBoard, state, currentMove, myTeam, bestMoveStr, bestFrom, bestTo, isGameOver, isMyKingCheck, isOppKingCheck)
		}

		if enableHistory {
			recorder.RecordFrame(stableBoard, state, currentMove, myTeam, bestMoveStr, bestFrom, bestTo, isGameOver, bestMoveStr, cfg.Duolingo.GameType)
		}

		elapsed := time.Since(start)
		if elapsed < sleepDuration {
			time.Sleep(sleepDuration - elapsed)
		}
	}
}
