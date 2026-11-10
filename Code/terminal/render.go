package terminal

import "fmt"

func RenderTerminal(board [8][8]string, state int, move, myTeam, bestMove string, autoPlay, isGameOver bool, autoClaim, autoStart bool, gameType string) {
	fmt.Print("\033[H\033[2J")

	// Determine file and rank headers based on orientation:
	// When playing Black, bottom is rank 8, top is rank 1, file h..a
	// But in standard 2D array representation:
	filesHeader := "       a    b    c    d    e    f    g    h"
	if myTeam == "Black" {
		filesHeader = "       h    g    f    e    d    c    b    a"
	}

	fmt.Println(filesHeader)
	fmt.Println("    +----+----+----+----+----+----+----+----+")
	for r := 0; r < 8; r++ {
		rankNum := 8 - r
		if myTeam == "Black" {
			rankNum = r + 1
		}
		fmt.Printf("  %d |", rankNum)
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
		fmt.Printf(" %d\n", rankNum)
		fmt.Println("    +----+----+----+----+----+----+----+----+")
	}
	fmt.Println(filesHeader)
	fmt.Println()
	fmt.Printf("  Move:    %s\n", move)
	fmt.Printf("  My Team: %s\n", myTeam)
	fmt.Printf("  Best:    %s\n", bestMove)
}
