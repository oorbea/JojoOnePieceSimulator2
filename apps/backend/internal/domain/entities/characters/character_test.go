package characters_test

import (
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/characters"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

func mustCharacterID(t *testing.T, b byte) characters.CharacterID {
	t.Helper()
	var id characters.CharacterID
	id[15] = b
	return id
}

func TestNewCharacter_ValidatesFields(t *testing.T) {
	validID := mustCharacterID(t, 1)

	cases := []struct {
		name        string
		id          characters.CharacterID
		manga       enums.Manga
		charName    string
		rarity      enums.PowerRarity
		description string
		wantErr     bool
	}{
		{"nil id", characters.NilCharacterID, enums.Jojo, "Jotaro", enums.Rare, "desc", true},
		{"invalid manga", validID, enums.Manga(99), "Jotaro", enums.Rare, "desc", true},
		{"empty name", validID, enums.Jojo, "", enums.Rare, "desc", true},
		{"invalid rarity", validID, enums.Jojo, "Jotaro", enums.PowerRarity(99), "desc", true},
		{"empty description", validID, enums.Jojo, "Jotaro", enums.Rare, "", true},
		{"valid", validID, enums.Jojo, "Jotaro", enums.Rare, "desc", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := characters.NewCharacter(c.id, c.manga, c.charName, c.rarity, c.description, "")
			if (err != nil) != c.wantErr {
				t.Errorf("NewCharacter(%+v) err = %v, wantErr %v", c, err, c.wantErr)
			}
		})
	}
}

func TestCharacter_SetPictureRenditions(t *testing.T) {
	c, err := characters.NewCharacter(mustCharacterID(t, 1), enums.Jojo, "Jotaro", enums.Rare, "desc", "")
	if err != nil {
		t.Fatalf("NewCharacter: %v", err)
	}

	c.SetPictureRenditions("main.webp", "thumb.webp", "card.webp", "data:...", enums.PictureReady)

	if c.Picture() != "main.webp" || c.PictureThumb() != "thumb.webp" || c.PictureCard() != "card.webp" || c.PictureLqip() != "data:..." {
		t.Errorf("picture fields not set: %+v", c)
	}
	if c.PictureStatus() != enums.PictureReady {
		t.Errorf("PictureStatus() = %v, want READY", c.PictureStatus())
	}
}

func TestCharacter_SetMediaID(t *testing.T) {
	c, err := characters.NewCharacter(mustCharacterID(t, 1), enums.Jojo, "Jotaro", enums.Rare, "desc", "")
	if err != nil {
		t.Fatalf("NewCharacter: %v", err)
	}
	if c.PictureMediaID() != "" {
		t.Fatalf("PictureMediaID() = %q, want empty before SetMediaID", c.PictureMediaID())
	}
	c.SetMediaID("group-1")
	if c.PictureMediaID() != "group-1" {
		t.Errorf("PictureMediaID() = %q, want group-1", c.PictureMediaID())
	}
}
