package model

import (
	crypto "github.com/tendermint/tendermint/crypto"

	"github.com/lino-network/lino/types"
)

// AccountRow account related information when migrate, pk: Username
type AccountRow struct {
	Username            types.AccountKey    `json:"username"`
	Info                AccountInfo         `json:"info"`
	Bank                AccountBank         `json:"bank"`
	Meta                AccountMeta         `json:"meta"`
	PendingCoinDayQueue PendingCoinDayQueue `json:"pending_coin_day_queue"`
}

// ToIR -
func (a AccountRow) ToIR() AccountRowIR {
	return AccountRowIR{
		Username:            a.Username,
		Info:                a.Info.ToIR(),
		Bank:                a.Bank.ToIR(),
		Meta:                a.Meta.ToIR(),
		PendingCoinDayQueue: a.PendingCoinDayQueue.ToIR(),
	}
}

// GrantPubKeyRow also in account, pk: (Username, pubKey)
type GrantPubKeyRow struct {
	Username    types.AccountKey `json:"username"`
	PubKey      crypto.PubKey    `json:"pub_key"`
	GrantPubKey GrantPubKey      `json:"grant_pub_key"`
}

// GrantPubKeyRowSliceToIR -
func GrantPubKeyRowSliceToIR(origin []GrantPubKeyRow) (ir []GrantPubKeyRowIR) {
	for _, v := range origin {
		ir = append(ir, v.ToIR())
	}
	return
}

// ToIR - int to string and internal conversions
func (g GrantPubKeyRow) ToIR() GrantPubKeyRowIR {
	return GrantPubKeyRowIR{
		Username:    g.Username,
		PubKey:      g.PubKey,
		GrantPubKey: g.GrantPubKey.ToIR(),
	}
}

// AccountTables is the state of account storage, organized as a table.
type AccountTables struct {
	Accounts            []AccountRow     `json:"accounts"`
	AccountGrantPubKeys []GrantPubKeyRow `json:"account_grant_pub_keys"`
}

// ToIR -
func (a AccountTables) ToIR() *AccountTablesIR {
	tables := &AccountTablesIR{}
	for _, v := range a.Accounts {
		tables.Accounts = append(tables.Accounts, v.ToIR())
	}
	tables.AccountGrantPubKeys = GrantPubKeyRowSliceToIR(a.AccountGrantPubKeys)
	return tables
}
