package duolingo

import "image"

// IsSuperAdScreen detects the Super Duolingo promotion modal ("No thanks" screen).
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
