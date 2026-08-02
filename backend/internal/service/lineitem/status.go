package lineitem

import "backend/db"

// AllStatuses lists every status in lifecycle order. It is the authoritative
// set for iterating the lifecycle (tests, and later the API surface that tells
// the frontend which actions to render).
var AllStatuses = []db.LineItemStatus{
	db.LineItemStatusPrDraft,
	db.LineItemStatusPrPending,
	db.LineItemStatusPrApproved,
	db.LineItemStatusPrRejected,
	db.LineItemStatusReimbDraft,
	db.LineItemStatusReimbPending,
	db.LineItemStatusReimbApproved,
	db.LineItemStatusReimbRejected,
}

// allowedTransitions maps each status to the statuses it may legally move to.
// It is held as data rather than as conditionals so that the legal next states
// can be served to callers, not just enforced.
//
// A line item starts in pr_draft and is submitted for approval. Approval moves
// it into the reimbursement phase, where the buyer records what was actually
// spent and submits again. An entry with no outgoing edges is terminal.
var allowedTransitions = map[db.LineItemStatus][]db.LineItemStatus{
	// Purchase request phase.
	db.LineItemStatusPrDraft: {db.LineItemStatusPrPending},
	db.LineItemStatusPrPending: {
		db.LineItemStatusPrApproved,
		db.LineItemStatusPrRejected,
	},
	db.LineItemStatusPrApproved: {db.LineItemStatusReimbDraft},
	db.LineItemStatusPrRejected: {},

	// Reimbursement phase.
	db.LineItemStatusReimbDraft: {db.LineItemStatusReimbPending},
	db.LineItemStatusReimbPending: {
		db.LineItemStatusReimbApproved,
		db.LineItemStatusReimbRejected,
	},
	db.LineItemStatusReimbApproved: {},
	db.LineItemStatusReimbRejected: {},
}

// CanTransition reports whether moving from one status to another is a legal
// edge of the lifecycle. Unknown statuses have no edges, so they always
// report false.
func CanTransition(from, to db.LineItemStatus) bool {
	for _, next := range allowedTransitions[from] {
		if next == to {
			return true
		}
	}
	return false
}

// NextStates returns the statuses reachable from the given status. The result
// is a copy, so callers cannot mutate the transition table through it. A
// terminal or unknown status yields an empty slice.
func NextStates(from db.LineItemStatus) []db.LineItemStatus {
	next := allowedTransitions[from]
	out := make([]db.LineItemStatus, len(next))
	copy(out, next)
	return out
}

// IsTerminal reports whether the status is a known end state, meaning nothing
// further can happen to the line item. Unknown statuses report false: they are
// not terminal, they are invalid.
func IsTerminal(s db.LineItemStatus) bool {
	next, known := allowedTransitions[s]
	return known && len(next) == 0
}

// IsReimbursementPhase reports whether the status belongs to the reimbursement
// half of the lifecycle, which is where the actual_* fields become meaningful.
func IsReimbursementPhase(s db.LineItemStatus) bool {
	switch s {
	case db.LineItemStatusReimbDraft,
		db.LineItemStatusReimbPending,
		db.LineItemStatusReimbApproved,
		db.LineItemStatusReimbRejected:
		return true
	default:
		return false
	}
}
