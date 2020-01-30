package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/danicat/simpleansi"
	"golang.org/x/net/websocket"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/fatih/color"
)

type bike struct {
	Player 		[]sprite
	Direction 	string
	Winner		bool
}

type sprite struct {
	Row  int
	Col  int
	Here bool
}

type Message struct {
	PlayerA bike
	PlayerB bike
}

type inComing struct {
	Player	string
	Command string
}

var (
	maze      []string
	PlayerA   bike
	PlayerB   bike
	ServerA		bike
	ServerB		bike
	maxLength = 150
	exit      bool
	port      = flag.String("port", "9000", "port used for ws connection")
	serverIP  = "10.190.159.32"
)

type hub struct {
	clients          map[string]*websocket.Conn
	addClientChan    chan *websocket.Conn
	removeClientChan chan *websocket.Conn
	broadcastChan    chan Message
}

// run receives from the hub channels and calls the appropriate hub method
func (h *hub) run() {
	for {
		select {
		case conn := <-h.addClientChan:
			h.addClient(conn)
		case conn := <-h.removeClientChan:
			h.removeClient(conn)
		case m := <-h.broadcastChan:
			h.broadcastMessage(m)
		}
	}
}

// removeClient removes a conn from the pool
func (h *hub) removeClient(conn *websocket.Conn) {
	delete(h.clients, conn.LocalAddr().String())
}

// addClient adds a conn to the pool
func (h *hub) addClient(conn *websocket.Conn) {
	h.clients[conn.RemoteAddr().String()] = conn
}

// broadcastMessage sends a message to all client conns in the pool
func (h *hub) broadcastMessage(m Message) {
	for _, conn := range h.clients {
		err := websocket.JSON.Send(conn, m)
		if err != nil {
			fmt.Println("Error broadcasting message: ", err)
			return
		}
	}
}

// newHub returns a new hub object
func newHub() *hub {
	return &hub{
		clients:          make(map[string]*websocket.Conn),
		addClientChan:    make(chan *websocket.Conn),
		removeClientChan: make(chan *websocket.Conn),
		broadcastChan:    make(chan Message),
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

// server creates a websocket server at port <port> and registers the sole handler
func server(port string) error {
	h := newHub()
	mux := http.NewServeMux()
	mux.Handle("/", websocket.Handler(func(ws *websocket.Conn) {
		handler(ws, h)
	}))

	s := http.Server{Addr: ":" + port, Handler: mux}
	return s.ListenAndServe()
}

// handler registers a new chat client conn;
// It runs the hub, adds the client to the connection pool
// and broadcasts received message
func handler(ws *websocket.Conn, h *hub) {
	go h.run()

	h.addClientChan <- ws

	input := make(chan string)
	go func(ch chan<- string) {
		for {
			var m inComing
			err := websocket.JSON.Receive(ws, &m)
			if err != nil {
				//h.broadcastChan <- Message{"ERROR",err.Error()}
				h.removeClient(ws)
				return
			}
			ch <- m.Command
		}
	}(input)

	for {
		//TODO: Currently it doesn't matter who send the command because you still use 'wasd' v arrows
		select {
		case m := <-input:
			playerDirection(m)
		default:
		}

		var crash bool
		var winner string
		crash, winner = updateLogic(ServerA, ServerB)
		if crash {
			switch winner {
			case "Arrows":
				ServerA.Winner = true
			case "WASD":
				ServerB.Winner = true
			default:
				ServerA.Winner = false
				ServerB.Winner = false
			}
		}
		h.broadcastChan <- Message{PlayerA:ServerA, PlayerB:ServerB}
		time.Sleep(100*time.Millisecond)
	}
}

func main() {
	// Start server (required for both local & networked play)
	go func() {
		flag.Parse()
		log.Fatal(server(*port))
	}()

	// Load the Maze
	err := loadMaze("maze.txt")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

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
			fmt.Println("Enter Host IP")
			fmt.Scan(&serverIP)
			fmt.Println("Joined host: "+serverIP+". Enter any value to begin.")
			fmt.Scan(&host)
		} else {
			hostIP := GetOutboundIP()
			fmt.Println("Host IP is: ", hostIP)
			fmt.Println("Enter any value to begin.")
			fmt.Scan(&host)
		}

	default:
		fmt.Println("Invalid option. Goodbye.")
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

	updates := make(chan Message)
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
				ServerA = PlayerA
			case 'b':
				PlayerB.Player = append(PlayerB.Player, sprite{row, col, true})
				ServerB = PlayerB
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

func updateLogic(pA, pB bike) (bool, string) {
	var crash bool
	if ServerA, crash = playerMovement(pA); crash {
		return true, "WASD"
	}
	if ServerB, crash = playerMovement(pB); crash {
		return true, "Arrows"
	}

	if collisionDetection(pA, pB.Player[0]) {
		color.Red("Arrows Wins")
		return true, "Arrows"
	}
	if collisionDetection(pB, pA.Player[0]) {
		return true, "WASD"
	}

	return false, ""
}

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

// Get preferred outbound ip of this machine
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}