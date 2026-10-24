package engine

import "fmt"

type Move struct {
	FromR, FromC int
	ToR, ToC     int
	Piece        string
}

func SquareName(r, c int) string {
	file := string(rune('a' + c))
	rank := fmt.Sprintf("%d", 8-r)
	return file + rank
}

func (m Move) String() string {
	if m.Piece == "" {
		return "-"
	}
	return fmt.Sprintf("%s (%s -> %s)", m.Piece, SquareName(m.FromR, m.FromC), SquareName(m.ToR, m.ToC))
}

var PieceValues = map[byte]int{
	'P': 100,
	'N': 320,
	'B': 330,
	'R': 500,
	'Q': 900,
	'K': 20000,
}
