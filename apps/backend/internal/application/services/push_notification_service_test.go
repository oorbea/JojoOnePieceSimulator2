package services_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

type fakePushSubs struct {
	mu      sync.Mutex
	byUser  map[user.UserID][]ports.PushSubscription
	deleted []string
}

func (f *fakePushSubs) Upsert(context.Context, ports.PushSubscription) error { return nil }
func (f *fakePushSubs) DeleteForUser(context.Context, user.UserID, string) error {
	return nil
}
func (f *fakePushSubs) DeleteByEndpoint(_ context.Context, endpoint string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deleted = append(f.deleted, endpoint)
	return nil
}
func (f *fakePushSubs) ListByUser(_ context.Context, id user.UserID) ([]ports.PushSubscription, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.byUser[id], nil
}

type sentPush struct {
	sub ports.PushSubscription
	msg ports.PushMessage
}

type fakePushSender struct {
	mu   sync.Mutex
	sent []sentPush
	errs map[string]error // by endpoint
}

func (f *fakePushSender) Send(_ context.Context, sub ports.PushSubscription, msg ports.PushMessage) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, sentPush{sub: sub, msg: msg})
	return f.errs[sub.Endpoint]
}

type fakePushUsers map[user.UserID]*user.User

func (f fakePushUsers) FindByID(_ context.Context, id user.UserID) (*user.User, error) {
	if u, ok := f[id]; ok {
		return u, nil
	}
	return nil, ports.ErrUserNotFound
}

func pushTestUser(t *testing.T, seed byte, language enums.Locale) *user.User {
	t.Helper()
	var id user.UserID
	id[0] = seed
	name := fmt.Sprintf("player%d", seed)
	u, err := user.NewUser(id, "sub-"+name, name+"@example.com", name, "", "", enums.Regular)
	if err != nil {
		t.Fatalf("NewUser: %v", err)
	}
	if err := u.ChangeLanguage(language); err != nil {
		t.Fatalf("ChangeLanguage: %v", err)
	}
	return u
}

func pushTestGameID() game.GameID {
	var id game.GameID
	id[0] = 0xAB
	return id
}

func TestPushNotificationService_SendsLocalizedMessageToEveryDevice(t *testing.T) {
	spanish := pushTestUser(t, 1, enums.EsES)
	english := pushTestUser(t, 2, enums.EnGB)
	subs := &fakePushSubs{byUser: map[user.UserID][]ports.PushSubscription{
		spanish.ID(): {{UserID: spanish.ID(), Endpoint: "https://fcm.googleapis.com/a"}, {UserID: spanish.ID(), Endpoint: "https://fcm.googleapis.com/b"}},
		english.ID(): {{UserID: english.ID(), Endpoint: "https://fcm.googleapis.com/c"}},
	}}
	sender := &fakePushSender{}
	svc := services.NewPushNotificationService(subs, sender, fakePushUsers{spanish.ID(): spanish, english.ID(): english})

	gameID := pushTestGameID()
	svc.NotifyGame(ports.GameNotification{GameID: gameID, Kind: ports.NotifyVotingOpened, Recipients: []user.UserID{spanish.ID(), english.ID()}})
	svc.Wait()

	if len(sender.sent) != 3 {
		t.Fatalf("sent %d pushes, want 3 (two devices + one device)", len(sender.sent))
	}
	byEndpoint := map[string]ports.PushMessage{}
	for _, s := range sender.sent {
		byEndpoint[s.sub.Endpoint] = s.msg
	}
	if got := byEndpoint["https://fcm.googleapis.com/a"].Title; got != "Te toca votar" {
		t.Errorf("spanish title = %q", got)
	}
	if got := byEndpoint["https://fcm.googleapis.com/c"].Title; got != "Time to vote" {
		t.Errorf("english title = %q", got)
	}
	msg := byEndpoint["https://fcm.googleapis.com/a"]
	if msg.URL != "/play/"+gameID.String() || msg.Tag != "game-"+gameID.String() {
		t.Errorf("url/tag = %q / %q, want the game's room and a per-game tag", msg.URL, msg.Tag)
	}
}

