package game

import "github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities"

type Team struct {
	name    string
	color   uint32
	players map[*entities.Player]struct{}
}

func NewTeam(name string, color uint32, players map[*entities.Player]struct{}) *Team {
	return &Team{
		name:    name,
		color:   color,
		players: players,
	}
}

func (t *Team) Name() string {
	return t.name
}

func (t *Team) Color() uint32 {
	return t.color
}

func (t *Team) Players() *map[*entities.Player]struct{} {
	return &t.players
}
