package game

// RoundResult is the outcome of tallying a Round's Ballot.
type RoundResult struct {
	Winner            OptionID
	DecidedByCoinFlip bool
}

// Round is one iteration of assign -> vote -> resolve.
type Round struct {
	Index        int
	Stage        Stage
	Ballot       *Ballot
	TiebreakUsed bool
	Result       *RoundResult
	// TiedVotes is a snapshot of every vote cast before a tie forced a
	// revote (see Game.CloseVoting), keyed by participant - nil until the
	// first tie for this round. Ballot.Reset() wipes the live ballot for
	// the revote, so this is the only place the tied vote breakdown
	// survives to be shown to clients.
	TiedVotes map[ParticipantID]OptionID
}
