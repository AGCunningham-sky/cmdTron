package main

import (
	"fmt"
	"github.com/danicat/simpleansi"
	"github.com/fatih/color"
	"os"
)

// controls movement updates and collision detect for all players
func updateLogic(pA, pB bike) (bool, string) {
	var crash bool
	if ServerA, crash = playerMovement(pA); crash {
		return true, "WASD"
	}
	if ServerB, crash = playerMovement(pB); crash {
		return true, "Arrows"
	}

	if collisionDetection(pA, pB.Player[0]) {
		return true, "Arrows"
	}
	if collisionDetection(pB, pA.Player[0]) {
		return true, "WASD"
	}

	return false, ""
}

// determine collisions per player
func collisionDetection(user bike, opp sprite) bool {
	for ind, seg := range user.Player {
		if ind != 0 {
			if seg == opp {
				return true
			}
		}
	}

	return false
}

// updates player movement per player
func playerMovement(Player bike) (bike, bool) {
	if Player.Direction != "" {
		var newRow sprite

		switch Player.Direction {
		case "UP": newRow = sprite{Player.Player[0].Row - 1, Player.Player[0].Col, true}
		case "DOWN": newRow = sprite{Player.Player[0].Row + 1, Player.Player[0].Col, true}
		case "LEFT": newRow = sprite{Player.Player[0].Row, Player.Player[0].Col + 1, true}
		case "RIGHT": newRow = sprite{Player.Player[0].Row, Player.Player[0].Col - 1, true}
		}
		if newRow.Here != false {
			Player.Player = append([]sprite{newRow}, Player.Player...)
		}

		if len(Player.Player) > maxLength {
			Player.Player = Player.Player[:len(Player.Player)-1]
		}

		if Player.Player[0].Row >= len(maze)-1 {
			Player.Player[0].Row = 1
		} else if Player.Player[0].Row <= 0 {
			Player.Player[0].Row = len(maze)-2
		}
		if Player.Player[0].Col > len(maze[0])-1 {
			Player.Player[0].Col = 0
		} else if Player.Player[0].Col < 0 {
			Player.Player[0].Col = len(maze[0])
		}

		for ind, seg := range Player.Player {
			if ind != 0 {
				if Player.Player[0] == seg {
					return Player, true
				}
			}
		}
	}
	return Player, false
}

// updates player direction based on input
func playerDirection(input string) {
	switch input {
	case "UP":
		ServerA.Direction = "UP"
	case "DOWN":
		ServerA.Direction = "DOWN"
	case "RIGHT":
		ServerA.Direction = "LEFT"
	case "LEFT":
		ServerA.Direction = "RIGHT"
	case "w":
		ServerB.Direction = "UP"
	case "s":
		ServerB.Direction = "DOWN"
	case "d":
		ServerB.Direction = "LEFT"
	case "a":
		ServerB.Direction = "RIGHT"
	}
}

// read input from keyboard
func readInput() (string, error) {
	buffer := make([]byte, 100)

	cnt, err := os.Stdin.Read(buffer)
	if err != nil {
		return "", err
	}

	if cnt == 1 {
		if buffer[0] == 0x1b {
			return "ESC", nil
		} else {
			switch buffer[0] {
			case 'w':
				return "w", nil
			case 'a':
				return "a", nil
			case 's':
				return "s", nil
			case 'd':
				return "d", nil
			}
		}
	} else if cnt >= 3 {
		if buffer[0] == 0x1b && buffer[1] == '[' {
			switch buffer[2] {
			case 'A':
				return "UP", nil
			case 'B':
				return "DOWN", nil
			case 'C':
				return "RIGHT", nil
			case 'D':
				return "LEFT", nil
			}
		}
	}

	return "", nil
}

// re-print the screen with updated positions
func printScreen() {
	simpleansi.ClearScreen()
	for _, line := range maze {
		for _, chr := range line {
			switch chr {
			case '-':
				fmt.Print("#")
			default:
				fmt.Print(" ")
			}
		}
		fmt.Println()
	}

	for _,segment := range PlayerA.Player {
		simpleansi.MoveCursor(segment.Row, segment.Col)
		color.Red("a")
	}
	for _,segment := range PlayerB.Player {
		simpleansi.MoveCursor(segment.Row, segment.Col)
		color.Blue("b")
	}

	simpleansi.MoveCursor(len(maze), 0)
}