package logging

import (
	"context"
	"log/slog"
)
// Log template for editing events
func EventEdit (ctx context.Context, eventID, societyID, actorID string) {
	slog.InfoContext(ctx, "event edit",
	"action", "event.edit",
	"event_id", eventID,
	"society_id", societyID,
	"actor_id", actorID,
	)
}

// template for when edit fails
func EventEditFailed (ctx context.Context, eventID, societyID, actorID string, err error) {
	slog.ErrorContext(ctx, "event edit failed",
	"action", "event.edit",
	"event_id", eventID,
	"society_id", societyID,
	"actor_id", actorID,
	"err", err,
	)
}

func EventDelete (ctx context.Context, eventID, societyID, actorID string) {
	slog.InfoContext(ctx, "event delete",
	"action", "event.delete",
	"event_id", eventID,
	"society_id", societyID,
	"actor_id", actorID,
	)
}

func EventDeleteFailed (ctx context.Context, eventID, societyID, actorID string, err error) {
	slog.InfoContext(ctx, "event delete",
	"action", "event.delete",
	"event_id", eventID,
	"society_id", societyID,
	"actor_id", actorID,
	"err", err,
	)
}

func EventClose (ctx context.Context, eventID, societyID, actorID string) {
	slog.InfoContext(ctx, "event close",
	"action", "event.close",
	"event_id", eventID,
	"society_id", societyID,
	"actor_id", actorID,
	)
}

func EventCloseFailed (ctx context.Context, eventID, societyID, actorID string, err error) {
	slog.InfoContext(ctx, "event close",
	"action", "event.close",
	"event_id", eventID,
	"society_id", societyID,
	"actor_id", actorID,
	"err", err,
	)
}


