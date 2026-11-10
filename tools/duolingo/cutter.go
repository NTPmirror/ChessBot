package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"
)

type PieceCoord struct {
	Name string
	Col  int
	Row  int
}

func main() {

	filePath := "Figures/Duolingo/screen.png"
	file, err := os.Open(filePath)
	if err != nil {
		filePath = "Figures/Duolingo/screen.jpg"
		file, err = os.Open(filePath)
		if err != nil {
			log.Fatalf("Не найден файл скриншота в Figures/Duolingo/ (screen.png или screen.jpg)")
		}
	}
	defer file.Close()

	src, _, err := image.Decode(file)
	if err != nil {
		log.Fatalf("Ошибка декодирования: %v", err)
	}

	w := src.Bounds().Dx()
	h := src.Bounds().Dy()


	startY := int(float64(h) * 0.355)
	cellW := w / 8
	cellH := cellW


	targets := []PieceCoord{
		{"wR", 0, 0},
		{"wK", 3, 0},
		{"wN", 6, 0},
		{"wP", 0, 1},
		{"wQ", 7, 3},
		{"wB", 3, 4},

		{"bP", 0, 6},
		{"bK", 3, 6},
		{"bR", 0, 7},
		{"bN", 1, 7},
		{"bB", 2, 7},
		{"bQ", 4, 7},
	}

	type subImager interface {
		SubImage(r image.Rectangle) image.Image
	}

	si, ok := src.(subImager)
	if !ok {
		log.Fatal("Изображение не поддерживает SubImage")
	}

	for _, t := range targets {
		rect := image.Rect(
			t.Col*cellW,
			startY+t.Row*cellH,
			(t.Col+1)*cellW,
			startY+(t.Row+1)*cellH,
		)

		crop := si.SubImage(rect)
		outPath := fmt.Sprintf("Figures/Duolingo/%s.png", t.Name)
		outF, err := os.Create(outPath)
		if err != nil {
			log.Printf("Ошибка создания %s: %v", outPath, err)
			continue
		}

		if err := png.Encode(outF, crop); err != nil {
			log.Printf("Ошибка сохранения %s: %v", outPath, err)
		} else {
			fmt.Printf("✓ Сохранено: %s\n", outPath)
		}
		outF.Close()
	}

	fmt.Println("Готово! Все шаблоны сформированы в папке Figures/Duolingo.")
}
