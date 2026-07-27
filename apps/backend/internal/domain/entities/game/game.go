package game

type Game struct {
	gameMode *IGameMode
	teams    []Team
	turn     byte
}
