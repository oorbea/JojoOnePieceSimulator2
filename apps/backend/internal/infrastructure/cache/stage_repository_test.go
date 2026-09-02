package cache_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	infracache "github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/cache"
)

// countingStageRepository is the Stage counterpart of
// countingStandRepository: an in-memory store counting calls per method, so
// a test can assert a cache hit never reached it. It satisfies
// infracache.StageStore, i.e. both the admin CRUD port and the gameplay
// catalogue port, because the decorator serves both.
type countingStageRepository struct {
	mu             sync.Mutex
	stages         map[game.StageID]game.Stage
	stagesCalls    int
	listCalls      int
	filterCalls    int
	findByIDCalls  int
	translateCalls int
	updatePicCalls int
	// saveErr, when set, makes Save fail - for the "a failed write must not
	// invalidate" case.
	saveErr error
}

func newCountingStageRepository() *countingStageRepository {
	return &countingStageRepository{stages: make(map[game.StageID]game.Stage)}
}

func (r *countingStageRepository) Stages(_ context.Context, manga enums.Manga) ([]game.Stage, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stagesCalls++
	var results []game.Stage
	for _, s := range r.stages {
		if s.Manga() == manga {
			results = append(results, s)
		}
	}
	return results, nil
}

func (r *countingStageRepository) List(_ context.Context, _ enums.Locale) ([]game.Stage, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.listCalls++
	all := make([]game.Stage, 0, len(r.stages))
	for _, s := range r.stages {
		all = append(all, s)
	}
	return all, nil
}

func (r *countingStageRepository) Filter(_ context.Context, filters ports.StageFilters, _ enums.Locale) ([]game.Stage, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.filterCalls++
	var results []game.Stage
	for _, s := range r.stages {
		if filters.Manga != nil && s.Manga() != *filters.Manga {
			continue
		}
		results = append(results, s)
	}
	return results, nil
}

func (r *countingStageRepository) FindByID(_ context.Context, id game.StageID, _ enums.Locale) (game.Stage, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.findByIDCalls++
	s, ok := r.stages[id]
	if !ok {
		return game.Stage{}, ports.ErrStageNotFound
	}
	return s, nil
}

func (r *countingStageRepository) Save(_ context.Context, s game.Stage, _ ports.StageTranslations) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.saveErr != nil {
		return r.saveErr
	}
	r.stages[s.ID()] = s
	return nil
}

func (r *countingStageRepository) Delete(_ context.Context, id game.StageID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.stages, id)
	return nil
}

func (r *countingStageRepository) Translations(_ context.Context, id game.StageID) (ports.StageTranslations, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.translateCalls++
	s, ok := r.stages[id]
	if !ok {
		return nil, ports.ErrStageNotFound
	}
	return ports.StageTranslations{enums.EnGB: s.Description()}, nil
}

func (r *countingStageRepository) UpdatePicture(_ context.Context, id game.StageID, main, thumb *string, status enums.PictureStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.updatePicCalls++
	s, ok := r.stages[id]
	if !ok {
		return ports.ErrStageNotFound
	}
	newMain, newThumb := s.Picture(), s.PictureThumb()
	if main != nil {
		newMain = *main
	}
	if thumb != nil {
		newThumb = *thumb
	}
	s.SetPictureRenditions(newMain, newThumb, status)
	r.stages[id] = s
	return nil
}

var _ infracache.StageStore = (*countingStageRepository)(nil)

func newTestStage(t *testing.T, name string, manga enums.Manga, order int) game.Stage {
	t.Helper()
	var id game.StageID
	id[15] = byte(len(name)) // cheap distinct id per test
	id[14] = byte(order)
	stage, err := game.NewStage(id, manga, order, name, "desc", "")
	if err != nil {
		t.Fatalf("NewStage: %v", err)
	}
	return stage
}

func newStageRepo(t *testing.T) (*infracache.StageRepository, *countingStageRepository) {
	t.Helper()
	next := newCountingStageRepository()
	return infracache.NewStageRepository(next, newFakeCache(), time.Minute, time.Second), next
}

func seedStage(t *testing.T, next *countingStageRepository, stage game.Stage) {
	t.Helper()
	if err := next.Save(context.Background(), stage, ports.StageTranslations{enums.EnGB: stage.Description()}); err != nil {
		t.Fatalf("seeding stage: %v", err)
	}
}

