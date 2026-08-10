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
}
