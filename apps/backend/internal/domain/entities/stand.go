package entities

import "github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"

type Stand struct {
	Power
	power       enums.StandStat
	speed       enums.StandStat
	attackRange enums.StandStat
	endurance   enums.StandStat
	precision   enums.StandStat
	potential   enums.StandStat
	evolvesFrom *Stand
}