func TestStageRepository_FindByID_CachesOnMiss(t *testing.T) {
	repo, next := newStageRepo(t)
	stage := newTestStage(t, "Stardust Crusaders", enums.Jojo, 3)
	seedStage(t, next, stage)
	ctx := context.Background()

	first, err := repo.FindByID(ctx, stage.ID(), enums.EnGB)
	if err != nil {
		t.Fatalf("first FindByID: %v", err)
	}
	if first.Name() != stage.Name() {
		t.Fatalf("first FindByID name = %q, want %q", first.Name(), stage.Name())
	}
	second, err := repo.FindByID(ctx, stage.ID(), enums.EnGB)
	if err != nil {
		t.Fatalf("second FindByID: %v", err)
	}
	if second.Name() != stage.Name() || second.Manga() != stage.Manga() || second.Order() != stage.Order() {
		t.Errorf("cached FindByID = %+v, want a faithful round trip of %+v", second, stage)
	}

	if next.findByIDCalls != 1 {
		t.Errorf("underlying FindByID calls = %d, want 1 (second call should hit cache)", next.findByIDCalls)
	}
}

func TestStageRepository_FindByID_NotFoundIsCachedAsTombstone(t *testing.T) {
	repo, next := newStageRepo(t)
	ctx := context.Background()

	var missing game.StageID
	missing[15] = 99

	got, err := repo.FindByID(ctx, missing, enums.EnGB)
	if !errors.Is(err, ports.ErrStageNotFound) {
		t.Fatalf("first FindByID err = %v, want ErrStageNotFound", err)
	}
	got, err = repo.FindByID(ctx, missing, enums.EnGB)
	if !errors.Is(err, ports.ErrStageNotFound) {
		t.Fatalf("second FindByID err = %v, want ErrStageNotFound", err)
	}
	if got != (game.Stage{}) {
		t.Errorf("tombstone hit returned %+v, want the zero Stage", got)
	}

	if next.findByIDCalls != 1 {
		t.Errorf("underlying FindByID calls = %d, want 1 (404 should be cached)", next.findByIDCalls)
	}
}

func TestStageRepository_Save_InvalidatesCache(t *testing.T) {
	repo, next := newStageRepo(t)
	seedStage(t, next, newTestStage(t, "Alabasta", enums.OnePiece, 1))
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		if _, err := repo.List(ctx, enums.EnGB); err != nil {
			t.Fatalf("List #%d: %v", i, err)
		}
	}
	if next.listCalls != 1 {
		t.Fatalf("underlying List calls before Save = %d, want 1", next.listCalls)
	}

	other := newTestStage(t, "Skypiea", enums.OnePiece, 2)
	if err := repo.Save(ctx, other, ports.StageTranslations{enums.EnGB: other.Description()}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if _, err := repo.List(ctx, enums.EnGB); err != nil {
		t.Fatalf("List after Save: %v", err)
	}
	if next.listCalls != 2 {
		t.Errorf("underlying List calls after Save = %d, want 2 (Save should invalidate)", next.listCalls)
	}
}

func TestStageRepository_Save_InvalidatesTheCatalogue(t *testing.T) {
	repo, next := newStageRepo(t)
	seedStage(t, next, newTestStage(t, "Alabasta", enums.OnePiece, 1))
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		if _, err := repo.Stages(ctx, enums.OnePiece); err != nil {
			t.Fatalf("Stages #%d: %v", i, err)
		}
	}
	if next.stagesCalls != 1 {
		t.Fatalf("underlying Stages calls before Save = %d, want 1", next.stagesCalls)
	}

	other := newTestStage(t, "Skypiea", enums.OnePiece, 2)
	if err := repo.Save(ctx, other, ports.StageTranslations{enums.EnGB: other.Description()}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if _, err := repo.Stages(ctx, enums.OnePiece); err != nil {
		t.Fatalf("Stages after Save: %v", err)
	}
	if next.stagesCalls != 2 {
		t.Errorf("underlying Stages calls after Save = %d, want 2 (a write must reach the gameplay catalogue too)", next.stagesCalls)
	}
}

