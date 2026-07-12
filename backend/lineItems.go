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

type lineItems struct {
	estimatedQuantity uint64;
	estimatedCostPerItem uint64;
	estimatedUnitCost uint64;

	// IDs are not explicitly entered by the user
	reimbursementRequestID string;
	purchaseRequestID string;
	
	actualQuantity uint64;
	actualCostPerItem uint64;
	actualUnitCost uint64;
	
	description string;
	status StatusEnum;
}