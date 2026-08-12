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

	// MinVotingWindowSeconds and MaxVotingWindowSeconds bound a lobby's
	// configurable voting-window duration.
	MinVotingWindowSeconds = 5
	MaxVotingWindowSeconds = 180

	// DefaultVotingWindowSeconds is the fallback used by Restore for a
	// legacy (pre-per-game-window) Snapshot, and by the application layer
	// when a CreateGame request doesn't specify one.
	DefaultVotingWindowSeconds = 30
)

var (
	// ErrEmptyMangas is returned when a Config selects no manga at all.
	ErrEmptyMangas = errors.New("at least one manga must be selected")

	// ErrInvalidTeamSize is returned when Config's team size is outside
	// the bounds for its GameModeKind.
	ErrInvalidTeamSize = errors.New("invalid team size for this game mode")

	// ErrInvalidVotingWindow is returned when Config's voting window falls
	// outside [MinVotingWindowSeconds, MaxVotingWindowSeconds].
	ErrInvalidVotingWindow = errors.New("invalid voting window")
)

// Config is a lobby's setup: which mode, which manga(s) abilities are drawn
// from, how abilities are sourced, how many players per team, whether bots
// may fill empty Versus slots, how long a round's voting window lasts,
// whether the lobby is browsable, and which Stands/DevilFruits are eligible
// to be drawn. Config itself is still a plain immutable value object - the
// host changes it by having Game.Reconfigure swap in a whole new Config,
// never by mutating one in place.
type Config struct {
	poolFilter          PoolFilter
	mangas              []enums.Manga
	teamSize            int
	votingWindowSeconds int
	mode                enums.GameModeKind
	abilitySource       enums.AbilitySource
	allowBots           bool
	visibility          enums.LobbyVisibility
}

// NewConfig validates and builds a Config.
func NewConfig(
	mode enums.GameModeKind,
	mangas []enums.Manga,
	abilitySource enums.AbilitySource,
	teamSize int,
	allowBots bool,
	visibility enums.LobbyVisibility,
	votingWindowSeconds int,
	poolFilter PoolFilter,
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
	if !visibility.IsValid() {
		return Config{}, enums.ErrInvalidLobbyVisibility
	}
	if votingWindowSeconds < MinVotingWindowSeconds || votingWindowSeconds > MaxVotingWindowSeconds {
		return Config{}, ErrInvalidVotingWindow
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
		mode:                mode,
		mangas:              uniqueMangas,
		abilitySource:       abilitySource,
		teamSize:            teamSize,
		allowBots:           allowBots,
		visibility:          visibility,
		votingWindowSeconds: votingWindowSeconds,
		poolFilter:          poolFilter,
	}, nil
}

func (c Config) Mode() enums.GameModeKind           { return c.mode }
func (c Config) AbilitySource() enums.AbilitySource { return c.abilitySource }
func (c Config) TeamSize() int                      { return c.teamSize }
func (c Config) AllowBots() bool                    { return c.allowBots }
func (c Config) Visibility() enums.LobbyVisibility  { return c.visibility }
func (c Config) VotingWindowSeconds() int           { return c.votingWindowSeconds }
func (c Config) PoolFilter() PoolFilter             { return c.poolFilter }

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
