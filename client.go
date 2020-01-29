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
}

type sprite struct {
	row 	int
	col 	int
	here 	bool
}

var (
	maze []string
	PlayerA player
	PlayerB player
	maxLength = 150
	exit bool
)

func main() {
	initialise()
	defer cleanup()

	exit = false

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

		select {
		case inp := <-input:
			if inp == "ESC" {
				color.Cyan("Game exited")
				exit = true
			}
			// Send move
		default:
		}
		if exit {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
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