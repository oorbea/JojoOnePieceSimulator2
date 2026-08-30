package redis_test

import (
	"context"
	"os"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	gameredis "github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/gamestore/redis"
)

// newTestStore connects to TEST_REDIS_URL, skipping the test entirely when
// it is unset - same convention as infrastructure/cache/redis's tests.
func newTestStore(t *testing.T) *gameredis.Store {
	t.Helper()
	url := os.Getenv("TEST_REDIS_URL")
	if url == "" {
		t.Skip("TEST_REDIS_URL not set, skipping redis-backed game store test")
	}
	s, err := gameredis.New(context.Background(), gameredis.Config{
		URL: url, DialTimeout: 2 * time.Second, OpTimeout: 2 * time.Second, TTL: time.Minute,
	})
	if err != nil {
		t.Fatalf("connecting to redis: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// rawClient is used only to assert on TTLs directly - the Store itself
// exposes no such introspection, which is correct for the port but not for
// this test.
func rawClient(t *testing.T) *goredis.Client {
	t.Helper()
	url := os.Getenv("TEST_REDIS_URL")
	opts, err := goredis.ParseURL(url)
	if err != nil {
		t.Fatalf("parsing TEST_REDIS_URL: %v", err)
	}
	c := goredis.NewClient(opts)
	t.Cleanup(func() { _ = c.Close() })
	return c
}

// newFreshGame builds a minimal, valid Gauntlet game with a fresh random id
// and a random-ish join code, so concurrent test runs never collide.
func newFreshGame(t *testing.T, seed byte) (*game.Game, string) {
	t.Helper()
	cfg, err := game.NewConfig(enums.Gauntlet, []enums.Manga{enums.Jojo}, []enums.Manga{enums.Jojo}, enums.Random, game.MaxGauntletPlayers, false, enums.Private, 30, game.PoolFilter{}, enums.Normal, game.DefaultSummaryDurationSeconds)
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}
	var idBytes [16]byte
	idBytes[0] = seed
	idBytes[1] = 0xAA
	host, err := game.NewHumanParticipant(game.ParticipantID{seed, 1}, user.UserID{seed, 2}, "host", game.TeamID{seed, 10})
	if err != nil {
		t.Fatalf("NewHumanParticipant: %v", err)
	}
	team, err := game.NewTeam(game.TeamID{seed, 10}, "Squad", 0)
	if err != nil {
		t.Fatalf("NewTeam: %v", err)
	}
	stage, err := game.NewStage(game.StageID{seed, 20}, enums.Jojo, 0, "Phantom Blood", "a test stage", "")
	if err != nil {
		t.Fatalf("NewStage: %v", err)
	}
	g, err := game.NewGame(game.GameID(idBytes), cfg, host, []*game.Team{team}, []game.Stage{stage})
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	code := game.GameID(idBytes).String()[:6]
	return g, code
}

func TestStore_CreateGetSave_RoundTrip(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	g, code := newFreshGame(t, 1)
	t.Cleanup(func() { _ = s.Delete(ctx, g.ID()) })

	if err := s.Create(ctx, code, g); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := s.Get(ctx, g.ID())
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID() != g.ID() {
		t.Errorf("Get: ID mismatch got %s want %s", got.ID(), g.ID())
	}

	byCode, err := s.GetByCode(ctx, code)
	if err != nil {
		t.Fatalf("GetByCode: %v", err)
	}
	if byCode.ID() != g.ID() {
		t.Errorf("GetByCode: ID mismatch got %s want %s", byCode.ID(), g.ID())
	}

	gotCode, err := s.Code(ctx, g.ID())
	if err != nil {
		t.Fatalf("Code: %v", err)
	}
	if gotCode != code {
		t.Errorf("Code = %q, want %q", gotCode, code)
	}
}

func TestStore_Create_DuplicateCodeFails(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	g1, code := newFreshGame(t, 2)
	g2, _ := newFreshGame(t, 3)
	t.Cleanup(func() {
		_ = s.Delete(ctx, g1.ID())
		_ = s.Delete(ctx, g2.ID())
	})

	if err := s.Create(ctx, code, g1); err != nil {
		t.Fatalf("Create(g1): %v", err)
	}
	if err := s.Create(ctx, code, g2); err != ports.ErrGameCodeTaken {
		t.Fatalf("Create(g2) with duplicate code = %v, want ErrGameCodeTaken", err)
	}
}

func TestStore_UnknownGame_ReturnsNotFound(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	unknown := game.GameID{0xFF, 0xEE, 0xDD}

	if _, err := s.Get(ctx, unknown); err != ports.ErrGameNotFound {
		t.Errorf("Get(unknown) = %v, want ErrGameNotFound", err)
	}
	if _, err := s.GetByCode(ctx, "ZZZZZZ"); err != ports.ErrGameNotFound {
		t.Errorf("GetByCode(unknown) = %v, want ErrGameNotFound", err)
	}
	if _, err := s.Code(ctx, unknown); err != ports.ErrGameNotFound {
		t.Errorf("Code(unknown) = %v, want ErrGameNotFound", err)
	}
}

func TestStore_Save_BeforeCreate_ReturnsNotFound(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	g, _ := newFreshGame(t, 4)

	if err := s.Save(ctx, g); err != ports.ErrGameNotFound {
		t.Fatalf("Save before Create = %v, want ErrGameNotFound", err)
	}
}

func TestStore_Save_RefreshesTTLOnAllThreeKeys(t *testing.T) {
	s := newTestStore(t)
	client := rawClient(t)
	ctx := context.Background()
	g, code := newFreshGame(t, 5)
	t.Cleanup(func() { _ = s.Delete(ctx, g.ID()) })

	if err := s.Create(ctx, code, g); err != nil {
		t.Fatalf("Create: %v", err)
	}

	idKey := "jojo:game:id:" + g.ID().String()
	codeKey := "jojo:game:code:" + code
	codeOfKey := "jojo:game:codeof:" + g.ID().String()

	// Let the initial TTL tick down a little, then Save and confirm it
	// jumped back up close to the full TTL on every key.
	time.Sleep(1100 * time.Millisecond)
	if err := s.Save(ctx, g); err != nil {
		t.Fatalf("Save: %v", err)
	}
	for _, k := range []string{idKey, codeKey, codeOfKey} {
		ttl, err := client.PTTL(ctx, k).Result()
		if err != nil {
			t.Fatalf("PTTL(%s): %v", k, err)
		}
		if ttl < 50*time.Second {
			t.Errorf("PTTL(%s) = %s, expected close to the full 1m TTL after Save refreshed it", k, ttl)
		}
	}
}

func TestStore_Delete_RemovesAllThreeKeys(t *testing.T) {
	s := newTestStore(t)
	client := rawClient(t)
	ctx := context.Background()
	g, code := newFreshGame(t, 6)

	if err := s.Create(ctx, code, g); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := s.Delete(ctx, g.ID()); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	// Deleting an already-absent game must not error.
	if err := s.Delete(ctx, g.ID()); err != nil {
		t.Fatalf("Delete (again): %v", err)
	}

	for _, k := range []string{
		"jojo:game:id:" + g.ID().String(),
		"jojo:game:code:" + code,
		"jojo:game:codeof:" + g.ID().String(),
	} {
		n, err := client.Exists(ctx, k).Result()
		if err != nil {
			t.Fatalf("Exists(%s): %v", k, err)
		}
		if n != 0 {
			t.Errorf("key %s still exists after Delete", k)
		}
	}
}

func TestStore_DeleteExpired_IsANoOp(t *testing.T) {
	s := newTestStore(t)
	if n := s.DeleteExpired(context.Background(), time.Hour); n != 0 {
		t.Errorf("DeleteExpired = %d, want 0 (expiry is delegated to Redis TTL)", n)
	}
}

func TestStore_Get_CorruptPayload_ErrorsAndKeySurvives(t *testing.T) {
	s := newTestStore(t)
	client := rawClient(t)
	ctx := context.Background()
	g, code := newFreshGame(t, 7)
	t.Cleanup(func() { _ = s.Delete(ctx, g.ID()) })

	if err := s.Create(ctx, code, g); err != nil {
		t.Fatalf("Create: %v", err)
	}

	idKey := "jojo:game:id:" + g.ID().String()
	if err := client.Set(ctx, idKey, "not json at all", time.Minute).Err(); err != nil {
		t.Fatalf("corrupting payload: %v", err)
	}

	if _, err := s.Get(ctx, g.ID()); err == nil {
		t.Fatal("expected an error reading a corrupt payload, not a silent miss")
	}
	n, err := client.Exists(ctx, idKey).Result()
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if n != 1 {
		t.Error("corrupt key should survive a failed Get (fail-closed, kept for debugging)")
	}
}

// newFreshPublicGame is newFreshGame, but PUBLIC - so it's eligible for the
// jojo:game:public index.
func newFreshPublicGame(t *testing.T, seed byte) (*game.Game, string) {
	t.Helper()
	cfg, err := game.NewConfig(enums.Gauntlet, []enums.Manga{enums.Jojo}, []enums.Manga{enums.Jojo}, enums.Random, game.MaxGauntletPlayers, false, enums.Public, 30, game.PoolFilter{}, enums.Normal, game.DefaultSummaryDurationSeconds)
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}
	var idBytes [16]byte
	idBytes[0] = seed
	idBytes[1] = 0xBB
	host, err := game.NewHumanParticipant(game.ParticipantID{seed, 1}, user.UserID{seed, 2}, "host", game.TeamID{seed, 10})
	if err != nil {
		t.Fatalf("NewHumanParticipant: %v", err)
	}
	team, err := game.NewTeam(game.TeamID{seed, 10}, "Squad", 0)
	if err != nil {
		t.Fatalf("NewTeam: %v", err)
	}
	stage, err := game.NewStage(game.StageID{seed, 20}, enums.Jojo, 0, "Phantom Blood", "a test stage", "")
	if err != nil {
		t.Fatalf("NewStage: %v", err)
	}
	g, err := game.NewGame(game.GameID(idBytes), cfg, host, []*game.Team{team}, []game.Stage{stage})
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	code := game.GameID(idBytes).String()[:6]
	return g, code
}

func TestStore_ListPublic_ReturnsOnlyPublicGames(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	pub, pubCode := newFreshPublicGame(t, 10)
	priv, privCode := newFreshGame(t, 11)
	t.Cleanup(func() {
		_ = s.Delete(ctx, pub.ID())
		_ = s.Delete(ctx, priv.ID())
	})

	if err := s.Create(ctx, pubCode, pub); err != nil {
		t.Fatalf("Create(pub): %v", err)
	}
	if err := s.Create(ctx, privCode, priv); err != nil {
		t.Fatalf("Create(priv): %v", err)
	}

	games, err := s.ListPublic(ctx, 10)
	if err != nil {
		t.Fatalf("ListPublic: %v", err)
	}
	found := false
	for _, g := range games {
		if g.ID() == priv.ID() {
			t.Fatalf("expected the private game to be excluded from ListPublic")
		}
		if g.ID() == pub.ID() {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected the public game to be included in ListPublic")
	}
}

func TestStore_ListPublic_RemovedAfterDelete(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	pub, code := newFreshPublicGame(t, 12)

	if err := s.Create(ctx, code, pub); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := s.Delete(ctx, pub.ID()); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	games, err := s.ListPublic(ctx, 50)
	if err != nil {
		t.Fatalf("ListPublic: %v", err)
	}
	for _, g := range games {
		if g.ID() == pub.ID() {
			t.Fatalf("expected the deleted game to be removed from the public index")
		}
	}
}
