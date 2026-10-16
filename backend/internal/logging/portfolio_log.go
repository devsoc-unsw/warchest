package logging

import (
	"context"
	"log/slog"
)

// template for portfolio budget reset - wrapper function that calls slog
// uses slog.InfoContext to accept context.Context
func BudgetReset (ctx context.Context, portfolioID, societyID, actorID string, oldBudget, newBudget int) {
	slog.InfoContext(ctx, "portfolio budget reset",
	"action", "portfolio.budget.reset",
	"portfolio_id", portfolioID,
	"society_id", societyID,
	"actor_id", actorID,
	"old_budget", oldBudget,
	"new_budget", newBudget,
	)
}

// template for when reset fails
func BudgetResetFailed (ctx context.Context, portfolioID, societyID, actorID string,
	err error) {
	slog.ErrorContext(ctx, "portfolio budget reset failed",
	"action", "portfolio.budget.reset",
	"portfolio_id", portfolioID,
	"society_id", societyID,
	"actor_id", actorID,
	"err", err,
	)
}

// call from service layer
// logging.BudgetReset(ctx, portfolioID, societyID, userID, oldBudget, newBudget)
