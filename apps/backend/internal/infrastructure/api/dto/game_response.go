package dto

import (
	"context"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// GameConfigResponse mirrors game.Config.
type GameConfigResponse struct {
	// StageMangas and PowerMangas are independent - see game.Config's doc
	// comment.
	StageMangas         []string           `json:"stageMangas"`
	PowerMangas         []string           `json:"powerMangas"`
	AbilitySource       string             `json:"abilitySource"`
	TeamSize            int                `json:"teamSize"`
	AllowBots           bool               `json:"allowBots"`
	Visibility          string             `json:"visibility"`
	VotingWindowSeconds int                `json:"votingWindowSeconds"`
	PoolFilter          PoolFilterResponse `json:"poolFilter"`
}

// PoolFilterResponse mirrors game.PoolFilter. Empty arrays mean "no
// restriction", exactly like the domain type.
type PoolFilterResponse struct {
	StandRarities []string `json:"standRarities"`
	FruitRarities []string `json:"fruitRarities"`
	FruitTypes    []string `json:"fruitTypes"`
	Banned        []string `json:"banned"`
}

func newPoolFilterResponse(f game.PoolFilter) PoolFilterResponse {
	standRarities := make([]string, 0)
	for _, r := range f.StandRarities() {
		standRarities = append(standRarities, r.String())
	}
	fruitRarities := make([]string, 0)
	for _, r := range f.FruitRarities() {
		fruitRarities = append(fruitRarities, r.String())
	}
	fruitTypes := make([]string, 0)
	for _, t := range f.FruitTypes() {
		fruitTypes = append(fruitTypes, t.String())
	}
	banned := make([]string, 0)
	for _, id := range f.Banned() {
		banned = append(banned, id.String())
	}
	return PoolFilterResponse{
		StandRarities: standRarities, FruitRarities: fruitRarities,
		FruitTypes: fruitTypes, Banned: banned,
	}
}

// GameTeamResponse mirrors game.Team.
type GameTeamResponse struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Color     uint32   `json:"color"`
	MemberIDs []string `json:"memberIds"`
}

// GameLoadoutResponse mirrors game.Loadout. Every participant's loadout is
// visible to every other participant: in Gauntlet the vote is whether the
// whole squad survives a Stage, in Versus it's which team wins the round -
// judging either without seeing the powers in play is arbitrary, and the
// domain's own bot-voting already assumes full information (see
// Game.optionScores). Only the vote itself is hidden until a round
// resolves - see GameRoundResponse.
type GameLoadoutResponse struct {
	Stand           *StandResponse      `json:"stand,omitempty"`
	DevilFruit      *DevilFruitResponse `json:"devilFruit,omitempty"`
	Spin            string              `json:"spin"`
	Hamon           string              `json:"hamon"`
	FruitMastery    string              `json:"fruitMastery"`
	ArmamentHaki    string              `json:"armamentHaki"`
	ObservationHaki string              `json:"observationHaki"`
	ConquerorHaki   string              `json:"conquerorHaki"`
	PhysicalForm    string              `json:"physicalForm"`
}

// GameParticipantResponse mirrors game.Participant. AvatarThumb is resolved
// at serialization time (own upload if the participant's user has one, else
// their Google-synced picture, "" for a bot) - see resolveParticipantAvatar.
type GameParticipantResponse struct {
	ID          string               `json:"id"`
	UserID      *string              `json:"userId,omitempty"`
	DisplayName string               `json:"displayName"`
	TeamID      string               `json:"teamId"`
	Kind        string               `json:"kind"`
	Connected   bool                 `json:"connected"`
	AvatarThumb string               `json:"avatarThumb"`
	Loadout     *GameLoadoutResponse `json:"loadout,omitempty"`
}

