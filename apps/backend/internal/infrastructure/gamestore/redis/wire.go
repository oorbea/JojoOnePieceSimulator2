// Package redis adapts ports.IGameStore onto Redis via go-redis. Unlike
// infrastructure/cache/redis (which is the only package allowed to import
// the Redis client for caching), this package holds the source of truth for
// live lobbies/matches and is fail-closed: every error surfaces to the
// caller, nothing is silently treated as a miss. See store.go's package doc
// for the key scheme and fail-closed contract.
package redis

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/powersnap"
)

// snapshotVersion is bumped whenever wireGame's shape changes in a
// backward-incompatible way. A version mismatch on read is a hard error
// (see envelope.hydrate), never a silent zero value: a live lobby that
// cannot be read after a deploy is a visible GAME_NOT_FOUND, not a
// corrupted-looking Game. The TTL (game_service.go's Config.GameLobbyTTL,
// default 2h) means any such breakage self-heals within that window.
//
// The lobby-management tanda (Locked, Config.Visibility/VotingWindowSeconds/
// PoolFilter) deliberately did NOT bump this: every new field is
// `omitempty` and game.Restore already defaults an absent Visibility to
// PRIVATE, an absent/zero VotingWindowSeconds to
// game.DefaultVotingWindowSeconds, an absent PoolFilter to "no
// restriction", and an absent Locked to false - so a lobby saved by the
// previous build decodes cleanly instead of dying with a version mismatch
// mid-TTL. Only a change with no safe default would justify bumping to 2.
const snapshotVersion = 1

// envelope wraps a wireGame with a version tag and a last-write timestamp
// (kept for observability/debugging; DeleteExpired does not use it - see
// store.go).
type envelope struct {
	Version   int       `json:"v"`
	UpdatedAt time.Time `json:"updatedAt"`
	Game      wireGame  `json:"game"`
}

// wireGame mirrors game.Snapshot field-for-field with JSON tags. It is kept
// separate from the domain type on purpose: the domain's ports.IGameStore
// doc explicitly says it carries no snapshot/rehydration API, and JSON tags
// are a wire-format concern that belongs in infrastructure (same reasoning
// as api/dto and infrastructure/cache's snapshot types).
type wireGame struct {
	ID           [16]byte          `json:"id"`
	State        string            `json:"state"`
	HostID       [16]byte          `json:"hostId"`
	Locked       bool              `json:"locked,omitempty"`
	Config       wireConfig        `json:"config"`
	Participants []wireParticipant `json:"participants"`
	Teams        []wireTeam        `json:"teams"`
	Stages       []wireStage       `json:"stages"`
	Rounds       []wireRound       `json:"rounds"`
}

type wireConfig struct {
	Mode                string         `json:"mode"`
	Mangas              []string       `json:"mangas"`
	AbilitySource       string         `json:"abilitySource"`
	TeamSize            int            `json:"teamSize"`
	AllowBots           bool           `json:"allowBots"`
	Visibility          string         `json:"visibility,omitempty"`
	VotingWindowSeconds int            `json:"votingWindowSeconds,omitempty"`
	PoolFilter          wirePoolFilter `json:"poolFilter,omitempty"`
}

type wirePoolFilter struct {
	StandRarities []string   `json:"standRarities,omitempty"`
	FruitRarities []string   `json:"fruitRarities,omitempty"`
	FruitTypes    []string   `json:"fruitTypes,omitempty"`
	Banned        [][16]byte `json:"banned,omitempty"`
}

// AvatarThumbKey/GooglePicture are additive `omitempty` fields with a safe
// empty-string default - see snapshotVersion's doc for why that means no
// version bump: a payload written before this field existed just decodes
// with both as "", exactly like a participant who never called SetAvatar.
type wireParticipant struct {
	ID             [16]byte     `json:"id"`
	UserID         *[16]byte    `json:"userId,omitempty"`
	DisplayName    string       `json:"displayName"`
	TeamID         [16]byte     `json:"teamId"`
	Kind           string       `json:"kind"`
	Connected      bool         `json:"connected"`
	Loadout        *wireLoadout `json:"loadout,omitempty"`
	AvatarThumbKey string       `json:"avatarThumbKey,omitempty"`
	GooglePicture  string       `json:"googlePicture,omitempty"`
}

