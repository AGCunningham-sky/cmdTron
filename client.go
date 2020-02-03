// --- Inspiration Repos
// https://github.com/stinkyfingers/chat/blob/master/readme.md
// https://github.com/danicat/pacgo

package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"golang.org/x/net/websocket"
	"log"
	"os"
	"os/exec"
)

func main() {
	//TODO: Make it so more players adding can dynamically increase the number of bikes

	// Start server (required for both local & networked play)
	go func() {
		// Load the Maze
		err := loadMaze(mazePath)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		initA = ServerA.BikeTrail
		initB = ServerB.BikeTrail

		flag.Parse()
		log.Fatal(server(*port))
	}()

	var mode string
	fmt.Println("-- Welcome to Tron --")
	fmt.Println("Enter '1' for local play or '2' for networked play")
	fmt.Scan(&mode)

	switch mode {
	case "1":
	case "2":
		fmt.Println("Press '1' to host or any key to join")
		var host string
		fmt.Scan(&host)
		if host != "1" {
			dirControl = "host"
			fmt.Println("Enter Host IP")
			fmt.Scan(&serverIP)
			fmt.Println("Joined host: "+serverIP+". Enter any value to begin.")
			fmt.Scan(&host)
		} else {
			dirControl = "slave"
			serverIP = getOutboundIP().String()
			fmt.Println("Host IP is: ", serverIP)
			fmt.Println("Enter any value to begin.")
			fmt.Scan(&host)
		}
	default:
		fmt.Println("Invalid option. Goodbye.")
		cleanup()
		os.Exit(0)
	}

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

	updates := make(chan serverToClients)
	go func(ch chan<- serverToClients) {
		for {
			var m serverToClients
			err := websocket.JSON.Receive(ws, &m)
			if err != nil {
				fmt.Println("Error receiving serverToClients: ", err.Error())
				break
			}
			ch <- m
		}
	}(updates)

	for {
		select {
		case inp := <-input:
			if inp == "ESC" {
				color.Cyan("Game exited")
				exit = true
			}
			m := clientToServer{
				Player: dirControl,
				Command: inp,
			}
			err = websocket.JSON.Send(ws, m)
			if err != nil {
				fmt.Println("Error sending serverToClients: ", err.Error())
				break
			}
		case m:= <- updates:
			PlayerA = m.ServA
			PlayerB = m.ServB

			// Only redraw the screen when the data is updated
			printScreen()
		default:
		}

		if PlayerA.Winner {
			color.Red("Arrows Win!")
			break
		} else if PlayerB.Winner {
			color.Blue("WASD Win!")
			break
		} else if exit {
			break
		}
	}
}

// load the maze from file
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
				PlayerA.BikeTrail = append(PlayerA.BikeTrail, sprite{row, col, true})
				ServerA = PlayerA
			case 'b':
				PlayerB.BikeTrail = append(PlayerB.BikeTrail, sprite{row, col, true})
				ServerB = PlayerB
			}
		}
	}

	ServerA.Lives = startLives
	ServerB.Lives = startLives

	return nil
}

// configure the terminal to not echo and enable cbreak mode
func initialise() {
	cbTerm := exec.Command("stty", "cbreak", "-echo")
	cbTerm.Stdin = os.Stdin

	err := cbTerm.Run()
	if err != nil {
		log.Fatalln("unable to activate cbreak mode:", err)
	}
}

// reverse initialise to return terminal to expected use
func cleanup() {
	cookedTerm := exec.Command("stty", "-cbreak", "echo")
	cookedTerm.Stdin = os.Stdin

	err := cookedTerm.Run()
	if err != nil {
		log.Fatalln("unable to restore cooked mode:", err)
	}
}