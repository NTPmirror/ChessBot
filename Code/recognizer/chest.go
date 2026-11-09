package recognizer

import "image"


func IsChestScreen(frame image.Image) bool {
	b := frame.Bounds()
	w, h := b.Dx(), b.Dy()

	samplePoints := [][2]int{
		{int(float64(w) * 0.50), int(float64(h) * 0.20)},
		{int(float64(w) * 0.50), int(float64(h) * 0.80)},
		{int(float64(w) * 0.10), int(float64(h) * 0.50)},
		{int(float64(w) * 0.90), int(float64(h) * 0.50)},
	}

	orangeMatches := 0
	for _, pt := range samplePoints {
		c := frame.At(pt[0], pt[1])
		red, grn, blu, _ := c.RGBA()
		r8, g8, b8 := int(red>>8), int(grn>>8), int(blu>>8)


		if r8 > 210 && g8 >= 100 && g8 <= 190 && b8 < 60 {
			orangeMatches++
		}
	}

	return orangeMatches >= 3
}


func GetChestTapCoordinates(screenW, screenH int) (int, int) {

	chestX := screenW / 2
	chestY := int(float64(screenH) * 0.41)
	return chestX, chestY
}
