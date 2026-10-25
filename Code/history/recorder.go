package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"chessbot/Code/board"
)

type HistoryFrame struct {
	Timestamp   int64        `json:"timestamp"`
	Board       [8][8]string `json:"board"`
	State       int          `json:"state"`
	Move        string       `json:"move"`
	MyTeam      string       `json:"my_team"`
	BestMove    string       `json:"best_move"`
	From        [2]int       `json:"from"`
	To          [2]int       `json:"to"`
	IsGameOver  bool         `json:"is_game_over"`
	StatusText  string       `json:"status_text,omitempty"`
}

type MatchHistory struct {
	ID        string         `json:"id"`
	Title     string         `json:"title"`
	Date      string         `json:"date"`
	StartTime time.Time      `json:"start_time"`
	GameType  string         `json:"game_type"`
	MyTeam    string         `json:"my_team"`
	Result    string         `json:"result"`
	Frames    []HistoryFrame `json:"frames"`
}

type Recorder struct {
	mu           sync.Mutex
	enabled      bool
	currentMatch *MatchHistory
	lastBoard    [8][8]string
	lastState    int
	lastMove     string
	baseDir      string
	currentPath  string
}

func NewRecorder(enabled bool) *Recorder {
	return &Recorder{
		enabled: enabled,
		baseDir: filepath.Join("DATA", "History"),
	}
}

func (r *Recorder) IsEnabled() bool {
	return r.enabled
}

func (r *Recorder) StartMatch(gameType, myTeam string) {
	if !r.enabled {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	dateStr := now.Format("2006-01-02")
	timeStr := now.Format("15-04-05")
	id := fmt.Sprintf("game_%s", timeStr)

	dir := filepath.Join(r.baseDir, dateStr)
	_ = os.MkdirAll(dir, 0755)

	teamDisplay := myTeam
	if teamDisplay == "" {
		teamDisplay = "Bot"
	}

	r.currentPath = filepath.Join(dir, id+".json")
	r.currentMatch = &MatchHistory{
		ID:        id,
		Title:     fmt.Sprintf("%s vs %s (%s)", teamDisplay, gameType, now.Format("15:04:05")),
		Date:      dateStr,
		StartTime: now,
		GameType:  gameType,
		MyTeam:    myTeam,
		Result:    "In Progress",
		Frames:    make([]HistoryFrame, 0, 100),
	}
	r.lastBoard = [8][8]string{}
	r.lastState = -1
	r.lastMove = ""
	r.saveCurrent()
}

func (r *Recorder) RecordFrame(b [8][8]string, state int, move, myTeam, bestMove string, from, to [2]int, isGameOver bool, statusText string, gameType ...string) {
	if !r.enabled {
		return
	}

	pieceCount := board.CountBoardPieces(b)
	if pieceCount < 4 {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.currentMatch == nil {
		now := time.Now()
		dateStr := now.Format("2006-01-02")
		timeStr := now.Format("15-04-05")
		id := fmt.Sprintf("game_%s", timeStr)
		dir := filepath.Join(r.baseDir, dateStr)
		_ = os.MkdirAll(dir, 0755)
		r.currentPath = filepath.Join(dir, id+".json")

		teamDisplay := myTeam
		if teamDisplay == "" {
			teamDisplay = "Bot"
		}

		gt := ""
		if len(gameType) > 0 {
			gt = gameType[0]
		}
		oppName := gt
		if oppName == "" {
			oppName = "Game"
		}

		r.currentMatch = &MatchHistory{
			ID:        id,
			Title:     fmt.Sprintf("%s vs %s (%s)", teamDisplay, oppName, now.Format("15:04:05")),
			Date:      dateStr,
			StartTime: now,
			GameType:  gt,
			MyTeam:    myTeam,
			Result:    "In Progress",
			Frames:    make([]HistoryFrame, 0, 100),
		}
		r.lastBoard = [8][8]string{}
		r.lastState = -1
		r.lastMove = ""
	}

	if myTeam != "" && r.currentMatch.MyTeam == "" {
		r.currentMatch.MyTeam = myTeam
	}

	boardChanged := !board.IsBoardEqual(r.lastBoard, b)
	stateChanged := r.lastState != state
	moveChanged := r.lastMove != move

	// Don't record duplicate consecutive idle frames unless game over or significant change
	if !boardChanged && !stateChanged && !moveChanged && !isGameOver && len(r.currentMatch.Frames) > 0 {
		return
	}

	frame := HistoryFrame{
		Timestamp:  time.Since(r.currentMatch.StartTime).Milliseconds(),
		Board:      b,
		State:      state,
		Move:       move,
		MyTeam:     myTeam,
		BestMove:   bestMove,
		From:       from,
		To:         to,
		IsGameOver: isGameOver,
		StatusText: statusText,
	}

	r.currentMatch.Frames = append(r.currentMatch.Frames, frame)
	r.lastBoard = b
	r.lastState = state
	r.lastMove = move

	if isGameOver && r.currentMatch.Result == "In Progress" {
		if bestMove != "" && bestMove != "-" {
			r.currentMatch.Result = bestMove
		} else {
			r.currentMatch.Result = "Game Over"
		}
	}

	r.saveCurrent()
}

func (r *Recorder) FinishMatch(result string) {
	if !r.enabled {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.currentMatch != nil {
		if result != "" {
			r.currentMatch.Result = result
		} else if r.currentMatch.Result == "In Progress" {
			r.currentMatch.Result = "Completed"
		}
		r.saveCurrent()
		r.currentMatch = nil
		r.currentPath = ""
		r.lastBoard = [8][8]string{}
		r.lastState = -1
		r.lastMove = ""
	}
}

func (r *Recorder) saveCurrent() {
	if r.currentMatch == nil || r.currentPath == "" {
		return
	}
	data, err := json.MarshalIndent(r.currentMatch, "", "  ")
	if err == nil {
		_ = os.WriteFile(r.currentPath, data, 0644)
	}
}
