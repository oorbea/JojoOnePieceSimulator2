package game

import (
	"errors"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

const (
	// MinGauntletPlayers and MaxGauntletPlayers bound a Gauntlet's single
	// team.
	MinGauntletPlayers = 1
	MaxGauntletPlayers = 10

	// MinVersusTeamSize and MaxVersusTeamSize bound each of a Versus
	// match's two teams - both teams always share the same size.
	MinVersusTeamSize = 1
	MaxVersusTeamSize = 5

	// VersusTeamCount is the fixed number of teams in a Versus match.
	VersusTeamCount = 2

	// VersusRounds is the fixed number of rounds a Versus match plays.
	VersusRounds = 3
)

var (
	// ErrEmptyMangas is returned when a Config selects no manga at all.
	ErrEmptyMangas = errors.New("at least one manga must be selected")

	// ErrInvalidTeamSize is returned when Config's team size is outside
	// the bounds for its GameModeKind.
	ErrInvalidTeamSize = errors.New("invalid team size for this game mode")
)

// Config is a Game's immutable setup, fixed by the host at creation time:
// which mode, which manga(s) abilities are drawn from, how abilities are
// sourced, how many players per team, and whether bots may fill empty
// Versus slots.
type Config struct {
	mode          enums.GameModeKind
	mangas        []enums.Manga
	abilitySource enums.AbilitySource
	teamSize      int
	allowBots     bool
}

// NewConfig validates and builds a Config.
func NewConfig(
	mode enums.GameModeKind,
	mangas []enums.Manga,
	abilitySource enums.AbilitySource,
	teamSize int,
	allowBots bool,
) (Config, error) {
	if !mode.IsValid() {
		return Config{}, enums.ErrInvalidGameModeKind
	}
	if !abilitySource.IsValid() {
		return Config{}, enums.ErrInvalidAbilitySource
	}
	if abilitySource == enums.Inventory {
		return Config{}, ErrInventoryNotSupported
	}
	if len(mangas) == 0 {
		return Config{}, ErrEmptyMangas
	}

	seen := make(map[enums.Manga]struct{}, len(mangas))
	for _, m := range mangas {
		if !m.IsValid() {
			return Config{}, enums.ErrInvalidManga
		}
		seen[m] = struct{}{}
	}
	uniqueMangas := make([]enums.Manga, 0, len(seen))
	for _, m := range enums.Mangas() {
		if _, ok := seen[m]; ok {
			uniqueMangas = append(uniqueMangas, m)
		}
	}

	switch mode {
	case enums.Gauntlet:
		if allowBots {
			return Config{}, ErrBotsNotAllowed
		}
		if teamSize < MinGauntletPlayers || teamSize > MaxGauntletPlayers {
			return Config{}, ErrInvalidTeamSize
		}
	case enums.Versus:
		if teamSize < MinVersusTeamSize || teamSize > MaxVersusTeamSize {
			return Config{}, ErrInvalidTeamSize
		}
	}

	return Config{
		mode:          mode,
		mangas:        uniqueMangas,
		abilitySource: abilitySource,
		teamSize:      teamSize,
		allowBots:     allowBots,
	}, nil
}

func (c Config) Mode() enums.GameModeKind           { return c.mode }
func (c Config) AbilitySource() enums.AbilitySource { return c.abilitySource }
func (c Config) TeamSize() int                      { return c.teamSize }
func (c Config) AllowBots() bool                    { return c.allowBots }

// Mangas returns a copy of the selected mangas, in enums.Mangas() order.
func (c Config) Mangas() []enums.Manga {
	return append([]enums.Manga(nil), c.mangas...)
}

// HasManga reports whether m is selected for this Config.
func (c Config) HasManga(m enums.Manga) bool {
	for _, x := range c.mangas {
		if x == m {
			return true
		}
	}
	return false
}
