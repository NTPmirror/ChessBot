package recognizer

import (
	"image"
	"image/color"
	"math"
)

func To32x32Grayscale(src image.Image) [32 * 32]float64 {
	var out [32 * 32]float64
	b := src.Bounds()
	stepX := float64(b.Dx()) / 32.0
	stepY := float64(b.Dy()) / 32.0

	cornerColor := color.GrayModel.Convert(src.At(b.Min.X+1, b.Min.Y+1)).(color.Gray)
	bgFallback := float64(cornerColor.Y)

	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			srcX := b.Min.X + int((float64(x)+0.5)*stepX)
			srcY := b.Min.Y + int((float64(y)+0.5)*stepY)

			pixel := src.At(srcX, srcY)
			red, green, blue, _ := pixel.RGBA()
			r8, g8, b8 := red>>8, green>>8, blue>>8

			// Yellow highlight filter (both Duolingo and Chess.com)
			// Highlight is saturated yellow (R > 150, G > 140, and R - B > 70)
			if r8 > 150 && g8 > 140 && int(r8)-int(b8) > 70 {
				out[y*32+x] = bgFallback
			} else {
				c := color.GrayModel.Convert(pixel).(color.Gray)
				out[y*32+x] = float64(c.Y)
			}
		}
	}
	return out
}

func IsSquareEmpty(cell [32 * 32]float64, bgLum float64) bool {
	fgPixels := 0
	for y := 6; y < 26; y++ {
		for x := 6; x < 26; x++ {
			if math.Abs(cell[y*32+x]-bgLum) > 20.0 {
				fgPixels++
			}
		}
	}
	return fgPixels < 35
}

func ComparePieceMask(cell [32 * 32]float64, cellBg float64, tmpl PieceTemplate) float64 {
	var sum, count float64
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			// Mask out top-right corner (rank numbers in col h) and bottom-left corner (file letters in rank 1)
			if (y < 8 && x > 23) || (y > 23 && x < 8) {
				continue
			}
			i := y*32 + x
			cellDiff := math.Abs(cell[i] - cellBg)
		tmplDiff := math.Abs(tmpl.Grays[i] - tmpl.BgLum)
		if cellDiff > 16.0 || tmplDiff > 16.0 {
				d := cell[i] - tmpl.Grays[i]
				sum += d * d
				count++
			}
		}
	}
	if count < 30 {
		return math.MaxFloat64
	}
	return sum / count
}

func subImage(src image.Image, rect image.Rectangle) image.Image {
	type subImager interface {
		SubImage(r image.Rectangle) image.Image
	}
	if si, ok := src.(subImager); ok {
		return si.SubImage(rect)
	}
	return nil
}

func Process(frame image.Image, templates []PieceTemplate, yRatio float64, maxError float64) [8][8]string {
	var board [8][8]string
	w := frame.Bounds().Dx()
	h := frame.Bounds().Dy()

	cellW := float64(w) / 8.0
	cellH := cellW
	startX := 0.0
	startY := float64(h) * yRatio

	if yRatio < 0.25 && w >= 1000 {
		cellW = float64(w) * 0.985 / 8.0
		cellH = cellW
		startX = float64(w) * 0.0074
	}

	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			rect := image.Rect(
				int(startX+float64(col)*cellW),
				int(startY+float64(row)*cellH),
				int(startX+float64(col+1)*cellW),
				int(startY+float64(row+1)*cellH),
			)

			cellCrop := subImage(frame, rect)
			cellGrays := To32x32Grayscale(cellCrop)
			cellBg := CalcCornerBg(cellGrays)

			if IsSquareEmpty(cellGrays, cellBg) {
				board[row][col] = "."
				continue
			}

			bestBase := "."
			minDiff := math.MaxFloat64

			for _, tmpl := range templates {
				diff := ComparePieceMask(cellGrays, cellBg, tmpl)
				if diff < minDiff {
					minDiff = diff
					bestBase = tmpl.BasePiece
				}
			}

			effectiveMaxErr := maxError
			offsets := [][2]int{
				{8, 8},
				{int(cellW) - 8, 8},
				{8, int(cellH) - 8},
				{int(cellW) - 8, int(cellH) - 8},
			}
			for _, off := range offsets {
				px := int(startX+float64(col)*cellW) + off[0]
				py := int(startY+float64(row)*cellH) + off[1]
				cSample := frame.At(px, py)
				sr, sg, sb, _ := cSample.RGBA()
				sr8, sg8, sb8 := int(sr>>8), int(sg>>8), int(sb>>8)
				if (sr8 > 60 && sr8 >= sb8 && (sr8+sg8) > 130) || (sr8 > 175 && sg8 > 135 && sb8 < 90 && (sr8-sb8 > 95)) {
					effectiveMaxErr += 380.0
					break
				}
			}

			if minDiff > effectiveMaxErr {
				board[row][col] = "."
			} else {
				board[row][col] = bestBase
			}
		}
	}
	return board
}