func TestPushNotificationService_DeletesSubscriptionsThePushServiceReportsGone(t *testing.T) {
	player := pushTestUser(t, 3, enums.CaES)
	subs := &fakePushSubs{byUser: map[user.UserID][]ports.PushSubscription{
		player.ID(): {{UserID: player.ID(), Endpoint: "https://fcm.googleapis.com/dead"}, {UserID: player.ID(), Endpoint: "https://fcm.googleapis.com/alive"}},
	}}
	sender := &fakePushSender{errs: map[string]error{"https://fcm.googleapis.com/dead": ports.ErrPushSubscriptionGone}}
	svc := services.NewPushNotificationService(subs, sender, fakePushUsers{player.ID(): player})

	svc.NotifyGame(ports.GameNotification{GameID: pushTestGameID(), Kind: ports.NotifyGameFinished, Recipients: []user.UserID{player.ID()}})
	svc.Wait()

	if len(subs.deleted) != 1 || subs.deleted[0] != "https://fcm.googleapis.com/dead" {
		t.Errorf("deleted = %v, want only the dead endpoint", subs.deleted)
	}
	if len(sender.sent) != 2 {
		t.Errorf("sent %d, want both devices attempted", len(sender.sent))
	}
}

func TestPushNotificationService_KeepsSubscriptionOnTransientFailure(t *testing.T) {
	player := pushTestUser(t, 4, enums.EsES)
	subs := &fakePushSubs{byUser: map[user.UserID][]ports.PushSubscription{
		player.ID(): {{UserID: player.ID(), Endpoint: "https://fcm.googleapis.com/flaky"}},
	}}
	sender := &fakePushSender{errs: map[string]error{"https://fcm.googleapis.com/flaky": errors.New("push service answered HTTP 503")}}
	svc := services.NewPushNotificationService(subs, sender, fakePushUsers{player.ID(): player})

	svc.NotifyGame(ports.GameNotification{GameID: pushTestGameID(), Kind: ports.NotifyRoundResolved, Recipients: []user.UserID{player.ID()}})
	svc.Wait()

	if len(subs.deleted) != 0 {
		t.Errorf("deleted = %v, a transient failure must keep the subscription", subs.deleted)
	}
}

func TestPushNotificationService_FallsBackToDefaultLocaleForUnknownUser(t *testing.T) {
	var ghost user.UserID
	ghost[0] = 9
	subs := &fakePushSubs{byUser: map[user.UserID][]ports.PushSubscription{
		ghost: {{UserID: ghost, Endpoint: "https://fcm.googleapis.com/x"}},
	}}
	sender := &fakePushSender{}
	svc := services.NewPushNotificationService(subs, sender, fakePushUsers{})

	svc.NotifyGame(ports.GameNotification{GameID: pushTestGameID(), Kind: ports.NotifyGameStarted, Recipients: []user.UserID{ghost}})
	svc.Wait()

	if len(sender.sent) != 1 || sender.sent[0].msg.Title != "¡Empieza la partida!" {
		t.Errorf("sent = %+v, want one message in the default locale", sender.sent)
	}
}

func TestPushNotificationService_SkipsRecipientsWithoutDevices(t *testing.T) {
	player := pushTestUser(t, 5, enums.EsES)
	sender := &fakePushSender{}
	svc := services.NewPushNotificationService(&fakePushSubs{}, sender, fakePushUsers{player.ID(): player})

	svc.NotifyGame(ports.GameNotification{GameID: pushTestGameID(), Kind: ports.NotifyTiebreak, Recipients: []user.UserID{player.ID()}})
	svc.NotifyGame(ports.GameNotification{GameID: pushTestGameID(), Kind: ports.NotifyTiebreak})
	svc.Wait()

	if len(sender.sent) != 0 {
		t.Errorf("sent %d pushes, want none", len(sender.sent))
	}
}

func TestPushNotificationService_EveryKindHasWordingInEveryLocale(t *testing.T) {
	kinds := []ports.GameNotificationKind{
		ports.NotifyGameStarted, ports.NotifyVotingOpened, ports.NotifyTiebreak,
		ports.NotifyRoundResolved, ports.NotifyGameFinished,
	}
	for _, locale := range []enums.Locale{enums.EsES, enums.EnGB, enums.CaES} {
		player := pushTestUser(t, 6, locale)
		subs := &fakePushSubs{byUser: map[user.UserID][]ports.PushSubscription{
			player.ID(): {{UserID: player.ID(), Endpoint: "https://fcm.googleapis.com/k"}},
		}}
		sender := &fakePushSender{}
		svc := services.NewPushNotificationService(subs, sender, fakePushUsers{player.ID(): player})
		for _, kind := range kinds {
			svc.NotifyGame(ports.GameNotification{GameID: pushTestGameID(), Kind: kind, Recipients: []user.UserID{player.ID()}})
		}
		svc.Wait()

		if len(sender.sent) != len(kinds) {
			t.Fatalf("%s: sent %d, want %d", locale, len(sender.sent), len(kinds))
		}
		for _, s := range sender.sent {
			if s.msg.Title == "" || s.msg.Body == "" {
				t.Errorf("%s: empty wording in %+v", locale, s.msg)
			}
		}
	}
}
