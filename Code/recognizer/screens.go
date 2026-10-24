package recognizer

import "image"

func IsLobbyScreen(frame image.Image) bool {
	b := frame.Bounds()
	w, h := b.Dx(), b.Dy()

	cTop := frame.At(w/2, int(float64(h)*0.08))
	tr, tg, tb, _ := cTop.RGBA()
	tr8, tg8, tb8 := int(tr>>8), int(tg>>8), int(tb>>8)
	isTopTeal := (tg8 > 155 && tb8 > 110 && tr8 < 60)

	cBottom := frame.At(int(float64(w)*0.95), int(float64(h)*0.60))
	br, bg, bb, _ := cBottom.RGBA()
	br8, bg8, bb8 := int(br>>8), int(bg>>8), int(bb>>8)
	isBottomDark := (br8 < 45 && bg8 < 50 && bb8 < 55)

	return isTopTeal && isBottomDark
}

func IsMatchmakingScreen(frame image.Image) bool {
	b := frame.Bounds()
	w, h := b.Dx(), b.Dy()

	points := [][2]int{
		{int(float64(w) * 0.15), int(float64(h) * 0.20)},
		{int(float64(w) * 0.85), int(float64(h) * 0.20)},
		{int(float64(w) * 0.15), int(float64(h) * 0.40)},
	}

	matches := 0
	for _, pt := range points {
		c := frame.At(pt[0], pt[1])
		red, grn, blu, _ := c.RGBA()
		r8, g8, b8 := int(red>>8), int(grn>>8), int(blu>>8)

		if r8 < 50 && g8 > 165 && b8 > 115 && (g8-b8 > 15) {
			matches++
		}
	}
	return matches >= 2
}

func DetectActionButton(frame image.Image) (string, int, int) {
	b := frame.Bounds()
	w, h := b.Dx(), b.Dy()

	sampleX := int(float64(w) * 0.25)
	minY := int(float64(h) * 0.60)
	maxY := int(float64(h) * 0.96)

	type colorSpan struct {
		btnType string
		startY  int
		endY    int
	}

	var activeSpan *colorSpan
	var bestSpan *colorSpan
	maxSpanLen := 0

	for y := minY; y <= maxY; y += 4 {
		c := frame.At(sampleX, y)
		red, grn, blu, _ := c.RGBA()
		r8, g8, b8 := int(red>>8), int(grn>>8), int(blu>>8)

		currentType := ""
		if g8 > 155 && g8 > r8+35 && b8 < 100 {
			currentType = "continue"
		} else if b8 > 180 && g8 > 120 && r8 < 100 && (b8 > r8+70) {
			currentType = "claim_xp"
		} else if r8 > 190 && g8 > 140 && b8 < 75 && (r8-b8 > 120) {
			currentType = "undo"
		}

		if currentType != "" {
			if activeSpan == nil {
				activeSpan = &colorSpan{btnType: currentType, startY: y, endY: y}
			} else if activeSpan.btnType == currentType {
				activeSpan.endY = y
			} else {
				spanLen := activeSpan.endY - activeSpan.startY
				if spanLen > maxSpanLen && spanLen >= 20 {
					maxSpanLen = spanLen
					bestSpan = activeSpan
				}
				activeSpan = &colorSpan{btnType: currentType, startY: y, endY: y}
			}
		} else {
			if activeSpan != nil {
				spanLen := activeSpan.endY - activeSpan.startY
				if spanLen > maxSpanLen && spanLen >= 20 {
					maxSpanLen = spanLen
					bestSpan = activeSpan
				}
				activeSpan = nil
			}
		}
	}

	if activeSpan != nil {
		spanLen := activeSpan.endY - activeSpan.startY
		if spanLen > maxSpanLen && spanLen >= 20 {
			maxSpanLen = spanLen
			bestSpan = activeSpan
		}
	}

	if bestSpan != nil {
		targetX := w / 2
		targetY := (bestSpan.startY + bestSpan.endY) / 2
		return bestSpan.btnType, targetX, targetY
	}

	return "", 0, 0
}

func GetPromotionQueenTap(screenW, screenH int, yRatio float64, toR, toC int) (int, int) {
	cellW := float64(screenW) / 8.0
	cellH := cellW
	startY := float64(screenH) * yRatio

	queenX := int((float64(toC) + 0.5) * cellW)

	var queenY int
	if toR == 0 {
		queenY = int(startY + 0.5*cellH)
	} else {
		queenY = int(startY + 7.5*cellH)
	}

	return queenX, queenY
}
