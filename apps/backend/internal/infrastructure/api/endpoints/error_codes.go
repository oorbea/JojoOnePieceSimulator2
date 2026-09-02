package endpoints

import (
	"errors"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/valueobjects"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/apierr"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/dto"
)

// errorCode maps a domain/service error onto a stable, machine-readable
// code the frontend can key a localized message off of (see
// shared/lib/toast.ts's showErrorToast). Every branch here mirrors
// handleError's switch in errors.go - the two stay in lockstep since a
// status without a code (or vice versa) would leave the frontend showing
// the backend's English fallback text instead of a translated message.
//
// The codes themselves are apierr constants, not string literals, so
// cmd/typegen can import apierr and emit the TypeScript ErrorCode union
// from apierr.Codes directly instead of AST-scraping this switch.
func errorCode(err error) string {
	var validationErr *dto.ValidationError
	switch {
	case errors.Is(err, ports.ErrStandNotFound):
		return apierr.StandNotFound
	case errors.Is(err, ports.ErrUserNotFound):
		return apierr.UserNotFound
	case errors.Is(err, ports.ErrDevilFruitNotFound):
		return apierr.DevilFruitNotFound
	case errors.Is(err, ports.ErrStandAlreadyExists):
		return apierr.StandAlreadyExists
	case errors.Is(err, ports.ErrUserAlreadyExists):
		return apierr.UserAlreadyExists
	case errors.Is(err, ports.ErrDevilFruitAlreadyExists):
		return apierr.DevilFruitAlreadyExists
	case errors.Is(err, services.ErrLastAdmin):
		return apierr.LastAdmin
	case errors.As(err, &validationErr):
		return apierr.ValidationFailed
	case errors.Is(err, enums.ErrInvalidRarity):
		return apierr.InvalidRarity
	case errors.Is(err, enums.ErrInvalidStandStat):
		return apierr.InvalidStandStat
	case errors.Is(err, enums.ErrInvalidFruitType):
		return apierr.InvalidFruitType
	case errors.Is(err, enums.ErrInvalidUserRole):
		return apierr.InvalidUserRole
	case errors.Is(err, user.ErrInvalidUsername):
		return apierr.InvalidUsername
	case errors.Is(err, valueobjects.ErrInvalidID):
		return apierr.InvalidID
	case errors.Is(err, services.ErrSelfEvolution):
		return apierr.SelfEvolution
	case errors.Is(err, services.ErrPictureRequired):
		return apierr.PictureRequired
	case errors.Is(err, services.ErrUnsupportedPictureType):
		return apierr.UnsupportedPictureType
	case errors.Is(err, ports.ErrInvalidImage):
		return apierr.InvalidImage
	case errors.Is(err, ports.ErrEmailNotVerified):
		return apierr.EmailNotVerified
	case errors.Is(err, ports.ErrConstraintViolation):
		return apierr.ConstraintViolation
	case errors.Is(err, ports.ErrUnauthenticated), errors.Is(err, ports.ErrInvalidGoogleToken):
		return apierr.Unauthenticated
	case errors.Is(err, ports.ErrForbidden):
		return apierr.Forbidden
	case errors.Is(err, services.ErrPictureTooLarge):
		return apierr.PictureTooLarge
	case errors.Is(err, ports.ErrRateLimited):
		return apierr.RateLimited
	case errors.Is(err, services.ErrPictureQueueFull):
		return apierr.PictureQueueFull
	case errors.Is(err, ports.ErrGameNotFound):
		return apierr.GameNotFound
	case errors.Is(err, game.ErrNotHost):
		return apierr.NotHost
	case errors.Is(err, game.ErrGameFull):
		return apierr.GameFull
	case errors.Is(err, game.ErrTeamFull):
		return apierr.TeamFull
	case errors.Is(err, game.ErrDuplicateParticipant):
		return apierr.DuplicateParticipant
	case errors.Is(err, services.ErrAlreadyInGame):
		return apierr.AlreadyInGame
	case errors.Is(err, ports.ErrGameCodeTaken):
		return apierr.GameCodeTaken
	case errors.Is(err, game.ErrInvalidStateTransition):
		return apierr.InvalidStateTransition
	case errors.Is(err, game.ErrVotingClosed):
		return apierr.VotingClosed
	case errors.Is(err, game.ErrInvalidBallotOption):
		return apierr.InvalidBallotOption
	case errors.Is(err, game.ErrTeamSizeMismatch):
		return apierr.TeamSizeMismatch
	case errors.Is(err, game.ErrNotEnoughPlayers):
		return apierr.NotEnoughPlayers
	case errors.Is(err, game.ErrTeamNotFound):
		return apierr.TeamNotFound
	case errors.Is(err, game.ErrParticipantNotFound):
		return apierr.ParticipantNotFound
	case errors.Is(err, game.ErrBotsNotAllowed):
		return apierr.BotsNotAllowed
	case errors.Is(err, services.ErrNotABot):
		return apierr.NotABot
	case errors.Is(err, services.ErrGameNotOver):
		return apierr.GameNotOver
	case errors.Is(err, game.ErrNoStagesAvailable):
		return apierr.NoStagesAvailable
	case errors.Is(err, game.ErrEmptyStageMangas):
		return apierr.EmptyStageMangas
	case errors.Is(err, game.ErrEmptyPowerMangas):
		return apierr.EmptyPowerMangas
	case errors.Is(err, game.ErrInvalidTeamSize):
		return apierr.InvalidTeamSize
	case errors.Is(err, game.ErrFruitMasteryMismatch):
		return apierr.FruitMasteryMismatch
	case errors.Is(err, game.ErrSpin4Required):
		return apierr.Spin4Required
	case errors.Is(err, game.ErrPowerPoolExhausted):
		return apierr.PowerPoolExhausted
	case errors.Is(err, services.ErrCodeGenerationFailed):
		return apierr.GameCodeGenerationFailed
	case errors.Is(err, enums.ErrInvalidGameModeKind):
		return apierr.InvalidGameMode
	case errors.Is(err, enums.ErrInvalidAbilitySource):
		return apierr.InvalidAbilitySource
	case errors.Is(err, enums.ErrInvalidManga):
		return apierr.InvalidManga
	case errors.Is(err, enums.ErrInvalidGameState):
		return apierr.InvalidGameState
	case errors.Is(err, game.ErrInventoryNotSupported):
		return apierr.InventoryNotSupported
	case errors.Is(err, ports.ErrStageNotFound):
		return apierr.StageNotFound
	case errors.Is(err, ports.ErrStageAlreadyExists):
		return apierr.StageAlreadyExists
	case errors.Is(err, game.ErrEmptyTeamName):
		return apierr.EmptyTeamName
	case errors.Is(err, enums.ErrInvalidParticipantKind):
		return apierr.InvalidParticipantKind
	case errors.Is(err, enums.ErrInvalidSquadVerdict):
		return apierr.InvalidSquadVerdict
	case errors.Is(err, errUnknownCommand):
		return apierr.UnknownCommand
	case errors.Is(err, game.ErrLobbyLocked):
		return apierr.LobbyLocked
	case errors.Is(err, game.ErrConfigWouldEvictPlayers):
		return apierr.ConfigWouldEvictPlayers
	case errors.Is(err, game.ErrCannotKickSelf):
		return apierr.CannotKickSelf
	case errors.Is(err, game.ErrPoolTooSmall):
		return apierr.PoolTooSmall
	case errors.Is(err, game.ErrInvalidVotingWindow):
		return apierr.InvalidVotingWindow
	case errors.Is(err, game.ErrInvalidPoolFilter):
		return apierr.InvalidPoolFilter
	case errors.Is(err, enums.ErrInvalidLobbyVisibility):
		return apierr.InvalidLobbyVisibility
	case errors.Is(err, game.ErrLobbyPrivate):
		return apierr.LobbyPrivate
	default:
		return apierr.Internal
	}
}
