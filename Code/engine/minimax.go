package engine

import (
	"math/rand"
	"runtime"
	"sort"
	"sync"
)

var (
	zobristPieces [8][8][12]uint64
	zobristBlack  uint64
	zobristInit   bool
)

func initZobrist() {
	if zobristInit {
		return
	}
	r := rand.New(rand.NewSource(133742))
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			for p := 0; p < 12; p++ {
				zobristPieces[row][col][p] = r.Uint64()
			}
		}
	}
	zobristBlack = r.Uint64()
	zobristInit = true
}

func pieceToIndex(p string) int {
	if len(p) < 2 {
		return -1
	}
	offset := 0
	if p[0] == 'b' {
		offset = 6
	}
	switch p[1] {
	case 'P':
		return offset + 0
	case 'N':
		return offset + 1
	case 'B':
		return offset + 2
	case 'R':
		return offset + 3
	case 'Q':
		return offset + 4
	case 'K':
		return offset + 5
	}
	return -1
}

func ComputeZobristHash(b [8][8]string, isBlackTurn bool) uint64 {
	initZobrist()
	var h uint64
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			idx := pieceToIndex(b[r][c])
			if idx != -1 {
				h ^= zobristPieces[r][c][idx]
			}
		}
	}
	if isBlackTurn {
		h ^= zobristBlack
	}
	return h
}

type TTEntry struct {
	depth int
	score int
	flag  int
	bestM Move
}

const ttShards = 16

type ShardedTT struct {
	mu     sync.RWMutex
	table  map[uint64]TTEntry
	maxLen int
}

var tt [ttShards]ShardedTT

func init() {
	for i := 0; i < ttShards; i++ {
		tt[i].table = make(map[uint64]TTEntry, 4096)
		tt[i].maxLen = 20000
	}
}

func getTTEntry(hash uint64) (TTEntry, bool) {
	idx := hash % ttShards
	tt[idx].mu.RLock()
	entry, ok := tt[idx].table[hash]
	tt[idx].mu.RUnlock()
	return entry, ok
}

func putTTEntry(hash uint64, entry TTEntry) {
	idx := hash % ttShards
	tt[idx].mu.Lock()
	if len(tt[idx].table) >= tt[idx].maxLen {
		tt[idx].table = make(map[uint64]TTEntry, 4096)
	}
	tt[idx].table[hash] = entry
	tt[idx].mu.Unlock()
}

var (
	killerMoves [32][2]Move
	killerMu    sync.RWMutex
)

func getKillers(ply int) (Move, Move) {
	if ply < 0 || ply >= 32 {
		return Move{}, Move{}
	}
	killerMu.RLock()
	k1, k2 := killerMoves[ply][0], killerMoves[ply][1]
	killerMu.RUnlock()
	return k1, k2
}

func storeKiller(ply int, m Move) {
	if ply < 0 || ply >= 32 {
		return
	}
	killerMu.Lock()
	if killerMoves[ply][0] != m {
		killerMoves[ply][1] = killerMoves[ply][0]
		killerMoves[ply][0] = m
	}
	killerMu.Unlock()
}

func scoreMove(b [8][8]string, m Move, ttBest Move, k1, k2 Move) int {
	if m.FromR == ttBest.FromR && m.FromC == ttBest.FromC && m.ToR == ttBest.ToR && m.ToC == ttBest.ToC {
		return 1000000
	}

	target := b[m.ToR][m.ToC]
	if target != "." && len(target) >= 2 {
		victimVal := PieceValues[target[1]]
		aggressorVal := PieceValues[m.Piece[1]]
		return 100000 + (victimVal*10 - aggressorVal)
	}

	if m == k1 {
		return 50000
	}
	if m == k2 {
		return 40000
	}

	if m.Piece[1] == 'P' && (m.ToR == 1 || m.ToR == 6) {
		return 20000
	}

	return 0
}

func OrderMoves(b [8][8]string, moves []Move, ttBest Move, ply int) {
	k1, k2 := getKillers(ply)
	sort.Slice(moves, func(i, j int) bool {
		return scoreMove(b, moves[i], ttBest, k1, k2) > scoreMove(b, moves[j], ttBest, k1, k2)
	})
}

func Quiescence(b [8][8]string, alpha, beta int, isMax bool, myTeam string, qDepth int) int {
	standPat := EvaluateBoard(b, myTeam)
	if qDepth <= 0 {
		return standPat
	}

	if isMax {
		if standPat >= beta {
			return beta
		}
		if standPat > alpha {
			alpha = standPat
		}
	} else {
		if standPat <= alpha {
			return alpha
		}
		if standPat < beta {
			beta = standPat
		}
	}

	oppTeam := "Black"
	if myTeam == "Black" {
		oppTeam = "White"
	}
	currentForTeam := myTeam
	if !isMax {
		currentForTeam = oppTeam
	}

	captures := GetLegalMoves(b, currentForTeam, myTeam, true)
	if len(captures) == 0 {
		return standPat
	}

	OrderMoves(b, captures, Move{}, 0)

	if isMax {
		for _, m := range captures {
			nextB := ApplyMove(b, m)
			score := Quiescence(nextB, alpha, beta, false, myTeam, qDepth-1)
			if score >= beta {
				return beta
			}
			if score > alpha {
				alpha = score
			}
		}
		return alpha
	} else {
		for _, m := range captures {
			nextB := ApplyMove(b, m)
			score := Quiescence(nextB, alpha, beta, true, myTeam, qDepth-1)
			if score <= alpha {
				return alpha
			}
			if score < beta {
				beta = score
			}
		}
		return beta
	}
}

