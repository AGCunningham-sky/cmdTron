package main

import (
	"flag"
	"golang.org/x/net/websocket"
)

var (
	maze      []string
	PlayerA   bike
	PlayerB   bike
	ServerA		bike
	ServerB		bike
	maxLength = 150
	exit      bool
	port      = flag.String("port", "9000", "port used for ws connection")
	serverIP string
	mazePath = "maze.txt"
	startLives = 3
)

type bike struct {
	Player 		[]sprite
	Direction 	string
	Winner		bool
	Lives		int
}

type sprite struct {
	Row  int
	Col  int
	Here bool
}

type serverToClients struct {
	//TODO: Figure out why I couldn't change this
	PlayerA bike
	PlayerB bike
}

type clientToServer struct {
	Player	string
	Command string
}

type hub struct {
	clients          map[string]*websocket.Conn
	addClientChan    chan *websocket.Conn
	removeClientChan chan *websocket.Conn
	broadcastChan    chan serverToClients
}