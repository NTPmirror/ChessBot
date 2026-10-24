package engine

func IsKingInCheck(b [8][8]string, kingTeam, myTeam string) bool {
	kingTarget := "wK"
	oppPrefix := "b"
	if kingTeam == "Black" {
		kingTarget = "bK"
		oppPrefix = "w"
	}

	kR, kC := -1, -1
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			if b[r][c] == kingTarget {
				kR, kC = r, c
				break
			}
		}
		if kR != -1 {
			break
		}
	}
	if kR == -1 {
		return false
	}

	return IsSquareAttacked(b, kR, kC, oppPrefix, myTeam)
}

func IsSquareAttacked(b [8][8]string, targetR, targetC int, oppPrefix string, myTeam string) bool {
	pawnDir := 1
	oppTeam := "White"
	if oppPrefix == "b" {
		oppTeam = "Black"
	}
	if oppTeam == myTeam {
		pawnDir = -1
	}
	pawnAttackFromR := targetR - pawnDir
	if pawnAttackFromR >= 0 && pawnAttackFromR < 8 {
		for _, dc := range []int{-1, 1} {
			nc := targetC + dc
			if nc >= 0 && nc < 8 {
				p := b[pawnAttackFromR][nc]
				if len(p) >= 2 && p[0] == oppPrefix[0] && p[1] == 'P' {
					return true
				}
			}
		}
	}

	knightDeltas := [][2]int{
		{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2},
		{1, -2}, {1, 2}, {2, -1}, {2, 1},
	}
	for _, d := range knightDeltas {
		nr, nc := targetR+d[0], targetC+d[1]
		if nr >= 0 && nr < 8 && nc >= 0 && nc < 8 {
			p := b[nr][nc]
			if len(p) >= 2 && p[0] == oppPrefix[0] && p[1] == 'N' {
				return true
			}
		}
	}

	diagDirs := [][2]int{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}}
	for _, d := range diagDirs {
		nr, nc := targetR+d[0], targetC+d[1]
		for nr >= 0 && nr < 8 && nc >= 0 && nc < 8 {
			p := b[nr][nc]
			if p != "." {
				if len(p) >= 2 && p[0] == oppPrefix[0] && (p[1] == 'B' || p[1] == 'Q') {
					return true
				}
				break
			}
			nr += d[0]
			nc += d[1]
		}
	}

	orthoDirs := [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	for _, d := range orthoDirs {
		nr, nc := targetR+d[0], targetC+d[1]
		for nr >= 0 && nr < 8 && nc >= 0 && nc < 8 {
			p := b[nr][nc]
			if p != "." {
				if len(p) >= 2 && p[0] == oppPrefix[0] && (p[1] == 'R' || p[1] == 'Q') {
					return true
				}
				break
			}
			nr += d[0]
			nc += d[1]
		}
	}

	kingDirs := [][2]int{
		{-1, -1}, {-1, 0}, {-1, 1},
		{0, -1}, {0, 1},
		{1, -1}, {1, 0}, {1, 1},
	}
	for _, d := range kingDirs {
		nr, nc := targetR+d[0], targetC+d[1]
		if nr >= 0 && nr < 8 && nc >= 0 && nc < 8 {
			p := b[nr][nc]
			if len(p) >= 2 && p[0] == oppPrefix[0] && p[1] == 'K' {
				return true
			}
		}
	}

	return false
}

