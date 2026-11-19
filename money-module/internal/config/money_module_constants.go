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
