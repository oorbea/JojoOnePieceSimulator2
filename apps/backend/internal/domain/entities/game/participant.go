package game

import (
	"errors"

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
		id:          id,
		userID:      &uid,
		displayName: displayName,
		teamID:      teamID,
		kind:        enums.Human,
		connected:   true,
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

// SetAvatar records where this participant's avatar picture comes from -
// their own uploaded thumbnail key (presigned at serialization time) and/or
// their Google-synced picture URL (already a full external URL). A bot
// never calls this and so always resolves to no avatar.
func (p *Participant) SetAvatar(avatarThumbKey, googlePicture string) {
	p.avatarThumbKey = avatarThumbKey
	p.googlePicture = googlePicture
}

// Disconnect marks the participant as no longer reachable. It does not
// remove them from the Game - Game.Disconnect handles the follow-on host
// reassignment / abort checks.
func (p *Participant) Disconnect() { p.connected = false }

// Reconnect marks the participant as reachable again.
func (p *Participant) Reconnect() { p.connected = true }

// AssignLoadout replaces this participant's current abilities.
func (p *Participant) AssignLoadout(l *Loadout) { p.loadout = l }

// setTeam moves this participant to teamID. Deliberately unexported: only
// Game may reseat a participant, since a Team's member list must change in
// lockstep (see Game.SwitchTeam / Game.Reconfigure).
func (p *Participant) setTeam(id TeamID) { p.teamID = id }
