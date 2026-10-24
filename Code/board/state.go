package board

import "strings"

func CountBoardPieces(b [8][8]string) int {
	cnt := 0
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			if b[r][c] != "." && b[r][c] != "" {
				cnt++
			}
		}
	}
	return cnt
}

func IsBoardEqual(a, b [8][8]string) bool {
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			if a[r][c] != b[r][c] {
				return false
			}
		}
	}
	return true
}

func BoardToKey(b [8][8]string) string {
	var sb strings.Builder
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			sb.WriteString(b[r][c])
		}
	}
	return sb.String()
}

func DetermineMyTeam(board [8][8]string) string {
	wBottom, bBottom := 0, 0
	for r := 5; r < 8; r++ {
		for c := 0; c < 8; c++ {
			if strings.HasPrefix(board[r][c], "w") {
				wBottom++
			} else if strings.HasPrefix(board[r][c], "b") {
				bBottom++
			}
		}
	}
	if wBottom == 0 && bBottom == 0 {
		return ""
	}
	if bBottom > wBottom {
		return "Black"
	}
	return "White"
}

func FindPendingPromotion(board [8][8]string, myTeam string) (int, int, bool) {
	targetPiece := "wP"
	if myTeam == "Black" {
		targetPiece = "bP"
	}

	for c := 0; c < 8; c++ {
		if board[0][c] == targetPiece {
			return 0, c, true
		}
	}
	return -1, -1, false
}
