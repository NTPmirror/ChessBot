package main

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
)

// Нарезка шаблонов фигур из РЕАЛЬНОГО скриншота игры с людьми
// (Figures/screen_game.png). Суффикс _p → загрузится как доп. вариант шаблона.
type Coord struct {
	Name string
	Col  int // 0..7
	Row  int // 0..7 (0 = верхняя горизонталь)
}

func main() {
	file, err := os.Open("Figures/screen_game.png")
	if err != nil {
		log.Fatal("Нет Figures/screen_game.png")
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	w := img.Bounds().Dx()
	h := img.Bounds().Dy()

	type subImager interface {
		SubImage(r image.Rectangle) image.Image
	}
	impl, ok := img.(subImager)
	if !ok {
		log.Fatal("нет SubImage")
	}

	// Доска начинается на ~0.336 высоты экрана (проверено диагностикой)
	startY := int(0.336 * float64(h))
	cellW := w / 8
	cellH := cellW

	targets := []Coord{
		// Верх: чёрные фигуры соперника
		{"bR_p", 0, 0},
		{"bN_p", 1, 0},
		{"bB_p", 2, 0},
		{"bQ_p", 3, 0},
		{"bK_p", 4, 0},
		{"bB2_p", 5, 0},
		{"bR2_p", 7, 0},
		{"bP_p", 0, 1},
		{"bN2_p", 5, 2}, // конь f6
		{"bP2_p", 7, 1}, // пешка h7
		// Низ: наши белые фигуры
		{"wR_p", 0, 7},
		{"wN_p", 1, 7},
		{"wB_p", 2, 7},
		{"wQ_p", 3, 7},
		{"wK_p", 4, 7},
		{"wB2_p", 5, 7},
		{"wN2_p", 6, 7},
		{"wR2_p", 7, 7},
		{"wP_p", 0, 6},
	}

	for _, t := range targets {
		rect := image.Rect(
			t.Col*cellW, startY+t.Row*cellH,
			(t.Col+1)*cellW, startY+(t.Row+1)*cellH,
		)
		crop := impl.SubImage(rect)

		out, err := os.Create(fmt.Sprintf("Figures/%s.png", t.Name))
		if err != nil {
			log.Printf("не создать %s: %v", t.Name, err)
			continue
		}
		if err := png.Encode(out, crop); err != nil {
			log.Printf("не сохранить %s: %v", t.Name, err)
		} else {
			fmt.Printf("OK %s.png\n", t.Name)
		}
		out.Close()
	}
}
