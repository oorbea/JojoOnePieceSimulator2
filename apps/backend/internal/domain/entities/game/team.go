package game

import (
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
)

type Team struct {
	name    string
	color   uint32
	players map[*user.Player]struct{}
}

func NewTeam(name string, color uint32, players map[*user.Player]struct{}) *Team {
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

func (t *Team) Players() *map[*user.Player]struct{} {
	return &t.players
}
