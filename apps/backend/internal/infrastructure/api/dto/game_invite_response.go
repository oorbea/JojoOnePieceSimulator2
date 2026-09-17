package dto

import (
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// GameInviteResponse is the JSON body returned by POST /games/{id}/invite:
// a short-lived, multi-use token the caller turns into a share link
// (https://<origin>/join/<token>) - see share.ts on the frontend.
type GameInviteResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// InviteStatusResponse is the JSON body returned by the PUBLIC
// GET /games/invite/{token}/status route. It exists to answer exactly one
// question - "is this link still good" - for a caller who has proven
// nothing but possession of a URL, possibly not even a human (an unfurl
// bot). Do not add another field here: anything beyond Status is a lobby
// fact leaked to every such caller. See services.GameService.InviteStatus's
// doc for what is deliberately withheld and why.
type InviteStatusResponse struct {
	Status string `json:"status" ts:"InviteStatus"`
}

// NewInviteStatusResponse builds an InviteStatusResponse from an
// enums.InviteStatus.
func NewInviteStatusResponse(status enums.InviteStatus) InviteStatusResponse {
	return InviteStatusResponse{Status: status.String()}
}

// InvitePreviewResponse is GET /games/invite/{token}'s response body - the
// authenticated counterpart of LobbyPreviewResponse, reachable with an
// invite token instead of a join code. Deliberately carries no Code field:
// an invite link must never hand its holder the raw join code, only a way
// to join through the token itself.
type InvitePreviewResponse struct {
	PublicLobbyResponse
	Visibility string `json:"visibility" ts:"LobbyVisibility"`
}

// NewInvitePreviewResponse builds an InvitePreviewResponse.
func NewInvitePreviewResponse(l services.LobbyListing) InvitePreviewResponse {
	return InvitePreviewResponse{
		PublicLobbyResponse: NewPublicLobbyResponse(l),
		Visibility:          l.Visibility.String(),
	}
}
