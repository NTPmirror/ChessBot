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
	Col  int // 0..7 (a..h)
	Row  int // 0..7 (сверху вниз: 0 = 8-я горизонталь, 7 = 1-я)
}

func main() {
	// Поддерживает и screen.png, и screen.jpg
	filePath := "Figures/screen.png"
	file, err := os.Open(filePath)
	if err != nil {
		filePath = "Figures/screen.jpg"
		file, err = os.Open(filePath)
		if err != nil {
			log.Fatalf("Не найден файл скриншота в Figures/ (screen.png или screen.jpg)")
		}
	}
	defer file.Close()

	src, _, err := image.Decode(file)
	if err != nil {
		log.Fatalf("Ошибка декодирования: %v", err)
	}

	w := src.Bounds().Dx()
	h := src.Bounds().Dy()

	// Начало доски сразу под диалоговым облачком персонажа
	startY := int(float64(h) * 0.355)
	cellW := w / 8
	cellH := cellW

	// Координаты всех 12 фигур, присутствующих на скриншоте
	targets := []PieceCoord{
		{"wR", 0, 0}, // a8 — белая ладья
		{"wK", 3, 0}, // d8 — белый король
		{"wN", 6, 0}, // g8 — белый конь
		{"wP", 0, 1}, // a7 — белая пешка
		{"wQ", 7, 3}, // h5 — белый ферзь
		{"wB", 3, 4}, // d4 — белый слон

		{"bP", 0, 6}, // a2 — черная пешка
		{"bK", 3, 6}, // d2 — черный король
		{"bR", 0, 7}, // a1 — черная ладья
		{"bN", 1, 7}, // b1 — черный конь
		{"bB", 2, 7}, // c1 — черный слон
		{"bQ", 4, 7}, // e1 — черный ферзь
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
		outPath := fmt.Sprintf("Figures/%s.png", t.Name)
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

	fmt.Println("Готово! Все шаблоны сформированы в папке Figures.")
}