package game

type IGameMode interface {
	NextRound(g *Game, winnerOfTheRound *Team) error
}
