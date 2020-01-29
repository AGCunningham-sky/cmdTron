package main

import (
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
	exit = false

	for {
		//TODO: recieve data from each side & calculate updated positions

		var crash bool
		if PlayerA, crash = playerMovement(PlayerA); crash {
			color.Blue("WASD Wins")
			break
		}
		if PlayerB, crash = playerMovement(PlayerB); crash {
			color.Red("Arrows Wins")
			break
		}

		if collisionDetection(PlayerA, PlayerB.Player[0]) {
			color.Red("Arrows Wins")
			break
		}
		if collisionDetection(PlayerB, PlayerA.Player[0]) {
			color.Blue("WASD Wins")
			break
		}

		PlayerA, PlayerB = playerDirection(PlayerA, PlayerB, inp)

		if exit {
			break
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

func playerDirection(user1, user2 player, input string) (player, player) {
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
		var newRow sprite

		switch Player.Direction {
		case "UP": newRow = sprite{Player.Player[0].row - 1, Player.Player[0].col, true}
		case "DOWN": newRow = sprite{Player.Player[0].row + 1, Player.Player[0].col, true}
		case "LEFT": newRow = sprite{Player.Player[0].row, Player.Player[0].col + 1, true}
		case "RIGHT": newRow = sprite{Player.Player[0].row, Player.Player[0].col - 1, true}
		}
		if newRow.here != false {
			Player.Player = append([]sprite{newRow}, Player.Player...)
		}

		if len(Player.Player) > maxLength {
			Player.Player = Player.Player[:len(Player.Player)-1]
		}

		if Player.Player[0].row >= len(maze)-1 {
			Player.Player[0].row = 1
		} else if Player.Player[0].row <= 0 {
			Player.Player[0].row = len(maze)-2
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