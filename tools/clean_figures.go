package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
)

// Flood fill or corner sampling to remove background from piece crops
func removeBackground(src image.Image) *image.NRGBA {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	out := image.NewNRGBA(image.Rect(0, 0, w, h))

	// Find background color from corners
	corners := []color.Color{
		src.At(b.Min.X, b.Min.Y),
		src.At(b.Max.X-1, b.Min.Y),
		src.At(b.Min.X, b.Max.Y-1),
		src.At(b.Max.X-1, b.Max.Y-1),
		src.At(b.Min.X+2, b.Min.Y+2),
		src.At(b.Max.X-3, b.Min.Y+2),
		src.At(b.Min.X+2, b.Max.Y-3),
		src.At(b.Max.X-3, b.Max.Y-3),
	}

	var sumR, sumG, sumB float64
	for _, c := range corners {
		r, g, b, _ := c.RGBA()
		sumR += float64(r >> 8)
		sumG += float64(g >> 8)
		sumB += float64(b >> 8)
	}
	bgR := sumR / float64(len(corners))
	bgG := sumG / float64(len(corners))
	bgB := sumB / float64(len(corners))

	// Breadth-first flood fill from edges to only remove the external background,
	// preserving any inside parts of pieces that might resemble the board color
	visited := make([][]bool, h)
	for y := range visited {
		visited[y] = make([]bool, w)
	}

	isBgColor := func(r, g, b uint8) bool {
		dr := float64(r) - bgR
		dg := float64(g) - bgG
		db := float64(b) - bgB
		dist := math.Sqrt(dr*dr + dg*dg + db*db)

		// Check if it's the board cell color (dark #17242a or light #273941 or similar)
		// Or within distance of corners
		if dist < 22.0 {
			return true
		}
		// Also check standard board cell tones
		// dark: ~ (23, 36, 42), light: ~ (39, 57, 65)
		dDark := math.Sqrt(math.Pow(float64(r)-23, 2) + math.Pow(float64(g)-36, 2) + math.Pow(float64(b)-42, 2))
		dLight := math.Sqrt(math.Pow(float64(r)-39, 2) + math.Pow(float64(g)-57, 2) + math.Pow(float64(b)-65, 2))
		return dDark < 16.0 || dLight < 16.0
	}

	type pt struct{ x, y int }
	var queue []pt

	// Initialize queue with border pixels
	for x := 0; x < w; x++ {
		queue = append(queue, pt{x, 0}, pt{x, h - 1})
		visited[0][x] = true
		visited[h-1][x] = true
	}
	for y := 1; y < h-1; y++ {
		queue = append(queue, pt{0, y}, pt{w - 1, y})
		visited[y][0] = true
		visited[y][w-1] = true
	}

	isBgPixel := make([][]bool, h)
	for y := range isBgPixel {
		isBgPixel[y] = make([]bool, w)
	}

	head := 0
	for head < len(queue) {
		p := queue[head]
		head++

		col := src.At(b.Min.X+p.x, b.Min.Y+p.y)
		r, g, bl, _ := col.RGBA()
		r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(bl>>8)

		if isBgColor(r8, g8, b8) {
			isBgPixel[p.y][p.x] = true

			// Neighbors
			dxs := []int{-1, 1, 0, 0}
			dys := []int{0, 0, -1, 1}
			for i := 0; i < 4; i++ {
				nx, ny := p.x+dxs[i], p.y+dys[i]
				if nx >= 0 && nx < w && ny >= 0 && ny < h && !visited[ny][nx] {
					visited[ny][nx] = true
					queue = append(queue, pt{nx, ny})
				}
			}
		}
	}

	// Copy to out with transparency
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if isBgPixel[y][x] {
				out.SetNRGBA(x, y, color.NRGBA{0, 0, 0, 0})
			} else {
				col := src.At(b.Min.X+x, b.Min.Y+y)
				r, g, bl, a := col.RGBA()
				out.SetNRGBA(x, y, color.NRGBA{uint8(r >> 8), uint8(g >> 8), uint8(bl >> 8), uint8(a >> 8)})
			}
		}
	}

	return out
}

func main() {
	inDir := "Figures"
	outDir := "Figures/WEB"

	entries, err := os.ReadDir(inDir)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		return
	}

	count := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".png") {
			continue
		}
		if strings.HasPrefix(strings.ToLower(e.Name()), "screen") {
			continue
		}

		inPath := filepath.Join(inDir, e.Name())
		f, err := os.Open(inPath)
		if err != nil {
			continue
		}

		img, _, err := image.Decode(f)
		f.Close()
		if err != nil {
			continue
		}

		cleanImg := removeBackground(img)

		outPath := filepath.Join(outDir, e.Name())
		outF, err := os.Create(outPath)
		if err != nil {
			fmt.Printf("Create err %s: %v\n", outPath, err)
			continue
		}

		if err := png.Encode(outF, cleanImg); err != nil {
			fmt.Printf("Encode err %s: %v\n", outPath, err)
		} else {
			count++
			fmt.Printf("✓ Очищен: %s\n", e.Name())
		}
		outF.Close()
	}

	fmt.Printf("\nУспешно очищено и сохранено в %s: %d файлов\n", outDir, count)
}
