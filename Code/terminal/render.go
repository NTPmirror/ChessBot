package terminal

import "fmt"

func RenderTerminal(board [8][8]string, state int, move, myTeam, bestMove string, autoPlay, isGameOver bool, autoClaim, autoStart bool, gameType string) {
	fmt.Print("\033[H\033[2J")
	fmt.Println("       a    b    c    d    e    f    g    h")
	fmt.Println("    +----+----+----+----+----+----+----+----+")
	for r := 0; r < 8; r++ {
		fmt.Printf("  %d |", 8-r)
		for c := 0; c < 8; c++ {
			val := board[r][c]
			switch len(val) {
			case 1:
				fmt.Printf("  %s |", val)
			case 2:
				fmt.Printf(" %s |", val)
			default:
				fmt.Printf("%4.4s|", val)
			}
		}
		fmt.Printf(" %d\n", 8-r)
		fmt.Println("    +----+----+----+----+----+----+----+----+")
	}
	fmt.Println("       a    b    c    d    e    f    g    h")
	fmt.Println()
	fmt.Printf("  Move:    %s\n", move)
	fmt.Printf("  My Team: %s\n", myTeam)
}