func TestStageRepository_UpdatePicture_InvalidatesCache(t *testing.T) {
	repo, next := newStageRepo(t)
	stage := newTestStage(t, "Alabasta", enums.OnePiece, 1)
	seedStage(t, next, stage)
	ctx := context.Background()

	if _, err := repo.FindByID(ctx, stage.ID(), enums.EnGB); err != nil {
		t.Fatalf("FindByID before UpdatePicture: %v", err)
	}

	// The background picture worker's path: publishing READY once a
	// transcode finishes must be visible to readers immediately.
	main, thumb := "stages/main.webp", "stages/thumb.webp"
	if err := repo.UpdatePicture(ctx, stage.ID(), &main, &thumb, enums.PictureReady); err != nil {
		t.Fatalf("UpdatePicture: %v", err)
	}

	got, err := repo.FindByID(ctx, stage.ID(), enums.EnGB)
	if err != nil {
		t.Fatalf("FindByID after UpdatePicture: %v", err)
	}
	if next.findByIDCalls != 2 {
		t.Errorf("underlying FindByID calls after UpdatePicture = %d, want 2 (UpdatePicture should invalidate)", next.findByIDCalls)
	}
	if got.Picture() != main || got.PictureThumb() != thumb || got.PictureStatus() != enums.PictureReady {
		t.Errorf("post-invalidation read = (%q, %q, %v), want (%q, %q, %v)",
			got.Picture(), got.PictureThumb(), got.PictureStatus(), main, thumb, enums.PictureReady)
	}
}

func TestStageRepository_FailedSave_DoesNotInvalidate(t *testing.T) {
	repo, next := newStageRepo(t)
	seedStage(t, next, newTestStage(t, "Alabasta", enums.OnePiece, 1))
	ctx := context.Background()

	if _, err := repo.List(ctx, enums.EnGB); err != nil {
		t.Fatalf("List: %v", err)
	}

	boom := errors.New("write failed")
	next.saveErr = boom
	other := newTestStage(t, "Skypiea", enums.OnePiece, 2)
	if err := repo.Save(ctx, other, ports.StageTranslations{enums.EnGB: other.Description()}); !errors.Is(err, boom) {
		t.Fatalf("Save err = %v, want %v", err, boom)
	}
	next.saveErr = nil

	if _, err := repo.List(ctx, enums.EnGB); err != nil {
		t.Fatalf("List after failed Save: %v", err)
	}
	if next.listCalls != 1 {
		t.Errorf("underlying List calls after a failed Save = %d, want 1 (a failed write must not invalidate)", next.listCalls)
	}
}

func TestStageRepository_KeysNeverCrossLocales(t *testing.T) {
	repo, next := newStageRepo(t)
	stage := newTestStage(t, "Alabasta", enums.OnePiece, 1)
	seedStage(t, next, stage)
	ctx := context.Background()

	for _, locale := range []enums.Locale{enums.EsES, enums.CaES, enums.EnGB} {
		if _, err := repo.FindByID(ctx, stage.ID(), locale); err != nil {
			t.Fatalf("FindByID(%v): %v", locale, err)
		}
	}
	if next.findByIDCalls != 3 {
		t.Fatalf("underlying FindByID calls = %d, want 3 (one per locale)", next.findByIDCalls)
	}

	if _, err := repo.FindByID(ctx, stage.ID(), enums.EsES); err != nil {
		t.Fatalf("repeated FindByID(es-ES): %v", err)
	}
	if next.findByIDCalls != 3 {
		t.Errorf("underlying FindByID calls after repeating es-ES = %d, want 3 (that locale is cached)", next.findByIDCalls)
	}

	for _, locale := range []enums.Locale{enums.EsES, enums.CaES, enums.EnGB} {
		if _, err := repo.List(ctx, locale); err != nil {
			t.Fatalf("List(%v): %v", locale, err)
		}
	}
	if next.listCalls != 3 {
		t.Errorf("underlying List calls = %d, want 3 (one per locale)", next.listCalls)
	}
}

func TestStageRepository_Filter_DistinctSearchesGetDistinctKeys(t *testing.T) {
	repo, next := newStageRepo(t)
	seedStage(t, next, newTestStage(t, "Alabasta", enums.OnePiece, 1))
	ctx := context.Background()

	alabasta, skypiea := "alabasta", "skypiea"
	if _, err := repo.Filter(ctx, ports.StageFilters{Search: &alabasta}, enums.EnGB); err != nil {
		t.Fatalf("Filter(alabasta): %v", err)
	}
	if _, err := repo.Filter(ctx, ports.StageFilters{Search: &skypiea}, enums.EnGB); err != nil {
		t.Fatalf("Filter(skypiea): %v", err)
	}
	if next.filterCalls != 2 {
		t.Fatalf("underlying Filter calls = %d, want 2 (different searches must not share a slot)", next.filterCalls)
	}

	if _, err := repo.Filter(ctx, ports.StageFilters{Search: &alabasta}, enums.EnGB); err != nil {
		t.Fatalf("repeated Filter(alabasta): %v", err)
	}
	if next.filterCalls != 2 {
		t.Errorf("underlying Filter calls after repeating the first search = %d, want 2 (it is cached)", next.filterCalls)
	}
}

