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

			if r8 > 60 && r8 >= b8 && (r8+g8) > 130 {
				out[y*32+x] = bgFallback
			} else if r8 > 175 && g8 > 135 && b8 < 90 && (int(r8)-int(b8) > 95) {
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
	for i := 0; i < 32*32; i++ {
		cellDiff := math.Abs(cell[i] - cellBg)
		tmplDiff := math.Abs(tmpl.Grays[i] - tmpl.BgLum)
		if cellDiff > 16.0 || tmplDiff > 16.0 {
			d := cell[i] - tmpl.Grays[i]
			sum += d * d
			count++
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
	startY := float64(h) * yRatio

	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			rect := image.Rect(
				int(float64(col)*cellW),
				int(startY+float64(row)*cellH),
				int(float64(col+1)*cellW),
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
			px := int(float64(col)*cellW) + 8
			py := int(startY+float64(row)*cellH) + 8
			cSample := frame.At(px, py)
			sr, sg, sb, _ := cSample.RGBA()
			sr8, sg8, sb8 := int(sr>>8), int(sg>>8), int(sb>>8)
			if sr8 > 60 && sr8 >= sb8 && (sr8+sg8) > 130 {
				effectiveMaxErr += 350.0
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
