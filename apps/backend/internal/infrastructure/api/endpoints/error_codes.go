package endpoints

import (
	"errors"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/valueobjects"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/dto"
)

// errorCode maps a domain/service error onto a stable, machine-readable
// code the frontend can key a localized message off of (see
// shared/lib/toast.ts's showErrorToast). Every branch here mirrors
// handleError's switch in errors.go - the two stay in lockstep since a
// status without a code (or vice versa) would leave the frontend showing
// the backend's English fallback text instead of a translated message.
func errorCode(err error) string {
	var validationErr *dto.ValidationError
	switch {
	case errors.Is(err, ports.ErrStandNotFound):
		return "STAND_NOT_FOUND"
	case errors.Is(err, ports.ErrUserNotFound):
		return "USER_NOT_FOUND"
	case errors.Is(err, ports.ErrDevilFruitNotFound):
		return "DEVIL_FRUIT_NOT_FOUND"
	case errors.Is(err, ports.ErrStandAlreadyExists):
		return "STAND_ALREADY_EXISTS"
	case errors.Is(err, ports.ErrUserAlreadyExists):
		return "USER_ALREADY_EXISTS"
	case errors.Is(err, ports.ErrDevilFruitAlreadyExists):
		return "DEVIL_FRUIT_ALREADY_EXISTS"
	case errors.Is(err, services.ErrLastAdmin):
		return "LAST_ADMIN"
	case errors.As(err, &validationErr):
		return "VALIDATION_FAILED"
	case errors.Is(err, enums.ErrInvalidRarity):
		return "INVALID_RARITY"
	case errors.Is(err, enums.ErrInvalidStandStat):
		return "INVALID_STAND_STAT"
	case errors.Is(err, enums.ErrInvalidFruitType):
		return "INVALID_FRUIT_TYPE"
	case errors.Is(err, enums.ErrInvalidUserRole):
		return "INVALID_USER_ROLE"
	case errors.Is(err, user.ErrInvalidUsername):
		return "INVALID_USERNAME"
	case errors.Is(err, valueobjects.ErrInvalidID):
		return "INVALID_ID"
	case errors.Is(err, services.ErrSelfEvolution):
		return "SELF_EVOLUTION"
	case errors.Is(err, services.ErrPictureRequired):
		return "PICTURE_REQUIRED"
	case errors.Is(err, services.ErrUnsupportedPictureType):
		return "UNSUPPORTED_PICTURE_TYPE"
	case errors.Is(err, ports.ErrInvalidImage):
		return "INVALID_IMAGE"
	case errors.Is(err, ports.ErrEmailNotVerified):
		return "EMAIL_NOT_VERIFIED"
	case errors.Is(err, ports.ErrConstraintViolation):
		return "CONSTRAINT_VIOLATION"
	case errors.Is(err, ports.ErrUnauthenticated), errors.Is(err, ports.ErrInvalidGoogleToken):
		return "UNAUTHENTICATED"
	case errors.Is(err, ports.ErrForbidden):
		return "FORBIDDEN"
	case errors.Is(err, services.ErrPictureTooLarge):
		return "PICTURE_TOO_LARGE"
	case errors.Is(err, ports.ErrRateLimited):
		return "RATE_LIMITED"
	case errors.Is(err, services.ErrPictureQueueFull):
		return "PICTURE_QUEUE_FULL"
	default:
		return "INTERNAL"
	}
}