func GenerateRawMoves(b [8][8]string, forTeam, myTeam string, onlyCaptures bool) []Move {
	moves := make([]Move, 0, 36)

	prefix := "w"
	oppPrefix := "b"
	if forTeam == "Black" {
		prefix = "b"
		oppPrefix = "w"
	}

	pawnDir := 1
	pawnStartRow := 1
	if forTeam == myTeam {
		pawnDir = -1
		pawnStartRow = 6
	}

	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			p := b[r][c]
			if len(p) < 2 || p[0] != prefix[0] {
				continue
			}
			pType := p[1]

			switch pType {
			case 'P':
				if !onlyCaptures {
					nr := r + pawnDir
					if nr >= 0 && nr < 8 && b[nr][c] == "." {
						moves = append(moves, Move{FromR: r, FromC: c, ToR: nr, ToC: c, Piece: p})
						if r == pawnStartRow && b[r+2*pawnDir][c] == "." {
							moves = append(moves, Move{FromR: r, FromC: c, ToR: r + 2*pawnDir, ToC: c, Piece: p})
						}
					}
				}
				for _, dc := range []int{-1, 1} {
					nc := c + dc
					nr := r + pawnDir
					if nr >= 0 && nr < 8 && nc >= 0 && nc < 8 {
						target := b[nr][nc]
						if len(target) >= 2 && target[0] == oppPrefix[0] {
							moves = append(moves, Move{FromR: r, FromC: c, ToR: nr, ToC: nc, Piece: p})
						}
					}
				}

			case 'N':
				deltas := [][2]int{
					{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2},
					{1, -2}, {1, 2}, {2, -1}, {2, 1},
				}
				for _, d := range deltas {
					nr, nc := r+d[0], c+d[1]
					if nr >= 0 && nr < 8 && nc >= 0 && nc < 8 {
						target := b[nr][nc]
						isCapture := len(target) >= 2 && target[0] == oppPrefix[0]
						if isCapture || (!onlyCaptures && target == ".") {
							moves = append(moves, Move{FromR: r, FromC: c, ToR: nr, ToC: nc, Piece: p})
						}
					}
				}

			case 'B', 'R', 'Q':
				var dirs [][2]int
				if pType == 'B' || pType == 'Q' {
					dirs = append(dirs, [][2]int{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}}...)
				}
				if pType == 'R' || pType == 'Q' {
					dirs = append(dirs, [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}...)
				}
				for _, d := range dirs {
					nr, nc := r+d[0], c+d[1]
					for nr >= 0 && nr < 8 && nc >= 0 && nc < 8 {
						target := b[nr][nc]
						if target == "." {
							if !onlyCaptures {
								moves = append(moves, Move{FromR: r, FromC: c, ToR: nr, ToC: nc, Piece: p})
							}
						} else {
							if len(target) >= 2 && target[0] == oppPrefix[0] {
								moves = append(moves, Move{FromR: r, FromC: c, ToR: nr, ToC: nc, Piece: p})
							}
							break
						}
						nr += d[0]
						nc += d[1]
					}
				}

			case 'K':
				dirs := [][2]int{
					{-1, -1}, {-1, 0}, {-1, 1},
					{0, -1}, {0, 1},
					{1, -1}, {1, 0}, {1, 1},
				}
				for _, d := range dirs {
					nr, nc := r+d[0], c+d[1]
					if nr >= 0 && nr < 8 && nc >= 0 && nc < 8 {
						target := b[nr][nc]
						isCapture := len(target) >= 2 && target[0] == oppPrefix[0]
						if isCapture || (!onlyCaptures && target == ".") {
							moves = append(moves, Move{FromR: r, FromC: c, ToR: nr, ToC: nc, Piece: p})
						}
					}
				}
			}
		}
	}
	return moves
}

func ApplyMove(b [8][8]string, m Move) [8][8]string {
	nb := b
	nb[m.ToR][m.ToC] = m.Piece
	nb[m.FromR][m.FromC] = "."

	if len(m.Piece) >= 2 && m.Piece[1] == 'P' {
		if m.ToR == 0 || m.ToR == 7 {
			nb[m.ToR][m.ToC] = string(m.Piece[0]) + "Q"
		}
	}
	return nb
}

func GetLegalMoves(b [8][8]string, forTeam, myTeam string, onlyCaptures bool) []Move {
	raw := GenerateRawMoves(b, forTeam, myTeam, onlyCaptures)
	legal := make([]Move, 0, len(raw))
	for _, m := range raw {
		nextB := ApplyMove(b, m)
		if !IsKingInCheck(nextB, forTeam, myTeam) {
			legal = append(legal, m)
		}
	}
	return legal
}
