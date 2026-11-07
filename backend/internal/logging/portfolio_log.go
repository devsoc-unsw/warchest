package logging

import (
	"context"
	"log/slog"
)

// template for portfolio budget reset - wrapper function that calls slog
// uses slog.InfoContext to accept context.Context
func BudgetReset(ctx context.Context, budgetID, moneyBudgetID, societyID, actorID string, oldBudget, newBudget int) {
	slog.InfoContext(ctx, "portfolio budget reset",
		"action", "portfolio.budget.reset",
		"budget_id", budgetID,
		"budget_type", "PORTFOLIO",
		"money_budget_id", moneyBudgetID,
		"society_id", societyID,
		"actor_id", actorID,
	)
}

// template for when reset fails
func BudgetResetFailed(ctx context.Context, budgetID, moneyBudgetID, societyID, actorID string,
	err error) {
	slog.ErrorContext(ctx, "portfolio budget reset failed",
		"action", "portfolio.budget.reset",
		"budget_id", budgetID,
		"budget_type", "PORTFOLIO",
		"money_budget_id", moneyBudgetID,
		"society_id", societyID,
		"actor_id", actorID,
		"err", err,
	)
}

// call from service layer
// logging.BudgetReset(ctx, budgetID, societyID, userID, oldBudget, newBudget)
