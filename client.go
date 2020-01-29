package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/danicat/simpleansi"
	"golang.org/x/net/websocket"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"github.com/fatih/color"
)

type bike struct {
	Player []sprite
	Direction string
}

type sprite struct {
	Row  int
	Col  int
	Here bool
}

type Message struct {
	PlayerA bike `json:"playerA"`
	PlayerB bike `json:"playerB"`
}

type inComing struct {
	Player	string `json:"bike"`
	Command string	`json:"command"`
}

var (
	maze      []string
	PlayerA   bike
	PlayerB   bike
	maxLength = 150
	exit      bool
	port      = flag.String("port", "9000", "port used for ws connection")
	serverIP  = "10.190.159.32"
)

func main() {
	initialise()
	defer cleanup()

	flag.Parse()

	// connect
	ws, err := connect()
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	exit = false

	err = loadMaze("maze.txt")
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

	updates := make(chan Message)
	// receive
	go func(ch chan<- Message) {
		for {
			var m Message
			err := websocket.JSON.Receive(ws, &m)
			if err != nil {
				fmt.Println("Error receiving message: ", err.Error())
				break
			}
			ch <- m
		}
	}(updates)

	for {
		printScreen()

		select {
		case inp := <-input:
			if inp == "ESC" {
				color.Cyan("Game exited")
				exit = true
			}
			m := inComing{
				Player: "A",
				Command: inp,
			}
			err = websocket.JSON.Send(ws, m)
			if err != nil {
				fmt.Println("Error sending message: ", err.Error())
				break
			}
			case m:= <- updates:
				PlayerA = m.PlayerA
				PlayerB = m.PlayerB
		default:
		}
		if exit {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// connect connects to the local chat server at port <port>
func connect() (*websocket.Conn, error) {
	return websocket.Dial(fmt.Sprintf("ws://"+serverIP+":%s", *port), "", mockedIP())
}

// mockedIP is a demo-only utility that generates a random IP address for this client
func mockedIP() string {
	var arr [4]int
	for i := 0; i < 4; i++ {
		rand.Seed(time.Now().UnixNano())
		arr[i] = rand.Intn(256)
	}
	return fmt.Sprintf("http://%d.%d.%d.%d", arr[0], arr[1], arr[2], arr[3])
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
		simpleansi.MoveCursor(segment.Row, segment.Col)
		color.Red("a")
	}
	for _,segment := range PlayerB.Player {
		simpleansi.MoveCursor(segment.Row, segment.Col)
		color.Blue("b")
	}

	simpleansi.MoveCursor(len(maze), 0)
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