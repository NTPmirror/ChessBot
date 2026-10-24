package engine

import (
	"math"
)

var pawnTable = [8][8]int{
	{0, 0, 0, 0, 0, 0, 0, 0},
	{50, 50, 50, 50, 50, 50, 50, 50},
	{10, 10, 20, 30, 30, 20, 10, 10},
	{5, 5, 10, 25, 25, 10, 5, 5},
	{0, 0, 0, 20, 20, 0, 0, 0},
	{5, -5, -10, 0, 0, -10, -5, 5},
	{5, 10, 10, -20, -20, 10, 10, 5},
	{0, 0, 0, 0, 0, 0, 0, 0},
}

var knightTable = [8][8]int{
	{-50, -40, -30, -30, -30, -30, -40, -50},
	{-40, -20, 0, 0, 0, 0, -20, -40},
	{-30, 0, 10, 15, 15, 10, 0, -30},
	{-30, 5, 15, 20, 20, 15, 5, -30},
	{-30, 0, 15, 20, 20, 15, 0, -30},
	{-30, 5, 10, 15, 15, 10, 5, -30},
	{-40, -20, 0, 5, 5, 0, -20, -40},
	{-50, -40, -30, -30, -30, -30, -40, -50},
}

var bishopTable = [8][8]int{
	{-20, -10, -10, -10, -10, -10, -10, -20},
	{-10, 0, 0, 0, 0, 0, 0, -10},
	{-10, 0, 5, 10, 10, 5, 0, -10},
	{-10, 5, 5, 10, 10, 5, 5, -10},
	{-10, 0, 10, 10, 10, 10, 0, -10},
	{-10, 10, 10, 10, 10, 10, 10, -10},
	{-10, 5, 0, 0, 0, 0, 5, -10},
	{-20, -10, -10, -10, -10, -10, -10, -20},
}

