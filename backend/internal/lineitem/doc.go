// Package lineitem implements the business rules for purchase request line
// items: the status lifecycle, the field mutability rules that fall out of it,
// and input validation.
//
// All money values handled by this package are minor units (cents). Nothing
// here converts between minor and major units; callers pass and receive cents.
//
// The three cost fields describe a purchase at different granularities. For
// two bags of ten paper cups at $10 a bag:
//
//	quantity      = 2   (bags bought)
//	cost per item = $10 (price of one bag)
//	unit cost     = $1  (price of one cup)
//
// Quantity times cost per item gives the total. Unit cost is entered by the
// user for reporting and cannot be derived from the other two, since the
// number of cups in a bag is not recorded; it is therefore bounds checked but
// never cross checked against them.
//
// Each field exists twice: the estimated_* values are what the requester
// expects to spend, submitted with the purchase request, and the actual_*
// values are what was really spent, submitted with the reimbursement request
// once the goods have been bought.
package lineitem
