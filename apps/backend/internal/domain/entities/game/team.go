package game

import "errors"

// ErrEmptyTeamName is returned when NewTeam is given an empty name.
var ErrEmptyTeamName = errors.New("team name is required")

// Team groups Participants under a name/color for Versus (2 teams) or
// Gauntlet (1 implicit team). Membership is tracked by ParticipantID,
// resolved against Game.Participants - a Team holds no player pointers, so
// nothing outside Game can mutate a Game's roster through a Team.
type Team struct {
	id      TeamID
	name    string
	color   uint32
	members []ParticipantID
}

// NewTeam validates and builds an empty Team; members are added via
// AddMember (Game.Join/AddBot do this).
func NewTeam(id TeamID, name string, color uint32) (*Team, error) {
	if id.IsNil() {
		return nil, errors.New("id is required")
	}
	if name == "" {
		return nil, ErrEmptyTeamName
	}
	return &Team{id: id, name: name, color: color}, nil
}

func (t *Team) ID() TeamID    { return t.id }
func (t *Team) Name() string  { return t.name }
func (t *Team) Color() uint32 { return t.color }
func (t *Team) Size() int     { return len(t.members) }

// Members returns a copy of the team's participant ids.
func (t *Team) Members() []ParticipantID {
	return append([]ParticipantID(nil), t.members...)
}

// HasMember reports whether id is currently a member of this team.
func (t *Team) HasMember(id ParticipantID) bool {
	for _, m := range t.members {
		if m == id {
			return true
		}
	}
	return false
}

// AddMember appends id if it isn't already a member.
func (t *Team) AddMember(id ParticipantID) {
	if t.HasMember(id) {
		return
	}
	t.members = append(t.members, id)
}

// RemoveMember removes id if present.
func (t *Team) RemoveMember(id ParticipantID) {
	for i, m := range t.members {
		if m == id {
			t.members = append(t.members[:i:i], t.members[i+1:]...)
			return
		}
	}
}
