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

	// Новый интерфейс лобби
	return IsNewLobbyScreen(frame)
}

// IsNewLobbyScreen проверяет, открыт ли новый интерфейс лобби
func IsNewLobbyScreen(frame image.Image) bool {
	b := frame.Bounds()
	w, h := b.Dx(), b.Dy()

	cTop := frame.At(w/2, int(float64(h)*0.08))
	tr, tg, tb, _ := cTop.RGBA()
	tr8, tg8, tb8 := int(tr>>8), int(tg>>8), int(tb>>8)

	// В новом интерфейсе фон шапки темный
	if tr8 < 60 && tg8 < 60 && tb8 < 60 {
		btnY := int(float64(h) * 0.391)
		samplePoints := []int{
			int(float64(w) * 0.30),
			int(float64(w) * 0.50),
			int(float64(w) * 0.70),
		}
		tealMatches := 0
		for _, px := range samplePoints {
			c := frame.At(px, btnY)
			red, grn, blu, _ := c.RGBA()
			r, g, bVal := int(red>>8), int(grn>>8), int(blu>>8)
			if g > 120 && bVal > 70 && g > r+35 {
				tealMatches++
			}
		}
		return tealMatches >= 2
	}
	return false
}

// GetNewLobbyCardSelection возвращает индекс выбранной карточки (0, 1 или 2), либо -1
func GetNewLobbyCardSelection(frame image.Image) int {
	b := frame.Bounds()
	w, h := b.Dx(), b.Dy()

	yTest := int(float64(h) * 0.163)
	cardX := []float64{0.20, 0.50, 0.80}

	for i, xFrac := range cardX {
		px := int(float64(w) * xFrac)
		c := frame.At(px, yTest)
		red, grn, _, _ := c.RGBA()
		r, g := int(red>>8), int(grn>>8)
		// У выбранной карточки верхняя рамка подсвечена зеленым (r < 15, g > 45)
		if r < 15 && g > 45 {
			return i
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
			// Зеленая кнопка "Продолжить" (старый UI)
			currentType = "continue"
		} else if b8 > 180 && g8 > 120 && r8 < 100 && (b8 > r8+70) {
			// Синяя кнопка "Забрать XP" (старый UI)
			currentType = "claim_xp"
		} else if r8 > 230 && g8 > 230 && b8 > 230 {
			// Белая кнопка "Продолжить" (новый UI, экран завершения/опыта)
			currentType = "continue"
		} else if g8 > 140 && b8 > 90 && g8 > r8+60 && r8 < 80 {
			// Бирюзовая кнопка действия (новый UI)
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
		return bestSpan.btnType, targetX, targetY
	}

	return "", 0, 0
}

func GetPromotionQueenTap(screenW, screenH int, yRatio float64, toR, toC int) (int, int) {
	// В Duolingo диалог превращения пешки ("ПРЕВРАЩЕНИЕ ПЕШКИ") центрируется относительно колонки превращения (toC),
	// либо фиксированно расположен по центру, если колонка неизвестна.
	// Ферзь всегда является первой фигурой слева в ряду выбора (Ферзь, Ладья, Слон, Конь).
	queenY := int(float64(screenH) * 0.494)

	if toC >= 0 && toC < 8 {
		cellW := float64(screenW) / 8.0
		colCenterX := (float64(toC) + 0.5) * cellW

		// Диалог шириной ~54% экрана не выходит за границы экрана
		minCenter := float64(screenW) * 0.28
		maxCenter := float64(screenW) * 0.72
		dialogCenter := colCenterX
		if dialogCenter < minCenter {
			dialogCenter = minCenter
		} else if dialogCenter > maxCenter {
			dialogCenter = maxCenter
		}

		// Ферзь смещен на ~17% экрана левее центра диалога
		queenX := int(dialogCenter - float64(screenW)*0.17)
		if queenX < int(float64(screenW)*0.08) {
			queenX = int(float64(screenW) * 0.08)
		}
		return queenX, queenY
	}

	// По умолчанию (если колонка не определена)
	queenX := int(float64(screenW) * 0.35)
	return queenX, queenY
}
