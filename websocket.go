package main

import (
	"fmt"
	"golang.org/x/net/websocket"
	"log"
	"math/rand"
	"net"
	"net/http"
	"time"
)

// get preferred outbound ip of this machine
func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
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

// broadcastMessage sends a serverToClients to all client conns in the pool
func (h *hub) broadcastMessage(m serverToClients) {
	for _, conn := range h.clients {
		err := websocket.JSON.Send(conn, m)
		if err != nil {
			log.Println("Error broadcasting serverToClients: ", err)
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
		broadcastChan:    make(chan serverToClients),
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

// controls server-side game logic
func handler(ws *websocket.Conn, h *hub) {
	go h.run()

	h.addClientChan <- ws

	input := make(chan string)
	go func(ch chan<- string) {
		for {
			var m clientToServer
			err := websocket.JSON.Receive(ws, &m)
			if err != nil {
				//h.broadcastChan <- serverToClients{"ERROR",err.Error()}
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

		crash := false
		var winner string
		crash, winner = updateLogic(ServerA, ServerB)
		if crash {
			switch winner {
			case "Arrows":
				ServerA.Lives--
				gameReset()
			case "WASD":
				ServerB.Lives--
				gameReset()
			default:
			}
		}

		if ServerA.Lives <= 0 {
			ServerA.Winner = true
		} else if ServerB.Lives <= 0 {
			ServerB.Winner = true
		}

		h.broadcastChan <- serverToClients{ServerA, ServerB}
		time.Sleep(100*time.Millisecond)
	}
}

func gameReset() {
	ServerA.BikeTrail = initA
	ServerA.BikeDirection = ""

	ServerB.BikeTrail = initB
	ServerB.BikeDirection = ""
}