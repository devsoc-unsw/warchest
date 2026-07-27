package warchest

// 2 bags of 10 paper cups. The bags cost $10 each.
// Quantity = 2
// Cost per item = $10
// Unit cost = $1 ($10 per bag, 10 cups in a bag)

// not sure if the status enum is supposed to be declared else where.
// if not, then it would look like this:
// credit: https://gobyexample.com/enums
type StatusEnum int;

const (
	LineItem StatusEnum = iota
	InReimbursement
	Archived
)

type LineItems struct {
	// primary key
	id string;

	estimatedQuantity int64;
	estimatedCostPerItem int64;
	estimatedUnitCost int64;

	// IDs are not explicitly entered by the user
	reimbursementRequestID string;
	purchaseRequestID string;
	
	actualQuantity int64;
	actualCostPerItem int64;
	actualUnitCost int64;
	
	description string;
	status StatusEnum;

	// logging
	// created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    // updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
}