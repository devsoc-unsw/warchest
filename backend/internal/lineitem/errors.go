package lineitem

import "errors"

// Sentinel errors returned by the service. They exist so that the HTTP layer
// can map failures to status codes with errors.Is, without inspecting strings
// or knowing anything about the rules that produced them.
//
// Wrap them with fmt.Errorf("%w: detail", ErrX) to add context.
var (
	// ErrNotFound is returned when no line item exists for the given id.
	// Maps to 404.
	ErrNotFound = errors.New("line item not found")

	// ErrValidation is returned when input fails a bounds or presence check.
	// Maps to 400.
	ErrValidation = errors.New("invalid line item")

	// ErrInvalidTransition is returned when a status change is not a legal
	// edge of the lifecycle. Maps to 409.
	ErrInvalidTransition = errors.New("invalid status transition")

	// ErrImmutableField is returned when a field is written in a status that
	// does not permit editing it. Maps to 409.
	ErrImmutableField = errors.New("field is not editable in the current status")
)
