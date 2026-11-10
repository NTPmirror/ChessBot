package duolingo

import "image"

// IsLobbyScreen checks if the current screen is a Duolingo chess lobby (either classic or new UI).
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

// IsNewLobbyScreen detects the modern Duolingo chess card-selection screen.
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

// GetNewLobbyCardSelection returns the index (0: Oscar, 1: Lily, 2: People) of the selected mode.
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
