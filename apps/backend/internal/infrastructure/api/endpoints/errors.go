package endpoints

import (
	"errors"
	"log"
	"net/http"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/valueobjects"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/dto"
)

// handleError maps a domain/service error onto the appropriate HTTP status
// and writes the response body, logging anything that maps to a 500 so the
// real cause isn't lost behind a generic message.
func handleError(w http.ResponseWriter, err error) {
	code := errorCode(err)
	var validationErr *dto.ValidationError
	switch {
	case errors.Is(err, ports.ErrStandNotFound), errors.Is(err, ports.ErrUserNotFound), errors.Is(err, ports.ErrDevilFruitNotFound):
		writeError(w, http.StatusNotFound, code, err.Error())
	case errors.Is(err, ports.ErrStandAlreadyExists), errors.Is(err, ports.ErrUserAlreadyExists), errors.Is(err, ports.ErrDevilFruitAlreadyExists),
		errors.Is(err, services.ErrLastAdmin):
		writeError(w, http.StatusConflict, code, err.Error())
	case errors.As(err, &validationErr):
		writeError(w, http.StatusBadRequest, code, "validation failed", validationErr.Errors...)
	case errors.Is(err, enums.ErrInvalidRarity),
		errors.Is(err, enums.ErrInvalidStandStat),
		errors.Is(err, enums.ErrInvalidFruitType),
		errors.Is(err, enums.ErrInvalidUserRole),
		errors.Is(err, user.ErrInvalidUsername),
		errors.Is(err, valueobjects.ErrInvalidID),
		errors.Is(err, services.ErrSelfEvolution),
		errors.Is(err, services.ErrPictureRequired),
		errors.Is(err, services.ErrUnsupportedPictureType),
		errors.Is(err, ports.ErrInvalidImage),
		errors.Is(err, ports.ErrEmailNotVerified),
		errors.Is(err, ports.ErrConstraintViolation):
		writeError(w, http.StatusBadRequest, code, err.Error())
	case errors.Is(err, ports.ErrUnauthenticated), errors.Is(err, ports.ErrInvalidGoogleToken):
		writeError(w, http.StatusUnauthorized, code, "unauthenticated")
	case errors.Is(err, ports.ErrForbidden):
		writeError(w, http.StatusForbidden, code, "forbidden")
	case errors.Is(err, services.ErrPictureTooLarge):
		writeError(w, http.StatusRequestEntityTooLarge, code, err.Error())
	case errors.Is(err, ports.ErrRateLimited):
		writeError(w, http.StatusTooManyRequests, code, "too many requests")
	case errors.Is(err, services.ErrPictureQueueFull):
		writeError(w, http.StatusServiceUnavailable, code, err.Error())
	default:
		log.Printf("internal error: %v", err)
		writeError(w, http.StatusInternalServerError, code, "internal server error")
	}
}
