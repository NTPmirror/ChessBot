package duolingo

import "image"

// DetectActiveClockTurn checks the header clocks to identify who has the active move timer.
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

// GetPromotionQueenTap calculates the exact tap position for the Queen selection in the promotion dialog.
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
