package web

import (
	"encoding/json"
	"strings"
	"sync"
	"time"
)

type WSHub struct {
	sync.RWMutex
	clients       map[chan []byte]struct{}
	lastBroadcast time.Time
	lastSig       string
}

func NewHub() *WSHub {
	return &WSHub{clients: make(map[chan []byte]struct{})}
}

func (h *WSHub) Register(ch chan []byte) {
	h.Lock()
	defer h.Unlock()
	h.clients[ch] = struct{}{}
}

func (h *WSHub) Unregister(ch chan []byte) {
	h.Lock()
	defer h.Unlock()
	delete(h.clients, ch)
	close(ch)
}

func boardSignature(board [8][8]string, state int, move, myTeam, bestMove string, from, to [2]int, isGameOver bool) string {
	var sb strings.Builder
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			sb.WriteString(board[r][c])
		}
	}
	sb.WriteString(move)
	sb.WriteString(myTeam)
	sb.WriteString(bestMove)
	return sb.String()
}

func (h *WSHub) Broadcast(board [8][8]string, state int, move, myTeam, bestMove string, from, to [2]int, isGameOver bool, isMyCheck bool, isOppCheck bool) {
	sig := boardSignature(board, state, move, myTeam, bestMove, from, to, isGameOver)
	if isMyCheck {
		sig += ":mycheck"
	}
	if isOppCheck {
		sig += ":oppcheck"
	}
	now := time.Now()

	if sig == h.lastSig && now.Sub(h.lastBroadcast) < 250*time.Millisecond {
		return
	}
	if sig != h.lastSig && now.Sub(h.lastBroadcast) < 40*time.Millisecond {
		return
	}

	h.lastSig = sig
	h.lastBroadcast = now

	data, err := json.Marshal(map[string]interface{}{
		"board":          board,
		"state":          state,
		"move":           move,
		"my_team":        myTeam,
		"best_move":      bestMove,
		"from":           from,
		"to":             to,
		"is_game_over":   isGameOver,
		"is_check":       isMyCheck,
		"is_my_check":    isMyCheck,
		"is_opp_check":   isOppCheck,
		"timestamp":      now.UnixMilli(),
	})
	if err != nil {
		return
	}

	frame := EncodeWSFrame(data)

	h.RLock()
	defer h.RUnlock()
	for ch := range h.clients {
		select {
		case ch <- frame:
		default:
			select {
			case <-ch:
			default:
			}
			select {
			case ch <- frame:
			default:
			}
		}
	}
}

func EncodeWSFrame(payload []byte) []byte {
	n := len(payload)
	var header []byte
	if n < 126 {
		header = []byte{0x81, byte(n)}
	} else if n <= 65535 {
		header = []byte{0x81, 126, byte(n >> 8), byte(n & 0xFF)}
	} else {
		header = []byte{0x81, 127, 0, 0, 0, 0, byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n & 0xFF)}
	}
	return append(header, payload...)
}
