package constants

import tb "github.com/tigerbeetle/tigerbeetle-go"

const RealAccountLedger uint32 = 67
const RealAccountCode uint16 = 67

type AccountType int32

const (
	ROOT    AccountType = 0
	BUDGET  AccountType = 1
	ACCOUNT AccountType = 2
)

type TransactionStatus int32

const (
	TransactionSuccess TransactionStatus = iota
	TransactionInvalidAmount
	TransactionSameAccount
	TransactionPayerNotFound
	TransactionPayeeNotFound
	TransactionPayerArchived
	TransactionPayeeArchived
	TransactionInvalidAccountType
	TransactionInsufficientFunds
	TransactionDuplicateTransfer
	TransactionTransferFailed
	TransactionInternalError
)

func (s TransactionStatus) String() string {
	switch s {
	case TransactionSuccess:
		return "SUCCESS"
	case TransactionInvalidAmount:
		return "INVALID_AMOUNT"
	case TransactionSameAccount:
		return "SAME_ACCOUNT"
	case TransactionPayerNotFound:
		return "PAYER_NOT_FOUND"
	case TransactionPayeeNotFound:
		return "PAYEE_NOT_FOUND"
	case TransactionPayerArchived:
		return "PAYER_ARCHIVED"
	case TransactionPayeeArchived:
		return "PAYEE_ARCHIVED"
	case TransactionInvalidAccountType:
		return "INVALID_ACCOUNT_TYPE"
	case TransactionInsufficientFunds:
		return "INSUFFICIENT_FUNDS"
	case TransactionDuplicateTransfer:
		return "DUPLICATE_TRANSFER"
	case TransactionTransferFailed:
		return "TRANSFER_FAILED"
	case TransactionInternalError:
		return "INTERNAL_ERROR"
	default:
		return "UNKNOWN"
	}
}

const VirtualAccountLedger uint32 = 69
const VirtualAccountCode uint16 = 69

var (
	realLedgerWorldAccBookUUID = [16]byte{
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00,
		0x00, 0x00,
		0x80, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	realLedgerLiquidityBookUUID = [16]byte{
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00,
		0x00, 0x00,
		0x80, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
	}
	virtualLedgerLiquidityBookUUID = [16]byte{
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00,
		0x00, 0x00,
		0x80, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x02,
	}
	realLedgerWorldBookId        tb.Uint128 = tb.BytesToUint128(realLedgerWorldAccBookUUID)
	realLedgerLiquidityBookId    tb.Uint128 = tb.BytesToUint128(realLedgerLiquidityBookUUID)
	virtualLedgerLiquidityBookId tb.Uint128 = tb.BytesToUint128(virtualLedgerLiquidityBookUUID)
)

// golang 'const' keyword only works on primitives. So we expose getters and keep data private so the data cannot be modified
func GetRealLedgerWorldBookId() tb.Uint128        { return realLedgerWorldBookId }
func GetRealLedgerLiquidityBookId() tb.Uint128    { return realLedgerLiquidityBookId }
func GetVirtualLedgerLiquidityBookId() tb.Uint128 { return virtualLedgerLiquidityBookId }