var rookTable = [8][8]int{
	{0, 0, 0, 0, 0, 0, 0, 0},
	{5, 10, 10, 10, 10, 10, 10, 5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{-5, 0, 0, 0, 0, 0, 0, -5},
	{0, 0, 0, 5, 5, 0, 0, 0},
}

var queenTable = [8][8]int{
	{-20, -10, -10, -5, -5, -10, -10, -20},
	{-10, 0, 0, 0, 0, 0, 0, -10},
	{-10, 0, 5, 5, 5, 5, 0, -10},
	{-5, 0, 5, 5, 5, 5, 0, -5},
	{0, 0, 5, 5, 5, 5, 0, -5},
	{-10, 5, 5, 5, 5, 5, 0, -10},
	{-10, 0, 5, 0, 0, 0, 0, -10},
	{-20, -10, -10, -5, -5, -10, -10, -20},
}

var kingMiddleTable = [8][8]int{
	{-30, -40, -40, -50, -50, -40, -40, -30},
	{-30, -40, -40, -50, -50, -40, -40, -30},
	{-30, -40, -40, -50, -50, -40, -40, -30},
	{-30, -40, -40, -50, -50, -40, -40, -30},
	{-20, -30, -30, -40, -40, -30, -30, -20},
	{-10, -20, -20, -20, -20, -20, -20, -10},
	{20, 20, 0, 0, 0, 0, 20, 20},
	{20, 30, 10, 0, 0, 10, 30, 20},
}

var kingEndTable = [8][8]int{
	{-50, -30, -30, -30, -30, -30, -30, -50},
	{-30, -10,   0,   0,   0,   0, -10, -30},
	{-30,  -5,  20,  30,  30,  20,  -5, -30},
	{-30,   0,  30,  40,  40,  30,   0, -30},
	{-30,   0,  30,  40,  40,  30,   0, -30},
	{-30,  -5,  20,  30,  30,  20,  -5, -30},
	{-30, -10,   5,  10,  10,   5, -10, -30},
	{-50, -30, -30, -30, -30, -30, -30, -50},
}

func getPSTValue(pType byte, r, c int, isMyPiece bool, isEndgame bool) int {
	mappedR := r
	if !isMyPiece {
		mappedR = 7 - r
	}

	switch pType {
	case 'P':
		return pawnTable[mappedR][c]
	case 'N':
		return knightTable[mappedR][c]
	case 'B':
		return bishopTable[mappedR][c]
	case 'R':
		return rookTable[mappedR][c]
	case 'Q':
		return queenTable[mappedR][c]
	case 'K':
		if isEndgame {
			return kingEndTable[mappedR][c]
		}
		return kingMiddleTable[mappedR][c]
	}
	return 0
}

func EvaluateBoard(b [8][8]string, myTeam string) int {
	myPrefix := byte('w')
	oppPrefix := byte('b')
	oppTeam := "Black"
	if myTeam == "Black" {
		myPrefix = 'b'
		oppPrefix = 'w'
		oppTeam = "White"
	}

	score := 0
	myMat := 0
	oppMat := 0
	oppKingR, oppKingC := -1, -1

	var myPawnsInCol [8]int
	var oppPawnsInCol [8]int
	myBishops := 0
	oppBishops := 0

	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			p := b[r][c]
			if len(p) < 2 {
				continue
			}
			val := PieceValues[p[1]]
			if p[0] == myPrefix {
				myMat += val
				if p[1] == 'P' {
					myPawnsInCol[c]++
				} else if p[1] == 'B' {
					myBishops++
				}
			} else {
				oppMat += val
				if p[1] == 'P' {
					oppPawnsInCol[c]++
				} else if p[1] == 'B' {
					oppBishops++
				} else if p[1] == 'K' {
					oppKingR, oppKingC = r, c
				}
			}
		}
	}

	isEndgame := (myMat+oppMat < 2800)

	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			p := b[r][c]
			if len(p) < 2 {
				continue
			}
			val := PieceValues[p[1]]

			if p[0] == myPrefix {
				pst := getPSTValue(p[1], r, c, true, isEndgame)
				score += val + pst

				if p[1] == 'R' {
					if myPawnsInCol[c] == 0 {
						if oppPawnsInCol[c] == 0 {
							score += 25
						} else {
							score += 12
						}
					}
				}
			} else {
				pst := getPSTValue(p[1], r, c, false, isEndgame)
				score -= val + pst

				if p[1] == 'R' {
					if oppPawnsInCol[c] == 0 {
						if myPawnsInCol[c] == 0 {
							score -= 25
						} else {
							score -= 12
						}
					}
				}
			}
		}
	}

	if myBishops >= 2 {
		score += 30
	}
	if oppBishops >= 2 {
		score -= 30
	}

	passedPawnBonus := [8]int{0, 10, 20, 35, 60, 100, 160, 0}
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			p := b[r][c]
			if len(p) < 2 || p[1] != 'P' {
				continue
			}

			if p[0] == myPrefix {
				isPassed := true
				for checkR := r - 1; checkR >= 0; checkR-- {
					for dc := -1; dc <= 1; dc++ {
						nc := c + dc
						if nc >= 0 && nc < 8 {
							chk := b[checkR][nc]
							if len(chk) >= 2 && chk[0] == oppPrefix && chk[1] == 'P' {
								isPassed = false
								break
							}
						}
					}
					if !isPassed {
						break
					}
				}
				if isPassed {
					rankAdvancement := 7 - r
					bonus := passedPawnBonus[rankAdvancement]
					if isEndgame {
						bonus = (bonus * 3) / 2
					}
					score += bonus
				}
			} else {
				isPassed := true
				for checkR := r + 1; checkR < 8; checkR++ {
					for dc := -1; dc <= 1; dc++ {
						nc := c + dc
						if nc >= 0 && nc < 8 {
							chk := b[checkR][nc]
							if len(chk) >= 2 && chk[0] == myPrefix && chk[1] == 'P' {
								isPassed = false
								break
							}
						}
					}
					if !isPassed {
						break
					}
				}
				if isPassed {
					rankAdvancement := r
					bonus := passedPawnBonus[rankAdvancement]
					if isEndgame {
						bonus = (bonus * 3) / 2
					}
					score -= bonus
				}
			}
		}
	}

	for c := 0; c < 8; c++ {
		if myPawnsInCol[c] > 1 {
			score -= 16 * (myPawnsInCol[c] - 1)
		}
		if oppPawnsInCol[c] > 1 {
			score += 16 * (oppPawnsInCol[c] - 1)
		}

		if myPawnsInCol[c] > 0 {
			hasLeft := (c > 0 && myPawnsInCol[c-1] > 0)
			hasRight := (c < 7 && myPawnsInCol[c+1] > 0)
			if !hasLeft && !hasRight {
				score -= 14
			}
		}
		if oppPawnsInCol[c] > 0 {
			hasLeft := (c > 0 && oppPawnsInCol[c-1] > 0)
			hasRight := (c < 7 && oppPawnsInCol[c+1] > 0)
			if !hasLeft && !hasRight {
				score += 14
			}
		}
	}

	if isEndgame && myMat > oppMat+300 && oppKingR != -1 {
		distFromCenter := int(math.Abs(float64(oppKingR)-3.5) + math.Abs(float64(oppKingC)-3.5))
		score += distFromCenter * 40
	}

	if IsKingInCheck(b, oppTeam, myTeam) {
		score += 70
	}

	_ = oppPrefix
	return score
}
