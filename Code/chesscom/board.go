package chesscom

import (
	"image"
)

// BoardBounds represents the pixel bounding box of the chessboard on screen.
type BoardBounds struct {
	X      int
	Y      int
	Width  int
	Height int
	CellW  float64
	CellH  float64
}



// DetectBoard scans the screen frame to find the exact chessboard bounds.
func DetectBoard(frame image.Image) (BoardBounds, bool) {
	bounds := frame.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// Default fallback for 1080x2400 (mobile web chess.com):
	// Start at y=154, width 1064, cell 133
	defaultBounds := BoardBounds{
		X:      int(float64(w) * 0.0074), // ~8px on 1080w
		Y:      int(float64(h) * 0.0642), // ~154px on 2400h
		Width:  int(float64(w) * 0.9852), // ~1064px on 1080w
		Height: int(float64(w) * 0.9852),
	}
	defaultBounds.CellW = float64(defaultBounds.Width) / 8.0
	defaultBounds.CellH = float64(defaultBounds.Height) / 8.0

	// Verify by sampling candidate squares across the board
	matches := 0
	for r := 2; r <= 5; r++ {
		for c := 0; c < 8; c++ {
			sampleX := defaultBounds.X + int((float64(c)+0.5)*defaultBounds.CellW)
			sampleY := defaultBounds.Y + int((float64(r)+0.5)*defaultBounds.CellH)

			px := frame.At(sampleX, sampleY)
			red, grn, blu, _ := px.RGBA()
			if isBoardSquareColor(int(red>>8), int(grn>>8), int(blu>>8)) {
				matches++
			}
		}
	}

	if matches >= 12 {
		return defaultBounds, true
	}

	return defaultBounds, false
}
