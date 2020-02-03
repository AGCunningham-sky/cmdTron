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

	if collisionDetection(pA, pB.BikeTrail[0]) {
		return true, "Arrows"
	}
	if collisionDetection(pB, pA.BikeTrail[0]) {
		return true, "WASD"
	}

	return false, ""
}

// determine collisions per player
func collisionDetection(user bike, opp sprite) bool {
	for ind, seg := range user.BikeTrail {
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
	if Player.BikeDirection != "" {
		var newRow sprite

		x, y := Player.BikeTrail[0].Col, Player.BikeTrail[0].Row
		switch Player.BikeDirection {
			case "UP": y--
			case "DOWN": y++
			case "LEFT": x--
			case "RIGHT": x++
		}
		if x != Player.BikeTrail[0].Col || y != Player.BikeTrail[0].Row  {
			newRow = sprite{y, x, true}
			Player.BikeTrail = append([]sprite{newRow}, Player.BikeTrail...)
		}

		if len(Player.BikeTrail) > maxLength {
			Player.BikeTrail = Player.BikeTrail[:len(Player.BikeTrail)-1]
		}

		if Player.BikeTrail[0].Row >= len(maze)-1 {
			Player.BikeTrail[0].Row = 1
		} else if Player.BikeTrail[0].Row <= 0 {
			Player.BikeTrail[0].Row = len(maze)-2
		}
		if Player.BikeTrail[0].Col > len(maze[0])-1 {
			Player.BikeTrail[0].Col = 0
		} else if Player.BikeTrail[0].Col < 0 {
			Player.BikeTrail[0].Col = len(maze[0])
		}

		for ind, seg := range Player.BikeTrail {
			if ind != 0 {
				if Player.BikeTrail[0] == seg {
					return Player, true
				}
			}
		}
	}
	return Player, false
}

// updates player direction based on input
func playerDirection(input clientToServer) {
	if input.Player == "host" {
		ServerA = dirSet(ServerA, input.Command)
	} else if input.Player == "slave" {
		ServerB = dirSet(ServerB, input.Command)
	} else {
		switch input.Command {
		case "UP":
			ServerA.BikeDirection = "UP"
		case "DOWN":
			ServerA.BikeDirection = "DOWN"
		case "RIGHT":
			ServerA.BikeDirection = "RIGHT"
		case "LEFT":
			ServerA.BikeDirection = "LEFT"
		case "w":
			ServerB.BikeDirection = "UP"
		case "s":
			ServerB.BikeDirection = "DOWN"
		case "d":
			ServerB.BikeDirection = "RIGHT"
		case "a":
			ServerB.BikeDirection = "LEFT"
		}
	}
}

func dirSet(input bike, command string) bike {
	output := input
	switch command {
		case "UP", "w":
			input.BikeDirection = "UP"
		case "DOWN", "s":
			input.BikeDirection = "DOWN"
		case "RIGHT", "d":
			input.BikeDirection = "RIGHT"
		case "LEFT", "a":
			input.BikeDirection = "LEFT"
	}
	return output
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

	for _,segment := range PlayerA.BikeTrail {
		simpleansi.MoveCursor(segment.Row, segment.Col)
		color.Red("a")
	}
	for _,segment := range PlayerB.BikeTrail {
		simpleansi.MoveCursor(segment.Row, segment.Col)
		color.Blue("b")
	}

	simpleansi.MoveCursor(len(maze), 0)
	fmt.Printf("Arrows: %d \tWASD: %d\n", PlayerA.Lives, PlayerB.Lives)
	simpleansi.MoveCursor(len(maze)+1, 0)
}