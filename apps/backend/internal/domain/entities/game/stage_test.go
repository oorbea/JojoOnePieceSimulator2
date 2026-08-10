package game_test

import (
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

func TestInterleave_MixedMangaWithLeftovers(t *testing.T) {
	jojo := []game.Stage{
		mustStage(t, 1, enums.Jojo, 0, "Phantom Blood"),
		mustStage(t, 2, enums.Jojo, 1, "Battle Tendency"),
	}
	onePiece := []game.Stage{
		mustStage(t, 10, enums.OnePiece, 0, "East Blue"),
		mustStage(t, 11, enums.OnePiece, 1, "Alabasta"),
		mustStage(t, 12, enums.OnePiece, 2, "Skypiea"),
	}

	result := game.Interleave(map[enums.Manga][]game.Stage{
		enums.Jojo:     jojo,
		enums.OnePiece: onePiece,
	})

	want := []string{"Phantom Blood", "East Blue", "Battle Tendency", "Alabasta", "Skypiea"}
	if len(result) != len(want) {
		t.Fatalf("expected %d stages, got %d: %+v", len(want), len(result), result)
	}
	for i, name := range want {
		if result[i].Name() != name {
			t.Fatalf("stage %d: expected %q, got %q", i, name, result[i].Name())
		}
	}
}

func TestInterleave_SingleManga(t *testing.T) {
	jojo := []game.Stage{mustStage(t, 1, enums.Jojo, 0, "Phantom Blood")}
	result := game.Interleave(map[enums.Manga][]game.Stage{enums.Jojo: jojo})
	if len(result) != 1 || result[0].Name() != "Phantom Blood" {
		t.Fatalf("expected single-manga passthrough, got %+v", result)
	}
}

func TestInterleave_EmptyCatalog(t *testing.T) {
	result := game.Interleave(map[enums.Manga][]game.Stage{})
	if len(result) != 0 {
		t.Fatalf("expected no stages, got %+v", result)
	}
}

func TestInterleave_UnsortedInputIsSortedByOrder(t *testing.T) {
	jojo := []game.Stage{
		mustStage(t, 2, enums.Jojo, 1, "Battle Tendency"),
		mustStage(t, 1, enums.Jojo, 0, "Phantom Blood"),
	}
	result := game.Interleave(map[enums.Manga][]game.Stage{enums.Jojo: jojo})
	if len(result) != 2 || result[0].Name() != "Phantom Blood" || result[1].Name() != "Battle Tendency" {
		t.Fatalf("expected input sorted by Order regardless of catalog order, got %+v", result)
	}
}
