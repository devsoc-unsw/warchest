package logging

import (
	"context"
	"log/slog"
)

// Log template for editing events
func EventEdit(ctx context.Context, budgetID, societyID, actorID string) {
	slog.InfoContext(ctx, "event edit",
		"action", "event.edit",
		"budget_id", budgetID,
		"budget_type", "EVENT",
		"society_id", societyID,
		"actor_id", actorID,
	)
}

// template for when edit fails
func EventEditFailed(ctx context.Context, budgetID, societyID,
		actorID string, err error) {
	slog.ErrorContext(ctx, "event edit failed",
		"action", "event.edit",
		"budget_id", budgetID,
		"budget_type", "EVENT",
		"society_id", societyID,
		"actor_id", actorID,
		"err", err,
	)
}

func EventDelete(ctx context.Context, budgetID, societyID, 
		actorID string) {
	slog.InfoContext(ctx, "event delete",
		"action", "event.delete",
		"budget_id", budgetID,
		"budget_type", "EVENT",
		"society_id", societyID,
		"actor_id", actorID,
	)
}

func EventDeleteFailed(ctx context.Context, budgetID, societyID, 
		actorID string, err error) {
	slog.ErrorContext(ctx, "event delete failed",
		"action", "event.delete",
		"budget_id", budgetID,
		"budget_type", "EVENT",
		"society_id", societyID,
		"actor_id", actorID,
		"err", err,
	)
}

func EventClose(ctx context.Context, budgetID, moneyBudgetID, societyID, 
		actorID string) {
	slog.InfoContext(ctx, "event close",
		"action", "event.close",
		"budget_id", budgetID,
		"budget_type", "EVENT",
		"money_budget_id", moneyBudgetID,
		"society_id", societyID,
		"actor_id", actorID,
	)
}

func EventCloseFailed(ctx context.Context, budgetID, moneyBudgetID, societyID, 
		actorID string, err error) {
	slog.ErrorContext(ctx, "event close failed",
		"action", "event.close",
		"budget_id", budgetID,
		"budget_type", "EVENT",
		"money_budget_id", moneyBudgetID,
		"society_id", societyID,
		"actor_id", actorID,
		"err", err,
	)
}
