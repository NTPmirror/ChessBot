package duolingo

import "image"

// IsMatchmakingScreen detects the Duolingo opponent search screen.
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

// DetectActionButton locates end-game and intermediate Duolingo action buttons ("continue", "claim_xp", "undo").
func DetectActionButton(frame image.Image) (string, int, int) {
	b := frame.Bounds()
	w, h := b.Dx(), b.Dy()

	if !IsLobbyScreen(frame) {
		yStart := int(float64(h) * 0.915)
		yEnd := int(float64(h) * 0.938)
		xStart := int(float64(w) * 0.35)
		xEnd := int(float64(w) * 0.65)

		brightCount := 0
		redBtnCount := 0
		blueBtnCount := 0
		totalSamples := 0
		for y := yStart; y <= yEnd; y += 4 {
			for x := xStart; x <= xEnd; x += 8 {
				c := frame.At(x, y)
				r, g, bl, _ := c.RGBA()
				r8, g8, b8 := int(r>>8), int(g>>8), int(bl>>8)
				lum := (r8*299 + g8*587 + b8*114) / 1000
				if lum > 140 {
					brightCount++
				}
				if r8 > 185 && r8 > g8+65 && r8 > b8+65 {
					redBtnCount++
				}
				if b8 > 175 && g8 > 135 && b8 > r8+70 {
					blueBtnCount++
				}
				totalSamples++
			}
		}

		if totalSamples > 0 && float64(brightCount)/float64(totalSamples) > 0.14 {
			targetX := w / 2
			targetY := int(float64(h) * 0.926)
			return "continue", targetX, targetY
		}
		if totalSamples > 0 && float64(redBtnCount)/float64(totalSamples) > 0.30 {
			targetX := int(float64(w) * 0.58)
			targetY := int(float64(h) * 0.925)
			return "continue", targetX, targetY
		}
		if totalSamples > 0 && float64(blueBtnCount)/float64(totalSamples) > 0.30 {
			targetX := w / 2
			targetY := int(float64(h) * 0.926)
			return "continue", targetX, targetY
		}
	}

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
		} else if r8 > 230 && g8 > 230 && b8 > 230 {
			currentType = "continue"
		} else if g8 > 140 && b8 > 90 && g8 > r8+60 && r8 < 80 {
			currentType = "continue"
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

		if targetY > int(float64(h)*0.88) {
			cLeft := frame.At(int(float64(w)*0.30), targetY)
			cRight := frame.At(int(float64(w)*0.70), targetY)
			lr, lg, lb, _ := cLeft.RGBA()
			rr, rg, rb, _ := cRight.RGBA()
			lLum := (int(lr>>8) + int(lg>>8) + int(lb>>8)) / 3
			rLum := (int(rr>>8) + int(rg>>8) + int(rb>>8)) / 3

			if lLum < 120 || rLum < 120 {
				return "", 0, 0
			}
		}

		return bestSpan.btnType, targetX, targetY
	}

	return "", 0, 0
}
