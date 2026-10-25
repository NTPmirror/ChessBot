package bot

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"chessbot/Code/board"
	"chessbot/Code/config"
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
		history.StartHistoryViewerServer(cfg.WebPort, cfg.FiguresDir, historyDir)
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

	activeYRatio := cfg.YRatio
	isPeopleMode := strings.EqualFold(cfg.GameType, "People")
	isOscarMode := strings.EqualFold(cfg.GameType, "Oscar")
	if isPeopleMode {
		if cfg.YRatioPeople > 0 {
			activeYRatio = cfg.YRatioPeople
		} else {
			activeYRatio = 0.336
		}
	} else {
		if cfg.YRatioOscar > 0 {
			activeYRatio = cfg.YRatioOscar
		} else {
			activeYRatio = 0.355
		}
	}

	adbPath := recognizer.FindADBPath()
	log.Printf("ADB binary ready: %s", adbPath)

	templates, err := recognizer.LoadTemplates(cfg.FiguresDir)
	if err != nil {
		log.Fatalf("Template load error: %v", err)
	}

	stabilizer := board.NewBoardStabilizer(cfg.StabilityThreshold)

	var hub *web.WSHub
	if enableWeb {
		hub = web.NewHub()
		web.StartWebServer(hub, cfg.WebPort, cfg.FiguresDir)
	}

	recorder := history.NewRecorder(enableHistory)
	if enableHistory {
		log.Printf("[History] Match recording enabled (saving to DATA/History/<date>/)")
	}

	log.Printf("ChessBot online (v%s by %s | Mode: %s | AutoPlay: %v | AutoXP: %v | AutoStart: %v)...",
		config.Version, config.Author, cfg.GameType, cfg.AutoPlay, cfg.AutoClaimXP, cfg.AutoStartGame)

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

		if isPeopleMode && recognizer.IsMatchmakingScreen(frame) {
			terminal.RenderTerminal([8][8]string{}, 0, "-", "-", "Searching for opponent...", cfg.AutoPlay, false, cfg.AutoClaimXP, cfg.AutoStartGame, cfg.GameType)
			if enableWeb && hub != nil {
				hub.Broadcast([8][8]string{}, 0, "-", "-", "Searching for opponent...", [2]int{-1, -1}, [2]int{-1, -1}, false, false, false)
			}
			resetMatchState()
			time.Sleep(700 * time.Millisecond)
			continue
		}

		rawBoard := recognizer.Process(frame, templates, activeYRatio, cfg.MaxErrorThreshold)
		pieceCount := board.CountBoardPieces(rawBoard)

		if pieceCount < 4 && recognizer.IsLobbyScreen(frame) {
			if cfg.AutoStartGame {
				var targetX, targetY int
				if isPeopleMode {
					targetX = int(float64(w) * 0.73)
					targetY = int(float64(h) * 0.31)
					terminal.RenderTerminal([8][8]string{}, 0, "-", "-", "Lobby: Joining online game...", cfg.AutoPlay, false, cfg.AutoClaimXP, cfg.AutoStartGame, cfg.GameType)
				} else {
					targetX = int(float64(w) * 0.27)
					targetY = int(float64(h) * 0.31)
					terminal.RenderTerminal([8][8]string{}, 0, "-", "-", "Lobby: Starting game with Oscar...", cfg.AutoPlay, false, cfg.AutoClaimXP, cfg.AutoStartGame, cfg.GameType)
				}
				_ = recognizer.Tap(adbPath, targetX, targetY)
				time.Sleep(1500 * time.Millisecond)
				resetMatchState()
				xpClaimed = false
				continue
			} else {
				terminal.RenderTerminal([8][8]string{}, 0, "-", "-", "Lobby: AutoStart is off", cfg.AutoPlay, false, cfg.AutoClaimXP, cfg.AutoStartGame, cfg.GameType)
				time.Sleep(500 * time.Millisecond)
				continue
			}
		}

		btnType, btnX, btnY := recognizer.DetectActionButton(frame)

		if pieceCount < 4 {
			if btnType == "continue" {
				if cfg.AutoStartGame || xpClaimed {
					terminal.RenderTerminal([8][8]string{}, 0, "-", "-", "Clicking Play / Continue...", cfg.AutoPlay, false, cfg.AutoClaimXP, cfg.AutoStartGame, cfg.GameType)
					_ = recognizer.Tap(adbPath, btnX, btnY)
					time.Sleep(1500 * time.Millisecond)
					resetMatchState()
					xpClaimed = false
					continue
				}
			} else if btnType == "claim_xp" {
				if cfg.AutoClaimXP {
					terminal.RenderTerminal([8][8]string{}, 0, "-", "-", "Claiming XP reward...", cfg.AutoPlay, false, cfg.AutoClaimXP, cfg.AutoStartGame, cfg.GameType)
					_ = recognizer.Tap(adbPath, btnX, btnY)
					time.Sleep(1800 * time.Millisecond)
					resetMatchState()
					xpClaimed = true
					continue
				}
			}

			terminal.RenderTerminal([8][8]string{}, 0, "-", "-", "Waiting for board to load...", cfg.AutoPlay, false, cfg.AutoClaimXP, cfg.AutoStartGame, cfg.GameType)
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

		if isOscarMode && state == 1 && currentMove != myTeam && myTeam != "" {
			if lastOpponentWaitStart.IsZero() {
				lastOpponentWaitStart = time.Now()
			} else if time.Since(lastOpponentWaitStart) > 2500*time.Millisecond {
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
			if cfg.AutoClaimXP {
				_ = recognizer.Tap(adbPath, btnX, btnY)
				time.Sleep(1500 * time.Millisecond)
			}

		} else if btnType == "claim_xp" {
			state = 0
			bestMoveStr = "Game complete. Claiming XP..."
			bestFrom = [2]int{-1, -1}
			bestTo = [2]int{-1, -1}

			if cfg.AutoClaimXP {
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
				qX, qY := recognizer.GetPromotionQueenTap(w, h, activeYRatio, promR, promC)
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

						_ = recognizer.ExecuteMove(adbPath, w, h, activeYRatio, m)
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

		terminal.RenderTerminal(stableBoard, state, currentMove, myTeam, bestMoveStr, cfg.AutoPlay, isGameOver, cfg.AutoClaimXP, cfg.AutoStartGame, cfg.GameType)

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
			recorder.RecordFrame(stableBoard, state, currentMove, myTeam, bestMoveStr, bestFrom, bestTo, isGameOver, bestMoveStr, cfg.GameType)
		}

		elapsed := time.Since(start)
		if elapsed < sleepDuration {
			time.Sleep(sleepDuration - elapsed)
		}
	}
}
