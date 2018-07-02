package types

// nolint
import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

// Transactions messages must fulfill the Msg
type Msg interface {
	sdk.Msg
	GetPermission() Permission
	GetCapacityLevel() CapacityLevel
}

// Register the lino message type
func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterInterface((*Msg)(nil), nil)
}
