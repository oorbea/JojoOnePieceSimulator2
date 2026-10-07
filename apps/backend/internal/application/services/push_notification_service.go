package services

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

const (
	// maxConcurrentPushBatches bounds how many game notifications are being
	// delivered at once. Each batch is a handful of HTTP calls to the push
	// services; a flood (a busy server, a slow push service) must not spawn
	// unbounded goroutines, so overflow is dropped - a late buzz about a game
	// moment that has already passed is no use anyway.
	maxConcurrentPushBatches = 64

	// pushBatchTimeout bounds one notification's whole fan-out.
	pushBatchTimeout = 30 * time.Second

	// pushSendTimeout bounds a single delivery attempt, so one slow push
	// service cannot hold up the rest of a recipient's devices.
	pushSendTimeout = 10 * time.Second
)

// pushUserLookup is the slice of ports.IUserRepository this service needs:
// the recipient's preferred language.
type pushUserLookup interface {
	FindByID(ctx context.Context, id user.UserID) (*user.User, error)
}

// PushNotificationService turns game moments into Web Push messages for the
// players' devices. It implements ports.IGameNotifier: GameService hands it a
// notification while holding the game's lock, so NotifyGame only enqueues and
// returns; the sending happens on a background goroutine.
type PushNotificationService struct {
	subs   ports.IPushSubscriptionRepository
	sender ports.IPushSender
	users  pushUserLookup

	slots chan struct{}
	wg    sync.WaitGroup
}

var _ ports.IGameNotifier = (*PushNotificationService)(nil)

// NewPushNotificationService builds the service.
func NewPushNotificationService(subs ports.IPushSubscriptionRepository, sender ports.IPushSender, users pushUserLookup) *PushNotificationService {
	return &PushNotificationService{
		subs:   subs,
		sender: sender,
		users:  users,
		slots:  make(chan struct{}, maxConcurrentPushBatches),
	}
}

// NotifyGame implements ports.IGameNotifier.
func (s *PushNotificationService) NotifyGame(n ports.GameNotification) {
	if len(n.Recipients) == 0 {
		return
	}
	select {
	case s.slots <- struct{}{}:
	default:
		log.Printf("push: dropping %s for game %s, delivery queue is full", n.Kind, n.GameID)
		return
	}
	s.wg.Add(1)
	go func() {
		defer func() {
			<-s.slots
			s.wg.Done()
		}()
		s.deliver(n)
	}()
}

// Wait blocks until every notification handed over so far has been
// delivered (or has failed). For tests and graceful shutdown.
func (s *PushNotificationService) Wait() {
	s.wg.Wait()
}

func (s *PushNotificationService) deliver(n ports.GameNotification) {
	ctx, cancel := context.WithTimeout(context.Background(), pushBatchTimeout)
	defer cancel()

	for _, recipient := range n.Recipients {
		s.deliverTo(ctx, n, recipient)
	}
}

func (s *PushNotificationService) deliverTo(ctx context.Context, n ports.GameNotification, recipient user.UserID) {
	subs, err := s.subs.ListByUser(ctx, recipient)
	if err != nil {
		log.Printf("push: listing subscriptions of user %s: %v", recipient, err)
		return
	}
	if len(subs) == 0 {
		return
	}

	locale := enums.DefaultLocale
	if u, err := s.users.FindByID(ctx, recipient); err == nil {
		locale = u.Language()
	}
	text, ok := pushTextFor(n.Kind, locale)
	if !ok {
		log.Printf("push: no wording for notification kind %q", n.Kind)
		return
	}
	msg := ports.PushMessage{
		Title: text.Title,
		Body:  text.Body,
		URL:   "/play/" + n.GameID.String(),
		// One notification per game: a newer moment replaces the previous
		// one instead of stacking a column of stale ones.
		Tag: "game-" + n.GameID.String(),
	}

	for _, sub := range subs {
		sendCtx, cancel := context.WithTimeout(ctx, pushSendTimeout)
		err := s.sender.Send(sendCtx, sub, msg)
		cancel()
		switch {
		case err == nil:
		case errors.Is(err, ports.ErrPushSubscriptionGone):
			if delErr := s.subs.DeleteByEndpoint(ctx, sub.Endpoint); delErr != nil {
				log.Printf("push: deleting gone subscription of user %s: %v", recipient, delErr)
			}
		default:
			log.Printf("push: sending %s to user %s: %v", n.Kind, recipient, err)
		}
	}
}