func Minimax(b [8][8]string, depth, alpha, beta int, isMax bool, myTeam string) int {
	oppTeam := "Black"
	if myTeam == "Black" {
		oppTeam = "White"
	}

	currentForTeam := myTeam
	if !isMax {
		currentForTeam = oppTeam
	}

	hash := ComputeZobristHash(b, currentForTeam == "Black")
	if entry, ok := getTTEntry(hash); ok && entry.depth >= depth {
		if entry.flag == 0 {
			return entry.score
		} else if entry.flag == 1 && entry.score <= alpha {
			return alpha
		} else if entry.flag == 2 && entry.score >= beta {
			return beta
		}
	}

	moves := GetLegalMoves(b, currentForTeam, myTeam, false)

	if len(moves) == 0 {
		inCheck := IsKingInCheck(b, currentForTeam, myTeam)
		if inCheck {
			if isMax {
				return -99999 + (10 - depth)
			} else {
				return 99999 - (10 - depth)
			}
		} else {
			return 0
		}
	}

	if depth <= 0 {
		return Quiescence(b, alpha, beta, isMax, myTeam, 2)
	}

	var ttBest Move
	if entry, ok := getTTEntry(hash); ok {
		ttBest = entry.bestM
	}
	OrderMoves(b, moves, ttBest, depth)

	origAlpha := alpha
	origBeta := beta
	bestScore := -1000000
	if !isMax {
		bestScore = 1000000
	}
	var bestMove Move

	if isMax {
		for _, m := range moves {
			nextB := ApplyMove(b, m)
			eval := Minimax(nextB, depth-1, alpha, beta, false, myTeam)
			if eval > bestScore {
				bestScore = eval
				bestMove = m
			}
			if eval > alpha {
				alpha = eval
			}
			if beta <= alpha {
				if b[m.ToR][m.ToC] == "." {
					storeKiller(depth, m)
				}
				break
			}
		}
	} else {
		for _, m := range moves {
			nextB := ApplyMove(b, m)
			eval := Minimax(nextB, depth-1, alpha, beta, true, myTeam)
			if eval < bestScore {
				bestScore = eval
				bestMove = m
			}
			if eval < beta {
				beta = eval
			}
			if beta <= alpha {
				if b[m.ToR][m.ToC] == "." {
					storeKiller(depth, m)
				}
				break
			}
		}
	}

	flag := 0
	if isMax {
		if bestScore <= origAlpha {
			flag = 1
		} else if bestScore >= beta {
			flag = 2
		}
	} else {
		if bestScore >= origBeta {
			flag = 2
		} else if bestScore <= alpha {
			flag = 1
		}
	}
	putTTEntry(hash, TTEntry{depth: depth, score: bestScore, flag: flag, bestM: bestMove})

	return bestScore
}

type moveResult struct {
	m     Move
	score int
}

func FindBestMove(b [8][8]string, myTeam string, depth int, blacklist map[string]bool) (Move, bool) {
	moves := GetLegalMoves(b, myTeam, myTeam, false)
	if len(moves) == 0 {
		return Move{}, false
	}

	hash := ComputeZobristHash(b, myTeam == "Black")
	var ttBest Move
	if entry, ok := getTTEntry(hash); ok {
		ttBest = entry.bestM
	}
	OrderMoves(b, moves, ttBest, depth)

	var validMoves []Move
	for _, m := range moves {
		if blacklist != nil && blacklist[m.String()] {
			continue
		}
		validMoves = append(validMoves, m)
	}
	if len(validMoves) == 0 {
		if len(moves) > 0 {
			return moves[0], true
		}
		return Move{}, false
	}

	searchDepth := depth
	if searchDepth < 4 {
		searchDepth = 4
	}

	numWorkers := runtime.GOMAXPROCS(0)
	if numWorkers > len(validMoves) {
		numWorkers = len(validMoves)
	}
	if numWorkers < 1 {
		numWorkers = 1
	}

	jobs := make(chan Move, len(validMoves))
	results := make(chan moveResult, len(validMoves))
	var wg sync.WaitGroup

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for m := range jobs {
				nextB := ApplyMove(b, m)
				score := Minimax(nextB, searchDepth-1, -1000000, 1000000, false, myTeam)
				results <- moveResult{m: m, score: score}
			}
		}()
	}

	for _, m := range validMoves {
		jobs <- m
	}
	close(jobs)

	wg.Wait()
	close(results)

	bestScore := -1000000
	var candidateMoves []Move

	for res := range results {
		if res.score > bestScore {
			bestScore = res.score
			candidateMoves = []Move{res.m}
		} else if res.score == bestScore {
			candidateMoves = append(candidateMoves, res.m)
		}
	}

	if len(candidateMoves) == 0 {
		return validMoves[0], true
	}

	return candidateMoves[0], true
}