func TestStageRepository_Filter_DistinctMangasGetDistinctKeys(t *testing.T) {
	repo, next := newStageRepo(t)
	seedStage(t, next, newTestStage(t, "Alabasta", enums.OnePiece, 1))
	seedStage(t, next, newTestStage(t, "Stardust", enums.Jojo, 3))
	ctx := context.Background()

	jojo, onePiece := enums.Jojo, enums.OnePiece
	if _, err := repo.Filter(ctx, ports.StageFilters{Manga: &jojo}, enums.EnGB); err != nil {
		t.Fatalf("Filter(JOJO): %v", err)
	}
	if _, err := repo.Filter(ctx, ports.StageFilters{Manga: &onePiece}, enums.EnGB); err != nil {
		t.Fatalf("Filter(ONE_PIECE): %v", err)
	}
	if next.filterCalls != 2 {
		t.Fatalf("underlying Filter calls = %d, want 2 (different mangas must not share a slot)", next.filterCalls)
	}

	if _, err := repo.Filter(ctx, ports.StageFilters{Manga: &jojo}, enums.EnGB); err != nil {
		t.Fatalf("repeated Filter(JOJO): %v", err)
	}
	if next.filterCalls != 2 {
		t.Errorf("underlying Filter calls after repeating JOJO = %d, want 2 (it is cached)", next.filterCalls)
	}
}

func TestStageRepository_Stages_CachesPerManga(t *testing.T) {
	repo, next := newStageRepo(t)
	seedStage(t, next, newTestStage(t, "Alabasta", enums.OnePiece, 1))
	seedStage(t, next, newTestStage(t, "Stardust", enums.Jojo, 3))
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		if _, err := repo.Stages(ctx, enums.Jojo); err != nil {
			t.Fatalf("Stages(JOJO) #%d: %v", i, err)
		}
	}
	if next.stagesCalls != 1 {
		t.Fatalf("underlying Stages calls = %d, want 1 (the second read is cached)", next.stagesCalls)
	}

	got, err := repo.Stages(ctx, enums.OnePiece)
	if err != nil {
		t.Fatalf("Stages(ONE_PIECE): %v", err)
	}
	if next.stagesCalls != 2 {
		t.Errorf("underlying Stages calls = %d, want 2 (a different manga is a different key)", next.stagesCalls)
	}
	if len(got) != 1 || got[0].Manga() != enums.OnePiece {
		t.Errorf("Stages(ONE_PIECE) = %+v, want exactly the One Piece stage", got)
	}
}

// TestStageRepository_StagesAndFilterDoNotShareASlot pins a deliberate
// design decision: IStageCatalog.Stages(manga) and an admin
// Filter{Manga: manga} happen to run the same query today, but they are
// separate contracts (gameplay vs admin, fixed EnGB vs caller locale) and
// must never answer each other from one cache slot.
func TestStageRepository_StagesAndFilterDoNotShareASlot(t *testing.T) {
	repo, next := newStageRepo(t)
	seedStage(t, next, newTestStage(t, "Alabasta", enums.OnePiece, 1))
	ctx := context.Background()

	if _, err := repo.Stages(ctx, enums.OnePiece); err != nil {
		t.Fatalf("Stages: %v", err)
	}

	onePiece := enums.OnePiece
	if _, err := repo.Filter(ctx, ports.StageFilters{Manga: &onePiece}, enums.EnGB); err != nil {
		t.Fatalf("Filter: %v", err)
	}
	if next.filterCalls != 1 {
		t.Errorf("underlying Filter calls = %d, want 1 (Filter must not be answered by the catalogue entry)", next.filterCalls)
	}

	if _, err := repo.Stages(ctx, enums.OnePiece); err != nil {
		t.Fatalf("second Stages: %v", err)
	}
	if next.stagesCalls != 1 {
		t.Errorf("underlying Stages calls = %d, want 1 (the catalogue entry survived the Filter read)", next.stagesCalls)
	}
}

func TestStageRepository_Translations_IsNeverCached(t *testing.T) {
	repo, next := newStageRepo(t)
	stage := newTestStage(t, "Alabasta", enums.OnePiece, 1)
	seedStage(t, next, stage)
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		if _, err := repo.Translations(ctx, stage.ID()); err != nil {
			t.Fatalf("Translations #%d: %v", i, err)
		}
	}
	if next.translateCalls != 2 {
		t.Errorf("underlying Translations calls = %d, want 2 (admin edit forms need fresh reads)", next.translateCalls)
	}
}
