package board

import "strings"

func DetectWhoMoved(prev, curr [8][8]string) string {
	whiteVacated, blackVacated := 0, 0
	whiteArrived, blackArrived := 0, 0

	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			p := prev[r][c]
			n := curr[r][c]
			if p == n {
				continue
			}

			if strings.HasPrefix(p, "w") && n == "." {
				whiteVacated++
			}
			if strings.HasPrefix(p, "b") && n == "." {
				blackVacated++
			}

			if strings.HasPrefix(n, "w") && !strings.HasPrefix(p, "w") {
				whiteArrived++
			}
			if strings.HasPrefix(n, "b") && !strings.HasPrefix(p, "b") {
				blackArrived++
			}
		}
	}

	if whiteArrived > 0 && blackArrived == 0 {
		return "White"
	}
	if blackArrived > 0 && whiteArrived == 0 {
		return "Black"
	}
	if whiteVacated > blackVacated {
		return "White"
	}
	if blackVacated > whiteVacated {
		return "Black"
	}
	return ""
}

func InferInitialTurn(board [8][8]string, myTeam string, lastMovedTeam string) string {
	if lastMovedTeam == "White" {
		return "Black"
	}
	if lastMovedTeam == "Black" {
		return "White"
	}

	if myTeam == "White" {
		whitePawnsMoved := 0
		for c := 0; c < 8; c++ {
			if !strings.HasPrefix(board[6][c], "wP") {
				whitePawnsMoved++
			}
		}
		blackPawnsMoved := 0
		for c := 0; c < 8; c++ {
			if !strings.HasPrefix(board[1][c], "bP") {
				blackPawnsMoved++
			}
		}

		if whitePawnsMoved == 0 && blackPawnsMoved == 0 {
			if board[7][1] != "wN" || board[7][6] != "wN" {
				return "Black"
			}
			return "White"
		}
		if whitePawnsMoved > blackPawnsMoved {
			return "Black"
		}
		return "White"
	}

	whitePawnsMoved := 0
	for c := 0; c < 8; c++ {
		if !strings.HasPrefix(board[1][c], "wP") {
			whitePawnsMoved++
		}
	}
	blackPawnsMoved := 0
	for c := 0; c < 8; c++ {
		if !strings.HasPrefix(board[6][c], "bP") {
			blackPawnsMoved++
		}
	}

	whiteAdvanced := 0
	for r := 2; r <= 5; r++ {
		for c := 0; c < 8; c++ {
			if strings.HasPrefix(board[r][c], "w") {
				whiteAdvanced++
			}
		}
	}

	whiteKnightsMoved := 0
	if board[0][1] == "." || !strings.HasPrefix(board[0][1], "wN") {
		whiteKnightsMoved++
	}
	if board[0][6] == "." || !strings.HasPrefix(board[0][6], "wN") {
		whiteKnightsMoved++
	}

	if (whitePawnsMoved > 0 || whiteAdvanced > 0 || whiteKnightsMoved > 0) && blackPawnsMoved == 0 {
		return "Black"
	}

	if whitePawnsMoved > blackPawnsMoved {
		return "Black"
	}

	if whitePawnsMoved == 0 && whiteAdvanced == 0 && whiteKnightsMoved == 0 {
		return "White"
	}

	return "White"
}
