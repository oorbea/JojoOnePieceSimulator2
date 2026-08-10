package game

import (
	"context"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// jojoPartNames/onePieceSagaNames are the round content ports.IStageCatalog
// serves today: JoJo's 8 parts, and the ~11 sagas the project owner chose
// for One Piece (not the finer ~31-arc breakdown - see
// ObsidianVault/gameplay-game-modes.md). Both are hardcoded here as a
// TODO-marked stub: there is no schema, migration, or admin CRUD for stage
// content yet (deliberately out of scope for this tanda, per the vault) -
// this exists only so the rest of the application layer is runnable
// end-to-end. Replace with a Postgres-backed repository once that CRUD
// exists.
var jojoPartNames = []string{
	"Phantom Blood",
	"Battle Tendency",
	"Stardust Crusaders",
	"Diamond is Unbreakable",
	"Golden Wind",
	"Stone Ocean",
	"Steel Ball Run",
	"JoJolion",
}

var onePieceSagaNames = []string{
	"East Blue",
	"Alabasta",
	"Sky Island",
	"Water Seven",
	"Thriller Bark",
	"Summit War",
	"Fish-Man Island",
	"Dressrosa",
	"Whole Cake Island",
	"Wano Country",
	"Egghead",
}

// StaticStageCatalog is a hardcoded, in-memory ports.IStageCatalog. See the
// TODO on jojoPartNames/onePieceSagaNames.
type StaticStageCatalog struct {
	stages map[enums.Manga][]game.Stage
}

// NewStaticStageCatalog builds a StaticStageCatalog. It panics if the
// hardcoded stage lists themselves fail to validate (game.NewStage never
// rejects a non-empty name and non-negative order, so this can only happen
// if this file's own constants are broken).
func NewStaticStageCatalog() *StaticStageCatalog {
	stages := map[enums.Manga][]game.Stage{
		enums.Jojo:     buildStages(0x01, enums.Jojo, jojoPartNames),
		enums.OnePiece: buildStages(0x02, enums.OnePiece, onePieceSagaNames),
	}
	return &StaticStageCatalog{stages: stages}
}

func buildStages(seed byte, manga enums.Manga, names []string) []game.Stage {
	out := make([]game.Stage, 0, len(names))
	for i, name := range names {
		var id game.StageID
		id[0] = seed
		id[1] = byte(i + 1)
		stage, err := game.NewStage(id, manga, i, name)
		if err != nil {
			panic("static stage catalog: " + err.Error())
		}
		out = append(out, stage)
	}
	return out
}

var _ ports.IStageCatalog = (*StaticStageCatalog)(nil)

func (c *StaticStageCatalog) Stages(_ context.Context, manga enums.Manga) ([]game.Stage, error) {
	stages := c.stages[manga]
	return append([]game.Stage(nil), stages...), nil
}