// GameStageResponse mirrors game.Stage. Description is NOT read off the
// domain Stage frozen into the Round - a live Game is one instance shared
// by every participant, so a Stage snapshotted at round-assignment time can
// only ever carry one baked-in locale. Instead, it's re-resolved per
// viewer's own configured language at serialization time - see
// StageTextResolver and NewGameStateResponse. Picture is locale-independent
// and does come straight off the domain Stage.
type GameStageResponse struct {
	ID            string `json:"id"`
	Manga         string `json:"manga"`
	Order         int    `json:"order"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Picture       string `json:"picture"`
	PictureThumb  string `json:"pictureThumb"`
	PictureStatus string `json:"pictureStatus"`
}

// StageTextResolver resolves a Stage's description for a specific viewer
// locale, bound by the caller (see endpoints.GameEndpoints.stageTextResolver)
// to whichever locale that viewer has configured on their account -
// unrelated to PictureURLResolver, which never depends on locale.
type StageTextResolver func(ctx context.Context, id game.StageID) (string, error)

// PowerTextResolver is StageTextResolver's analogue for a loadout's
// Stand/DevilFruit: a live Game's Loadout freezes its Stand/DevilFruit at
// whatever locale RepoPowerPool drew them in (en-GB - see
// infrastructure/game/repo_power_pool.go), since the Game is one instance
// shared by every participant. This resolver re-derives the description and
// skills per viewer at serialization time instead, bound by the caller (see
// endpoints.GameEndpoints.powerTextResolver) to that viewer's own locale.
type PowerTextResolver func(ctx context.Context, id powers.PowerID) (ports.PowerContent, error)

// GameRoundResultResponse mirrors game.RoundResult.
type GameRoundResultResponse struct {
	Winner            string `json:"winner"`
	DecidedByCoinFlip bool   `json:"decidedByCoinFlip"`
}

// GameRoundResponse mirrors game.Round. Votes are hidden until the round
// resolves: while voting is open, VotedParticipantIDs says who has voted
// (not what they voted), and Votes is omitted entirely; once Result is set,
// Votes is populated so a client resyncing after a round closed can still
// render who voted for what. This matches VOTE_CAST's transport framing
// (game_ws.go), which is anonymized the same way while a round is live.
// TiedVotes is the exception to "hidden while live": it is populated as
// soon as a tie opens a revote (state TIEBREAK), revealing what tied before
// the revote replaces it - an explicit owner call (2026-08-28) to let
// clients render the tied breakdown, not just a "tie - revote" label.
type GameRoundResponse struct {
	Index               int                      `json:"index"`
	Stage               GameStageResponse        `json:"stage"`
	Options             []string                 `json:"options"`
	TiebreakUsed        bool                     `json:"tiebreakUsed"`
	VotedParticipantIDs []string                 `json:"votedParticipantIds"`
	Votes               map[string]string        `json:"votes,omitempty"`
	TiedVotes           map[string]string        `json:"tiedVotes,omitempty"`
	Result              *GameRoundResultResponse `json:"result,omitempty"`
}

// GameResultResponse mirrors game.GameResult.
type GameResultResponse struct {
	Mode         string `json:"mode"`
	Winner       string `json:"winner"`
	RoundsPlayed int    `json:"roundsPlayed"`
	Aborted      bool   `json:"aborted"`
}

// GameSnapshotResponse is the shared, viewer-independent half of a game's
// state.
type GameSnapshotResponse struct {
	ID           string                    `json:"id"`
	Code         string                    `json:"code"`
	State        string                    `json:"state"`
	Mode         string                    `json:"mode"`
	HostID       string                    `json:"hostId"`
	Locked       bool                      `json:"locked"`
	Config       GameConfigResponse        `json:"config"`
	Teams        []GameTeamResponse        `json:"teams"`
	Participants []GameParticipantResponse `json:"participants"`
	Rounds       []GameRoundResponse       `json:"rounds"`
	Result       *GameResultResponse       `json:"result,omitempty"`
	// RevealEndsAt is set (RFC3339) only while the game is ASSIGNING with a
	// pending reveal - see NewGameStateResponse's deadlines param.
	RevealEndsAt *string `json:"revealEndsAt,omitempty"`
	// VotingEndsAt is set (RFC3339) only while the game is VOTING/TIEBREAK
	// with a pending window - the deadline a client reconnecting mid-vote
	// needs, since the VOTING_OPENED/TIEBREAK_OPENED frames that carry
	// closesAt are one-shot and long gone by then.
	VotingEndsAt *string `json:"votingEndsAt,omitempty"`
	// ResultEndsAt is set (RFC3339) only while the game is RESOLVING with a
	// pending result display - see GameService.ResultEndsAt.
	ResultEndsAt *string `json:"resultEndsAt,omitempty"`
}

// GameViewerResponse is the cheap, per-viewer convenience block - everything
// in it is derivable from GameSnapshotResponse plus the caller's own
// ParticipantID, computed once here so every client doesn't have to.
type GameViewerResponse struct {
	ParticipantID string  `json:"participantId"`
	TeamID        string  `json:"teamId"`
	IsHost        bool    `json:"isHost"`
	HasVoted      bool    `json:"hasVoted"`
	Vote          *string `json:"vote,omitempty"`
}

// GameStateResponse is the full authoritative payload sent as the WebSocket
// STATE frame and returned by the plain HTTP game routes.
type GameStateResponse struct {
	Game GameSnapshotResponse `json:"game"`
	You  GameViewerResponse   `json:"you"`
}

// GameStateDeadlines carries the in-process wall-clock deadlines GameService
// keeps for a Game, so a (re)connecting client resumes a countdown instead
// of restarting it. Each is nil unless that phase is actually in flight,
// and at most one is ever set (the reveal, voting, and result timers never
// coexist for the same GameID - see GameService's timers field doc).
// Grouped in a struct rather than passed as adjacent *time.Time parameters:
// same-typed pointers in a positional call are silently swappable, and the
// swap would still compile and still produce a plausible response.
type GameStateDeadlines struct {
	// RevealEndsAt is set only while the Game is ASSIGNING with a pending
	// reveal - see GameService.RevealEndsAt.
	RevealEndsAt *time.Time
	// VotingEndsAt is set only while the Game is VOTING/TIEBREAK with a
	// pending window - see GameService.VotingEndsAt.
	VotingEndsAt *time.Time
	// ResultEndsAt is set only while the Game is RESOLVING with a pending
	// result display - see GameService.ResultEndsAt.
	ResultEndsAt *time.Time
}

// NewGameStateResponse builds a GameStateResponse for self's point of view.
// resolveStand/resolveFruit resolve a Stand/DevilFruit's picture key into a
// URL, same signature as StandService.PictureURL/DevilFruitService.PictureURL.
// resolveStandText/resolveFruitText re-resolve a loadout's Stand/DevilFruit
// description+skills in self's own locale (see PowerTextResolver).
// resolveAvatarPicture resolves a participant's avatar thumb key into a URL,
// same signature as UserService.AvatarURL - "" resolves to "" without being
// called (see resolveParticipantAvatar). deadlines carries the in-flight
// reveal/voting deadlines for g (see GameStateDeadlines) - each nil unless
// that phase is actually in progress, which is what lets a (re)connecting
// client resume the countdown instead of restarting it from zero.
func NewGameStateResponse(
	ctx context.Context,
	g *game.Game,
	code string,
	self game.ParticipantID,
	resolveStand, resolveFruit, resolveStagePicture, resolveAvatarPicture PictureURLResolver,
	resolveStageDescription StageTextResolver,
	resolveStandText, resolveFruitText PowerTextResolver,
	deadlines GameStateDeadlines,
) (GameStateResponse, error) {
	teams := make([]GameTeamResponse, 0, len(g.Teams()))
	for _, t := range g.Teams() {
		members := make([]string, 0, len(t.Members()))
		for _, m := range t.Members() {
			members = append(members, m.String())
		}
		teams = append(teams, GameTeamResponse{
			ID: t.ID().String(), Name: t.Name(), Color: t.Color(), MemberIDs: members,
		})
	}

	participants := make([]GameParticipantResponse, 0, len(g.Participants()))
	var viewer GameViewerResponse
	for _, p := range g.Participants() {
		pr := GameParticipantResponse{
			ID:          p.ID().String(),
			DisplayName: p.DisplayName(),
			TeamID:      p.TeamID().String(),
			Kind:        p.Kind().String(),
			Connected:   p.Connected(),
		}
		if uid := p.UserID(); uid != nil {
			s := uid.String()
			pr.UserID = &s
		}
		avatarThumb, err := resolveParticipantAvatar(ctx, p.AvatarThumbKey(), p.GooglePicture(), resolveAvatarPicture)
		if err != nil {
			return GameStateResponse{}, err
		}
		pr.AvatarThumb = avatarThumb
		if l := p.Loadout(); l != nil {
			lr, err := newGameLoadoutResponse(ctx, l, resolveStand, resolveFruit, resolveStandText, resolveFruitText)
			if err != nil {
				return GameStateResponse{}, err
			}
			pr.Loadout = &lr
		}
		participants = append(participants, pr)

		if p.ID() == self {
			viewer = GameViewerResponse{
				ParticipantID: p.ID().String(),
				TeamID:        p.TeamID().String(),
				IsHost:        p.ID() == g.HostID(),
			}
		}
	}

	rounds := make([]GameRoundResponse, 0, len(g.Rounds()))
	for _, r := range g.Rounds() {
		options := make([]string, 0, len(r.Ballot.Options()))
		for _, o := range r.Ballot.Options() {
			options = append(options, string(o))
		}
		stageResp, err := newGameStageResponse(ctx, r.Stage, resolveStagePicture, resolveStageDescription)
		if err != nil {
			return GameStateResponse{}, err
		}
		rr := GameRoundResponse{
			Index:        r.Index,
			Stage:        stageResp,
			Options:      options,
			TiebreakUsed: r.TiebreakUsed,
		}
		if r.TiedVotes != nil {
			rr.TiedVotes = make(map[string]string, len(r.TiedVotes))
			for pid, opt := range r.TiedVotes {
				rr.TiedVotes[pid.String()] = string(opt)
			}
		}
		votes := r.Ballot.Votes()
		if r.Result != nil {
			// Round resolved: votes are no longer secret.
			rr.Votes = make(map[string]string, len(votes))
			for pid, opt := range votes {
				rr.Votes[pid.String()] = string(opt)
			}
			rr.Result = &GameRoundResultResponse{
				Winner: string(r.Result.Winner), DecidedByCoinFlip: r.Result.DecidedByCoinFlip,
			}
		} else {
			// Round still live: only who voted, not what they voted -
			// except the caller's own vote, which they already know.
			rr.VotedParticipantIDs = make([]string, 0, len(votes))
			for pid := range votes {
				rr.VotedParticipantIDs = append(rr.VotedParticipantIDs, pid.String())
			}
		}
		if r.Index == len(g.Rounds())-1 {
			if opt, ok := votes[self]; ok {
				viewer.HasVoted = true
				s := string(opt)
				viewer.Vote = &s
			}
		}
		rounds = append(rounds, rr)
	}

	var result *GameResultResponse
	if g.State() == enums.Finished || g.State() == enums.Aborted {
		if res, err := g.Result(); err == nil {
			result = &GameResultResponse{
				Mode: res.Mode.String(), Winner: string(res.Winner),
				RoundsPlayed: res.RoundsPlayed, Aborted: res.Aborted,
			}
		}
	}

	var revealEndsAtStr, votingEndsAtStr, resultEndsAtStr *string
	if deadlines.RevealEndsAt != nil {
		s := deadlines.RevealEndsAt.Format(time.RFC3339)
		revealEndsAtStr = &s
	}
	if deadlines.VotingEndsAt != nil {
		s := deadlines.VotingEndsAt.Format(time.RFC3339)
		votingEndsAtStr = &s
	}
	if deadlines.ResultEndsAt != nil {
		s := deadlines.ResultEndsAt.Format(time.RFC3339)
		resultEndsAtStr = &s
	}

	return GameStateResponse{
		Game: GameSnapshotResponse{
			ID:     g.ID().String(),
			Code:   code,
			State:  g.State().String(),
			Mode:   g.Config().Mode().String(),
			HostID: g.HostID().String(),
			Locked: g.Locked(),
			Config: GameConfigResponse{
				StageMangas:         mangaNames(g.Config().StageMangas()),
				PowerMangas:         mangaNames(g.Config().PowerMangas()),
				AbilitySource:       g.Config().AbilitySource().String(),
				TeamSize:            g.Config().TeamSize(),
				AllowBots:           g.Config().AllowBots(),
				Visibility:          g.Config().Visibility().String(),
				VotingWindowSeconds: g.Config().VotingWindowSeconds(),
				PoolFilter:          newPoolFilterResponse(g.Config().PoolFilter()),
			},
			Teams:        teams,
			Participants: participants,
			Rounds:       rounds,
			Result:       result,
			RevealEndsAt: revealEndsAtStr,
			VotingEndsAt: votingEndsAtStr,
			ResultEndsAt: resultEndsAtStr,
		},
		You: viewer,
	}, nil
}

func newGameStageResponse(ctx context.Context, s game.Stage, resolvePicture PictureURLResolver, resolveDescription StageTextResolver) (GameStageResponse, error) {
	pictureURL, err := resolvePicture(ctx, s.Picture())
	if err != nil {
		return GameStageResponse{}, err
	}
	thumbURL, err := resolvePicture(ctx, s.PictureThumb())
	if err != nil {
		return GameStageResponse{}, err
	}
	description, err := resolveDescription(ctx, s.ID())
	if err != nil {
		return GameStageResponse{}, err
	}
	return GameStageResponse{
		ID: s.ID().String(), Manga: s.Manga().String(), Order: s.Order(), Name: s.Name(),
		Description: description, Picture: pictureURL, PictureThumb: thumbURL,
		PictureStatus: s.PictureStatus().String(),
	}, nil
}

// newGameLoadoutResponse builds a GameLoadoutResponse for one viewer.
// resolveStandText/resolveFruitText override the Stand/DevilFruit's
// description+skills - RepoPowerPool freezes those to en-GB when the
// loadout was drawn (see repo_power_pool.go), so without this override
// every viewer would see the loadout's power text in English regardless of
// their own configured language.
func newGameLoadoutResponse(
	ctx context.Context,
	l *game.Loadout,
	resolveStand, resolveFruit PictureURLResolver,
	resolveStandText, resolveFruitText PowerTextResolver,
) (GameLoadoutResponse, error) {
	lr := GameLoadoutResponse{
		Spin:            l.Spin().String(),
		Hamon:           l.Hamon().String(),
		FruitMastery:    l.FruitMastery().String(),
		ArmamentHaki:    l.ArmamentHaki().String(),
		ObservationHaki: l.ObservationHaki().String(),
		ConquerorHaki:   l.ConquerorHaki().String(),
		PhysicalForm:    l.PhysicalForm().String(),
	}
	if s := l.Stand(); s != nil {
		sr, err := NewStandResponse(ctx, s, resolveStand)
		if err != nil {
			return GameLoadoutResponse{}, err
		}
		content, err := resolveStandText(ctx, s.ID())
		if err != nil {
			return GameLoadoutResponse{}, err
		}
		sr.Description = content.Description
		sr.Skills = nonNilSkills(content.Skills)
		lr.Stand = &sr
	}
	if f := l.DevilFruit(); f != nil {
		fr, err := NewDevilFruitResponse(ctx, f, resolveFruit)
		if err != nil {
			return GameLoadoutResponse{}, err
		}
		content, err := resolveFruitText(ctx, f.ID())
		if err != nil {
			return GameLoadoutResponse{}, err
		}
		fr.Description = content.Description
		fr.Skills = nonNilSkills(content.Skills)
		lr.DevilFruit = &fr
	}
	return lr, nil
}

func nonNilSkills(skills []string) []string {
	if skills == nil {
		return []string{}
	}
	return skills
}

// resolveParticipantAvatar resolves a game participant's avatar: their own
// uploaded thumbnail (presigned through resolve, an object-storage key) if
// one exists, else their Google-synced picture (already a full external
// URL - never passed through resolve), else "" for a bot or a human with
// neither. Unlike dto.resolveAvatar (the profile endpoint's version), there
// is no separate main/thumb pair here - game.Participant only carries the
// thumb key, since that's the only rendition the roster/reveal ever show.
func resolveParticipantAvatar(ctx context.Context, avatarThumbKey, googlePicture string, resolve PictureURLResolver) (string, error) {
	if avatarThumbKey == "" {
		return googlePicture, nil
	}
	return resolve(ctx, avatarThumbKey)
}

// PublicLobbyResponse is one entry in the public lobby browser - a
// services.LobbyListing flattened to JSON. Deliberately roster-free and
// join-code-free: browsing must never leak who's in a lobby, and the
// public browser joins by GameID, not by code.
type PublicLobbyResponse struct {
	GameID              string   `json:"gameId"`
	Mode                string   `json:"mode"`
	HostDisplayName     string   `json:"hostDisplayName"`
	AbilitySource       string   `json:"abilitySource"`
	Mangas              []string `json:"mangas"`
	PlayerCount         int      `json:"playerCount"`
	MaxPlayers          int      `json:"maxPlayers"`
	VotingWindowSeconds int      `json:"votingWindowSeconds"`
	AllowBots           bool     `json:"allowBots"`
	Locked              bool     `json:"locked"`
}

// NewPublicLobbyResponse flattens a services.LobbyListing for the browser.
func NewPublicLobbyResponse(l services.LobbyListing) PublicLobbyResponse {
	return PublicLobbyResponse{
		GameID:              l.GameID.String(),
		Mode:                l.Mode.String(),
		HostDisplayName:     l.HostDisplayName,
		PlayerCount:         l.PlayerCount,
		MaxPlayers:          l.MaxPlayers,
		Mangas:              mangaNames(l.Mangas),
		AbilitySource:       l.AbilitySource.String(),
		AllowBots:           l.AllowBots,
		VotingWindowSeconds: l.VotingWindowSeconds,
		Locked:              l.Locked,
	}
}

// PublicLobbyListResponse is GET /games/public's response body.
type PublicLobbyListResponse struct {
	Items []PublicLobbyResponse `json:"items"`
}

// LobbyPreviewResponse is GET /games/preview's response body - the same
// summary as PublicLobbyResponse plus the echoed code and visibility, since
// a preview also works for PRIVATE lobbies (the code itself is the
// credential).
type LobbyPreviewResponse struct {
	PublicLobbyResponse
	Code       string `json:"code"`
	Visibility string `json:"visibility"`
}

// NewLobbyPreviewResponse builds a LobbyPreviewResponse.
func NewLobbyPreviewResponse(code string, l services.LobbyListing) LobbyPreviewResponse {
	return LobbyPreviewResponse{
		PublicLobbyResponse: NewPublicLobbyResponse(l),
		Code:                code,
		Visibility:          l.Visibility.String(),
	}
}

func mangaNames(mangas []enums.Manga) []string {
	out := make([]string, len(mangas))
	for i, m := range mangas {
		out[i] = m.String()
	}
	return out
}
