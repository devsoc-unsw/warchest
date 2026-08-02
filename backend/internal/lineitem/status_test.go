package lineitem

import (
	"testing"

	"backend/db"
)

// legalEdges restates the lifecycle independently of allowedTransitions, so
// that the exhaustive test below checks the implementation against a spec
// rather than against itself. Any edge not listed here must be rejected.
var legalEdges = map[[2]db.LineItemStatus]bool{
	{db.LineItemStatusPrDraft, db.LineItemStatusPrPending}:          true,
	{db.LineItemStatusPrPending, db.LineItemStatusPrApproved}:       true,
	{db.LineItemStatusPrPending, db.LineItemStatusPrRejected}:       true,
	{db.LineItemStatusPrApproved, db.LineItemStatusReimbDraft}:      true,
	{db.LineItemStatusReimbDraft, db.LineItemStatusReimbPending}:    true,
	{db.LineItemStatusReimbPending, db.LineItemStatusReimbApproved}: true,
	{db.LineItemStatusReimbPending, db.LineItemStatusReimbRejected}: true,
}

var terminalStatuses = []db.LineItemStatus{
	db.LineItemStatusPrRejected,
	db.LineItemStatusReimbApproved,
	db.LineItemStatusReimbRejected,
}

// TestCanTransitionExhaustive checks every ordered pair of statuses, so a new
// edge cannot be added to the transition table without a matching change here.
func TestCanTransitionExhaustive(t *testing.T) {
	for _, from := range AllStatuses {
		for _, to := range AllStatuses {
			want := legalEdges[[2]db.LineItemStatus{from, to}]
			if got := CanTransition(from, to); got != want {
				t.Errorf("CanTransition(%q, %q) = %v, want %v", from, to, got, want)
			}
		}
	}
}

func TestCanTransitionRejectsSelfTransitions(t *testing.T) {
	for _, s := range AllStatuses {
		if CanTransition(s, s) {
			t.Errorf("CanTransition(%q, %q) = true, want false", s, s)
		}
	}
}

func TestCanTransitionRejectsUnknownStatus(t *testing.T) {
	unknown := db.LineItemStatus("not_a_status")

	if CanTransition(unknown, db.LineItemStatusPrPending) {
		t.Error("CanTransition from an unknown status = true, want false")
	}
	if CanTransition(db.LineItemStatusPrDraft, unknown) {
		t.Error("CanTransition to an unknown status = true, want false")
	}
}

func TestTerminalStatusesHaveNoOutgoingEdges(t *testing.T) {
	for _, s := range terminalStatuses {
		if !IsTerminal(s) {
			t.Errorf("IsTerminal(%q) = false, want true", s)
		}
		if next := NextStates(s); len(next) != 0 {
			t.Errorf("NextStates(%q) = %v, want empty", s, next)
		}
	}
}

func TestNonTerminalStatusesHaveOutgoingEdges(t *testing.T) {
	terminal := make(map[db.LineItemStatus]bool, len(terminalStatuses))
	for _, s := range terminalStatuses {
		terminal[s] = true
	}

	for _, s := range AllStatuses {
		if terminal[s] {
			continue
		}
		if IsTerminal(s) {
			t.Errorf("IsTerminal(%q) = true, want false", s)
		}
		if len(NextStates(s)) == 0 {
			t.Errorf("NextStates(%q) is empty, want at least one edge", s)
		}
	}
}

func TestIsTerminalRejectsUnknownStatus(t *testing.T) {
	if IsTerminal(db.LineItemStatus("not_a_status")) {
		t.Error("IsTerminal(unknown) = true, want false")
	}
}

// TestNextStatesAgreesWithCanTransition guards against the two accessors
// drifting apart.
func TestNextStatesAgreesWithCanTransition(t *testing.T) {
	for _, from := range AllStatuses {
		for _, to := range NextStates(from) {
			if !CanTransition(from, to) {
				t.Errorf("NextStates(%q) offers %q, CanTransition rejects", from, to)
			}
		}
	}
}

// TestNextStatesReturnsCopy ensures callers cannot corrupt the transition
// table through the returned slice.
func TestNextStatesReturnsCopy(t *testing.T) {
	const from = db.LineItemStatusPrPending

	first := NextStates(from)
	if len(first) == 0 {
		t.Fatalf("NextStates(%q) is empty, cannot test copy semantics", from)
	}
	first[0] = db.LineItemStatus("mutated")

	second := NextStates(from)
	if second[0] == db.LineItemStatus("mutated") {
		t.Error("NextStates returned a view of the transition table, want a copy")
	}
}

func TestIsReimbursementPhase(t *testing.T) {
	tests := map[db.LineItemStatus]bool{
		db.LineItemStatusPrDraft:        false,
		db.LineItemStatusPrPending:      false,
		db.LineItemStatusPrApproved:     false,
		db.LineItemStatusPrRejected:     false,
		db.LineItemStatusReimbDraft:     true,
		db.LineItemStatusReimbPending:   true,
		db.LineItemStatusReimbApproved:  true,
		db.LineItemStatusReimbRejected:  true,
		db.LineItemStatus("not_status"): false,
	}

	for status, want := range tests {
		if got := IsReimbursementPhase(status); got != want {
			t.Errorf("IsReimbursementPhase(%q) = %v, want %v", status, got, want)
		}
	}
}

// TestAllStatusesCoversTransitionTable keeps AllStatuses honest: if a status is
// added to the enum and the transition table but not to AllStatuses, the
// exhaustive test above would silently stop covering it.
func TestAllStatusesCoversTransitionTable(t *testing.T) {
	got, want := len(AllStatuses), len(allowedTransitions)
	if got != want {
		t.Errorf("AllStatuses has %d entries, transition table has %d", got, want)
	}

	for _, s := range AllStatuses {
		if _, ok := allowedTransitions[s]; !ok {
			t.Errorf("status %q in AllStatuses but not in transition table", s)
		}
	}
}
