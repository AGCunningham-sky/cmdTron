package main

import (
	"bufio"
	"fmt"
	"github.com/danicat/simpleansi"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/fatih/color"
)

type player struct {
	Player []sprite
	Direction string
	Lives int
}

type sprite struct {
	row 	int
	col 	int
	here 	bool
}

var maze []string
var PlayerA player
var PlayerB player
var maxLength = 150

func main() {
	initialise()
	defer cleanup()

	PlayerA.Lives = 1
	PlayerB.Lives = 1

	err := loadMaze("maze.txt")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	input := make(chan string)
	go func(ch chan<- string) {
		for {
			input, err := readInput()
			if err != nil {
				log.Println("error reading input:", err)
				ch <- "ESC"
			}
			ch <- input
		}
	}(input)

	for {
		printScreen()
		var crash bool
		PlayerA, crash = playerMovement(PlayerA)
		if crash {
			PlayerA.Lives--
			color.Blue("WASD Wins")
			break
		}
		PlayerB, crash = playerMovement(PlayerB)
		if crash {
			PlayerB.Lives--
			color.Red("Arrows Wins")
			break
		}

		crash = collisionDetection(PlayerA, PlayerB.Player[0])
		if crash {
			PlayerB.Lives--
			color.Red("Arrows Wins")
			break
		}
		crash = collisionDetection(PlayerB, PlayerA.Player[0])
		if crash {
			PlayerA.Lives--
			color.Blue("WASD Wins")
			break
		}

		select {
		case inp := <-input:
			if inp == "ESC" {
				PlayerA.Lives = 0
			}
			PlayerA, PlayerB = snakeMovement(PlayerA, PlayerB, inp)
		default:
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func collisionDetection(user player, opp sprite) bool {
	for ind, seg := range user.Player {
		if ind != 0 {
			if seg == opp {
				return true
			}
		}
	}

	return false
}

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

func snakeMovement(user1, user2 player, input string) (player, player) {
	switch input {
	case "UP":
		user1.Direction = "UP"
	case "DOWN":
		user1.Direction = "DOWN"
	case "RIGHT":
		user1.Direction = "LEFT"
	case "LEFT":
		user1.Direction = "RIGHT"
	case "w":
		user2.Direction = "UP"
	case "s":
		user2.Direction = "DOWN"
	case "d":
		user2.Direction = "LEFT"
	case "a":
		user2.Direction = "RIGHT"
	}

	return user1, user2
}

func playerMovement(Player player) (player, bool) {
	if Player.Direction != "" {
		if Player.Direction == "UP" {
			newRow := sprite{Player.Player[0].row - 1, Player.Player[0].col, true}
			Player.Player = append([]sprite{newRow}, Player.Player...)
		}
		if Player.Direction == "DOWN" {
			newRow := sprite{Player.Player[0].row + 1, Player.Player[0].col, true}
			Player.Player = append([]sprite{newRow}, Player.Player...)
		}
		if Player.Direction == "LEFT" {
			newRow := sprite{Player.Player[0].row, Player.Player[0].col + 1, true}
			Player.Player = append([]sprite{newRow}, Player.Player...)
		}
		if Player.Direction == "RIGHT" {
			newRow := sprite{Player.Player[0].row, Player.Player[0].col - 1, true}
			Player.Player = append([]sprite{newRow}, Player.Player...)
		}

		if len(Player.Player) > maxLength {
			Player.Player = Player.Player[:len(Player.Player)-1]
		}

		if Player.Player[0].row >= len(maze)-2 {
			Player.Player[0].row = 1
		} else if Player.Player[0].row <= 0 {
			Player.Player[0].row = len(maze) - 2
		}
		if Player.Player[0].col > len(maze[0])-1 {
			Player.Player[0].col = 0
		} else if Player.Player[0].col < 0 {
			Player.Player[0].col = len(maze[0])
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
		simpleansi.MoveCursor(segment.row, segment.col)
		color.Red("a")
	}
	for _,segment := range PlayerB.Player {
		simpleansi.MoveCursor(segment.row, segment.col)
		color.Blue("b")
	}

	simpleansi.MoveCursor(len(maze), 0)
}

func loadMaze(file string) error {
	f, err := os.Open(file)
	if err != nil{
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		maze = append(maze, scanner.Text())
	}

	for row, line := range maze {
		for col, chr := range line {
			switch chr {
			case 'a':
				PlayerA.Player = append(PlayerA.Player, sprite{row, col, true})
			case 'b':
				PlayerB.Player = append(PlayerB.Player, sprite{row, col, true})
			}
		}
	}

	return nil
}

func initialise() {
	cbTerm := exec.Command("stty", "cbreak", "-echo")
	cbTerm.Stdin = os.Stdin

	err := cbTerm.Run()
	if err != nil {
		log.Fatalln("unable to activate cbreak mode:", err)
	}
}

func cleanup() {
	cookedTerm := exec.Command("stty", "-cbreak", "echo")
	cookedTerm.Stdin = os.Stdin

	err := cookedTerm.Run()
	if err != nil {
		log.Fatalln("unable to restore cooked mode:", err)
	}
}