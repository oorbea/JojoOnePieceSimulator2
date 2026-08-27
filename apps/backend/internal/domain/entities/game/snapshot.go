package game

import (
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// Snapshot is a complete, exported view of a Game's private state, used by
// ports.IGameStore adapters that cannot keep the same *Game pointer alive
// across process restarts (unlike the in-memory adapter). It deliberately
// carries no JSON tags: the wire format is an infrastructure concern (see
// infrastructure/gamestore/redis), this type only defines what must survive
// a round trip.
//
// Snapshot is a pure read - it does not drain PullEvents. Every Save in
// application/services.GameService.withGame happens after publish(g), which
// always drains pending events first, so g.events is guaranteed empty at
// snapshot time in practice; Snapshot does not bother copying it.
//
// Enums travel as their String() form and come back through the existing
// enums.Parse* functions, so a Snapshot is human-readable JSON and immune to
// iota renumbering. IDs stay as [16]byte, same as every other ID type.
type Snapshot struct {
	ID           GameID
	State        string
	HostID       ParticipantID
	Locked       bool
	Config       ConfigSnapshot
	Participants []ParticipantSnapshot // in join order (== Game.order)
	Teams        []TeamSnapshot
	Stages       []StageSnapshot
	Rounds       []RoundSnapshot
}

// ConfigSnapshot mirrors Config. Visibility/VotingWindowSeconds/PoolFilter
// are additive fields on top of the original v1 shape - see wire.go's
// snapshotVersion doc for how a legacy payload missing them decodes.
//
// Mangas is a legacy field, kept only so a Snapshot written before
// StageMangas/PowerMangas split apart can still restore: snapshotConfig
// never populates it for a fresh Snapshot, but Restore falls back to it for
// either axis that comes back empty (see restoreConfig). A payload old
// enough to lack all three restores nothing and errors, same as before.
type ConfigSnapshot struct {
	Mode                string
	Mangas              []string
	StageMangas         []string
	PowerMangas         []string
	AbilitySource       string
	TeamSize            int
	AllowBots           bool
	Visibility          string
	VotingWindowSeconds int
	PoolFilter          PoolFilterSnapshot
}

// PoolFilterSnapshot mirrors PoolFilter. Empty slices mean "no
// restriction", exactly like PoolFilter itself.
type PoolFilterSnapshot struct {
	StandRarities []string
	FruitRarities []string
	FruitTypes    []string
	Banned        []powers.PowerID
}

// ParticipantSnapshot mirrors Participant. AvatarThumbKey/GooglePicture are
// "" for a bot and for a human who never called SetAvatar.
type ParticipantSnapshot struct {
	ID             ParticipantID
	UserID         *user.UserID // nil for bots
	DisplayName    string
	TeamID         TeamID
	Kind           string
	Connected      bool
	Loadout        *LoadoutSnapshot // nil before AssignLoadouts
	AvatarThumbKey string
	GooglePicture  string
}

// TeamSnapshot mirrors Team.
type TeamSnapshot struct {
	ID      TeamID
	Name    string
	Color   uint32
	Members []ParticipantID
}

// StageSnapshot mirrors Stage.
type StageSnapshot struct {
	ID            StageID
	Manga         string
	Order         int
	Name          string
	Description   string
	Picture       string
	PictureThumb  string
	PictureStatus string
}

// RoundSnapshot mirrors Round.
type RoundSnapshot struct {
	Index        int
	Stage        StageSnapshot
	Ballot       BallotSnapshot
	TiebreakUsed bool
	Result       *RoundResultSnapshot
}

// BallotSnapshot mirrors Ballot. Votes is a slice, not a map: [16]byte is
// not a valid JSON object key, and a slice serializes deterministically.
type BallotSnapshot struct {
	Options []OptionID
	Votes   []VoteSnapshot
}

// VoteSnapshot is a single cast vote.
type VoteSnapshot struct {
	ParticipantID ParticipantID
	Option        OptionID
}

// RoundResultSnapshot mirrors RoundResult.
type RoundResultSnapshot struct {
	Winner            OptionID
	DecidedByCoinFlip bool
}

// LoadoutSnapshot mirrors Loadout. Stand/DevilFruit are embedded in full
// (not by reference): a Loadout is documented as an immutable snapshot of
// abilities for a game/round, so an admin editing or deleting a power later
// must not retroactively change or brick a live lobby.
type LoadoutSnapshot struct {
	Stand           *powers.Stand
	DevilFruit      *powers.DevilFruit
	Spin            string
	Hamon           string
	FruitMastery    string
	ArmamentHaki    string
	ObservationHaki string
	ConquerorHaki   string
	PhysicalForm    string
}

// Snapshot captures g's complete state for out-of-process persistence. See
// the package-level doc on Snapshot for what it does and does not carry.
func (g *Game) Snapshot() Snapshot {
	s := Snapshot{
		ID:     g.id,
		State:  g.state.String(),
		HostID: g.hostID,
		Locked: g.locked,
		Config: ConfigSnapshot{
			Mode:                g.config.mode.String(),
			StageMangas:         mangaStrings(g.config.stageMangas),
			PowerMangas:         mangaStrings(g.config.powerMangas),
			AbilitySource:       g.config.abilitySource.String(),
			TeamSize:            g.config.teamSize,
			AllowBots:           g.config.allowBots,
			Visibility:          g.config.visibility.String(),
			VotingWindowSeconds: g.config.votingWindowSeconds,
			PoolFilter:          snapshotPoolFilter(g.config.poolFilter),
		},
		Participants: make([]ParticipantSnapshot, 0, len(g.order)),
		Teams:        make([]TeamSnapshot, 0, len(g.teams)),
		Stages:       make([]StageSnapshot, 0, len(g.stages)),
		Rounds:       make([]RoundSnapshot, 0, len(g.rounds)),
	}

	for _, pid := range g.order {
		p := g.participants[pid]
		ps := ParticipantSnapshot{
			ID:             p.id,
			UserID:         p.userID,
			DisplayName:    p.displayName,
			TeamID:         p.teamID,
			Kind:           p.kind.String(),
			Connected:      p.connected,
			AvatarThumbKey: p.avatarThumbKey,
			GooglePicture:  p.googlePicture,
		}
		if p.loadout != nil {
			ls := snapshotLoadout(p.loadout)
			ps.Loadout = &ls
		}
		s.Participants = append(s.Participants, ps)
	}

	for _, t := range g.teams {
		s.Teams = append(s.Teams, TeamSnapshot{
			ID:      t.id,
			Name:    t.name,
			Color:   t.color,
			Members: append([]ParticipantID(nil), t.members...),
		})
	}

	for _, st := range g.stages {
		s.Stages = append(s.Stages, snapshotStage(st))
	}

	for _, r := range g.rounds {
		rs := RoundSnapshot{
			Index:        r.Index,
			Stage:        snapshotStage(r.Stage),
			Ballot:       snapshotBallot(r.Ballot),
			TiebreakUsed: r.TiebreakUsed,
		}
		if r.Result != nil {
			rs.Result = &RoundResultSnapshot{
				Winner:            r.Result.Winner,
				DecidedByCoinFlip: r.Result.DecidedByCoinFlip,
			}
		}
		s.Rounds = append(s.Rounds, rs)
	}

	return s
}

func snapshotStage(st Stage) StageSnapshot {
	return StageSnapshot{
		ID: st.id, Manga: st.manga.String(), Order: st.order, Name: st.name,
		Description: st.description, Picture: st.picture, PictureThumb: st.pictureThumb,
		PictureStatus: st.pictureStatus.String(),
	}
}

func snapshotBallot(b *Ballot) BallotSnapshot {
	votes := b.Votes()
	bs := BallotSnapshot{
		Options: append([]OptionID(nil), b.options...),
		Votes:   make([]VoteSnapshot, 0, len(votes)),
	}
	for pid, opt := range votes {
		bs.Votes = append(bs.Votes, VoteSnapshot{ParticipantID: pid, Option: opt})
	}
	return bs
}

func snapshotLoadout(l *Loadout) LoadoutSnapshot {
	return LoadoutSnapshot{
		Stand:           l.stand,
		DevilFruit:      l.devilFruit,
		Spin:            l.spin.String(),
		Hamon:           l.hamon.String(),
		FruitMastery:    l.fruitMastery.String(),
		ArmamentHaki:    l.armamentHaki.String(),
		ObservationHaki: l.observationHaki.String(),
		ConquerorHaki:   l.conquerorHaki.String(),
		PhysicalForm:    l.physicalForm.String(),
	}
}

func mangaStrings(mangas []enums.Manga) []string {
	out := make([]string, len(mangas))
	for i, m := range mangas {
		out[i] = m.String()
	}
	return out
}

func parseMangaList(raw []string) ([]enums.Manga, error) {
	out := make([]enums.Manga, len(raw))
	for i, m := range raw {
		parsed, err := enums.ParseManga(m)
		if err != nil {
			return nil, err
		}
		out[i] = parsed
	}
	return out, nil
}

func rarityStrings(rarities []enums.PowerRarity) []string {
	out := make([]string, len(rarities))
	for i, r := range rarities {
		out[i] = r.String()
	}
	return out
}

func fruitTypeStrings(types []enums.FruitType) []string {
	out := make([]string, len(types))
	for i, t := range types {
		out[i] = t.String()
	}
	return out
}

func snapshotPoolFilter(f PoolFilter) PoolFilterSnapshot {
	return PoolFilterSnapshot{
		StandRarities: rarityStrings(f.standRarities),
		FruitRarities: rarityStrings(f.fruitRarities),
		FruitTypes:    fruitTypeStrings(f.fruitTypes),
		Banned:        append([]powers.PowerID(nil), f.banned...),
	}
}

// restorePoolFilter rebuilds a PoolFilter from its snapshot. Unlike
// NewPoolFilter it does not reject an empty result - an empty
// PoolFilterSnapshot (legacy payload, or a lobby that never restricted its
// pool) restores to the "no restriction" PoolFilter{}.
func restorePoolFilter(fs PoolFilterSnapshot) (PoolFilter, error) {
	standRarities := make([]enums.PowerRarity, len(fs.StandRarities))
	for i, r := range fs.StandRarities {
		parsed, err := enums.ParsePowerRarity(r)
		if err != nil {
			return PoolFilter{}, err
		}
		standRarities[i] = parsed
	}
	fruitRarities := make([]enums.PowerRarity, len(fs.FruitRarities))
	for i, r := range fs.FruitRarities {
		parsed, err := enums.ParsePowerRarity(r)
		if err != nil {
			return PoolFilter{}, err
		}
		fruitRarities[i] = parsed
	}
	fruitTypes := make([]enums.FruitType, len(fs.FruitTypes))
	for i, t := range fs.FruitTypes {
		parsed, err := enums.ParseFruitType(t)
		if err != nil {
			return PoolFilter{}, err
		}
		fruitTypes[i] = parsed
	}
	return NewPoolFilter(standRarities, fruitRarities, fruitTypes, fs.Banned)
}

// Restore rebuilds a *Game from a Snapshot. Unlike NewGame/Join/AddBot, it
// bypasses every capacity/membership invariant (a full lobby must still be
// restorable) but re-validates every value object through its normal
// constructor, so a Snapshot produced by a different build of this package
// (or corrupted in transit) fails loudly instead of producing a Game that
// silently misbehaves.
//
// Restore always installs DefaultLoadoutEvaluator{} - the evaluator is
// behaviour, not data, and is not part of a Snapshot. Callers that need a
// different evaluator must call SetLoadoutEvaluator afterwards.
func Restore(s Snapshot) (*Game, error) {
	mode, err := enums.ParseGameModeKind(s.Config.Mode)
	if err != nil {
		return nil, err
	}
	abilitySource, err := enums.ParseAbilitySource(s.Config.AbilitySource)
	if err != nil {
		return nil, err
	}
	// StageMangas/PowerMangas are additive fields, split apart from the
	// original single Mangas field - a legacy Snapshot (or a fresher one
	// that just hasn't been re-saved since the split) predating them
	// decodes with both empty, so either axis coming back empty falls back
	// to the legacy Mangas list. See ConfigSnapshot's doc comment.
	stageMangaStrs := s.Config.StageMangas
	if len(stageMangaStrs) == 0 {
		stageMangaStrs = s.Config.Mangas
	}
	powerMangaStrs := s.Config.PowerMangas
	if len(powerMangaStrs) == 0 {
		powerMangaStrs = s.Config.Mangas
	}
	stageMangas, err := parseMangaList(stageMangaStrs)
	if err != nil {
		return nil, err
	}
	powerMangas, err := parseMangaList(powerMangaStrs)
	if err != nil {
		return nil, err
	}
	// Visibility/VotingWindowSeconds/PoolFilter are additive fields - a
	// legacy Snapshot predating them decodes with an empty Visibility
	// string and a zero VotingWindowSeconds, which fall back here to
	// PRIVATE and DefaultVotingWindowSeconds rather than failing to parse.
	// See wire.go's snapshotVersion doc for why the version wasn't bumped.
	visibility := enums.Private
	if s.Config.Visibility != "" {
		visibility, err = enums.ParseLobbyVisibility(s.Config.Visibility)
		if err != nil {
			return nil, err
		}
	}
	votingWindowSeconds := s.Config.VotingWindowSeconds
	if votingWindowSeconds == 0 {
		votingWindowSeconds = DefaultVotingWindowSeconds
	}
	poolFilter, err := restorePoolFilter(s.Config.PoolFilter)
	if err != nil {
		return nil, err
	}

	cfg := Config{
		mode:                mode,
		stageMangas:         stageMangas,
		powerMangas:         powerMangas,
		abilitySource:       abilitySource,
		teamSize:            s.Config.TeamSize,
		allowBots:           s.Config.AllowBots,
		visibility:          visibility,
		votingWindowSeconds: votingWindowSeconds,
		poolFilter:          poolFilter,
	}

	state, err := enums.ParseGameState(s.State)
	if err != nil {
		return nil, err
	}

	gameMode, err := modeFor(mode)
	if err != nil {
		return nil, err
	}

	teams := make([]*Team, 0, len(s.Teams))
	for _, ts := range s.Teams {
		t := &Team{
			id:      ts.ID,
			name:    ts.Name,
			color:   ts.Color,
			members: append([]ParticipantID(nil), ts.Members...),
		}
		teams = append(teams, t)
	}

	stages := make([]Stage, 0, len(s.Stages))
	for _, ss := range s.Stages {
		st, err := restoreStage(ss)
		if err != nil {
			return nil, err
		}
		stages = append(stages, st)
	}

	g := &Game{
		id:           s.ID,
		config:       cfg,
		mode:         gameMode,
		hostID:       s.HostID,
		state:        state,
		locked:       s.Locked,
		participants: make(map[ParticipantID]*Participant, len(s.Participants)),
		order:        make([]ParticipantID, 0, len(s.Participants)),
		teams:        teams,
		stages:       stages,
		evaluator:    DefaultLoadoutEvaluator{},
	}

	for _, ps := range s.Participants {
		kind, err := enums.ParseParticipantKind(ps.Kind)
		if err != nil {
			return nil, err
		}
		p := &Participant{
			id:             ps.ID,
			userID:         ps.UserID,
			displayName:    ps.DisplayName,
			teamID:         ps.TeamID,
			kind:           kind,
			connected:      ps.Connected,
			avatarThumbKey: ps.AvatarThumbKey,
			googlePicture:  ps.GooglePicture,
		}
		if ps.Loadout != nil {
			loadout, err := restoreLoadout(*ps.Loadout)
			if err != nil {
				return nil, err
			}
			p.loadout = loadout
		}
		g.participants[p.id] = p
		g.order = append(g.order, p.id)
	}

	g.rounds = make([]Round, 0, len(s.Rounds))
	for _, rs := range s.Rounds {
		stage, err := restoreStage(rs.Stage)
		if err != nil {
			return nil, err
		}
		ballot, err := restoreBallot(rs.Ballot)
		if err != nil {
			return nil, err
		}
		r := Round{
			Index:        rs.Index,
			Stage:        stage,
			Ballot:       ballot,
			TiebreakUsed: rs.TiebreakUsed,
		}
		if rs.Result != nil {
			r.Result = &RoundResult{
				Winner:            rs.Result.Winner,
				DecidedByCoinFlip: rs.Result.DecidedByCoinFlip,
			}
		}
		g.rounds = append(g.rounds, r)
	}

	return g, nil
}

func restoreStage(ss StageSnapshot) (Stage, error) {
	manga, err := enums.ParseManga(ss.Manga)
	if err != nil {
		return Stage{}, err
	}
	st, err := NewStage(ss.ID, manga, ss.Order, ss.Name, ss.Description, ss.Picture)
	if err != nil {
		return Stage{}, err
	}
	status, err := enums.ParsePictureStatus(ss.PictureStatus)
	if err != nil {
		return Stage{}, err
	}
	st.SetPictureRenditions(ss.Picture, ss.PictureThumb, status)
	return st, nil
}

// restoreBallot rebuilds a Ballot without going through Cast's validation
// against a live options list drift - the options set is trusted as-is
// since it came from this same Snapshot.
func restoreBallot(bs BallotSnapshot) (*Ballot, error) {
	ballot, err := NewBallot(bs.Options)
	if err != nil {
		return nil, err
	}
	for _, v := range bs.Votes {
		ballot.votes[v.ParticipantID] = v.Option
	}
	return ballot, nil
}

func restoreLoadout(ls LoadoutSnapshot) (*Loadout, error) {
	// Legacy fallback: SpinLevel dropped its ADVANCED tier (see
	// ObsidianVault/gameplay-game-modes.md's V1-probabilities port) - a
	// snapshot written before that change can still carry the string.
	// Degrade to SpinBasic, the closest surviving tier below it, rather
	// than fail to restore an otherwise-valid in-flight game.
	spinStr := ls.Spin
	if spinStr == "ADVANCED" {
		spinStr = "BASIC"
	}
	spin, err := enums.ParseSpinLevel(spinStr)
	if err != nil {
		return nil, err
	}
	hamon, err := enums.ParseHamonLevel(ls.Hamon)
	if err != nil {
		return nil, err
	}
	fruitMastery, err := enums.ParseFruitMastery(ls.FruitMastery)
	if err != nil {
		return nil, err
	}
	armamentHaki, err := enums.ParseHakiLevel(ls.ArmamentHaki)
	if err != nil {
		return nil, err
	}
	observationHaki, err := enums.ParseHakiLevel(ls.ObservationHaki)
	if err != nil {
		return nil, err
	}
	conquerorHaki, err := enums.ParseHakiLevel(ls.ConquerorHaki)
	if err != nil {
		return nil, err
	}
	physicalForm, err := enums.ParsePhysicalForm(ls.PhysicalForm)
	if err != nil {
		return nil, err
	}
	return NewLoadout(
		ls.Stand, ls.DevilFruit,
		spin, hamon, fruitMastery,
		armamentHaki, observationHaki, conquerorHaki,
		physicalForm,
	)
}
