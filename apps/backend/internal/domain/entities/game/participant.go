package game

import (
	"errors"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// Participant is a seat in a Game: either a human tied to a user.UserID, or
// a bot (Versus-only, to fill uneven teams) with none. Loadout is nil
// until Game.AssignLoadouts runs. avatarThumbKey/googlePicture are empty for
// a bot and, for a human, set separately via SetAvatar right after
// construction (see GameService.CreateGame/joinLocked) rather than as
// constructor params - avatar is presentation-only and this keeps every
// existing NewHumanParticipant call site (tests included) unchanged.
type Participant struct {
	id             ParticipantID
	userID         *user.UserID
	displayName    string
	teamID         TeamID
	kind           enums.ParticipantKind
	connected      bool
	loadout        *Loadout
	avatarThumbKey string
	googlePicture  string
	avatarFocalX   float64
	avatarFocalY   float64
	disconnectedAt *time.Time
	abandoned      bool
}

// NewHumanParticipant builds a Participant backed by a registered user.
func NewHumanParticipant(id ParticipantID, userID user.UserID, displayName string, teamID TeamID) (*Participant, error) {
	if id.IsNil() {
		return nil, errors.New("id is required")
	}
	if userID.IsNil() {
		return nil, errors.New("user id is required")
	}
	if displayName == "" {
		return nil, errors.New("display name is required")
	}
	if teamID.IsNil() {
		return nil, errors.New("team id is required")
	}
	uid := userID
	return &Participant{
		id:           id,
		userID:       &uid,
		displayName:  displayName,
		teamID:       teamID,
		kind:         enums.Human,
		connected:    true,
		avatarFocalX: 0.5,
		avatarFocalY: 0.5,
	}, nil
}

// NewBotParticipant builds a bot Participant. Whether bots are allowed at
// all is enforced by Game.AddBot, not here.
func NewBotParticipant(id ParticipantID, displayName string, teamID TeamID) (*Participant, error) {
	if id.IsNil() {
		return nil, errors.New("id is required")
	}
	if displayName == "" {
		return nil, errors.New("display name is required")
	}
	if teamID.IsNil() {
		return nil, errors.New("team id is required")
	}
	return &Participant{
		id:          id,
		displayName: displayName,
		teamID:      teamID,
		kind:        enums.Bot,
		connected:   true,
	}, nil
}

func (p *Participant) ID() ParticipantID           { return p.id }
func (p *Participant) UserID() *user.UserID        { return p.userID }
func (p *Participant) DisplayName() string         { return p.displayName }
func (p *Participant) TeamID() TeamID              { return p.teamID }
func (p *Participant) Kind() enums.ParticipantKind { return p.kind }
func (p *Participant) IsBot() bool                 { return p.kind == enums.Bot }
func (p *Participant) Connected() bool             { return p.connected }
func (p *Participant) Loadout() *Loadout           { return p.loadout }
func (p *Participant) AvatarThumbKey() string      { return p.avatarThumbKey }
func (p *Participant) GooglePicture() string       { return p.googlePicture }
func (p *Participant) AvatarFocalX() float64       { return p.avatarFocalX }
func (p *Participant) AvatarFocalY() float64       { return p.avatarFocalY }
func (p *Participant) Abandoned() bool             { return p.abandoned }

// DisconnectedAt reports when this participant went unreachable, if they
// currently are. It is nil once Reconnect clears it.
func (p *Participant) DisconnectedAt() (time.Time, bool) {
	if p.disconnectedAt == nil {
		return time.Time{}, false
	}
	return *p.disconnectedAt, true
}

// SetAvatar records where this participant's avatar picture comes from -
// their own uploaded thumbnail key (presigned at serialization time) and/or
// their Google-synced picture URL (already a full external URL) - plus the
// focal point (0..1) the owner picked for that picture, so a card cropping
// this avatar to 'cover' crops the same spot as everywhere else. A bot
// never calls this and so always resolves to no avatar / centered focal.
func (p *Participant) SetAvatar(avatarThumbKey, googlePicture string, focalX, focalY float64) {
	p.avatarThumbKey = avatarThumbKey
	p.googlePicture = googlePicture
	p.avatarFocalX = focalX
	p.avatarFocalY = focalY
}

// Disconnect marks the participant as no longer reachable, stamping when
// this happened so a grace-period timer can be derived/re-derived from it
// later (see GameService's grace timers). It does not remove them from the
// Game - Game.Disconnect handles the follow-on host reassignment / abort
// checks.
func (p *Participant) Disconnect(at time.Time) {
	p.connected = false
	p.disconnectedAt = &at
}

// Reconnect marks the participant as reachable again, clearing both the
// disconnected-since timestamp and any abandoned flag - a returning player
// always regains full control of their seat, however long they were gone.
func (p *Participant) Reconnect() {
	p.connected = true
	p.disconnectedAt = nil
	p.abandoned = false
}

// MarkAbandoned flags a still-seated participant as no longer actively
// played - the seat, loadout and userID all stay put, but the seat is now
// treated as inactive (see Game.hasActiveHuman) and, in modes that allow it,
// auto-votes like a bot (see IGameMode.AutoVotesForAbandoned). Only
// GameService's grace timer calls this, once Disconnect's grace period has
// elapsed without a Reconnect.
func (p *Participant) MarkAbandoned() { p.abandoned = true }

// AssignLoadout replaces this participant's current abilities.
func (p *Participant) AssignLoadout(l *Loadout) { p.loadout = l }

// setTeam moves this participant to teamID. Deliberately unexported: only
// Game may reseat a participant, since a Team's member list must change in
// lockstep (see Game.SwitchTeam / Game.Reconfigure).
func (p *Participant) setTeam(id TeamID) { p.teamID = id }
