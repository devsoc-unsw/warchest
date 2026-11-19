package constants

const RealAccountLedger uint32 = 67
const RealAccountCode uint16 = 67

type AccountType int32

const (
	ROOT    AccountType = 0
	BUDGET  AccountType = 1
	ACCOUNT AccountType = 2
)
