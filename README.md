# cmdTron
Old School Tron, but for Terminal

# To Play
*NB.* Doesn't work in IntelliJ terminal, must be run in Terminal using `go run *.go`

- There are options to play locally or across a network (ensure you are on the same subnet prior to attempting networking multiplayer)
- If you are hosting a game the `hostIP` will be output and this can then be used to join
- Ensure the server has started the game before the client joins

# Rules

- The game is over when a player hits their own tail or crashed into an opponents
- The arrow keys control character 'a' (RED tail)
- 'WASD' keys control character 'b' (BLUE tail)

If you wish to exit prematurely press 'ESC'

# Configurables
Global var `maxLength` can be configured to increase the maximum tail length
