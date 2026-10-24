package recognizer

import (
	"image"
	"strings"
)

func DetectLastMoveHighlight(frame image.Image, yRatio float64, board [8][8]string) (whoMoved string, toR, toC int) {
	w := frame.Bounds().Dx()
	h := frame.Bounds().Dy()
	startY := float64(h) * yRatio
	cellW := float64(w) / 8.0
	cellH := cellW

	bestScore := 0
	bestR, bestC := -1, -1

	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			offsets := [][2]int{
				{8, 8},
				{int(cellW) - 8, 8},
				{8, int(cellH) - 8},
				{int(cellW) - 8, int(cellH) - 8},
			}

			warmCount := 0
			for _, off := range offsets {
				px := int(float64(c)*cellW) + off[0]
				py := int(startY+float64(r)*cellH) + off[1]
				col := frame.At(px, py)
				red, grn, blu, _ := col.RGBA()
				r8, g8, b8 := int(red>>8), int(grn>>8), int(blu>>8)

				if r8 > 60 && r8 >= b8 && (r8+g8) > 130 {
					warmCount++
				}
			}

			if warmCount >= 2 && warmCount > bestScore {
				bestScore = warmCount
				bestR = r
				bestC = c
			}
		}
	}

	if bestR != -1 {
		piece := board[bestR][bestC]
		if strings.HasPrefix(piece, "w") {
			return "White", bestR, bestC
		} else if strings.HasPrefix(piece, "b") {
			return "Black", bestR, bestC
		}
		return "", bestR, bestC
	}

	return "", -1, -1
}
