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

	if isTopTeal && isBottomDark {
		return true
	}


	return IsNewLobbyScreen(frame)
}


func IsNewLobbyScreen(frame image.Image) bool {
	b := frame.Bounds()
	w, h := b.Dx(), b.Dy()

	cTop := frame.At(w/2, int(float64(h)*0.08))
	tr, tg, tb, _ := cTop.RGBA()
	tr8, tg8, tb8 := int(tr>>8), int(tg>>8), int(tb>>8)

	if tr8 < 60 && tg8 < 60 && tb8 < 60 {
		btnY := int(float64(h) * 0.375)
		samplePoints := []int{
			int(float64(w) * 0.20),
			int(float64(w) * 0.35),
			int(float64(w) * 0.50),
			int(float64(w) * 0.65),
			int(float64(w) * 0.80),
		}
		btnMatches := 0
		for _, px := range samplePoints {
			c := frame.At(px, btnY)
			red, grn, blu, _ := c.RGBA()
			r, g, bVal := int(red>>8), int(grn>>8), int(blu>>8)
			isTeal := (g > 120 && bVal > 70 && g > r+35)
			isPurple := (r > 170 && bVal > 170 && r > g+35 && bVal > g+35)
			if isTeal || isPurple {
				btnMatches++
			}
		}
		if btnMatches >= 3 {
			darkCount := 0
			for _, fx := range []float64{0.35, 0.50, 0.65, 0.80} {
				cTab := frame.At(int(float64(w)*fx), int(float64(h)*0.926))
				dr, dg, db, _ := cTab.RGBA()
				if (dr>>8) < 55 && (dg>>8) < 55 && (db>>8) < 55 {
					darkCount++
				}
			}
			return darkCount >= 2
		}
	}
	return false
}


func GetNewLobbyCardSelection(frame image.Image) int {
	b := frame.Bounds()
	w, h := b.Dx(), b.Dy()

	cardX := []float64{0.20, 0.50, 0.80}

	for i, xFrac := range cardX {
		px := int(float64(w) * xFrac)

		for _, yFrac := range []float64{0.158, 0.160, 0.162, 0.164, 0.166} {
			py := int(float64(h) * yFrac)
			c := frame.At(px, py)
			red, grn, blu, _ := c.RGBA()
			r, g, bVal := int(red>>8), int(grn>>8), int(blu>>8)
			isGreen := (g > 45 && g > r+25)
			isPurple := (r > 120 && bVal > 140 && r > g+25 && bVal > g+25)
			if isGreen || isPurple {
				return i
			}
		}
	}
	return -1
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

func GetPromotionQueenTap(screenW, screenH int, yRatio float64, toR, toC int) (int, int) {


	queenY := int(float64(screenH) * 0.494)

	if toC >= 0 && toC < 8 {
		cellW := float64(screenW) / 8.0
		colCenterX := (float64(toC) + 0.5) * cellW


		minCenter := float64(screenW) * 0.28
		maxCenter := float64(screenW) * 0.72
		dialogCenter := colCenterX
		if dialogCenter < minCenter {
			dialogCenter = minCenter
		} else if dialogCenter > maxCenter {
			dialogCenter = maxCenter
		}


		queenX := int(dialogCenter - float64(screenW)*0.17)
		if queenX < int(float64(screenW)*0.08) {
			queenX = int(float64(screenW) * 0.08)
		}
		return queenX, queenY
	}


	queenX := int(float64(screenW) * 0.35)
	return queenX, queenY
}


func DetectActiveClockTurn(frame image.Image) string {
	b := frame.Bounds()
	w, h := b.Dx(), b.Dy()

	y := int(float64(h) * 0.092)
	xLeft := int(float64(w) * 0.23)
	xRight := int(float64(w) * 0.51)

	cL := frame.At(xLeft, y)
	cR := frame.At(xRight, y)

	rL, gL, bL, _ := cL.RGBA()
	rR, gR, bR, _ := cR.RGBA()

	lumL := (int(rL>>8) + int(gL>>8) + int(bL>>8)) / 3
	lumR := (int(rR>>8) + int(gR>>8) + int(bR>>8)) / 3

	if lumL > 185 && lumR < 175 && (lumL-lumR > 25) {
		return "White"
	}
	if lumR > 185 && lumL < 175 && (lumR-lumL > 25) {
		return "Black"
	}
	return ""
}


func IsSuperAdScreen(frame image.Image) bool {
	b := frame.Bounds()
	w, h := b.Dx(), b.Dy()

	yStart := int(float64(h) * 0.90)
	yEnd := int(float64(h) * 0.95)
	xStart := int(float64(w) * 0.20)
	xEnd := int(float64(w) * 0.80)

	whiteCount := 0
	blueCount := 0
	total := 0

	for y := yStart; y <= yEnd; y += 4 {
		for x := xStart; x <= xEnd; x += 8 {
			c := frame.At(x, y)
			r, g, bl, _ := c.RGBA()
			r8, g8, b8 := int(r>>8), int(g>>8), int(bl>>8)

			if r8 > 240 && g8 > 240 && b8 > 240 {
				whiteCount++
			} else if b8 > 190 && r8 < 120 && g8 < 130 {
				blueCount++
			}
			total++
		}
	}

	return total > 0 && float64(whiteCount)/float64(total) > 0.75 && float64(blueCount)/float64(total) >= 0.02
}
