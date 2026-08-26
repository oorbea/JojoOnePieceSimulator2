package game

// OptionID is an opaque ballot option: enums.SquadVerdict's string form for
// Gauntlet ("SURVIVE"/"FALL"), or a TeamID.String() for Versus. Ballot does
// not need to know which.
type OptionID string

// Ballot counts a single round's votes. It is deliberately dumb about what
// the options mean - each IGameMode maps its own OptionIDs onto it, so the
// majority/tie/null-vote algorithm lives in exactly one place.
type Ballot struct {
	options []OptionID
	votes   map[ParticipantID]OptionID
}

// NewBallot builds a Ballot with the given fixed set of options.
func NewBallot(options []OptionID) (*Ballot, error) {
	if len(options) < 2 {
		return nil, ErrInvalidBallotOption
	}
	return &Ballot{
		options: append([]OptionID(nil), options...),
		votes:   make(map[ParticipantID]OptionID, len(options)),
	}, nil
}

// Options returns the ballot's fixed set of options, in their original
// order.
func (b *Ballot) Options() []OptionID {
	return append([]OptionID(nil), b.options...)
}

// Cast records p's vote for option o, overwriting any previous vote by p -
// the last vote cast before the window closes is the one that counts.
func (b *Ballot) Cast(p ParticipantID, o OptionID) error {
	if !b.isValidOption(o) {
		return ErrInvalidBallotOption
	}
	b.votes[p] = o
	return nil
}

func (b *Ballot) isValidOption(o OptionID) bool {
	for _, opt := range b.options {
		if opt == o {
			return true
		}
	}
	return false
}

// HasVoted reports whether p has cast a vote that has not been superseded
// by never voting again (there is no "retract"; only overwrite).
func (b *Ballot) HasVoted(p ParticipantID) bool {
	_, ok := b.votes[p]
	return ok
}

// Count returns how many distinct participants have voted so far.
func (b *Ballot) Count() int {
	return len(b.votes)
}

// Votes returns a copy of every vote cast so far, keyed by participant. It
// exists for two callers that both need the actual choice rather than just
// HasVoted/Count: Snapshot (a mid-round Ballot cannot otherwise be
// serialized) and the application/transport layer revealing votes once a
// round has resolved (see the "votes hidden until close" transport rule -
// nothing in the domain hides them, the transport chooses when to read this).
func (b *Ballot) Votes() map[ParticipantID]OptionID {
	out := make(map[ParticipantID]OptionID, len(b.votes))
	for p, o := range b.votes {
		out[p] = o
	}
	return out
}

// Reset clears every cast vote while keeping the fixed option set - used by
// Game.CloseVoting when a tie opens a revote window, so the revote starts
// from a genuinely empty ballot instead of carrying every vote over (which
// would let a single changed vote resolve the whole revote instantly).
func (b *Ballot) Reset() {
	b.votes = make(map[ParticipantID]OptionID, len(b.options))
}

// Tally applies plurality over emitted votes only - a participant who
// never votes (disconnected, or too slow) simply does not count towards
// the total. Zero emitted votes, or two-or-more options tied for the
// lead, is reported via tied=true; the caller (Game.CloseVoting) must then
// fall back to a revote window and, ultimately, an externally-decided
// tiebreak.
func (b *Ballot) Tally() (winner OptionID, tied bool) {
	if len(b.votes) == 0 {
		return "", true
	}
	counts := make(map[OptionID]int, len(b.options))
	for _, v := range b.votes {
		counts[v]++
	}
	best := -1
	leaders := 0
	for _, o := range b.options {
		c := counts[o]
		switch {
		case c > best:
			best = c
			winner = o
			leaders = 1
		case c == best:
			leaders++
		}
	}
	if leaders > 1 {
		return "", true
	}
	return winner, false
}
