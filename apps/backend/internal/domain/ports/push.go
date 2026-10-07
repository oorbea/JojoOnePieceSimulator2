package ports

import (
	"context"
	"errors"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
)

// ErrPushSubscriptionGone is returned by IPushSender.Send when the push
// service says the subscription no longer exists (HTTP 404/410: the user
// uninstalled the app, cleared site data or revoked the permission). The
// caller must delete the subscription - retrying it can never succeed.
var ErrPushSubscriptionGone = errors.New("push subscription is gone")

// PushSubscription is one device's Web Push registration: the endpoint the
// browser's push service issued plus the keys a payload is encrypted with.
type PushSubscription struct {
	UserID     user.UserID
	Endpoint   string
	P256dh     string
	Auth       string
	LastSeenAt time.Time
}

// PushMessage is what the service worker receives and shows. URL is the
// in-app path a tap opens; Tag makes a newer notification replace an older
// one for the same game instead of stacking.
type PushMessage struct {
	Title string
	Body  string
	URL   string
	Tag   string
}

// IPushSubscriptionRepository persists push subscriptions. A subscription is
// identified by its endpoint: registering an endpoint that already exists
// (e.g. another account logged in on the same device) moves it to the new
// owner.
type IPushSubscriptionRepository interface {
	// Upsert registers or refreshes a subscription and trims the owner's
	// oldest ones beyond a per-user cap.
	Upsert(ctx context.Context, sub PushSubscription) error
	// DeleteForUser removes one of the user's own subscriptions; removing one
	// that does not exist is not an error.
	DeleteForUser(ctx context.Context, userID user.UserID, endpoint string) error
	// DeleteByEndpoint removes a subscription whatever its owner - used when
	// the push service reports it gone.
	DeleteByEndpoint(ctx context.Context, endpoint string) error
	ListByUser(ctx context.Context, userID user.UserID) ([]PushSubscription, error)
}

// IPushSender delivers one message to one subscription. Returns
// ErrPushSubscriptionGone for a dead subscription; any other error is
// transient or a delivery failure the caller should only log.
type IPushSender interface {
	Send(ctx context.Context, sub PushSubscription, msg PushMessage) error
}

// GameNotificationKind is a game moment worth waking a player's phone for.
type GameNotificationKind string

const (
	NotifyGameStarted   GameNotificationKind = "GAME_STARTED"
	NotifyVotingOpened  GameNotificationKind = "VOTING_OPENED"
	NotifyTiebreak      GameNotificationKind = "TIEBREAK_OPENED"
	NotifyRoundResolved GameNotificationKind = "ROUND_RESOLVED"
	NotifyGameFinished  GameNotificationKind = "GAME_FINISHED"
)

// GameNotification asks for Kind to be announced to Recipients (registered
// users only - bots have no devices) about GameID.
type GameNotification struct {
	GameID     game.GameID
	Kind       GameNotificationKind
	Recipients []user.UserID
}

// IGameNotifier is how GameService announces game moments to players who may
// not have the app open. NotifyGame MUST NOT block: it is called while the
// game's lock is held, so an implementation hands the work off and returns
// immediately. Delivery is best-effort.
type IGameNotifier interface {
	NotifyGame(n GameNotification)
}
