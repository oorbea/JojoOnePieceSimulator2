package endpoints

import (
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/dto"
)

// wsFrameConstants and wsCommandConstants are the Frame*/Command* string
// values as declared in game_ws.go, kept here (not AST-scraped) since they
// change rarely and this test's job is to catch dto.FramePayloads/
// CommandPayloads drifting out of sync with them, not to rediscover them.
var wsFrameConstants = []string{
	dto.FrameState, dto.FramePlayerJoined, dto.FramePlayerLeft, dto.FrameHostReassigned,
	dto.FrameGameStarted, dto.FrameLoadoutsAssigned, dto.FrameVotingOpened, dto.FrameVoteCast,
	dto.FrameTiebreakOpened, dto.FrameRoundResolved, dto.FrameGameFinished, dto.FrameGameAborted,
	dto.FrameError, dto.FrameResyncRequired, dto.FrameTeamChanged, dto.FramePlayerKicked,
	dto.FrameLobbyLockChanged, dto.FrameConfigUpdated, dto.FrameRevealReadyChanged,
	dto.FrameSummaryOpened, dto.FrameSummaryReadyChanged, dto.FrameRematchReady,
}

var wsCommandConstants = []string{
	dto.CommandLeave, dto.CommandAddBot, dto.CommandRemoveBot, dto.CommandStart, dto.CommandAbort,
	dto.CommandVote, dto.CommandResync, dto.CommandSwitchTeam, dto.CommandMovePlayer, dto.CommandKick,
	dto.CommandTransferHost, dto.CommandSetLock, dto.CommandUpdateConfig, dto.CommandRevealReady,
	dto.CommandSummaryReady, dto.CommandRematch,
}

// TestFramePayloads_CoversEveryFrameConstant asserts dto.FramePayloads'
// key set exactly equals wsFrameConstants, in both directions - a frame
// added to the Frame* const block without a FramePayloads entry, or vice
// versa, fails here instead of silently producing an incomplete (or stale)
// generated ServerFrame union.
func TestFramePayloads_CoversEveryFrameConstant(t *testing.T) {
	want := make(map[string]bool, len(wsFrameConstants))
	for _, f := range wsFrameConstants {
		want[f] = true
	}

	got := make(map[string]bool, len(dto.FramePayloads))
	for _, spec := range dto.FramePayloads {
		if got[spec.Type] {
			t.Errorf("dto.FramePayloads has a duplicate entry for %q", spec.Type)
		}
		got[spec.Type] = true
		if !want[spec.Type] {
			t.Errorf("dto.FramePayloads has an entry for %q, which is not a declared Frame* constant", spec.Type)
		}
	}
	for f := range want {
		if !got[f] {
			t.Errorf("Frame* constant %q has no dto.FramePayloads entry", f)
		}
	}
}

// TestCommandPayloads_CoversEveryCommandConstant is
// TestFramePayloads_CoversEveryFrameConstant's mirror for the client->server
// direction.
func TestCommandPayloads_CoversEveryCommandConstant(t *testing.T) {
	want := make(map[string]bool, len(wsCommandConstants))
	for _, c := range wsCommandConstants {
		want[c] = true
	}

	got := make(map[string]bool, len(dto.CommandPayloads))
	for _, spec := range dto.CommandPayloads {
		if got[spec.Type] {
			t.Errorf("dto.CommandPayloads has a duplicate entry for %q", spec.Type)
		}
		got[spec.Type] = true
		if !want[spec.Type] {
			t.Errorf("dto.CommandPayloads has an entry for %q, which is not a declared Command* constant", spec.Type)
		}
	}
	for c := range want {
		if !got[c] {
			t.Errorf("Command* constant %q has no dto.CommandPayloads entry", c)
		}
	}
}