type wireTeam struct {
	ID      [16]byte     `json:"id"`
	Name    string       `json:"name"`
	Color   uint32       `json:"color"`
	Members []([16]byte) `json:"members"`
}

type wireStage struct {
	ID            [16]byte `json:"id"`
	Manga         string   `json:"manga"`
	Order         int      `json:"order"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Picture       string   `json:"picture"`
	PictureThumb  string   `json:"pictureThumb"`
	PictureStatus string   `json:"pictureStatus"`
}

type wireRound struct {
	Index        int              `json:"index"`
	Stage        wireStage        `json:"stage"`
	Ballot       wireBallot       `json:"ballot"`
	TiebreakUsed bool             `json:"tiebreakUsed"`
	Result       *wireRoundResult `json:"result,omitempty"`
}

type wireBallot struct {
	Options []string   `json:"options"`
	Votes   []wireVote `json:"votes"`
}

type wireVote struct {
	ParticipantID [16]byte `json:"participantId"`
	Option        string   `json:"option"`
}

type wireRoundResult struct {
	Winner            string `json:"winner"`
	DecidedByCoinFlip bool   `json:"decidedByCoinFlip"`
}

type wireLoadout struct {
	Stand           *powersnap.StandSnapshot      `json:"stand,omitempty"`
	DevilFruit      *powersnap.DevilFruitSnapshot `json:"devilFruit,omitempty"`
	Spin            string                        `json:"spin"`
	Hamon           string                        `json:"hamon"`
	FruitMastery    string                        `json:"fruitMastery"`
	ArmamentHaki    string                        `json:"armamentHaki"`
	ObservationHaki string                        `json:"observationHaki"`
	ConquerorHaki   string                        `json:"conquerorHaki"`
	PhysicalForm    string                        `json:"physicalForm"`
}

// toWire converts a domain Snapshot to its wire form.
func toWire(s game.Snapshot) wireGame {
	w := wireGame{
		ID:     s.ID,
		State:  s.State,
		HostID: s.HostID,
		Locked: s.Locked,
		Config: wireConfig{
			Mode:                s.Config.Mode,
			Mangas:              append([]string(nil), s.Config.Mangas...),
			AbilitySource:       s.Config.AbilitySource,
			TeamSize:            s.Config.TeamSize,
			AllowBots:           s.Config.AllowBots,
			Visibility:          s.Config.Visibility,
			VotingWindowSeconds: s.Config.VotingWindowSeconds,
			PoolFilter:          toWirePoolFilter(s.Config.PoolFilter),
		},
		Participants: make([]wireParticipant, 0, len(s.Participants)),
		Teams:        make([]wireTeam, 0, len(s.Teams)),
		Stages:       make([]wireStage, 0, len(s.Stages)),
		Rounds:       make([]wireRound, 0, len(s.Rounds)),
	}

	for _, p := range s.Participants {
		wp := wireParticipant{
			ID:             p.ID,
			DisplayName:    p.DisplayName,
			TeamID:         p.TeamID,
			Kind:           p.Kind,
			Connected:      p.Connected,
			AvatarThumbKey: p.AvatarThumbKey,
			GooglePicture:  p.GooglePicture,
		}
		if p.UserID != nil {
			id := [16]byte(*p.UserID)
			wp.UserID = &id
		}
		if p.Loadout != nil {
			wl := toWireLoadout(*p.Loadout)
			wp.Loadout = &wl
		}
		w.Participants = append(w.Participants, wp)
	}

	for _, t := range s.Teams {
		members := make([][16]byte, len(t.Members))
		for i, m := range t.Members {
			members[i] = [16]byte(m)
		}
		w.Teams = append(w.Teams, wireTeam{ID: t.ID, Name: t.Name, Color: t.Color, Members: members})
	}

	for _, st := range s.Stages {
		w.Stages = append(w.Stages, toWireStage(st))
	}

	for _, r := range s.Rounds {
		wr := wireRound{
			Index:        r.Index,
			Stage:        toWireStage(r.Stage),
			Ballot:       toWireBallot(r.Ballot),
			TiebreakUsed: r.TiebreakUsed,
		}
		if r.Result != nil {
			wr.Result = &wireRoundResult{Winner: string(r.Result.Winner), DecidedByCoinFlip: r.Result.DecidedByCoinFlip}
		}
		w.Rounds = append(w.Rounds, wr)
	}

	return w
}

func toWirePoolFilter(f game.PoolFilterSnapshot) wirePoolFilter {
	banned := make([][16]byte, len(f.Banned))
	for i, id := range f.Banned {
		banned[i] = [16]byte(id)
	}
	return wirePoolFilter{
		StandRarities: append([]string(nil), f.StandRarities...),
		FruitRarities: append([]string(nil), f.FruitRarities...),
		FruitTypes:    append([]string(nil), f.FruitTypes...),
		Banned:        banned,
	}
}

func fromWirePoolFilter(w wirePoolFilter) game.PoolFilterSnapshot {
	banned := make([]powers.PowerID, len(w.Banned))
	for i, id := range w.Banned {
		banned[i] = powers.PowerID(id)
	}
	return game.PoolFilterSnapshot{
		StandRarities: append([]string(nil), w.StandRarities...),
		FruitRarities: append([]string(nil), w.FruitRarities...),
		FruitTypes:    append([]string(nil), w.FruitTypes...),
		Banned:        banned,
	}
}

func toWireStage(st game.StageSnapshot) wireStage {
	return wireStage{
		ID: st.ID, Manga: st.Manga, Order: st.Order, Name: st.Name,
		Description: st.Description, Picture: st.Picture, PictureThumb: st.PictureThumb,
		PictureStatus: st.PictureStatus,
	}
}

func toWireBallot(b game.BallotSnapshot) wireBallot {
	options := make([]string, len(b.Options))
	for i, o := range b.Options {
		options[i] = string(o)
	}
	votes := make([]wireVote, len(b.Votes))
	for i, v := range b.Votes {
		votes[i] = wireVote{ParticipantID: v.ParticipantID, Option: string(v.Option)}
	}
	return wireBallot{Options: options, Votes: votes}
}

func toWireLoadout(l game.LoadoutSnapshot) wireLoadout {
	wl := wireLoadout{
		Spin:            l.Spin,
		Hamon:           l.Hamon,
		FruitMastery:    l.FruitMastery,
		ArmamentHaki:    l.ArmamentHaki,
		ObservationHaki: l.ObservationHaki,
		ConquerorHaki:   l.ConquerorHaki,
		PhysicalForm:    l.PhysicalForm,
	}
	if l.Stand != nil {
		s := powersnap.OfStand(l.Stand)
		wl.Stand = &s
	}
	if l.DevilFruit != nil {
		f := powersnap.OfDevilFruit(l.DevilFruit)
		wl.DevilFruit = &f
	}
	return wl
}

// fromWire converts a wire payload back to a domain Snapshot.
func fromWire(w wireGame) game.Snapshot {
	s := game.Snapshot{
		ID:     w.ID,
		State:  w.State,
		HostID: w.HostID,
		Locked: w.Locked,
		Config: game.ConfigSnapshot{
			Mode:                w.Config.Mode,
			Mangas:              append([]string(nil), w.Config.Mangas...),
			AbilitySource:       w.Config.AbilitySource,
			TeamSize:            w.Config.TeamSize,
			AllowBots:           w.Config.AllowBots,
			Visibility:          w.Config.Visibility,
			VotingWindowSeconds: w.Config.VotingWindowSeconds,
			PoolFilter:          fromWirePoolFilter(w.Config.PoolFilter),
		},
		Participants: make([]game.ParticipantSnapshot, 0, len(w.Participants)),
		Teams:        make([]game.TeamSnapshot, 0, len(w.Teams)),
		Stages:       make([]game.StageSnapshot, 0, len(w.Stages)),
		Rounds:       make([]game.RoundSnapshot, 0, len(w.Rounds)),
	}

	for _, wp := range w.Participants {
		p := game.ParticipantSnapshot{
			ID:             wp.ID,
			DisplayName:    wp.DisplayName,
			TeamID:         wp.TeamID,
			Kind:           wp.Kind,
			Connected:      wp.Connected,
			AvatarThumbKey: wp.AvatarThumbKey,
			GooglePicture:  wp.GooglePicture,
		}
		if wp.UserID != nil {
			uid := user.UserID(*wp.UserID)
			p.UserID = &uid
		}
		if wp.Loadout != nil {
			ls := fromWireLoadout(*wp.Loadout)
			p.Loadout = &ls
		}
		s.Participants = append(s.Participants, p)
	}

	for _, wt := range w.Teams {
		members := make([]game.ParticipantID, len(wt.Members))
		for i, m := range wt.Members {
			members[i] = game.ParticipantID(m)
		}
		s.Teams = append(s.Teams, game.TeamSnapshot{ID: wt.ID, Name: wt.Name, Color: wt.Color, Members: members})
	}

	for _, wst := range w.Stages {
		s.Stages = append(s.Stages, fromWireStage(wst))
	}

	for _, wr := range w.Rounds {
		r := game.RoundSnapshot{
			Index:        wr.Index,
			Stage:        fromWireStage(wr.Stage),
			Ballot:       fromWireBallot(wr.Ballot),
			TiebreakUsed: wr.TiebreakUsed,
		}
		if wr.Result != nil {
			r.Result = &game.RoundResultSnapshot{
				Winner:            game.OptionID(wr.Result.Winner),
				DecidedByCoinFlip: wr.Result.DecidedByCoinFlip,
			}
		}
		s.Rounds = append(s.Rounds, r)
	}

	return s
}

func fromWireStage(w wireStage) game.StageSnapshot {
	return game.StageSnapshot{
		ID: w.ID, Manga: w.Manga, Order: w.Order, Name: w.Name,
		Description: w.Description, Picture: w.Picture, PictureThumb: w.PictureThumb,
		PictureStatus: w.PictureStatus,
	}
}

func fromWireBallot(w wireBallot) game.BallotSnapshot {
	options := make([]game.OptionID, len(w.Options))
	for i, o := range w.Options {
		options[i] = game.OptionID(o)
	}
	votes := make([]game.VoteSnapshot, len(w.Votes))
	for i, v := range w.Votes {
		votes[i] = game.VoteSnapshot{ParticipantID: v.ParticipantID, Option: game.OptionID(v.Option)}
	}
	return game.BallotSnapshot{Options: options, Votes: votes}
}

func fromWireLoadout(w wireLoadout) game.LoadoutSnapshot {
	ls := game.LoadoutSnapshot{
		Spin:            w.Spin,
		Hamon:           w.Hamon,
		FruitMastery:    w.FruitMastery,
		ArmamentHaki:    w.ArmamentHaki,
		ObservationHaki: w.ObservationHaki,
		ConquerorHaki:   w.ConquerorHaki,
		PhysicalForm:    w.PhysicalForm,
	}
	if w.Stand != nil {
		if s, err := w.Stand.Hydrate(); err == nil {
			ls.Stand = s
		}
	}
	if w.DevilFruit != nil {
		if f, err := w.DevilFruit.Hydrate(); err == nil {
			ls.DevilFruit = f
		}
	}
	return ls
}

// encode marshals g's Snapshot into the versioned envelope's JSON bytes.
func encode(g *game.Game, now time.Time) ([]byte, error) {
	env := envelope{Version: snapshotVersion, UpdatedAt: now, Game: toWire(g.Snapshot())}
	data, err := json.Marshal(env)
	if err != nil {
		return nil, fmt.Errorf("marshaling game %s: %w", g.ID(), err)
	}
	return data, nil
}

// decode unmarshals payload back into a live *game.Game. A version mismatch
// or any structural problem is a hard error - see snapshotVersion's doc.
func decode(payload []byte) (*game.Game, error) {
	var env envelope
	if err := json.Unmarshal(payload, &env); err != nil {
		return nil, fmt.Errorf("unmarshaling game payload: %w", err)
	}
	if env.Version != snapshotVersion {
		return nil, fmt.Errorf("game payload has snapshot version %d, expected %d", env.Version, snapshotVersion)
	}
	snap := fromWire(env.Game)
	g, err := game.Restore(snap)
	if err != nil {
		return nil, fmt.Errorf("restoring game %s: %w", snap.ID, err)
	}
	return g, nil
}
