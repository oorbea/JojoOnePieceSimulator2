package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestWikiTitle(t *testing.T) {
	cases := map[string]string{
		"Hira Hira no mi":               "Hira Hira no Mi",
		"Inu Inu no mi: Model Hound":    "Inu Inu no Mi, Model: Hound",
		"Tori Tori no mi: Model Falcon": "Tori Tori no Mi, Model: Falcon",
	}
	for in, want := range cases {
		if got := wikiTitle(in); got != want {
			t.Errorf("wikiTitle(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestImageStem(t *testing.T) {
	if got := imageStem("Inu Inu no mi: Model Hound"); got != "Inu_Inu_no_Mi" {
		t.Errorf("imageStem = %q", got)
	}
}

func TestMissingFruitsExcludesPicturedAndReservedRarities(t *testing.T) {
	rows := []fruitRow{
		{fruit: fruit{ID: "1", Name: "A", Rarity: "COMMON"}},
		{fruit: fruit{ID: "2", Name: "B", Rarity: "COMMON"}, Picture: "/x"},
		{fruit: fruit{ID: "3", Name: "C", Rarity: "legendary"}},
		{fruit: fruit{ID: "4", Name: "D", Rarity: "RARE"}},
	}
	got := missingFruits(rows, splitSet("EPIC,LEGENDARY,MYTHICAL"))
	if len(got) != 2 || got[0].ID != "1" || got[1].ID != "4" {
		t.Fatalf("got %+v", got)
	}
}

func TestIsPlaceholderAndImageFile(t *testing.T) {
	if !isPlaceholder("https://x/NoPicAvailable.png/revision") {
		t.Error("placeholder not detected")
	}
	if !isImageFile("A_b.PNG") || isImageFile("clip.ogg") {
		t.Error("isImageFile wrong")
	}
}

func TestReviewKeepsFruitIDInCandidateRadios(t *testing.T) {
	var buf bytes.Buffer
	err := reviewTmpl.Execute(&buf, []candidateEntry{{
		fruit:      fruit{ID: "abc", Name: "Hira Hira no mi", Rarity: "COMMON"},
		Candidates: []candidate{{File: "candidates/abc/1.png", Page: "Hira Hira no Mi"}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `name="abc" value="candidates/abc/1.png"`) {
		t.Fatalf("radio missing fruit id:\n%s", buf.String())
	}
}
